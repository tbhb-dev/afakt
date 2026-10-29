// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
	"github.com/tbhb-dev/afakt/internal/temporal"
)

func TestTemporalUIComposition(t *testing.T) {
	t.Parallel()
	for _, prefix := range []string{"", "/tools/temporal"} {
		t.Run(prefix, func(t *testing.T) {
			t.Parallel()
			wantTemporalUI(t, prefix)
		})
	}
}

func wantTemporalUI(t *testing.T, prefix string) {
	t.Helper()
	cfg := workerConfig(temporal.Config{Mode: temporal.ModeEmbedded})
	cfg.Targets = []string{TargetAPI, TargetDashboard, TargetWorker}
	cfg.TemporalUIPath = prefix
	s := startServer(t, cfg, hooks{})
	waitReady(t, s.addr(t))
	if prefix == "" {
		prefix = "/temporal"
	}
	for _, target := range []string{prefix, prefix + "/", prefix + "/namespaces/default/workflows"} {
		code, _, body := request(t, http.MethodGet, s.addr(t), target)
		if code != http.StatusOK || !strings.Contains(body, "<html") {
			t.Errorf("GET %s = %d %q, want the UI page", target, code, body)
		}
	}
	code, _, body := request(t, http.MethodGet, s.addr(t), prefix+"/api/v1/namespaces")
	if code != http.StatusOK || !strings.Contains(body, `"name": "default"`) {
		t.Errorf("namespaces = %d %s, want embedded default namespace", code, body)
	}
	wantAPIRoot(t, s.addr(t))
	wantDashboardPaths(t, s.addr(t))
	wantHealth(t, s.addr(t), true)
	stopCleanly(t, s)
	// Shared HTTP, four Temporal gRPC ports, and the frontend HTTP reservation.
	wantBoundAddresses(t, s, 6)
}

func TestTemporalUIAbsent(t *testing.T) {
	t.Parallel()
	frontend := startFrontend(t, temporal.DefaultNamespace)
	for name, cfg := range map[string]Config{
		"no worker": testConfig(),
		"external":  workerConfig(external(frontend.Address(), temporal.DefaultNamespace)),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			wantTemporalUIAbsent(t, cfg)
		})
	}
}

func wantTemporalUIAbsent(t *testing.T, cfg Config) {
	t.Helper()
	s := startServer(t, cfg, hooks{})
	waitReady(t, s.addr(t))
	for _, target := range []string{"/temporal", "/temporal/", "/temporal/api/v1/namespaces"} {
		if code, _, _ := request(t, http.MethodGet, s.addr(t), target); code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, code)
		}
	}
	// Both configurations leave the UI service out of the composition.
	if slices.Contains(s.events.matching("starting"), "temporal-ui") {
		t.Fatal("inactive UI started a service")
	}
	stopCleanly(t, s)
	wantBoundAddresses(t, s, 1)
}

// wantBoundAddresses checks actual successful binds after shutdown, independently of logs.
// The embedded frontend privately rebinds its recorded HTTP reservation (issue #73).
func wantBoundAddresses(t *testing.T, s *harness, count int) {
	t.Helper()
	bound := s.events.matching("bound")
	wantCount := len(bound) == count
	if count == 6 {
		// Each construction attempt binds five Temporal ports and reuses the shared HTTP port.
		wantCount = len(bound) >= 6 && len(bound) <= 16 && (len(bound)-1)%5 == 0
	}
	if !wantCount || bound[0] != s.addr(t) {
		t.Fatalf("bound addresses = %q, want shared HTTP %q and %d total", bound, s.addr(t), count)
	}
	seen := make(map[string]bool)
	for i, address := range bound {
		if count == 6 && i > 1 && (i-1)%5 == 0 {
			// A new attempt may reuse a closed port, but never the shared listener.
			seen = map[string]bool{bound[0]: true}
		}
		host, port, err := net.SplitHostPort(address)
		if err != nil || host != "127.0.0.1" || port == "0" || port == "" || seen[address] {
			t.Errorf("bound address = %q, want a distinct chosen loopback port: %v", address, err)
		}
		seen[address] = true
	}
}

func TestTemporalUIHealthStates(t *testing.T) {
	t.Parallel()
	frontend := startFrontend(t, temporal.DefaultNamespace)
	for _, tc := range []temporal.Config{
		{Mode: temporal.ModeEmbedded}, external(frontend.Address(), temporal.DefaultNamespace),
	} {
		p := newProbe()
		startEntered, startRelease := make(chan struct{}), make(chan struct{})
		stopEntered, stopRelease := make(chan struct{}), make(chan struct{})
		p.fake.StartFunc = lifecycletest.Block(startEntered, startRelease)
		p.fake.StopFunc = lifecycletest.Block(stopEntered, stopRelease)
		s := startServer(t, workerConfig(tc), hooks{services: []lifecycle.Registration{p.regOn("temporal-client")}})
		receive(t, startEntered, "starting")
		wantHealth(t, s.addr(t), false)
		close(startRelease)
		waitReady(t, s.addr(t))
		wantHealth(t, s.addr(t), true)
		s.cancel()
		receive(t, stopEntered, "draining")
		wantHealth(t, s.addr(t), false)
		close(stopRelease)
		if err := s.wait(t); err != nil {
			t.Fatal(err)
		}
		wantClosed(t, s.listener)
	}
}

func TestTemporalUIConflicts(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	for _, prefix := range []string{"/api/v1", "/dashboard", "/t\u00e9st"} {
		cfg := workerConfig(temporal.Config{Mode: temporal.ModeEmbedded})
		cfg.TemporalUIPath = prefix
		if err := Run(t.Context(), cfg, logger); err == nil || !strings.Contains(err.Error(), prefix) {
			t.Errorf("Run() = %v, want conflict naming %s", err, prefix)
		}
	}
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	cfg := workerConfig(temporal.Config{Mode: temporal.ModeEmbedded})
	cfg.Listen = ln.Addr().String()
	if err = Run(t.Context(), cfg, logger); err == nil || !strings.Contains(err.Error(), cfg.Listen) {
		t.Errorf("Run() = %v, want address conflict naming %s", err, cfg.Listen)
	}
}
