// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package dashboard

import (
	"bytes"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

// get records a request to the dashboard without a router or redirects.
func get(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody))
	return w
}

func TestHashedAssetCache(t *testing.T) {
	t.Parallel()
	const name = "assets/app-Abc_1234.js"
	h := newHandler(fstest.MapFS{
		contentRoot + "/" + name: &fstest.MapFile{Data: []byte("export {};\n")},
	}, entry)
	w := get(t, h, "/dashboard/"+name)
	if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Errorf("hashed asset = %d %q", w.Code, w.Header())
	}
	wantSecurity(t, w)
}

func TestEntryAndBoundaries(t *testing.T) {
	t.Parallel()
	h := New()
	for _, target := range []string{
		"/dashboard", "/dashboard/", "/dashboard/index.html", "/dashboard/settings",
		"/dashboard/v1.2/settings", "/dashboard/report.json/", "/dashboard/a/b/",
	} {
		w := get(t, h, target)
		if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") ||
			!bytes.Equal(w.Body.Bytes(), entry) {
			t.Errorf("GET %s = %d %q %q", target, w.Code, w.Header(), w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-cache" {
			t.Errorf("GET %s allows entry reuse without revalidation", target)
		}
	}
	for _, target := range []string{
		"/", "/api/v1/", "/dashboardish", "/dashboard.html", "/dashboard/missing.js",
		"/dashboard/users/jane.doe", "/dashboard/assets/missing.css",
	} {
		if w := get(t, h, target); w.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, w.Code)
		}
	}
}

func TestBuiltAssets(t *testing.T) {
	t.Parallel()
	h := New()
	err := fs.WalkDir(content, contentRoot, func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		wantAsset(t, h, name)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Vite emits both module scripts and styles as absolute references under the base.
	refs := regexp.MustCompile(`(?:src|href)="([^"]+)"`).FindAllSubmatch(entry, -1)
	for _, ref := range refs {
		target := string(ref[1])
		if !strings.HasPrefix(target, "/dashboard/") || get(t, h, target).Code != http.StatusOK {
			t.Errorf("entry references an unavailable asset: %s", target)
		}
	}
}

// wantAsset compares one embedded file with its HTTP representation.
func wantAsset(t *testing.T, h http.Handler, name string) {
	t.Helper()
	body, err := content.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	w := get(t, h, "/dashboard/"+strings.TrimPrefix(name, contentRoot+"/"))
	wantType := mime.TypeByExtension(path.Ext(name))
	if wantType == "" {
		wantType = http.DetectContentType(body)
	}
	if w.Code != http.StatusOK || !bytes.Equal(w.Body.Bytes(), body) || w.Header().Get("Content-Type") != wantType {
		t.Errorf("asset %s = %d %q; bytes equal: %v", name, w.Code, w.Header(), bytes.Equal(w.Body.Bytes(), body))
	}
	if regexp.MustCompile(`-[A-Za-z0-9_-]{8,}\.`).MatchString(path.Base(name)) &&
		w.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Errorf("hashed asset %s requires revalidation: %q", name, w.Header())
	}
	wantSecurity(t, w)
}

// wantSecurity checks the framing, sniffing and script/style source restrictions.
func wantSecurity(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Header().Get("X-Frame-Options") != "SAMEORIGIN" || w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing security headers: %q", w.Header())
	}
	csp := w.Header().Get("Content-Security-Policy")
	for _, directive := range []string{"default-src 'none'", "script-src 'self'", "style-src 'self'", "frame-ancestors 'self'"} {
		if !strings.Contains(csp, directive) {
			t.Errorf("CSP %q lacks %q", csp, directive)
		}
	}
}

func TestSecurityAndRevalidation(t *testing.T) {
	t.Parallel()
	h := New()
	for _, target := range []string{"/dashboard", "/dashboard/route", "/dashboard/missing.js"} {
		wantSecurity(t, get(t, h, target))
	}
	w := get(t, h, "/dashboard/")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/dashboard/route", http.NoBody)
	req.Header.Set("If-None-Match", w.Header().Get("Etag"))
	conditional := httptest.NewRecorder()
	h.ServeHTTP(conditional, req)
	if conditional.Code != http.StatusNotModified || w.Header().Get("Etag") == "" {
		t.Errorf("entry revalidation = %d, headers %q", conditional.Code, w.Header())
	}
	post := httptest.NewRecorder()
	h.ServeHTTP(post, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/dashboard/", http.NoBody))
	if post.Code != http.StatusMethodNotAllowed || post.Header().Get("Allow") != "GET, HEAD" {
		t.Errorf("POST /dashboard/ = %d %q", post.Code, post.Header())
	}
	wantSecurity(t, post)
}
