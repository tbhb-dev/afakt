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

// hooks adds services and routes to the composition. Tests use it to hold the manager in
// a chosen state and to keep a request in flight.
type hooks struct {
	services []lifecycle.Spec
	routes   func(mux *http.ServeMux)
}

// Run starts the services of the selected targets and serves until the context ends or a
// required service fails. Then it stops every service within the configured shutdown
// timeout. A clean shutdown that the context started returns nil. A failed start, a failed
// service, or a shutdown that didn't finish in time returns an error.
func Run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	return run(ctx, cfg, logger, hooks{})
}

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

func compose(manager *lifecycle.Manager, cfg Config, logger *slog.Logger, h hooks) error {
	var specs []lifecycle.Spec
	for _, target := range cfg.Targets {
		if target != TargetAPI {
			return fmt.Errorf("%w '%s'", ErrNotComposable, target)
		}
		ln := &listener{address: cfg.Listen, logger: logger}
		specs = append(specs,
			lifecycle.Spec{Name: "api-listener", Service: ln},
			lifecycle.Spec{
				Name:      "api",
				Service:   &api{listener: ln, manager: manager, routes: h.routes},
				DependsOn: []string{"api-listener"},
			},
		)
	}
	// Services from the hooks come last, after the services of every target.
	for _, spec := range append(specs, h.services...) {
		if err := manager.Register(spec); err != nil {
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

func (l *listener) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (l *listener) Stop(context.Context) error {
	if err := l.ln.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("close listener: %w", err)
	}
	return nil
}

// api serves the health endpoints on the listener.
type api struct {
	listener *listener
	manager  *lifecycle.Manager
	routes   func(mux *http.ServeMux)
	server   *http.Server
}

func (a *api) Start(context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", livez)
	mux.HandleFunc("GET /readyz", a.readyz)
	if a.routes != nil {
		a.routes(mux)
	}
	a.server = &http.Server{Handler: mux, ReadHeaderTimeout: readHeaderTimeout}
	return nil
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

func livez(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintln(w, "ok")
}

// readyz answers 200 when every required service runs and 503 otherwise. Its body lists
// the state of each service. The other health endpoint answers 200 whenever the HTTP server
// can serve a request.
func (a *api) readyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if a.manager.Ready() {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ready")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintln(w, "not ready")
	}
	for _, status := range a.manager.Statuses() {
		_, _ = fmt.Fprintf(w, "%s %s\n", status.Name, status.State)
	}
}
