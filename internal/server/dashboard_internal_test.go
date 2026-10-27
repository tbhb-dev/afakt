// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
)

func TestDashboardComposition(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"dashboard", "api,dashboard", "dashboard,webhook,api"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := newProbe()
			s := startServer(t, targetConfigs()[name], hooks{services: []lifecycle.Registration{p.reg()}})
			receive(t, p.up, "the probe to start")
			waitReady(t, s.addr(t))
			wantDashboardPaths(t, s.addr(t))
			wantDashboardAPI(t, s.addr(t), name != "dashboard")
			s.cancel()
			if err := s.wait(t); err != nil {
				t.Fatal(err)
			}
			// Check the owned socket: another parallel test can reuse its released address.
			if err := s.listener.Close(); !errors.Is(err, net.ErrClosed) {
				t.Fatalf("listener Close() = %v, want already closed", err)
			}
		})
	}
}

func TestDashboardCanonicalSecurity(t *testing.T) {
	t.Parallel()
	for name, targets := range map[string][]string{
		"dashboard": {TargetDashboard},
		"all":       {TargetAPI, TargetWebhook, TargetDashboard},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := testConfig()
			cfg.Targets = targets
			s := startServer(t, cfg, hooks{})
			waitReady(t, s.addr(t))
			for _, path := range []string{"/dashboard//", "/dashboard/../", "/dashboard/./settings"} {
				wantDashboardCanonicalSecurity(t, s.addr(t), path)
			}
			s.cancel()
			if err := s.wait(t); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// wantDashboardCanonicalSecurity checks security headers on the 404 response for a path
// that the server rejects before routing.
func wantDashboardCanonicalSecurity(t *testing.T, address, path string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+address+path, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET %s = %d, want 404", path, resp.StatusCode)
	}
	if resp.Header.Get("X-Frame-Options") != "SAMEORIGIN" || resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("GET %s lacks security headers: %q", path, resp.Header)
	}
	csp := resp.Header.Get("Content-Security-Policy")
	for _, directive := range []string{"default-src 'none'", "script-src 'self'", "style-src 'self'", "frame-ancestors 'self'"} {
		if !strings.Contains(csp, directive) {
			t.Errorf("GET %s: CSP %q lacks %q", path, csp, directive)
		}
	}
}

// wantDashboardAPI checks that only a composition including the API serves its endpoints.
func wantDashboardAPI(t *testing.T, address string, enabled bool) {
	t.Helper()
	if enabled {
		wantAPIRoot(t, address)
		wantCheck(t, address, "/healthz", http.StatusOK,
			`{"status":"pass","checks":{"livez":{"status":"pass"},"readyz":{"status":"pass"}}}`)
		return
	}
	for _, target := range []string{"/api/v1/", "/healthz"} {
		if code, _, _ := request(t, http.MethodGet, address, target); code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, code)
		}
	}
}

// wantDashboardPaths checks a running composition. Entry routes serve the same page,
// while paths outside the dashboard and missing assets return 404.
func wantDashboardPaths(t *testing.T, address string) {
	t.Helper()
	_, _, entry := request(t, http.MethodGet, address, "/dashboard/")
	for _, target := range []string{
		"/dashboard", "/dashboard/", "/dashboard/settings", "/dashboard/v1.2/settings",
		"/dashboard/report.json/", "/dashboard/a/b/",
	} {
		code, contentType, body := request(t, http.MethodGet, address, target)
		if code != http.StatusOK || !strings.HasPrefix(contentType, "text/html") || body != entry {
			t.Errorf("GET %s = %d %q %q", target, code, contentType, body)
		}
	}
	for _, target := range []string{
		"/", "/api", "/dashboardish", "/dashboard.html", "/dashboard/missing.js",
		"/dashboard/users/jane.doe", "/dashboard//", "/dashboard/../",
	} {
		if code, _, _ := request(t, http.MethodGet, address, target); code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, code)
		}
	}
}
