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
	"slices"
	"strings"
	"time"

	"github.com/tbhb-dev/afakt/internal/dashboard"
	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/temporal"
	"github.com/tbhb-dev/afakt/internal/temporal/embedded"
	"github.com/tbhb-dev/afakt/internal/temporal/ui"
)

// readHeaderTimeout bounds how long a client may take to send request headers.
const readHeaderTimeout = 10 * time.Second

// The names of the services every composition shares. Each target also runs one service
// under its own name, which depends on the HTTP server.
const (
	listenerService = "http-listener"
	httpService     = "http"
	temporalService = "temporal-server"
)

// ErrNotComposable reports a target that Run has no services for.
var ErrNotComposable = errors.New("no services for target")

// Config is a checked server configuration. Targets comes from ParseTargets. Temporal applies
// only to the worker target.
type Config struct {
	Targets         []string
	Listen          string
	ShutdownTimeout time.Duration
	Temporal        temporal.Config
	TemporalUIPath  string
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
	// Cancellation starts one budget, including cleanup of a service whose Start fails.
	deadline := make(chan time.Time, 1)
	stopDeadline := context.AfterFunc(ctx, func() { deadline <- time.Now().Add(cfg.ShutdownTimeout) })
	defer stopDeadline()
	manager := lifecycle.New(logger)
	if err := compose(ctx, manager, cfg, logger, h); err != nil {
		return err
	}
	err := manager.Start(ctx)
	if err == nil {
		logger.InfoContext(ctx, "server ready", "targets", cfg.Targets)
		err = manager.Wait(ctx)
	}
	logger.InfoContext(ctx, "shutting down", "timeout", cfg.ShutdownTimeout.String())
	end := time.Now().Add(cfg.ShutdownTimeout)
	if ctx.Err() != nil {
		end = <-deadline
	}
	stopCtx, cancel := context.WithDeadline(context.WithoutCancel(ctx), end)
	defer cancel()
	return errors.Join(err, manager.Stop(stopCtx))
}

// compose registers the services of the selected targets with manager. Every target shares
// one listener and one HTTP server, which serves the health routes, and the HTTP server
// starts first so that readiness answers while the other services start. A target without
// services returns ErrNotComposable.
func compose(ctx context.Context, manager *lifecycle.Manager, cfg Config, logger *slog.Logger, h hooks) error {
	if cfg.TemporalUIPath == "" {
		cfg.TemporalUIPath = ui.DefaultPath
	}
	if err := ui.CheckPath(cfg.TemporalUIPath); err != nil {
		return fmt.Errorf("compose: %w", err)
	}
	ln := &listener{address: cfg.Listen, logger: logger}
	web := &httpServer{listener: ln, routes: []func(*http.ServeMux){nil}}
	var checks []func(context.Context) error
	regs := []lifecycle.Registration{
		{Name: listenerService, Service: ln},
		{Name: httpService, Service: web, DependsOn: []string{listenerService}},
	}
	for _, target := range cfg.Targets {
		switch target {
		case TargetAPI:
			web.routes = append(web.routes, apiRoutes)
		case TargetWebhook:
			// The webhook target has no receivers yet.
		case TargetDashboard:
			web.dashboard = true
			web.routes = append(web.routes, dashboardRoutes)
		case TargetWorker:
			services, check := workerServices(cfg.Temporal, logger, cfg.ShutdownTimeout, cfg.TemporalUIPath, web)
			regs = append(regs, services...)
			checks = append(checks, check)
			continue
		default:
			return fmt.Errorf("%w '%s'", ErrNotComposable, target)
		}
		regs = append(regs, lifecycle.Registration{Name: target, Service: idle{}, DependsOn: []string{httpService}})
	}
	if h.routes != nil {
		web.routes = append(web.routes, h.routes)
	}
	probes := newHealth(ctx, manager, checks...)
	web.routes[0] = probes.probeRoutes
	if slices.Contains(cfg.Targets, TargetAPI) || slices.Contains(cfg.Targets, TargetWorker) {
		web.routes[0] = probes.routes
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

// workerServices returns the Temporal services of the worker target and adds the client's
// check to readiness. In embedded mode the client depends on the embedded server, so the
// server starts first and stops last, and the client connects to the address the server
// bound.
func workerServices(
	cfg temporal.Config,
	logger *slog.Logger,
	shutdownTimeout time.Duration,
	uiPath string,
	web *httpServer,
) ([]lifecycle.Registration, func(context.Context) error) {
	var regs []lifecycle.Registration
	var dependsOn []string
	address := func() string { return cfg.Address }
	if cfg.Mode == temporal.ModeEmbedded {
		server := embedded.New(cfg.EmbeddedDir, cfg.Namespace, logger, shutdownTimeout)
		regs = append(regs, lifecycle.Registration{Name: temporalService, Service: server})
		dependsOn = []string{temporalService}
		address = server.Address
		view := ui.New(address, cfg.Namespace, uiPath)
		web.routes = append(web.routes, func(mux *http.ServeMux) {
			mux.Handle(uiPath, view)
			mux.Handle(uiPath+"/", view)
		})
		regs = append(regs, lifecycle.Registration{
			Name: "temporal-ui", Service: view, DependsOn: []string{temporalService, httpService},
		})
	}
	client := temporal.NewClient(address, cfg.Namespace, logger)
	return append(
		regs,
		lifecycle.Registration{Name: "temporal-client", Service: client, DependsOn: dependsOn},
	), client.Check
}

// listener binds the TCP address of the HTTP server. On shutdown, the HTTP server closes the
// listener. Stop closes it too, for the case where the HTTP server never served.
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
	listener  *listener
	routes    []func(mux *http.ServeMux)
	server    *http.Server
	dashboard bool
}

// Start builds the routes and the HTTP server without serving. Serving starts in Run.
func (s *httpServer) Start(context.Context) error {
	mux := http.NewServeMux()
	for _, add := range s.routes {
		add(mux)
	}
	handler := canonical(mux)
	if s.dashboard {
		handler = dashboard.SecurityHeaders(handler)
	}
	s.server = &http.Server{Handler: handler, ReadHeaderTimeout: readHeaderTimeout}
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
