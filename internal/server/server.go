// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package server composes the services of the selected targets and runs them through the
// lifecycle package until a signal or a failure ends the process.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/tbhb-dev/afakt/internal/dashboard"
	"github.com/tbhb-dev/afakt/internal/lifecycle"
)

// readHeaderTimeout bounds how long a client may take to send request headers.
const readHeaderTimeout = 10 * time.Second

// The names of the services every composition shares. Each target also runs one service
// under its own name, which depends on the HTTP server.
const (
	listenerService = "http-listener"
	httpService     = "http"
)

// ErrNotComposable reports a target that Run has no services for.
var ErrNotComposable = errors.New("no services for target")

// Config is a checked server configuration. Targets comes from ParseTargets.
type Config struct {
	Targets         []string
	Listen          string
	ShutdownTimeout time.Duration
}

// hooks adds services and routes to the composition. Tests use it to hold a service in a
// chosen state and to keep a request in flight.
type hooks struct {
	services []lifecycle.Registration
	routes   func(mux *http.ServeMux)
}

// Run starts the services of the selected targets and serves until the context ends or a
// required service fails. Then it stops every service within the configured shutdown
// timeout. A clean shutdown that the context started returns nil. A failed start, a failed
// service, or a shutdown that didn't finish in time returns an error.
func Run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	return run(ctx, cfg, logger, hooks{})
}

// run is Run with hooks for tests. The server is ready from the moment every service has
// started until the wait ends, so it reports not ready while it starts, after a required
// service fails, and while it drains. The shutdown context keeps the values of the
// caller's context but not its cancellation, because a signal that cancels the caller's
// context starts the shutdown and the services still need time to stop.
func run(ctx context.Context, cfg Config, logger *slog.Logger, h hooks) error {
	manager := lifecycle.New(logger)
	if err := compose(manager, cfg, logger, newHealth(ctx, manager), h); err != nil {
		return err
	}
	err := manager.Start(ctx)
	if err == nil {
		logger.InfoContext(ctx, "server ready", "targets", cfg.Targets)
		err = manager.Wait(ctx)
	}
	logger.InfoContext(ctx, "shutting down", "timeout", cfg.ShutdownTimeout.String())
	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
	defer cancel()
	return errors.Join(err, manager.Stop(stopCtx))
}

// compose registers the services of the selected targets with manager. Every target serves
// its routes through one listener and one HTTP server. The health checks cover the whole
// composition. A target without services returns ErrNotComposable.
func compose(manager *lifecycle.Manager, cfg Config, logger *slog.Logger, checks health, h hooks) error {
	ln := &listener{address: cfg.Listen, logger: logger}
	srv := &httpServer{listener: ln, routes: []func(*http.ServeMux){checks.probeRoutes}}
	regs := []lifecycle.Registration{
		{Name: listenerService, Service: ln},
		{Name: httpService, Service: srv, DependsOn: []string{listenerService}},
	}
	for _, target := range cfg.Targets {
		switch target {
		case TargetAPI:
			srv.routes[0] = checks.routes
			srv.routes = append(srv.routes, apiRoutes)
		case TargetWebhook:
			// The webhook target has no receivers, so it serves only the health routes.
		case TargetDashboard:
			srv.routes = append(srv.routes, dashboardRoutes)
		default:
			return fmt.Errorf("%w '%s'", ErrNotComposable, target)
		}
		regs = append(regs, lifecycle.Registration{Name: target, Service: idle{}, DependsOn: []string{httpService}})
	}
	if h.routes != nil {
		srv.routes = append(srv.routes, h.routes)
	}
	// Services from the hooks come last, after the services of every target.
	for _, reg := range append(regs, h.services...) {
		if err := manager.Register(reg); err != nil {
			return fmt.Errorf("compose: %w", err)
		}
	}
	return nil
}

// dashboardRoutes adds only the dashboard's exact root and its subtree.
func dashboardRoutes(mux *http.ServeMux) {
	handler := dashboard.New()
	mux.Handle("/dashboard", handler)
	mux.Handle("/dashboard/", handler)
}

// listener binds the TCP address of the HTTP server. On shutdown, the HTTP server closes
// the listener. Stop closes it too, for the case where the HTTP server never served. It
// has no work of its own while it runs, so it takes Run from idle.
type listener struct {
	idle

	address string
	logger  *slog.Logger
	ln      net.Listener
}

// Start binds the address and logs the address it bound. With port 0 in the configuration,
// that log line shows the port the system chose.
func (l *listener) Start(ctx context.Context) error {
	var config net.ListenConfig
	ln, err := config.Listen(ctx, "tcp", l.address)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	l.ln = ln
	l.logger.InfoContext(ctx, "listening", "address", ln.Addr().String())
	return nil
}

// Stop closes the listener. A listener that the HTTP server already closed isn't an error.
func (l *listener) Stop(context.Context) error {
	if err := l.ln.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("close listener: %w", err)
	}
	return nil
}

// httpServer serves the routes of the selected targets on the listener.
type httpServer struct {
	listener *listener
	routes   []func(mux *http.ServeMux)
	server   *http.Server
}

// Start builds the routes and the HTTP server without serving. Serving starts in Run.
func (s *httpServer) Start(context.Context) error {
	mux := http.NewServeMux()
	for _, add := range s.routes {
		add(mux)
	}
	s.server = &http.Server{Handler: canonical(mux), ReadHeaderTimeout: readHeaderTimeout}
	return nil
}

// canonical answers 404 for a path that isn't in clean form, such as one with a doubled
// slash or a dot segment. The router would otherwise redirect it to the clean form.
func canonical(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.EscapedPath()
		clean := path.Clean("/" + p)
		if strings.HasSuffix(p, "/") && clean != "/" {
			clean += "/"
		}
		if p != clean {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Run serves until Stop shuts the HTTP server down. Cancellation doesn't reach the HTTP
// server, because Stop drains its connections.
func (s *httpServer) Run(context.Context) error {
	err := s.server.Serve(s.listener.ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("serve: %w", err)
}

// Stop drains open connections until the deadline passes, then closes the rest.
func (s *httpServer) Stop(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		_ = s.server.Close()
		return fmt.Errorf("drain connections: %w", err)
	}
	return nil
}

// idle is the service of a target whose work happens in the routes it adds to the HTTP
// server. It runs until the manager stops it.
type idle struct{}

// Start has nothing to set up.
func (idle) Start(context.Context) error { return nil }

// Run waits until the manager cancels it, since a return before then counts as a failure.
func (idle) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Stop has nothing to release.
func (idle) Stop(context.Context) error { return nil }

// apiRoutes adds the API routes.
func apiRoutes(mux *http.ServeMux) {
	// The {$} anchor limits the route to the root itself. Every other path gets 404.
	mux.HandleFunc("GET /api/v1/{$}", apiRoot)
	// Without this route, a request for the root's path without the slash gets a redirect
	// to the root instead of 404.
	mux.Handle("/api/v1", http.NotFoundHandler())
}

// apiRoot answers the API root with an empty JSON object.
func apiRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprint(w, "{}")
}
