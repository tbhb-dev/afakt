// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tbhb-dev/afakt/internal/lifecycle"
)

// health holds the checks behind the health endpoints. live backs /livez, ready backs
// /readyz, and /healthz reports both. ready receives the context of the request.
type health struct {
	live  func() bool
	ready func(ctx context.Context) bool
}

// newHealth returns the checks for a server whose services manager runs. The live check
// passes whenever a handler runs, since the HTTP server is then answering. The ready check
// passes only while `ctx` is live, every required service of manager is running, and each
// of `checks` returns nil. It fails as soon as a signal cancels `ctx`, before the stop
// sequence begins. A target adds its own readiness conditions, such as a connection it
// depends on, through `checks`. Each one runs on every request with the context of that
// request, so it must return promptly once that context ends.
func newHealth(ctx context.Context, manager *lifecycle.Manager, checks ...func(context.Context) error) health {
	return health{
		live: func() bool { return true },
		ready: func(reqCtx context.Context) bool {
			if ctx.Err() != nil || !manager.Ready() {
				return false
			}
			for _, check := range checks {
				if check(reqCtx) != nil {
					return false
				}
			}
			return true
		},
	}
}

// routes adds the health endpoints at the root of `mux`, outside the versioned API prefix.
func (h health) routes(mux *http.ServeMux) {
	h.probeRoutes(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		live, ready := h.live(), h.ready(r.Context())
		body := fmt.Sprintf(healthzFormat, checkStatus(live && ready), checkStatus(live), checkStatus(ready))
		writeHealth(w, live && ready, body)
	})
}

// probeRoutes adds /livez and /readyz.
func (h health) probeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		live := h.live()
		writeHealth(w, live, fmt.Sprintf(`{"status":%q}`, checkStatus(live)))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ready := h.ready(r.Context())
		writeHealth(w, ready, fmt.Sprintf(`{"status":%q}`, checkStatus(ready)))
	})
}

// healthzFormat is the /healthz body, with the combined result first and then the result of
// each check.
const healthzFormat = `{"status":%q,"checks":{"livez":{"status":%q},"readyz":{"status":%q}}}`

// checkStatus names the result of a check as the health bodies report it.
func checkStatus(pass bool) string {
	if pass {
		return "pass"
	}
	return "fail"
}

// writeHealth writes a JSON health body with 200 when the check passes and 503 when it
// fails.
func writeHealth(w http.ResponseWriter, pass bool, body string) {
	w.Header().Set("Content-Type", "application/json")
	if !pass {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_, _ = fmt.Fprint(w, body)
}
