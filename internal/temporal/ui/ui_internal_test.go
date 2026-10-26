// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package ui

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"testing/synctest"

	"github.com/temporalio/ui-server/v2/server/config"
	"google.golang.org/grpc"
)

func TestCheckPath(t *testing.T) {
	t.Parallel()
	for _, prefix := range []string{
		DefaultPath, "/tools/temporal", "/apiary", "/dashboard-tools",
		"/ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~$&+:=@",
	} {
		if err := CheckPath(prefix); err != nil {
			t.Errorf("CheckPath(%q) = %v", prefix, err)
		}
	}
	for _, prefix := range []string{
		"", "/", "temporal", "/temporal/", "/a/../b", "/a//b", "/a?b", "/a#b", "/a%20b", "/a b", "/a/{b}",
		"/t\u00e9st", "/a;b", "/a,b", "/a!b", "/a'b", "/a(b", "/a)b", "/a*b", "/a[b", "/a]b", "/a\\b", "/a\"b", "/a<b", "/a>b", "/a|b", "/a^b", "/a`b",
		"/api", "/api/v1", "/dashboard", "/dashboard/settings", "/livez", "/readyz", "/healthz", "/readyz/child",
	} {
		if err := CheckPath(prefix); err == nil || !strings.Contains(err.Error(), strconv.Quote(prefix)) {
			t.Errorf("CheckPath(%q) = %v, want an error naming the path", prefix, err)
		}
	}
}

func TestService(t *testing.T) {
	t.Parallel()
	s := New(func() string { return "127.0.0.1:1" }, "default", "/tools/temporal")
	wantResponse(t, s, "/tools/temporal/", http.StatusNotFound, "404")
	if err := s.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	wantResponse(t, s, "/tools/temporal/", http.StatusOK, "<html")
	wantResponse(t, s, "/tools/temporal/api/v1/settings", http.StatusOK, `"DisableNewsFetch":true`)
	wantResponse(t, s, "/tools/temporal/api/v1/ui-extensions", http.StatusOK, `"Enabled":false`)
	wantResponse(t, s, "/tools/temporal/_app/missing.js", http.StatusNotFound, "404")
	if err := s.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	wantResponse(t, s, "/tools/temporal/", http.StatusNotFound, "404")
	if err := s.Stop(t.Context()); err == nil {
		t.Fatal("second Stop() succeeded, want a closed-connection error")
	}
}

func TestServiceRun(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		s := New(nil, "default", DefaultPath)
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() { done <- s.Run(ctx) }()
		synctest.Wait()
		select {
		case err := <-done:
			t.Fatalf("Run() returned before cancellation: %v", err)
		default:
		}
		cancel()
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	})
}

func TestStartFailures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		files fs.FS
		want  string
	}{
		{"connection", content(), "connect Temporal UI"},
		{"assets", fstest.MapFS{}, "temporal UI assets"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := New(func() string { return "127.0.0.1:1" }, "default", DefaultPath)
			s.files = tc.files
			if tc.name == "connection" {
				s.dial = func(string, ...grpc.DialOption) (*grpc.ClientConn, error) {
					return nil, fs.ErrInvalid
				}
			}
			if err := s.Start(t.Context()); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Start() = %v, want %q", err, tc.want)
			}
			wantResponse(t, s, "/temporal/", http.StatusNotFound, "404")
		})
	}
}

type failedConfig struct{ err error }

func (f failedConfig) GetConfig() (*config.Config, error) { return nil, f.err }

func TestConfigurationFailure(t *testing.T) {
	t.Parallel()
	want := fs.ErrPermission
	if _, err := newHandler(failedConfig{want}, nil, DefaultPath, content()); !errors.Is(err, want) {
		t.Fatalf("newHandler() = %v, want %v", err, want)
	}
}

func TestAssetPackagingFailure(t *testing.T) {
	t.Parallel()
	defer func() {
		got, ok := recover().(error)
		if !ok || !errors.Is(got, fs.ErrNotExist) {
			t.Errorf("packaging panic = %v, want %v", got, fs.ErrNotExist)
		}
	}()
	mustFiles(nil, fs.ErrNotExist)
}

func wantResponse(t *testing.T, handler http.Handler, target string, status int, body string) {
	t.Helper()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody))
	if w.Code != status || !strings.Contains(w.Body.String(), body) {
		t.Errorf("GET %s = %d %q, want %d containing %q", target, w.Code, w.Body.String(), status, body)
	}
}
