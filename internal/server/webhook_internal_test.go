// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
	"github.com/tbhb-dev/afakt/internal/lifecycle/lifecycletest"
)

// targetConfigs returns single and combined target configurations, keyed by the target list.
func targetConfigs() map[string]Config {
	webhook, combined := testConfig(), testConfig()
	webhook.Targets = []string{TargetWebhook}
	combined.Targets = []string{TargetAPI, TargetWebhook}
	dashboard, apiDashboard, allHTTP := testConfig(), testConfig(), testConfig()
	dashboard.Targets = []string{"dashboard"}
	apiDashboard.Targets = []string{TargetAPI, "dashboard"}
	allHTTP.Targets = []string{"dashboard", TargetWebhook, TargetAPI}
	return map[string]Config{
		"webhook": webhook, "api,webhook": combined, "dashboard": dashboard,
		"api,dashboard": apiDashboard, "dashboard,webhook,api": allHTTP,
	}
}

// wantCheck checks that a GET request for path answers code with a JSON body.
func wantCheck(t *testing.T, address, path string, code int, body string) {
	t.Helper()
	gotCode, contentType, gotBody := request(t, http.MethodGet, address, path)
	if gotCode != code || contentType != "application/json" || gotBody != body {
		t.Errorf("GET %s = %d %q %q, want %d \"application/json\" %q", path, gotCode, contentType, gotBody, code, body)
	}
}

// wantNotReady checks that /livez passes and /readyz fails.
func wantNotReady(t *testing.T, address string) {
	t.Helper()
	wantCheck(t, address, "/livez", http.StatusOK, okBody)
	wantCheck(t, address, "/readyz", http.StatusServiceUnavailable, notOKBody)
}

// waitReady polls /readyz until it answers 200, and stops the test if that doesn't happen
// in time. The server reports ready just after its last service starts, so a test that saw
// that service start can still see 503 for a moment. Then it checks that /livez passes.
func waitReady(t *testing.T, address string) {
	t.Helper()
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		if code, _, _ := request(t, http.MethodGet, address, "/readyz"); code == http.StatusOK {
			wantCheck(t, address, "/readyz", http.StatusOK, okBody)
			wantCheck(t, address, "/livez", http.StatusOK, okBody)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("GET /readyz on %s never answered 200", address)
}

func TestWebhookRoutes(t *testing.T) {
	t.Parallel()
	p := newProbe()
	s := startServer(t, targetConfigs()["webhook"], hooks{services: []lifecycle.Registration{p.reg()}})
	receive(t, p.up, "the probe to start")
	waitReady(t, s.addr(t))
	methods := []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions,
	}
	// Without receivers, the webhook target answers 404 everywhere but the health checks,
	// whatever the method, so it doesn't accept a delivery.
	for _, method := range methods {
		for _, path := range []string{
			"/", "/api/v1/", "/api/v1", "/healthz", "/healthz/", "/livez/", "/readyz/", "/livez/x",
			"/webhook", "/webhooks", "/webhooks/github", "/github", "/api/v1/livez",
			"//livez", "/./readyz", "/x/../livez",
		} {
			if code, _, _ := request(t, method, s.addr(t), path); code != http.StatusNotFound {
				t.Errorf("%s %s = %d, want 404", method, path, code)
			}
		}
	}
	// The health checks answer 405 to the methods a delivery would use.
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		for _, path := range []string{"/livez", "/readyz"} {
			if code, _, _ := request(t, method, s.addr(t), path); code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s = %d, want 405", method, path, code)
			}
		}
	}
	s.cancel()
	if err := s.wait(t); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

// TestReadiness keeps the probe in Start and then in Stop. The HTTP server answers in both
// states, /livez passes throughout, and /readyz passes only between them.
func TestReadiness(t *testing.T) {
	t.Parallel()
	for name, cfg := range targetConfigs() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			startEntered, startRelease := make(chan struct{}), make(chan struct{})
			stopEntered, stopRelease := make(chan struct{}), make(chan struct{})
			p := newProbe()
			p.fake.StartFunc = lifecycletest.Block(startEntered, startRelease)
			p.fake.StopFunc = lifecycletest.Block(stopEntered, stopRelease)
			s := startServer(t, cfg, hooks{services: []lifecycle.Registration{p.reg()}})

			receive(t, startEntered, "the probe to start")
			wantNotReady(t, s.addr(t))
			close(startRelease)
			waitReady(t, s.addr(t))
			s.cancel()
			receive(t, stopEntered, "the probe to stop")
			wantNotReady(t, s.addr(t))

			close(stopRelease)
			if err := s.wait(t); err != nil {
				t.Fatalf("run() error = %v", err)
			}
			wantClosed(t, s.listener)
		})
	}
}

// TestReadinessAfterFailure fails the running probe and keeps it in Stop, so the HTTP server
// still answers after the failure.
func TestReadinessAfterFailure(t *testing.T) {
	t.Parallel()
	for name, cfg := range targetConfigs() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			stopEntered, stopRelease := make(chan struct{}), make(chan struct{})
			p := newProbe()
			p.fake.StopFunc = lifecycletest.Block(stopEntered, stopRelease)
			s := startServer(t, cfg, hooks{services: []lifecycle.Registration{p.reg()}})
			receive(t, p.up, "the probe to start")
			waitReady(t, s.addr(t))
			p.fake.Fail(errBoom)
			receive(t, stopEntered, "the failed probe to stop")
			wantNotReady(t, s.addr(t))

			close(stopRelease)
			if err := s.wait(t); !errors.Is(err, errBoom) {
				t.Errorf("run() error = %v, want %v", err, errBoom)
			}
		})
	}
}

// TestCombinedTargetsMatchAPI runs the API alone and the API with the webhook target on
// separate listeners, then repeats with the dashboard. The API responses stay the same.
func TestCombinedTargetsMatchAPI(t *testing.T) {
	t.Parallel()
	for _, target := range []string{"api,webhook", "api,dashboard", "dashboard,webhook,api"} {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			wantCombinedAPI(t, targetConfigs()[target])
		})
	}
}

// wantCombinedAPI compares API paths and methods with and without another target.
func wantCombinedAPI(t *testing.T, cfg Config) {
	t.Helper()
	apiProbe, combinedProbe := newProbe(), newProbe()
	apiOnly := startServer(t, testConfig(), hooks{services: []lifecycle.Registration{apiProbe.reg()}})
	combined := startServer(t, cfg,
		hooks{services: []lifecycle.Registration{combinedProbe.reg()}})
	receive(t, apiProbe.up, "the API-only probe to start")
	receive(t, combinedProbe.up, "the combined probe to start")
	waitReady(t, combined.addr(t))

	wantAPIRoot(t, combined.addr(t))
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete} {
		for _, path := range []string{
			apiRootPath, "/api/v1", "/api", "/api/v2/", "/api/v1/livez", "/api/v1/readyz",
			"/api//v1/", "/api/v1/./", "/", "/healthz", "/livez/", "/readyz/",
			"/webhook", "/webhooks", "/webhooks/github",
		} {
			wantCode, wantType, wantBody := request(t, method, apiOnly.addr(t), path)
			code, contentType, body := request(t, method, combined.addr(t), path)
			if code != wantCode || contentType != wantType || body != wantBody {
				t.Errorf("%s %s = %d %q %q, want %d %q %q as with the API alone",
					method, path, code, contentType, body, wantCode, wantType, wantBody)
			}
		}
	}
	for _, s := range []*harness{apiOnly, combined} {
		s.cancel()
		if err := s.wait(t); err != nil {
			t.Errorf("run() error = %v", err)
		}
	}
}
