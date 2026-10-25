// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
)

const (
	// testTimeout bounds each wait in these tests. It only matters when something hangs.
	testTimeout = 30 * time.Second

	apiRootPath = "/api/v1/"
)

var errBoom = errors.New("boom")

func testConfig() Config {
	// With port 0 the system chooses the port, and parallel tests can't collide on one.
	return Config{Targets: []string{TargetAPI}, Listen: "127.0.0.1:0", ShutdownTimeout: testTimeout}
}

// addressHandler drops log records, except that it reports the address of the listening
// record.
type addressHandler struct {
	addresses chan<- string
}

func (addressHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h addressHandler) WithAttrs([]slog.Attr) slog.Handler     { return h }
func (h addressHandler) WithGroup(string) slog.Handler          { return h }

func (h addressHandler) Handle(_ context.Context, record slog.Record) error {
	if record.Message == "listening" {
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "address" {
				h.addresses <- attr.Value.String()
			}
			return true
		})
	}
	return nil
}

// probe is a required service that starts after the API and stops before it. Its start
// closes up, at which point the API serves.
type probe struct {
	fake *lifecycletest.Fake
	up   chan struct{}
}

func newProbe() *probe {
	p := &probe{fake: lifecycletest.NewFake("probe", &lifecycletest.Recorder{}), up: make(chan struct{})}
	p.fake.StartFunc = func(context.Context) error {
		close(p.up)
		return nil
	}
	return p
}

func (p *probe) reg() lifecycle.Registration {
	return lifecycle.Registration{Name: "probe", Service: p.fake, DependsOn: []string{"api"}}
}

type harness struct {
	cancel    context.CancelFunc
	done      chan error
	addresses chan string
	address   string
}

func startServer(t *testing.T, cfg Config, h hooks) *harness {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	s := &harness{cancel: cancel, done: make(chan error, 1), addresses: make(chan string, 1)}
	logger := slog.New(addressHandler{addresses: s.addresses})
	go func() { s.done <- run(ctx, cfg, logger, h) }()
	t.Cleanup(cancel)
	return s
}

func (s *harness) addr(t *testing.T) string {
	t.Helper()
	// The listener logs its address once it binds, and this waits for that line.
	if s.address == "" {
		select {
		case s.address = <-s.addresses:
		case <-time.After(testTimeout):
			t.Fatal("the server never logged its address")
		}
	}
	return s.address
}

func (s *harness) wait(t *testing.T) error {
	t.Helper()
	select {
	case err := <-s.done:
		return err
	case <-time.After(testTimeout):
		t.Fatal("run did not return")
		return nil
	}
}

func receive(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(testTimeout):
		t.Fatalf("timed out waiting for %s", what)
	}
}

// request sends a request without credentials and returns the status, the media type,
// and the body.
func request(t *testing.T, method, address, path string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, "http://"+address+path, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	// The client doesn't follow redirects. A test sees the redirect status itself, not the
	// status of the page it points to.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s error = %v", method, path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return resp.StatusCode, resp.Header.Get("Content-Type"), string(body)
}

// wantAPIRoot checks that the API root answers 200 with an empty JSON object.
func wantAPIRoot(t *testing.T, address string) {
	t.Helper()
	code, contentType, body := request(t, http.MethodGet, address, apiRootPath)
	if code != http.StatusOK || contentType != "application/json" || body != "{}" {
		t.Errorf("GET %s = %d %q %q, want 200 \"application/json\" \"{}\"", apiRootPath, code, contentType, body)
	}
}

func wantClosed(t *testing.T, address string) {
	t.Helper()
	var config net.ListenConfig
	ln, err := config.Listen(t.Context(), "tcp", address)
	if err != nil {
		t.Fatalf("address %s is still in use: %v", address, err)
	}
	if err = ln.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestAPIRoutes(t *testing.T) {
	t.Parallel()
	p := newProbe()
	s := startServer(t, testConfig(), hooks{services: []lifecycle.Registration{p.reg()}})
	receive(t, p.up, "the probe to start")
	wantAPIRoot(t, s.addr(t))
	// The API serves its root and nothing else.
	for _, path := range []string{
		"/", "/livez", "/readyz", "/healthz", "/api", "/api/v1", "/api/v1/livez", "/api/v2/",
	} {
		if code, _, _ := request(t, http.MethodGet, s.addr(t), path); code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, code)
		}
	}
	if code, _, _ := request(t, http.MethodPost, s.addr(t), apiRootPath); code != http.StatusMethodNotAllowed {
		t.Errorf("POST %s = %d, want 405", apiRootPath, code)
	}
	s.cancel()
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

// TestAPIServesWhileStopping holds the probe in Stop. The API stops after the probe, so its
// root still answers until the probe has stopped.
func TestAPIServesWhileStopping(t *testing.T) {
	t.Parallel()
	p := newProbe()
	stopEntered, stopRelease := make(chan struct{}), make(chan struct{})
	p.fake.StopFunc = lifecycletest.Block(stopEntered, stopRelease)
	s := startServer(t, testConfig(), hooks{services: []lifecycle.Registration{p.reg()}})
	receive(t, p.up, "the probe to start")
	s.cancel()
	receive(t, stopEntered, "the probe to stop")
	wantAPIRoot(t, s.addr(t))
	close(stopRelease)
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	wantClosed(t, s.addr(t))
}

func TestRequiredFailureAfterStartup(t *testing.T) {
	t.Parallel()
	p := newProbe()
	entered, release := make(chan struct{}), make(chan struct{})
	p.fake.StopFunc = lifecycletest.Block(entered, release)
	s := startServer(t, testConfig(), hooks{services: []lifecycle.Registration{p.reg()}})
	receive(t, p.up, "the probe to start")
	// The failure starts the same shutdown that a signal starts.
	p.fake.Fail(errBoom)
	receive(t, entered, "the probe to stop")
	close(release)
	err := s.wait(t)
	if !errors.Is(err, errBoom) || err.Error() != `run "probe": boom` {
		t.Errorf("run() error = %v, want run \"probe\": boom", err)
	}
	wantClosed(t, s.addr(t))
}

func TestStartupFailureReleasesListener(t *testing.T) {
	t.Parallel()
	p := newProbe()
	p.fake.StartFunc = func(context.Context) error { return errBoom }
	s := startServer(t, testConfig(), hooks{services: []lifecycle.Registration{p.reg()}})
	err := s.wait(t)
	if !errors.Is(err, errBoom) || err.Error() != `start "probe": boom` {
		t.Errorf("run() error = %v, want start \"probe\": boom", err)
	}
	// The listener was bound before the probe failed, and the stop sequence closed it.
	wantClosed(t, s.addr(t))
}

func TestListenFailure(t *testing.T) {
	t.Parallel()
	var config net.ListenConfig
	taken, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer taken.Close()
	cfg := testConfig()
	cfg.Listen = taken.Addr().String()
	s := startServer(t, cfg, hooks{})
	err = s.wait(t)
	if !errors.Is(err, syscall.EADDRINUSE) || !strings.HasPrefix(err.Error(), `start "api-listener": listen: `) {
		t.Errorf("run() error = %v, want a listen failure with EADDRINUSE", err)
	}
}

// slowRoute adds /slow, which closes entered and answers once release closes or the
// connection ends.
func slowRoute(entered, release chan struct{}) func(mux *http.ServeMux) {
	return func(mux *http.ServeMux) {
		mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
			close(entered)
			select {
			case <-release:
				_, _ = fmt.Fprint(w, "done\n")
			case <-r.Context().Done():
			}
		})
	}
}

type response struct {
	code int
	body string
	err  error
}

func getAsync(t *testing.T, address, path string) <-chan response {
	t.Helper()
	result := make(chan response, 1)
	go func() {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+address+path, http.NoBody)
		if err != nil {
			result <- response{err: err}
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			result <- response{err: err}
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		result <- response{code: resp.StatusCode, body: string(body), err: err}
	}()
	return result
}

func waitRefused(t *testing.T, address string) {
	t.Helper()
	// A refused connection shows that the HTTP server has closed its listener to drain.
	deadline := time.Now().Add(testTimeout)
	var dialer net.Dialer
	for time.Now().Before(deadline) {
		conn, err := dialer.DialContext(t.Context(), "tcp", address)
		if err != nil {
			return
		}
		_ = conn.Close()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%s still accepts connections", address)
}

func TestShutdownDrainsRequests(t *testing.T) {
	t.Parallel()
	entered, release := make(chan struct{}), make(chan struct{})
	p := newProbe()
	s := startServer(t, testConfig(), hooks{
		services: []lifecycle.Registration{p.reg()},
		routes:   slowRoute(entered, release),
	})
	receive(t, p.up, "the probe to start")
	result := getAsync(t, s.addr(t), "/slow")
	receive(t, entered, "the slow request")
	s.cancel()
	// New connections fail from here on, but the slow request in flight still finishes.
	waitRefused(t, s.addr(t))
	close(release)
	got := <-result
	if got.err != nil || got.code != http.StatusOK || got.body != "done\n" {
		t.Errorf("slow request = %d %q %v, want 200 \"done\\n\"", got.code, got.body, got.err)
	}
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

func TestShutdownTimeout(t *testing.T) {
	t.Parallel()
	entered := make(chan struct{})
	p := newProbe()
	cfg := testConfig()
	cfg.ShutdownTimeout = 200 * time.Millisecond
	s := startServer(t, cfg, hooks{
		services: []lifecycle.Registration{p.reg()},
		routes:   slowRoute(entered, make(chan struct{})),
	})
	receive(t, p.up, "the probe to start")
	result := getAsync(t, s.addr(t), "/slow")
	receive(t, entered, "the slow request")
	// The slow request never finishes on its own. The shutdown has to end at its deadline.
	start := time.Now()
	s.cancel()
	err := s.wait(t)
	if elapsed := time.Since(start); elapsed < cfg.ShutdownTimeout {
		t.Errorf("shutdown took %s, want at least %s", elapsed, cfg.ShutdownTimeout)
	}
	want := `shutdown deadline exceeded while stopping "api"; not confirmed stopped: api, api-listener`
	if !errors.Is(err, lifecycle.ErrShutdownTimeout) || err.Error() != want {
		t.Errorf("run() error = %q, want %q", err, want)
	}
	if got := <-result; got.err == nil {
		t.Errorf("slow request = %d %q, want a closed connection", got.code, got.body)
	}
}

func TestRunRejectsComposition(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		targets []string
		want    error
	}{
		{name: "unavailable target", targets: []string{"webhook"}, want: ErrNotComposable},
		{name: "repeated target", targets: []string{TargetAPI, TargetAPI}, want: lifecycle.ErrDuplicateService},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := Config{Targets: tc.targets, Listen: "127.0.0.1:0", ShutdownTimeout: time.Second}
			if err := Run(t.Context(), cfg, slog.New(slog.DiscardHandler)); !errors.Is(err, tc.want) {
				t.Errorf("Run() error = %v, want %v", err, tc.want)
			}
		})
	}
}

type closeFailsListener struct {
	net.Listener
}

func (closeFailsListener) Close() error {
	return errBoom
}

func TestListenerStopReportsCloseError(t *testing.T) {
	t.Parallel()
	l := &listener{ln: closeFailsListener{}}
	if err := l.Stop(t.Context()); !errors.Is(err, errBoom) {
		t.Errorf("Stop() error = %v, want %v", err, errBoom)
	}
}

func TestAPIRunReportsServeError(t *testing.T) {
	t.Parallel()
	l := &listener{address: "127.0.0.1:0", logger: slog.New(slog.DiscardHandler)}
	if err := l.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	// Closing the listener before Serve makes Serve fail without any shutdown.
	if err := l.ln.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	a := &api{listener: l}
	if err := a.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := a.Run(t.Context()); !errors.Is(err, net.ErrClosed) {
		t.Errorf("Run() error = %v, want %v", err, net.ErrClosed)
	}
}
