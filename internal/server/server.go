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

	"github.com/tbhb-dev/afakt/internal/lifecycle"
)

// readHeaderTimeout bounds how long a client may take to send request headers.
const readHeaderTimeout = 10 * time.Second

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

// run is Run with hooks for tests. The shutdown context keeps the values of the caller's
// context but not its cancellation, because a signal that cancels the caller's context
// starts the shutdown and the services still need time to stop.
func run(ctx context.Context, cfg Config, logger *slog.Logger, h hooks) error {
	manager := lifecycle.New(logger)
	if err := compose(manager, cfg, logger, h); err != nil {
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

// compose registers the services of each target with manager. A target without services
// returns ErrNotComposable.
func compose(manager *lifecycle.Manager, cfg Config, logger *slog.Logger, h hooks) error {
	var regs []lifecycle.Registration
	for _, target := range cfg.Targets {
		if target != TargetAPI {
			return fmt.Errorf("%w '%s'", ErrNotComposable, target)
		}
		ln := &listener{address: cfg.Listen, logger: logger}
		regs = append(regs,
			lifecycle.Registration{Name: "api-listener", Service: ln},
			lifecycle.Registration{
				Name:      "api",
				Service:   &api{listener: ln, routes: h.routes},
				DependsOn: []string{"api-listener"},
			},
		)
	}
	// Services from the hooks come last, after the services of every target.
	for _, reg := range append(regs, h.services...) {
		if err := manager.Register(reg); err != nil {
			return fmt.Errorf("compose: %w", err)
		}
	}
	return nil
}

// listener binds the TCP address of the API. On shutdown, the HTTP server closes the
// listener. Stop closes it too, for the case where the API never served.
type listener struct {
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

// Run has no work of its own. It waits until the manager cancels it, since a return
// before then counts as a failure.
func (l *listener) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Stop closes the listener. A listener that the HTTP server already closed isn't an error.
func (l *listener) Stop(context.Context) error {
	if err := l.ln.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("close listener: %w", err)
	}
	return nil
}

// api serves the API routes on the listener.
type api struct {
	listener *listener
	routes   func(mux *http.ServeMux)
	server   *http.Server
}

// Start builds the routes and the HTTP server without serving. Serving starts in Run.
func (a *api) Start(context.Context) error {
	mux := http.NewServeMux()
	// The {$} anchor limits the route to the root itself. Every other path gets 404.
	mux.HandleFunc("GET /api/v1/{$}", apiRoot)
	// Without this route, a request for the root's path without the slash gets a redirect
	// to the root instead of 404.
	mux.Handle("/api/v1", http.NotFoundHandler())
	if a.routes != nil {
		a.routes(mux)
	}
	a.server = &http.Server{Handler: canonical(mux), ReadHeaderTimeout: readHeaderTimeout}
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
func (a *api) Run(context.Context) error {
	err := a.server.Serve(a.listener.ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("serve: %w", err)
}

// Stop drains open connections until the deadline passes, then closes the rest.
func (a *api) Stop(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		_ = a.server.Close()
		return fmt.Errorf("drain connections: %w", err)
	}
	return nil
}

// apiRoot answers the API root with an empty JSON object.
func apiRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprint(w, "{}")
}
