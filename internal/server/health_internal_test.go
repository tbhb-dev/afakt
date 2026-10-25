// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
)

const (
	okBody    = `{"status":"pass"}`
	notOKBody = `{"status":"fail"}`

	healthyBody  = `{"status":"pass","checks":{"livez":{"status":"pass"},"readyz":{"status":"pass"}}}`
	notReadyBody = `{"status":"fail","checks":{"livez":{"status":"pass"},"readyz":{"status":"fail"}}}`
	notLiveBody  = `{"status":"fail","checks":{"livez":{"status":"fail"},"readyz":{"status":"pass"}}}`
	neitherBody  = `{"status":"fail","checks":{"livez":{"status":"fail"},"readyz":{"status":"fail"}}}`
)

// errDisconnected is the error of a failing readiness check.
var errDisconnected = errors.New("disconnected")

// fixed returns a check with a fixed result.
func fixed(pass bool) func() bool {
	return func() bool { return pass }
}

// TestHealthResults serves the health routes with each combination of check results. Each
// case compares the whole body, so a body with any key beyond the checks fails it.
func TestHealthResults(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		live, ready bool
		path        string
		code        int
		body        string
	}{
		{name: "livez pass", live: true, ready: true, path: "/livez", code: http.StatusOK, body: okBody},
		{
			name:  "livez fail",
			live:  false,
			ready: true,
			path:  "/livez",
			code:  http.StatusServiceUnavailable,
			body:  notOKBody,
		},
		{name: "readyz pass", live: true, ready: true, path: "/readyz", code: http.StatusOK, body: okBody},
		{
			name:  "readyz fail",
			live:  true,
			ready: false,
			path:  "/readyz",
			code:  http.StatusServiceUnavailable,
			body:  notOKBody,
		},
		{name: "healthz both pass", live: true, ready: true, path: "/healthz", code: http.StatusOK, body: healthyBody},
		{
			name: "healthz readyz fails", live: true, ready: false, path: "/healthz",
			code: http.StatusServiceUnavailable, body: notReadyBody,
		},
		{
			name: "healthz livez fails", live: false, ready: true, path: "/healthz",
			code: http.StatusServiceUnavailable, body: notLiveBody,
		},
		{
			name: "healthz both fail", live: false, ready: false, path: "/healthz",
			code: http.StatusServiceUnavailable, body: neitherBody,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()
			ready := fixed(tc.ready)
			health{live: fixed(tc.live), ready: func(context.Context) bool { return ready() }}.routes(mux)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, http.NoBody))
			contentType := rec.Header().Get("Content-Type")
			if rec.Code != tc.code || contentType != "application/json" || rec.Body.String() != tc.body {
				t.Errorf("GET %s = %d %q %q, want %d \"application/json\" %q",
					tc.path, rec.Code, contentType, rec.Body.String(), tc.code, tc.body)
			}
		})
	}
}

// TestReadinessFollowsContext checks that readiness fails once the context ends, even while
// every service still runs. A signal makes the server unready before its stop sequence.
func TestReadinessFollowsContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	manager := lifecycle.New(slog.New(slog.DiscardHandler))
	if err := manager.Register(lifecycle.Registration{
		Name: "a", Service: lifecycletest.NewFake("a", &lifecycletest.Recorder{}),
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	checks := newHealth(ctx, manager)
	if checks.ready(t.Context()) {
		t.Error("ready() before Start = true, want false")
	}
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !checks.live() || !checks.ready(t.Context()) {
		t.Errorf("live(), ready() while running = %t, %t, want true, true", checks.live(), checks.ready(t.Context()))
	}
	cancel()
	if !manager.Ready() || checks.ready(t.Context()) {
		t.Errorf("Ready(), ready() after cancel = %t, %t, want true, false", manager.Ready(), checks.ready(t.Context()))
	}
	if !checks.live() {
		t.Error("live() after cancel = false, want true")
	}
	if err := manager.Stop(t.Context()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

// TestReadinessFollowsChecks checks that a target's own checks feed readiness, each with
// the context of the request.
func TestReadinessFollowsChecks(t *testing.T) {
	t.Parallel()
	manager := lifecycle.New(slog.New(slog.DiscardHandler))
	if err := manager.Register(lifecycle.Registration{
		Name: "a", Service: lifecycletest.NewFake("a", &lifecycletest.Recorder{}),
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := manager.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		if err := manager.Stop(context.Background()); err != nil {
			t.Errorf("Stop() error = %v", err)
		}
	})
	var connErr error
	type ctxKey struct{}
	reqCtx := context.WithValue(t.Context(), ctxKey{}, "request")
	var seen any
	checks := newHealth(t.Context(), manager,
		func(context.Context) error { return nil },
		func(ctx context.Context) error {
			seen = ctx.Value(ctxKey{})
			return connErr
		})
	if !checks.ready(reqCtx) {
		t.Error("ready() with passing checks = false, want true")
	}
	if seen != "request" {
		t.Error("check didn't receive the request context")
	}
	connErr = errDisconnected
	if checks.ready(reqCtx) {
		t.Error("ready() with a failing check = true, want false")
	}
	if !checks.live() {
		t.Error("live() with a failing check = false, want true")
	}
}

// wantHealth checks all three health endpoints on a server that answers. The live check
// always passes, and /readyz and /healthz pass only when ready.
func wantHealth(t *testing.T, address string, ready bool) {
	t.Helper()
	readyCode, readyBody, healthBody := http.StatusOK, okBody, healthyBody
	if !ready {
		readyCode, readyBody, healthBody = http.StatusServiceUnavailable, notOKBody, notReadyBody
	}
	for _, want := range []struct {
		path string
		code int
		body string
	}{
		{path: "/livez", code: http.StatusOK, body: okBody},
		{path: "/readyz", code: readyCode, body: readyBody},
		{path: "/healthz", code: readyCode, body: healthBody},
	} {
		code, contentType, body := request(t, http.MethodGet, address, want.path)
		if code != want.code || contentType != "application/json" || body != want.body {
			t.Errorf("GET %s = %d %q %q, want %d \"application/json\" %q",
				want.path, code, contentType, body, want.code, want.body)
		}
	}
}

func TestHealthWhileStarting(t *testing.T) {
	t.Parallel()
	p := newProbe()
	entered, release := make(chan struct{}), make(chan struct{})
	p.fake.StartFunc = lifecycletest.Block(entered, release)
	s := startServer(t, testConfig(), hooks{services: []lifecycle.Registration{p.reg()}})
	// The API serves while the probe, which starts after it, is still starting.
	receive(t, entered, "the probe to start")
	wantHealth(t, s.addr(t), false)
	close(release)
	s.cancel()
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

func TestHealthWhileRunning(t *testing.T) {
	t.Parallel()
	p := newProbe()
	s := startServer(t, testConfig(), hooks{services: []lifecycle.Registration{p.reg()}})
	receive(t, p.up, "the probe to start")
	wantHealth(t, s.addr(t), true)
	s.cancel()
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

// TestHealthWhileDraining sends each shutdown signal to the test process. The signal
// context ends the server as the server command's does, and the probe, held in Stop, keeps
// the API serving while it drains.
func TestHealthWhileDraining(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the drain tests need Unix signals")
	}
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			t.Parallel()
			ctx, stop := signal.NotifyContext(t.Context(), sig)
			defer stop()
			p := newProbe()
			entered, release := make(chan struct{}), make(chan struct{})
			p.fake.StopFunc = lifecycletest.Block(entered, release)
			s := startServerContext(ctx, t, testConfig(), hooks{services: []lifecycle.Registration{p.reg()}})
			receive(t, p.up, "the probe to start")
			wantHealth(t, s.addr(t), true)
			self, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Fatalf("FindProcess() error = %v", err)
			}
			if err = self.Signal(sig); err != nil {
				t.Fatalf("Signal(%s) error = %v", sig, err)
			}
			receive(t, entered, "the probe to stop")
			wantHealth(t, s.addr(t), false)
			close(release)
			if err = s.wait(t); err != nil {
				t.Fatalf("run() error = %v", err)
			}
		})
	}
}

// TestHealthAfterRequiredFailure fails the probe and holds it in the Stop that follows, so
// the API still answers after the failure.
func TestHealthAfterRequiredFailure(t *testing.T) {
	t.Parallel()
	p := newProbe()
	entered, release := make(chan struct{}), make(chan struct{})
	p.fake.StopFunc = lifecycletest.Block(entered, release)
	s := startServer(t, testConfig(), hooks{services: []lifecycle.Registration{p.reg()}})
	receive(t, p.up, "the probe to start")
	p.fake.Fail(errBoom)
	receive(t, entered, "the probe to stop")
	wantHealth(t, s.addr(t), false)
	close(release)
	if err := s.wait(t); err == nil {
		t.Error("run() error = nil, want the probe's failure")
	}
}
