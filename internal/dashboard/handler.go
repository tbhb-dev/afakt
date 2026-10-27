// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package dashboard serves the embedded web app below /dashboard.
package dashboard

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

// New returns a handler for the dashboard's entry page, assets and client-side routes.
func New() http.Handler {
	return newHandler(content, entry)
}

// SecurityHeaders applies the dashboard policy before the next handler can reject a path.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/dashboard" || strings.HasPrefix(r.URL.Path, "/dashboard/") {
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; "+
				"img-src 'self'; connect-src 'self'; base-uri 'none'; frame-ancestors 'self'")
		}
		next.ServeHTTP(w, r)
	})
}

// newHandler serves a fixed asset set and entry page, including in Go-only tests.
func newHandler(files fs.ReadFileFS, index []byte) http.Handler {
	// Vite uses an eight-character content hash in each generated asset name.
	hashed := regexp.MustCompile(`-[A-Za-z0-9_-]{8,}\.`)
	return SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dashboard" && !strings.HasPrefix(r.URL.Path, "/dashboard/") {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/dashboard"), "/")
		body, err := files.ReadFile(contentRoot + "/" + name)
		if err != nil {
			// Only the final segment determines whether this can be a client-side route.
			_, last := path.Split(name)
			if strings.Contains(last, ".") {
				http.NotFound(w, r)
				return
			}
			name, body = "index.html", index
		}
		w.Header().Set("Cache-Control", "no-cache")
		if name != "index.html" && hashed.MatchString(path.Base(name)) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		w.Header().Set("ETag", fmt.Sprintf(`"%x"`, sha256.Sum256(body)))
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(body))
	}))
}
