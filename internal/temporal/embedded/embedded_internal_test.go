// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package embedded

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/temporal"
	"go.uber.org/fx"

	sqliteschema "go.temporal.io/server/schema/sqlite"
)

var (
	errBuild  = errors.New("build failed")
	errSchema = errors.New("schema failed")
	errStart  = errors.New("start failed")
	errStop   = errors.New("stop failed")
)

// fakeServer is a Temporal server whose Start and Stop return the errors in its fields.
type fakeServer struct {
	start, stop error
	stopped     bool
}

func (f *fakeServer) Start(context.Context) error { return f.start }

func (f *fakeServer) Stop(context.Context) error {
	f.stopped = true
	return f.stop
}

// newTestServer returns a server for the default namespace that logs into the buffer.
func newTestServer(dir string, buf *bytes.Buffer) *Server {
	// Real Temporal service shutdown includes its mandatory drain. Clock tests below
	// separately verify short, exact deadlines without waiting for those drain timers.
	return New(dir, "default", slog.New(slog.NewTextHandler(buf, nil)), 15*time.Second, (&net.ListenConfig{}).Listen)
}

// startStop checks that s starts and serves before it stops s.
func startStop(t *testing.T, s *Server, buf *bytes.Buffer) {
	t.Helper()
	if err := s.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	want := `msg="temporal embedded server started" address=` + s.Address()
	if !strings.HasPrefix(s.Address(), loopback+":") || !strings.Contains(buf.String(), want) {
		t.Errorf("Address() = %q and log %q, want a loopback address that the log names", s.Address(), buf)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := s.Run(ctx); err != nil {
		t.Errorf("Run() error = %v", err)
	}
	if err := s.Stop(t.Context()); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	wantListenersClosed(t, s.listeners)
}

func TestServersShareAProcess(t *testing.T) {
	t.Parallel()
	dirs := []string{"", "", filepath.Join(t.TempDir(), "a"), filepath.Join(t.TempDir(), "b")}
	servers := startAll(t, dirs)
	addresses := map[string]bool{}
	for i, s := range servers {
		addresses[s.Address()] = true
		if dirs[i] != "" {
			if _, err := os.Stat(filepath.Join(dirs[i], DatabaseFile)); err != nil {
				t.Errorf("database file: %v", err)
			}
		}
	}
	if len(addresses) != len(servers) {
		t.Errorf("addresses = %v, want one for each server", addresses)
	}
	wantRuntimeCollectorsOnly(t)
	for _, s := range servers {
		if err := s.Stop(t.Context()); err != nil {
			t.Errorf("Stop() error = %v", err)
		}
	}
}

// startAll starts one server for each directory in parallel and returns them in the order of
// the directories.
func startAll(t *testing.T, dirs []string) []*Server {
	t.Helper()
	servers := make([]*Server, len(dirs))
	logs := make([]bytes.Buffer, len(dirs))
	var wg sync.WaitGroup
	for i, dir := range dirs {
		servers[i] = newTestServer(dir, &logs[i])
		wg.Go(func() {
			if err := servers[i].Start(t.Context()); err != nil {
				t.Errorf("Start(%q) error = %v", dir, err)
			}
		})
	}
	wg.Wait()
	return servers
}

// wantRuntimeCollectorsOnly checks that the default Prometheus registry has only the Go
// runtime and process collectors, which shows that the servers didn't register metrics there.
func wantRuntimeCollectorsOnly(t *testing.T) {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	for _, family := range families {
		if name := family.GetName(); !strings.HasPrefix(name, "go_") && !strings.HasPrefix(name, "process_") {
			t.Errorf("default registry has %q, want the runtime collectors only", name)
		}
	}
}

func TestServerRestartsOnItsFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for range 2 {
		var buf bytes.Buffer
		startStop(t, newTestServer(dir, &buf), &buf)
	}
}

func TestServerStartFailures(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	notDatabase := t.TempDir()
	if err := os.Mkdir(filepath.Join(notDatabase, DatabaseFile), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []startFailure{
		{name: "directory", dir: filepath.Join(file, "dir"), want: "create embedded Temporal directory: "},
		{name: "schema", dir: t.TempDir(), schema: errSchema, want: "set up embedded Temporal schema: schema failed"},
		{name: "namespace", dir: notDatabase, want: `create embedded Temporal namespace "default": `},
		{name: "ports", host: "256.0.0.1", want: "pick an embedded Temporal port: "},
		{name: "build", build: errBuild, want: "build embedded Temporal server: build failed"},
		{
			name: "start", server: &fakeServer{start: errStart}, stopped: true,
			want: "start embedded Temporal server: start failed",
		},
		{
			name: "start and stop", server: &fakeServer{start: errStart, stop: errStop}, stopped: true,
			want: "start embedded Temporal server: start failed\nstop embedded Temporal server: stop failed",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t)
		})
	}
}

// startFailure is a case of `TestServerStartFailures`. A set server or build error replaces
// the server library's constructor. A schema error replaces setup without relying on file permissions.
type startFailure struct {
	name, dir, host string
	server          runtimeServer
	build           error
	schema          error
	want            string
	stopped         bool
}

// check starts a server set up as the case says and checks the error and whether the fake
// server stopped.
func (tc startFailure) check(t *testing.T) {
	t.Helper()
	var buf bytes.Buffer
	s := newTestServer(tc.dir, &buf)
	listen := s.listen
	var listeners []net.Listener
	s.listen = func(ctx context.Context, network, address string) (net.Listener, error) {
		ln, err := listen(ctx, network, address)
		if err == nil {
			listeners = append(listeners, ln)
		}
		return ln, err
	}
	if tc.schema != nil {
		s.setupSchema = func(*config.SQL) error { return tc.schema }
	}
	if tc.host != "" {
		s.host = tc.host
	}
	if tc.server != nil || tc.build != nil {
		s.newServer = func(context.Context, []net.Listener, ...temporal.ServerOption) (runtimeServer, error) {
			return tc.server, tc.build
		}
	}
	err := s.Start(t.Context())
	if err == nil || !strings.HasPrefix(err.Error(), tc.want) {
		t.Errorf("Start() error = %v, want prefix %q", err, tc.want)
	}
	if tc.schema != nil && !errors.Is(err, tc.schema) {
		t.Errorf("Start() error = %v, want wrapped schema error %v", err, tc.schema)
	}
	if fake, ok := tc.server.(*fakeServer); ok && fake.stopped != tc.stopped {
		t.Errorf("stopped = %v, want %v", fake.stopped, tc.stopped)
	}
	wantListenersClosed(t, listeners)
}

func TestServerStopFailure(t *testing.T) {
	t.Parallel()
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", net.JoinHostPort(loopback, "0"))
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{server: &fakeServer{stop: errStop}, listeners: []net.Listener{ln}}
	if err = s.Stop(t.Context()); !errors.Is(err, errStop) {
		t.Errorf("Stop() error = %v, want %v", err, errStop)
	}
	wantListenersClosed(t, s.listeners)
}

func TestServerBindFailure(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := newTestServer("", &buf)
	listen := s.listen
	var acquired net.Listener
	var bindErr error
	s.listen = func(ctx context.Context, network, address string) (net.Listener, error) {
		if acquired == nil {
			var err error
			acquired, err = listen(ctx, network, address)
			return acquired, err
		}
		ln, err := listen(ctx, network, acquired.Addr().String())
		bindErr = err
		return ln, err
	}
	err := s.Start(t.Context())
	if bindErr == nil || !errors.Is(err, bindErr) || !errors.Is(err, syscall.EADDRINUSE) {
		t.Fatalf("Start() error = %v, want original bind error %v", err, bindErr)
	}
	wantListenersClosed(t, []net.Listener{acquired})
}

func TestServerHTTPBindFailure(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := newTestServer("", &buf)
	listen := s.listen
	var listeners []net.Listener
	s.listen = func(ctx context.Context, network, address string) (net.Listener, error) {
		ln, err := listen(ctx, network, address)
		listeners = append(listeners, ln)
		return ln, err
	}
	var conflict net.Listener
	s.newServer = func(ctx context.Context, owned []net.Listener, opts ...temporal.ServerOption) (runtimeServer, error) {
		i := len(owned) - 1
		owned[i] = &onCloseListener{Listener: owned[i], after: func() {
			var err error
			conflict, err = listen(ctx, "tcp", listeners[i].Addr().String())
			if err != nil {
				t.Error(err)
			}
		}}
		return newOwnedServer(ctx, owned, opts...)
	}
	if err := s.Start(t.Context()); !errors.Is(err, syscall.EADDRINUSE) {
		t.Fatalf("Start() error = %v, want wrapped bind error", err)
	}
	if conflict == nil {
		t.Fatal("HTTP reservation was not released")
	}
	_ = conflict.Close()
	wantListenersClosed(t, listeners)
}

type onCloseListener struct {
	net.Listener

	once  sync.Once
	after func()
}

func (l *onCloseListener) Close() error {
	err := l.Listener.Close()
	l.once.Do(l.after)
	if err != nil {
		return fmt.Errorf("close reservation: %w", err)
	}
	return nil
}

func TestConstructionRollback(t *testing.T) {
	t.Parallel()
	for fail := range 5 {
		t.Run(strconv.Itoa(fail), func(t *testing.T) {
			t.Parallel()
			checkConstructionRollback(t, fail)
		})
	}
}

func checkConstructionRollback(t *testing.T, fail int) {
	t.Helper()
	var buf bytes.Buffer
	s := newTestServer("", &buf)
	var acquired, released []int
	s.newServer = func(ctx context.Context, listeners []net.Listener, opts ...temporal.ServerOption) (runtimeServer, error) {
		next := 0
		return constructOwnedServer(ctx, listeners, func(options ...fx.Option) *fx.App {
			step := next
			next++
			return fx.New(fx.Invoke(func(lc fx.Lifecycle) {
				acquired = append(acquired, step)
				lc.Append(fx.StopHook(func() error { released = append(released, step); return errStop }))
			}), fx.Options(options...), fx.Invoke(func() error {
				if step == fail {
					return errBuild
				}
				return nil
			}))
		}, opts...)
	}
	if err := s.Start(t.Context()); !errors.Is(err, errBuild) || !errors.Is(err, errStop) {
		t.Fatalf("Start() = %v, want construction and cleanup errors", err)
	}
	// Check the actual acquisition order, including the outer app before its children.
	want := slices.Clone(acquired)
	slices.Reverse(want)
	if !slices.Equal(released, want) {
		t.Errorf("released = %v, want %v", released, want)
	}
	wantListenersClosed(t, s.listeners)
	// The private HTTP listener is released even when its app never started.
	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", s.listeners[len(s.listeners)-1].Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	_ = ln.Close()
}

func TestOwnedStartupRollback(t *testing.T) {
	t.Parallel()
	for fail := range 5 {
		t.Run(strconv.Itoa(fail), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			s := newTestServer("", &buf)
			s.newServer = func(ctx context.Context, listeners []net.Listener, opts ...temporal.ServerOption) (runtimeServer, error) {
				next := 0
				return constructOwnedServer(ctx, listeners, func(options ...fx.Option) *fx.App {
					step := next
					next++
					return fx.New(append(options, fx.Invoke(func(lc fx.Lifecycle) {
						lc.Append(fx.StartHook(func(context.Context) error {
							if step == fail {
								return errStart
							}
							return nil
						}))
					}))...)
				}, opts...)
			}
			if err := s.Start(t.Context()); !errors.Is(err, errStart) || errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("Start() = %v, want start error", err)
			}
			wantListenersClosed(t, s.listeners)
			ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", s.listeners[len(s.listeners)-1].Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			_ = ln.Close()
		})
	}
}

func TestOwnedHookStartupRollback(t *testing.T) {
	t.Parallel()
	for _, failure := range []struct {
		name  string
		start func() error
		want  error
	}{
		{"error", func() error { return errStart }, errStart},
		{"panic", func() error { panic(errStart) }, errLifecyclePanic},
	} {
		t.Run(failure.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				checkOwnedHookStartupRollback(t, failure.start, failure.want)
			})
		})
	}
}

func checkOwnedHookStartupRollback(t *testing.T, start func() error, want error) {
	t.Helper()
	var buf bytes.Buffer
	s := newTestServer("", &buf)
	s.createNamespaces = func(*config.SQL, ...*sqliteschema.NamespaceConfig) error { return nil }
	s.listen = func(context.Context, string, string) (net.Listener, error) { return &clockListener{}, nil }
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var release context.CancelFunc
	released := make(chan struct{})
	var stopped []int
	owner := &ownedApp{}
	owner.app = fx.New(fx.NopLogger, owner.options(), fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error { return nil },
			OnStop:  func(context.Context) error { stopped = append(stopped, 0); return nil },
		})
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				var resourceCtx context.Context
				resourceCtx, release = context.WithCancel(context.WithoutCancel(ctx))
				t.Cleanup(release)
				go func() { <-resourceCtx.Done(); close(released) }()
				cancel()
				return start()
			},
			OnStop: func(ctx context.Context) error {
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) != s.shutdownTimeout || ctx.Err() != nil {
					t.Errorf("cleanup context = %v, deadline = %v", ctx.Err(), deadline)
				}
				release()
				<-released
				stopped = append(stopped, 1)
				return errStop
			},
		})
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error { t.Error("later hook ran"); return nil },
			OnStop:  func(context.Context) error { t.Error("stopped a hook that never ran"); return nil },
		})
	}))
	s.newServer = func(context.Context, []net.Listener, ...temporal.ServerOption) (runtimeServer, error) {
		return &ownedServer{top: owner}, nil
	}
	err := s.Start(ctx)
	if !errors.Is(err, want) || !errors.Is(err, errStop) || !errors.Is(err, context.Canceled) ||
		!strings.Contains(err.Error(), errStart.Error()) {
		t.Errorf("Start() = %v, want original failure, cancellation, and cleanup error", err)
	}
	select {
	case <-released:
	default:
		t.Error("startup resource was not released")
	}
	if !slices.Equal(stopped, []int{1, 0}) {
		t.Errorf("stopped = %v, want failing hook then earlier hook", stopped)
	}
}

func TestOwnedHookCancellation(t *testing.T) {
	t.Parallel()
	owner := &ownedApp{}
	owner.app = fx.New(fx.NopLogger, owner.options(), fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.StartHook(func() { t.Error("cancelled hook ran") }))
	}))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	s := &ownedServer{top: owner}
	if err := s.Start(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Start() = %v", err)
	}
	if err := s.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestConstructionCancellation(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := newTestServer("", &buf)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	constructed := 0
	s.newServer = func(ctx context.Context, listeners []net.Listener, opts ...temporal.ServerOption) (runtimeServer, error) {
		return constructOwnedServer(ctx, listeners, func(options ...fx.Option) *fx.App {
			constructed++
			app := fx.New(options...)
			cancel()
			return app
		}, opts...)
	}
	if err := s.Start(ctx); !errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Start() = %v, want cancellation with completed rollback", err)
	}
	if constructed != 2 {
		t.Errorf("constructed %d apps, want outer app and matching only", constructed)
	}
	wantListenersClosed(t, s.listeners)
}

func TestStartAlreadyCancelled(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := newTestServer("", &buf)
	s.createNamespaces = func(*config.SQL, ...*sqliteschema.NamespaceConfig) error {
		t.Error("prepared database after cancellation")
		return nil
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := s.Start(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Start() = %v", err)
	}
	if len(s.listeners) != 0 {
		t.Error("acquired listeners after cancellation")
	}
}

func TestStartupCancellationCleanup(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		s := newTestServer("", &buf)
		s.createNamespaces = func(*config.SQL, ...*sqliteschema.NamespaceConfig) error { return nil }
		s.listen = func(context.Context, string, string) (net.Listener, error) { return &clockListener{}, nil }
		entered := make(chan struct{})
		s.newServer = func(context.Context, []net.Listener, ...temporal.ServerOption) (runtimeServer, error) {
			return controlledServer{start: func(ctx context.Context) error {
				close(entered)
				<-ctx.Done()
				return errors.Join(errStart, fmt.Errorf("controlled start: %w", ctx.Err()))
			}, stop: func(context.Context) error { return errStop }}, nil
		}
		ctx, cancel := context.WithCancel(t.Context())
		result := make(chan error, 1)
		go func() { result <- s.Start(ctx) }()
		<-entered
		cancel()
		err := <-result
		var interrupted *StartupError
		if !errors.As(err, &interrupted) || !errors.Is(err, context.Canceled) || !errors.Is(err, errStop) ||
			!strings.Contains(err.Error(), "embedded Temporal startup interrupted") {
			t.Fatalf("Start() = %v, want typed cancellation and cleanup error", err)
		}
		wantListenersClosed(t, s.listeners)
	})
}

type controlledServer struct {
	start func(context.Context) error
	stop  func(context.Context) error
}

func (s controlledServer) Start(ctx context.Context) error { return s.start(ctx) }
func (s controlledServer) Stop(ctx context.Context) error  { return s.stop(ctx) }

func TestStartupCancellationDeadline(t *testing.T) {
	t.Parallel()
	for _, phase := range []string{"construction", "startup", "cleanup"} {
		t.Run(phase, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) { checkCancellationDeadline(t, phase) })
		})
	}
}

func stalledServer(phase string, entered, release, stopped chan struct{}) newServerFunc {
	return func(context.Context, []net.Listener, ...temporal.ServerOption) (runtimeServer, error) {
		if phase == "construction" {
			close(entered)
			<-release
		}
		return controlledServer{
			start: func(ctx context.Context) error {
				if phase != "construction" {
					close(entered)
				}
				if phase == "startup" {
					<-release
				}
				<-ctx.Done()
				return errors.Join(errStart, fmt.Errorf("controlled start: %w", ctx.Err()))
			},
			stop: func(context.Context) error {
				defer close(stopped)
				if phase == "cleanup" {
					<-release
				}
				return errStop
			},
		}, nil
	}
}

func checkCancellationDeadline(t *testing.T, phase string) {
	t.Helper()
	var buf bytes.Buffer
	s := newTestServer("", &buf)
	s.createNamespaces = func(*config.SQL, ...*sqliteschema.NamespaceConfig) error { return nil }
	// Controlled listeners keep this clock test independent of network I/O.
	s.listen = func(context.Context, string, string) (net.Listener, error) { return &clockListener{}, nil }
	entered, release, stopped := make(chan struct{}), make(chan struct{}), make(chan struct{})
	s.newServer = stalledServer(phase, entered, release, stopped)
	ctx, cancel := context.WithCancel(t.Context())
	result := make(chan error, 1)
	go func() { result <- s.Start(ctx) }()
	<-entered
	start := time.Now()
	cancel()
	err := <-result
	var interrupted *StartupError
	if !errors.As(err, &interrupted) || !errors.Is(err, context.Canceled) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Start() = %v, want typed cancellation and deadline", err)
	}
	if elapsed := time.Since(start); elapsed != s.shutdownTimeout {
		t.Errorf("elapsed = %v", elapsed)
	}
	if phase == "cleanup" && !errors.Is(err, errStart) {
		t.Errorf("Start() = %v, lost original startup error before cleanup timeout", err)
	}
	if !strings.Contains(buf.String(), "cleanup exceeded shutdown deadline") {
		t.Errorf("missing overrun log: %s", &buf)
	}
	close(release)
	<-stopped
	synctest.Wait()
	wantListenersClosed(t, s.listeners)
}

type clockListener struct{ closed bool }

func (*clockListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (*clockListener) Addr() net.Addr            { return &net.TCPAddr{} }
func (l *clockListener) Close() error {
	if l.closed {
		return net.ErrClosed
	}
	l.closed = true
	return nil
}

func TestStartupSignalsExit(t *testing.T) {
	t.Parallel()
	if os.Getenv("AFAKT_TEST_EMBEDDED_STARTUP") == "1" {
		os.Exit(startupSignalChild())
	}
	for _, sig := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			t.Parallel()
			checkStartupSignalExit(t, sig)
		})
	}
}

func TestStartupPanicCleanup(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := newTestServer("", &buf)
	s.newServer = func(context.Context, []net.Listener, ...temporal.ServerOption) (runtimeServer, error) {
		panic("constructor failed")
	}
	if err := s.Start(t.Context()); err == nil || !strings.Contains(err.Error(), "constructor failed") {
		t.Fatalf("Start() = %v, want protected constructor panic", err)
	}
	wantListenersClosed(t, s.listeners)
}

func TestCleanupPanicContinues(t *testing.T) {
	t.Parallel()
	owner := &ownedApp{app: fx.New(fx.NopLogger)}
	released := false
	owner.cleanup = []func(context.Context) error{
		func(context.Context) error { released = true; return nil },
		func(context.Context) error { panic("cleanup failed") },
	}
	if err := owner.stop(t.Context()); err == nil || !strings.Contains(err.Error(), "cleanup failed") || !released {
		t.Fatalf("stop() = %v, released = %v", err, released)
	}
}

func startupSignalChild() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	s := New("", "default", logger, 100*time.Millisecond, (&net.ListenConfig{}).Listen)
	s.createNamespaces = func(*config.SQL, ...*sqliteschema.NamespaceConfig) error { return nil }
	s.listen = func(context.Context, string, string) (net.Listener, error) { return &clockListener{}, nil }
	s.newServer = func(context.Context, []net.Listener, ...temporal.ServerOption) (runtimeServer, error) {
		return controlledServer{start: func(context.Context) error {
			_, _ = fmt.Fprintln(os.Stdout, "starting")
			select {}
		}, stop: func(context.Context) error { return nil }}, nil
	}
	err := s.Start(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		return 1
	}
	return 2
}

func checkStartupSignalExit(t *testing.T, sig os.Signal) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestStartupSignalsExit$")
	cmd.Env = append(os.Environ(), "AFAKT_TEST_EMBEDDED_STARTUP=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "starting" {
		t.Fatalf("startup handshake: %s", &stderr)
	}
	start := time.Now()
	if err = cmd.Process.Signal(sig); err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("exit = %v: %s", err, &stderr)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("signal exit took %s", elapsed)
	}
	if !strings.Contains(stderr.String(), "cleanup exceeded shutdown deadline") {
		t.Errorf("stderr = %s", &stderr)
	}
}

// wantListenersClosed checks the actual sockets, including reservations from failed builds.
func wantListenersClosed(t *testing.T, listeners []net.Listener) {
	t.Helper()
	for _, ln := range listeners {
		if err := ln.Close(); !errors.Is(err, net.ErrClosed) {
			t.Errorf("Close() error = %v, want already closed listener", err)
		}
	}
}

func TestLoggerDropsDisabledLevels(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := newLogger(slog.New(slog.NewTextHandler(&buf, nil)))
	l.Debug("debug")
	l.Info("info", tag.NewStringTag("key", "value"))
	if got := buf.String(); got != "" {
		t.Errorf("log = %q, want nothing", got)
	}
}

func TestLogger(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := newLogger(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn", tag.NewStringTag("key", "value"))
	l.Error("error")
	l.DPanic("development")
	for _, fn := range []func(string, ...tag.Tag){l.Panic, l.Fatal} {
		func() {
			defer func() {
				if r := recover(); r != "fails" {
					t.Errorf("recover() = %v, want fails", r)
				}
			}()
			fn("fails")
		}()
	}
	want := []string{
		"level=DEBUG msg=debug", "level=DEBUG msg=info", "level=WARN msg=warn key=value", "level=ERROR msg=error",
		"level=ERROR msg=development", "level=ERROR msg=fails", "level=ERROR msg=fails",
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(want) {
		t.Fatalf("log = %q, want %q", lines, want)
	}
	for i, line := range lines {
		if !strings.Contains(line, want[i]) {
			t.Errorf("line %d = %q, want it to contain %q", i, line, want[i])
		}
	}
}
