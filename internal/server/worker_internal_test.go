// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
	"github.com/tbhb-dev/afakt/internal/temporal"
	"github.com/tbhb-dev/afakt/internal/temporal/embedded"
	"github.com/tbhb-dev/afakt/internal/temporal/temporaltest"
)

// workerConfig returns a configuration for the worker target on 127.0.0.1 port 0.
func workerConfig(tc temporal.Config) Config {
	if tc.Namespace == "" {
		tc.Namespace = temporal.DefaultNamespace
	}
	return Config{Targets: []string{TargetWorker}, Listen: "127.0.0.1:0", ShutdownTimeout: testTimeout, Temporal: tc}
}

// external returns the Temporal configuration for a frontend at address.
func external(address, namespace string) temporal.Config {
	return temporal.Config{Mode: temporal.ModeExternal, Address: address, Namespace: namespace}
}

// startFrontend starts a stand-in frontend for namespaces on a free loopback port.
func startFrontend(t *testing.T, namespaces ...string) *temporaltest.Frontend {
	t.Helper()
	var config net.ListenConfig
	ln, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	return temporaltest.Serve(t, ln, namespaces...)
}

// stopCleanly cancels the server and checks that run returns nil and releases its address.
func stopCleanly(t *testing.T, s *harness) {
	t.Helper()
	s.cancel()
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	wantClosed(t, s.addr(t))
}

func TestWorkerEmbedded(t *testing.T) {
	t.Parallel()
	s := startServer(t, workerConfig(temporal.Config{Mode: temporal.ModeEmbedded}), hooks{})
	waitReady(t, s.addr(t))
	wantHealth(t, s.addr(t), true)
	// The worker target alone doesn't serve the API.
	if code, _, _ := request(t, http.MethodGet, s.addr(t), apiRootPath); code != http.StatusNotFound {
		t.Errorf("GET %s = %d, want 404", apiRootPath, code)
	}
	stopCleanly(t, s)
	want := []string{"temporal-client", "temporal-server", "http", "http-listener"}
	if got := s.events.matching("stopping"); !slices.Equal(got, want) {
		t.Errorf("stop order = %q, want %q", got, want)
	}
}

func TestWorkerEmbeddedInstancesShareAProcess(t *testing.T) {
	t.Parallel()
	dirs := []string{"", filepath.Join(t.TempDir(), "one"), filepath.Join(t.TempDir(), "two")}
	servers := make([]*harness, len(dirs))
	for i, dir := range dirs {
		servers[i] = startServer(
			t,
			workerConfig(temporal.Config{Mode: temporal.ModeEmbedded, EmbeddedDir: dir}),
			hooks{},
		)
	}
	addresses := map[string]bool{}
	for _, s := range servers {
		waitReady(t, s.addr(t))
		addresses[s.addr(t)] = true
	}
	if len(addresses) != len(servers) {
		t.Errorf("addresses = %v, want one for each instance", addresses)
	}
	for _, dir := range dirs[1:] {
		if _, err := os.Stat(filepath.Join(dir, embedded.DatabaseFile)); err != nil {
			t.Errorf("database file: %v", err)
		}
	}
	for _, s := range servers {
		stopCleanly(t, s)
	}
}

func TestWorkerExternal(t *testing.T) {
	t.Parallel()
	for _, namespace := range []string{temporal.DefaultNamespace, "orders"} {
		t.Run(namespace, func(t *testing.T) {
			t.Parallel()
			frontend := startFrontend(t, namespace)
			s := startServer(t, workerConfig(external(frontend.Address(), namespace)), hooks{})
			waitReady(t, s.addr(t))
			stopCleanly(t, s)
			// External mode doesn't start an embedded server.
			if got, want := s.events.matching(
				"starting",
			), []string{
				"http-listener",
				"http",
				"temporal-client",
			}; !slices.Equal(
				got,
				want,
			) {
				t.Errorf("started %q, want %q", got, want)
			}
		})
	}
}

// TestWorkerReadiness walks /readyz and /livez through starting, running, a lost connection
// and draining. A probe that depends on the client blocks in its Start and in its Stop.
func TestWorkerReadiness(t *testing.T) {
	t.Parallel()
	frontend := startFrontend(t, temporal.DefaultNamespace)
	p := newProbe()
	startEntered, startRelease := make(chan struct{}), make(chan struct{})
	p.fake.StartFunc = lifecycletest.Block(startEntered, startRelease)
	stopEntered, stopRelease := make(chan struct{}), make(chan struct{})
	p.fake.StopFunc = lifecycletest.Block(stopEntered, stopRelease)
	s := startServer(t, workerConfig(external(frontend.Address(), temporal.DefaultNamespace)), hooks{
		services: []lifecycle.Registration{p.regOn("temporal-client")},
	})
	receive(t, startEntered, "the probe to start")
	// The client has connected, but the process is still starting.
	wantHealth(t, s.addr(t), false)
	close(startRelease)
	waitReady(t, s.addr(t))
	frontend.Stop()
	wantHealth(t, s.addr(t), false)
	s.cancel()
	receive(t, stopEntered, "the probe to stop")
	wantHealth(t, s.addr(t), false)
	close(stopRelease)
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

func TestWorkerStartFailures(t *testing.T) {
	t.Parallel()
	var config net.ListenConfig
	refused, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	refusedAddress := refused.Addr().String()
	if err = refused.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	file := filepath.Join(t.TempDir(), "file")
	if err = os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	frontend := startFrontend(t, temporal.DefaultNamespace)
	cases := []struct {
		name     string
		temporal temporal.Config
		want     string
		started  []string
	}{
		{
			name:     "unreachable address",
			temporal: external(refusedAddress, temporal.DefaultNamespace),
			want:     `start "temporal-client": connect to Temporal at ` + refusedAddress + ": ",
			started:  []string{"http-listener", "http", "temporal-client"},
		},
		{
			name:     "unknown namespace",
			temporal: external(frontend.Address(), "orders"),
			want:     `start "temporal-client": describe Temporal namespace "orders": `,
			started:  []string{"http-listener", "http", "temporal-client"},
		},
		{
			name:     "embedded server",
			temporal: temporal.Config{Mode: temporal.ModeEmbedded, EmbeddedDir: filepath.Join(file, "dir")},
			want:     `start "temporal-server": create embedded Temporal directory: `,
			started:  []string{"http-listener", "http", "temporal-server"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := startServer(t, workerConfig(tc.temporal), hooks{})
			runErr := s.wait(t)
			if runErr == nil || !strings.HasPrefix(runErr.Error(), tc.want) {
				t.Errorf("run() error = %v, want prefix %q", runErr, tc.want)
			}
			if got := s.events.matching("starting"); !slices.Equal(got, tc.started) {
				t.Errorf("started %q, want %q", got, tc.started)
			}
			// The HTTP server and the listener stopped, and the address is free again.
			if got, want := s.events.matching("stopping"), []string{"http", "http-listener"}; !slices.Equal(got, want) {
				t.Errorf("stopped %q, want %q", got, want)
			}
			wantClosed(t, s.addr(t))
		})
	}
}

func TestWorkerEmbeddedStartFailurePreservesError(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	s := startServer(t, workerConfig(temporal.Config{Mode: temporal.ModeEmbedded, EmbeddedDir: file}), hooks{})
	if err := s.wait(t); !errors.Is(err, syscall.ENOTDIR) && !errors.Is(err, syscall.EEXIST) {
		t.Errorf("run() error = %v, want the directory error from the system", err)
	}
}

func TestWorkerShutdownTimeout(t *testing.T) {
	t.Parallel()
	frontend := startFrontend(t, temporal.DefaultNamespace)
	p := newProbe()
	p.fake.StopFunc = func(ctx context.Context) error {
		<-ctx.Done()
		select {}
	}
	cfg := workerConfig(external(frontend.Address(), temporal.DefaultNamespace))
	cfg.ShutdownTimeout = 200 * time.Millisecond
	s := startServer(t, cfg, hooks{services: []lifecycle.Registration{p.regOn("temporal-client")}})
	receive(t, p.up, "the probe to start")
	s.cancel()
	err := s.wait(t)
	want := `shutdown deadline exceeded while stopping "probe"; not confirmed stopped: probe, temporal-client, http, http-listener`
	if !errors.Is(err, lifecycle.ErrShutdownTimeout) || err.Error() != want {
		t.Errorf("run() error = %q, want %q", err, want)
	}
}
