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
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
	"github.com/tbhb-dev/afakt/internal/temporal"
)

const (
	// testTimeout bounds each wait in these tests. It only matters when something hangs.
	testTimeout = 30 * time.Second

	apiRootPath = "/api/v1/"
)

var errBoom = errors.New("boom")

// testConfig returns a configuration for the API on 127.0.0.1 port 0. The system chooses
// the port, so parallel tests can't collide on one.
func testConfig() Config {
	return Config{Targets: []string{TargetAPI}, Listen: "127.0.0.1:0", ShutdownTimeout: testTimeout}
}

// recordHandler drops log records, except that it reports the address of the listening
// record and records each service transition as the service name, a space and the state.
type recordHandler struct {
	addresses chan<- string
	events    *events
}

func (recordHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h recordHandler) WithAttrs([]slog.Attr) slog.Handler     { return h }
func (h recordHandler) WithGroup(string) slog.Handler          { return h }

func (h recordHandler) Handle(_ context.Context, record slog.Record) error {
	record.Attrs(func(attr slog.Attr) bool {
		switch {
		case record.Message == "listening" && attr.Key == "address":
			h.addresses <- attr.Value.String()
		case strings.HasPrefix(record.Message, "service ") && attr.Key == "service":
			h.events.add(attr.Value.String() + " " + strings.TrimPrefix(record.Message, "service "))
		}
		return true
	})
	return nil
}

// events is the list of service transitions a server logged.
type events struct {
	mu   sync.Mutex
	list []string
}

func (e *events) add(event string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.list = append(e.list, event)
}

// matching returns the recorded transitions that end in state, in order.
func (e *events) matching(state string) []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	var names []string
	for _, event := range e.list {
		if name, ok := strings.CutSuffix(event, " "+state); ok {
			names = append(names, name)
		}
	}
	return names
}

// probe is a required service that starts after the service it depends on and stops before
// it. Its start closes up.
type probe struct {
	fake *lifecycletest.Fake
	up   chan struct{}
}

// newProbe returns a probe whose up hasn't closed yet.
func newProbe() *probe {
	p := &probe{fake: lifecycletest.NewFake("probe", &lifecycletest.Recorder{}), up: make(chan struct{})}
	p.fake.StartFunc = func(context.Context) error {
		close(p.up)
		return nil
	}
	return p
}

// reg returns the registration that makes the probe depend on the HTTP server.
func (p *probe) reg() lifecycle.Registration {
	return p.regOn("http")
}

// regOn returns the registration that makes the probe depend on service.
func (p *probe) regOn(service string) lifecycle.Registration {
	return lifecycle.Registration{Name: "probe", Service: p.fake, DependsOn: []string{service}}
}

// harness is a server that run serves in the background. cancel ends it like a signal, done
// receives what run returned, address caches the first address the listener logged, and
// events records the service transitions.
type harness struct {
	cancel    context.CancelFunc
	done      chan error
	addresses chan string
	address   string
	events    *events
	listener  net.Listener
	listeners []net.Listener
}

// startServer starts run in the background and returns at once. The test's
// cleanup cancels the server.
func startServer(t *testing.T, cfg Config, h hooks) *harness {
	t.Helper()
	return startServerContext(t.Context(), t, cfg, h)
}

// startServerContext is `startServer` under parent, whose end stops the server as a signal does.
func startServerContext(parent context.Context, t *testing.T, cfg Config, h hooks) *harness {
	t.Helper()
	ctx, cancel := context.WithCancel(parent)
	s := &harness{cancel: cancel, done: make(chan error, 1), addresses: make(chan string, 1), events: &events{}}
	h.listen = func(ctx context.Context, network, address string) (net.Listener, error) {
		ln, err := listenTCP(ctx, network, address)
		if err == nil {
			if s.listener == nil {
				s.listener = ln
			}
			s.listeners = append(s.listeners, ln)
			s.events.add(ln.Addr().String() + " bound")
		}
		return ln, err
	}
	logger := slog.New(recordHandler{addresses: s.addresses, events: s.events})
	go func() { s.done <- run(ctx, cfg, logger, h) }()
	t.Cleanup(cancel)
	return s
}

// addr returns the address the listener bound. The listener logs it once it binds, so the
// first call waits for that line and stops the test if it never comes.
func (s *harness) addr(t *testing.T) string {
	t.Helper()
	if s.address == "" {
		select {
		case s.address = <-s.addresses:
		case <-time.After(testTimeout):
			t.Fatal("the server never logged its address")
		}
	}
	return s.address
}

// wait returns what run returned, and stops the test if run doesn't return in time.
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

// receive waits for ch to close or send, and stops the test, naming what, if it doesn't in
// time.
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

// wantClosed checks that the server released address, by binding it again.
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
	wantHealth(t, s.addr(t), true)
	// The API serves its root and nothing else beside the health routes.
	for _, path := range []string{
		"/", "/api", "/api/v1", "/api/v2/",
		"/api/v1/livez", "/api/v1/readyz", "/api/v1/healthz", "/livez/", "/readyz/", "/healthz/",
		// Paths that clean to the root get 404, not a redirect to it.
		"/api//v1/", "/api/v1/./", "/api/v1//", "/api/v2/../v1/",
	} {
		if code, _, _ := request(t, http.MethodGet, s.addr(t), path); code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, code)
		}
	}
	for _, path := range []string{apiRootPath, "/livez", "/readyz", "/healthz"} {
		if code, _, _ := request(t, http.MethodPost, s.addr(t), path); code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s = %d, want 405", path, code)
		}
	}
	s.cancel()
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	// The API target alone doesn't start a Temporal service.
	if got, want := s.events.matching(
		"starting",
	), []string{
		"http-listener",
		"http",
		"api",
		"probe",
	}; !slices.Equal(
		got,
		want,
	) {
		t.Errorf("started %q, want %q", got, want)
	}
}

// TestAPIServesWhileStopping holds the probe in Stop. The HTTP server stops after the probe,
// so the API root and liveness still answer until the probe has stopped, and readiness fails
// from the start of the shutdown.
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
	wantHealth(t, s.addr(t), false)
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
	if !errors.Is(err, syscall.EADDRINUSE) || !strings.HasPrefix(err.Error(), `start "http-listener": listen: `) {
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

// response is the result of one background request.
type response struct {
	code int
	body string
	err  error
}

// getAsync sends a GET request in the background and returns a channel that receives its
// response, so the test can act while the request is in flight.
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

// waitRefused waits until address refuses connections, which shows that the HTTP server
// has closed its listener to drain. It stops the test if that doesn't happen in time.
func waitRefused(t *testing.T, address string) {
	t.Helper()
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
	for name, cfg := range targetConfigs() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			wantDrainedRequest(t, cfg)
		})
	}
}

// wantDrainedRequest holds a shared-listener request until shutdown has stopped accepting connections.
func wantDrainedRequest(t *testing.T, cfg Config) {
	t.Helper()
	entered, release := make(chan struct{}), make(chan struct{})
	p := newProbe()
	s := startServer(t, cfg, hooks{
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
	for _, targets := range [][]string{{TargetAPI}, {TargetWebhook}, {TargetDashboard}, {TargetAPI, TargetDashboard}} {
		t.Run(targets[0], func(t *testing.T) {
			t.Parallel()
			entered := make(chan struct{})
			p := newProbe()
			cfg := testConfig()
			cfg.Targets = targets
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
			want := `shutdown deadline exceeded while stopping "http"; not confirmed stopped: http, http-listener`
			if !errors.Is(err, lifecycle.ErrShutdownTimeout) || err.Error() != want {
				t.Errorf("run() error = %q, want %q", err, want)
			}
			if got := <-result; got.err == nil {
				t.Errorf("slow request = %d %q, want a closed connection", got.code, got.body)
			}
		})
	}
}

func TestRunRejectsComposition(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		targets []string
		want    error
	}{
		{name: "unknown target", targets: []string{"unknown"}, want: ErrNotComposable},
		{name: "repeated target", targets: []string{TargetAPI, TargetAPI}, want: lifecycle.ErrDuplicateService},
		{
			name:    "repeated webhook",
			targets: []string{TargetWebhook, TargetAPI, TargetWebhook},
			want:    lifecycle.ErrDuplicateService,
		},
		{name: "all", targets: []string{TargetWorker, "all"}, want: ErrNotComposable},
		{name: "repeated worker", targets: []string{TargetWorker, TargetWorker}, want: lifecycle.ErrDuplicateService},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := Config{
				Targets: tc.targets, Listen: "127.0.0.1:0", ShutdownTimeout: time.Second,
				Temporal: temporal.Config{Mode: temporal.ModeEmbedded, Namespace: temporal.DefaultNamespace},
			}
			if err := Run(t.Context(), cfg, slog.New(slog.DiscardHandler)); !errors.Is(err, tc.want) {
				t.Errorf("Run() error = %v, want %v", err, tc.want)
			}
		})
	}
}

// closeFailsListener is a listener whose Close always fails, for the error path of
// the listener's Stop.
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
	l := &listener{address: "127.0.0.1:0", logger: slog.New(slog.DiscardHandler), listen: listenTCP}
	if err := l.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	// Closing the listener before Serve makes Serve fail without any shutdown.
	if err := l.ln.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	a := &httpServer{listener: l}
	if err := a.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := a.Run(t.Context()); !errors.Is(err, net.ErrClosed) {
		t.Errorf("Run() error = %v, want %v", err, net.ErrClosed)
	}
}
