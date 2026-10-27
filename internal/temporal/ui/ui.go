// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package ui mounts the Temporal Web UI and its gRPC gateway on the shared HTTP server.
package ui

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"sync/atomic"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/temporalio/ui-server/v2/server/api"
	"github.com/temporalio/ui-server/v2/server/config"
	"github.com/temporalio/ui-server/v2/server/csrf"
	"github.com/temporalio/ui-server/v2/server/route"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DefaultPath is the root path for the UI on the shared listener.
const DefaultPath = "/temporal"

var errPath = errors.New(
	"expected a clean, non-root URL path without a trailing slash, " +
		"using only ASCII path-segment characters that need no percent-encoding and /",
)

var errPathConflict = errors.New("temporal UI path conflicts with a reserved path")

// CheckPath rejects paths that overlap another target or a health endpoint, even when that
// target isn't selected. A later addition of a target can't change where the UI routes go.
func CheckPath(prefix string) error {
	if prefix == "/" || !strings.HasPrefix(prefix, "/") || path.Clean(prefix) != prefix ||
		strings.ContainsFunc(prefix, func(r rune) bool {
			return !strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~$&+:=@/", r)
		}) {
		return fmt.Errorf("temporal UI path %q: %w", prefix, errPath)
	}
	for _, reserved := range []string{"/api", "/dashboard", "/livez", "/readyz", "/healthz"} {
		if prefix == reserved || strings.HasPrefix(prefix, reserved+"/") || strings.HasPrefix(reserved, prefix+"/") {
			return fmt.Errorf("%w: %q overlaps %q", errPathConflict, prefix, reserved)
		}
	}
	return nil
}

// Service manages a handler and its connection, but never opens an HTTP listener. The handler
// answers 404 until Start publishes it and after Stop removes it.
type Service struct {
	address   func() string
	namespace string
	prefix    string
	handler   atomic.Pointer[echo.Echo]
	conn      *grpc.ClientConn
	files     fs.FS
	dial      func(string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

// New returns the UI service for the embedded frontend. The frontend must start before it.
func New(address func() string, namespace, prefix string) *Service {
	return &Service{address: address, namespace: namespace, prefix: prefix, files: content(), dial: grpc.NewClient}
}

// Start creates a connection to the embedded frontend and mounts the upstream handlers.
func (s *Service) Start(context.Context) error {
	conn, err := s.dial(s.address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connect Temporal UI: %w", err)
	}
	cfg := &config.Config{
		TemporalGRPCAddress: s.address(), PublicPath: s.prefix, DefaultNamespace: s.namespace,
		EnableUI: true, DisableNewsFetch: true,
	}
	handler, err := newHandler(cfg, conn, s.prefix, s.files)
	if err != nil {
		_ = conn.Close()
		return err
	}
	s.conn = conn
	s.handler.Store(handler)
	return nil
}

// Run waits for the lifecycle manager to stop the service.
func (s *Service) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Stop removes the handler and closes the gateway connection without opening another port.
func (s *Service) Stop(context.Context) error {
	s.handler.Store(nil)
	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("close Temporal UI connection: %w", err)
	}
	return nil
}

// ServeHTTP delegates only while the UI is running.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := s.handler.Load(); handler != nil {
		handler.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

// newHandler uses the upstream UI routes and gateway with a connection that afakt manages. The
// upstream server's Start would own another listener and its route helper would hide the
// connection, so neither is used here.
func newHandler(provider config.ConfigProvider, conn *grpc.ClientConn, prefix string, files fs.FS) (*echo.Echo, error) {
	cfg, err := config.NewConfigProviderWithRefresh(provider)
	if err != nil {
		return nil, fmt.Errorf("temporal UI configuration: %w", err)
	}
	e := echo.New()
	e.Pre(route.PublicPath(prefix))
	e.Use(middleware.Recover(), middleware.Gzip(), middleware.Secure())
	e.Use(csrf.EnsureTokenCookie(false), middleware.CSRFWithConfig(csrf.MiddlewareConfig(cfg, false)))
	e.GET("/api/v1/settings", api.GetSettings(cfg))
	e.GET("/api/v1/ui-extensions", api.GetUIExtensions(cfg, nil))
	e.GET("/api/v1/releases", api.GetLatestReleases(cfg, api.NewReleaseChecker()))
	e.GET(api.WorkflowRawHistoryUrl, api.WorkflowRawHistoryHandler(workflowservice.NewWorkflowServiceClient(conn)))
	e.Any("/api/v1/*", api.TemporalAPIHandler(cfg, nil, conn))
	route.SetRenderRoute(e, "/")
	if err = route.SetUIRoutes(e, prefix, files, cfg); err != nil {
		return nil, fmt.Errorf("temporal UI assets: %w", err)
	}
	return e, nil
}

// mustFiles unwraps compile-time asset directories. A missing directory means packaging
// failed, independently of requests or server configuration.
func mustFiles(files fs.FS, err error) fs.FS {
	if err != nil {
		panic(err)
	}
	return files
}
