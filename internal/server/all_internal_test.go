// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"testing"
	"testing/synctest"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
	"github.com/tbhb-dev/afakt/internal/temporal"
)

func allConfig(t *testing.T, value string) Config {
	t.Helper()
	cfg := workerConfig(temporal.Config{Mode: temporal.ModeEmbedded})
	var err error
	cfg.Targets, err = ParseTargets(value)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestAllAndExplicitComposition(t *testing.T) {
	t.Parallel()
	alone := testConfig()
	alone.Targets = []string{TargetDashboard}
	dashboard := startServer(t, alone, hooks{})
	waitReady(t, dashboard.addr(t))
	_, _, entry := request(t, http.MethodGet, dashboard.addr(t), "/dashboard/")
	stopCleanly(t, dashboard)
	var started []string
	for _, value := range []string{"all", "api,webhook,dashboard,worker"} {
		s := startServer(t, allConfig(t, value), hooks{})
		waitReady(t, s.addr(t))
		wantHealth(t, s.addr(t), true)
		wantAPIRoot(t, s.addr(t))
		wantDashboardPaths(t, s.addr(t))
		code, _, body := request(t, http.MethodGet, s.addr(t), "/dashboard/")
		if code != http.StatusOK || body != entry {
			t.Error("combined dashboard differs from dashboard-only response")
		}
		got := s.events.matching("starting")
		if started != nil && !slices.Equal(got, started) {
			t.Errorf("services = %q, want %q", got, started)
		}
		started = got
		stopCleanly(t, s)
		wantBoundAddresses(t, s, 6)
		for _, address := range s.events.matching("bound") {
			wantClosed(t, address)
		}
		want := []string{
			"temporal-client",
			"temporal-ui",
			"temporal-server",
			"dashboard",
			"webhook",
			"api",
			"http",
			"http-listener",
		}
		if stopped := s.events.matching("stopping"); !slices.Equal(stopped, want) {
			t.Errorf("stop order = %q, want %q", stopped, want)
		}
	}
}

func TestAllInstancesAndStartupRollback(t *testing.T) {
	t.Parallel()
	servers := make([]*harness, 0, 2)
	for range 2 {
		cfg := allConfig(t, "all")
		cfg.Temporal.EmbeddedDir = filepath.Join(t.TempDir(), "data")
		s := startServer(t, cfg, hooks{})
		waitReady(t, s.addr(t))
		servers = append(servers, s)
	}
	for _, s := range servers {
		wantHealth(t, s.addr(t), true)
		wantAPIRoot(t, s.addr(t))
		wantDashboardPaths(t, s.addr(t))
	}
	for _, s := range servers {
		stopCleanly(t, s)
		for _, address := range s.events.matching("bound") {
			wantClosed(t, address)
		}
	}
	// Fail after all real services have acquired resources, including persistence.
	cfg := allConfig(t, "all")
	cfg.Temporal.EmbeddedDir = filepath.Join(t.TempDir(), "data")
	p := newProbe()
	p.fake.StartFunc = func(context.Context) error { return errBoom }
	s := startServer(t, cfg, hooks{services: []lifecycle.Registration{p.regOn("temporal-client")}})
	if err := s.wait(t); !errors.Is(err, errBoom) {
		t.Errorf("run() = %v, want startup failure", err)
	}
	for _, address := range s.events.matching("bound") {
		wantClosed(t, address)
	}
	// Reusing the data directory proves the failed run released its persistence handles.
	retry := startServer(t, cfg, hooks{})
	waitReady(t, retry.addr(t))
	stopCleanly(t, retry)
}

func TestAllRoleHealthStates(t *testing.T) {
	t.Parallel()
	for _, target := range availableTargets() {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				checkAllRoleHealthStates(t, target)
			})
		})
	}
}

func checkAllRoleHealthStates(t *testing.T, target string) {
	t.Helper()
	manager := lifecycle.New(slog.New(slog.DiscardHandler))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var held *lifecycletest.Fake
	start, stop := make(chan struct{}), make(chan struct{})
	for _, name := range availableTargets() {
		fake := lifecycletest.NewFake(name, &lifecycletest.Recorder{})
		if name == target {
			held = fake
			fake.StartFunc = func(context.Context) error { <-start; return nil }
			fake.StopFunc = func(context.Context) error { <-stop; return nil }
		}
		if err := manager.Register(lifecycle.Registration{Name: name, Service: fake}); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	newHealth(ctx, manager).routes(mux)
	go func() {
		if err := manager.Start(ctx); err != nil {
			t.Error(err)
		}
	}()
	synctest.Wait()
	wantAllRoleHealth(t, mux, false)
	close(start)
	synctest.Wait()
	wantAllRoleHealth(t, mux, true)
	held.Fail(errBoom)
	synctest.Wait()
	wantAllRoleHealth(t, mux, false)
	cancel()
	go func() {
		if err := manager.Stop(context.WithoutCancel(ctx)); !errors.Is(err, errBoom) {
			t.Error(err)
		}
	}()
	synctest.Wait()
	wantAllRoleHealth(t, mux, false)
	close(stop)
	synctest.Wait()
}

func wantAllRoleHealth(t *testing.T, handler http.Handler, ready bool) {
	t.Helper()
	for _, path := range []string{"/livez", "/readyz"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, http.NoBody))
		want := http.StatusOK
		if path == "/readyz" && !ready {
			want = http.StatusServiceUnavailable
		}
		if response.Code != want {
			t.Errorf("%s = %d, want %d", path, response.Code, want)
		}
	}
}
