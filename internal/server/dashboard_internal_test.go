// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
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
			wantClosed(t, s.addr(t))
		})
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

// wantDashboardPaths checks entry routes and boundaries on a running composition.
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
