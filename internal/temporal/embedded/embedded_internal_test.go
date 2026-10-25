// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package embedded

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/temporal"
)

var (
	errBuild = errors.New("build failed")
	errStart = errors.New("start failed")
	errStop  = errors.New("stop failed")
)

// fakeServer is a Temporal server whose Start and Stop return the errors in its fields.
type fakeServer struct {
	start, stop error
	stopped     bool
}

func (f *fakeServer) Start() error { return f.start }

func (f *fakeServer) Stop() error {
	f.stopped = true
	return f.stop
}

// newTestServer returns a server for the default namespace that logs into the buffer.
func newTestServer(dir string, buf *bytes.Buffer) *Server {
	return New(dir, "default", slog.New(slog.NewTextHandler(buf, nil)))
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
	readOnly := filepath.Join(t.TempDir(), "read-only")
	if err := os.Mkdir(readOnly, 0o500); err != nil {
		t.Fatal(err)
	}
	notDatabase := t.TempDir()
	if err := os.Mkdir(filepath.Join(notDatabase, DatabaseFile), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []startFailure{
		{name: "directory", dir: filepath.Join(file, "dir"), want: "create embedded Temporal directory: "},
		{name: "schema", dir: readOnly, want: "set up embedded Temporal schema: "},
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
// the server library's constructor.
type startFailure struct {
	name, dir, host string
	server          temporal.Server
	build           error
	want            string
	stopped         bool
}

// check starts a server set up as the case says and checks the error and whether the fake
// server stopped.
func (tc startFailure) check(t *testing.T) {
	t.Helper()
	var buf bytes.Buffer
	s := newTestServer(tc.dir, &buf)
	if tc.host != "" {
		s.host = tc.host
	}
	if tc.server != nil || tc.build != nil {
		s.newServer = func(...temporal.ServerOption) (temporal.Server, error) { return tc.server, tc.build }
	}
	err := s.Start(t.Context())
	if err == nil || !strings.HasPrefix(err.Error(), tc.want) {
		t.Errorf("Start() error = %v, want prefix %q", err, tc.want)
	}
	if fake, ok := tc.server.(*fakeServer); ok && fake.stopped != tc.stopped {
		t.Errorf("stopped = %v, want %v", fake.stopped, tc.stopped)
	}
}

func TestServerStopFailure(t *testing.T) {
	t.Parallel()
	s := &Server{server: &fakeServer{stop: errStop}}
	if err := s.Stop(t.Context()); !errors.Is(err, errStop) {
		t.Errorf("Stop() error = %v, want %v", err, errStop)
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
