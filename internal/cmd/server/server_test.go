// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tbhb-dev/afakt/internal/cmd/root"
	servercmd "github.com/tbhb-dev/afakt/internal/cmd/server"
	"github.com/tbhb-dev/afakt/internal/cmdutil"
	"github.com/tbhb-dev/afakt/internal/server"
)

const (
	runtimeExit = 1
	usageExit   = 2

	serverUsageLine = "afakt server [OPTIONS] --target <TARGET>"

	// serveTimeout bounds each wait on a server in these tests. It only matters when
	// something hangs.
	serveTimeout = 30 * time.Second
)

// newFactory returns a factory with empty input and a fixed version.
func newFactory(out, errOut io.Writer) *cmdutil.Factory {
	return &cmdutil.Factory{In: strings.NewReader(""), Out: out, ErrOut: errOut, Version: "1.2.3-test"}
}

// parseServer runs the server command on `args` with a run function that records the
// options, and returns them, or nil when the run function never ran.
func parseServer(args []string) (*servercmd.Options, error) {
	var got *servercmd.Options
	cmd := servercmd.NewCmdServer(newFactory(io.Discard, io.Discard), func(opts *servercmd.Options) error {
		got = opts
		return nil
	})
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err != nil {
		return got, fmt.Errorf("server %q: %w", args, err)
	}
	return got, nil
}

// clapError is the error text clap prints: the message, the usage line and the hint.
func clapError(message, usage string) string {
	return "error: " + message + "\n\nUsage: " + usage + "\n\nFor more information, try '--help'.\n"
}

func TestNewCmdServerAccepts(t *testing.T) {
	t.Parallel()
	api := []string{server.TargetAPI}
	webhook := []string{server.TargetWebhook}
	combined := []string{server.TargetAPI, server.TargetWebhook}
	cases := []struct {
		name string
		args []string
		want server.Config
	}{
		{
			name: "defaults",
			args: []string{"--target=api"},
			want: server.Config{Targets: api, Listen: "127.0.0.1:8080", ShutdownTimeout: 30 * time.Second},
		},
		{
			name: "equals form",
			args: []string{"--target=api", "--listen=127.0.0.1:9000", "--shutdown-timeout=1m30s"},
			want: server.Config{Targets: api, Listen: "127.0.0.1:9000", ShutdownTimeout: 90 * time.Second},
		},
		{
			name: "separate form",
			args: []string{"--shutdown-timeout", "5s", "--listen", "0.0.0.0:80", "--target", "api"},
			want: server.Config{Targets: api, Listen: "0.0.0.0:80", ShutdownTimeout: 5 * time.Second},
		},
		{
			name: "host name and ephemeral port",
			args: []string{"--target", "api", "--listen=localhost:0"},
			want: server.Config{Targets: api, Listen: "localhost:0", ShutdownTimeout: 30 * time.Second},
		},
		{
			name: "IPv6 host",
			args: []string{"--target=api", "--listen", "[::1]:65535", "--shutdown-timeout=1ns"},
			want: server.Config{Targets: api, Listen: "[::1]:65535", ShutdownTimeout: time.Nanosecond},
		},
		{
			name: "terminator at the end",
			args: []string{"--target=api", "--"},
			want: server.Config{Targets: api, Listen: "127.0.0.1:8080", ShutdownTimeout: 30 * time.Second},
		},
		{
			name: "webhook defaults",
			args: []string{"--target", "webhook"},
			want: server.Config{Targets: webhook, Listen: "127.0.0.1:8080", ShutdownTimeout: 30 * time.Second},
		},
		{
			name: "webhook equals form",
			args: []string{"--target=webhook", "--listen=127.0.0.1:9001", "--shutdown-timeout=2m"},
			want: server.Config{Targets: webhook, Listen: "127.0.0.1:9001", ShutdownTimeout: 2 * time.Minute},
		},
		{
			name: "webhook separate form",
			args: []string{"--listen", "[::1]:0", "--target", "webhook", "--shutdown-timeout", "250ms"},
			want: server.Config{Targets: webhook, Listen: "[::1]:0", ShutdownTimeout: 250 * time.Millisecond},
		},
		{
			name: "api and webhook",
			args: []string{"--target=api,webhook"},
			want: server.Config{Targets: combined, Listen: "127.0.0.1:8080", ShutdownTimeout: 30 * time.Second},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts, err := parseServer(tc.args)
			if err != nil || opts == nil {
				t.Fatalf("NewCmdServer rejected %q: %v", tc.args, err)
			}
			got := opts.Config
			if !slices.Equal(got.Targets, tc.want.Targets) || got.Listen != tc.want.Listen ||
				got.ShutdownTimeout != tc.want.ShutdownTimeout {
				t.Errorf("config = %+v, want %+v", got, tc.want)
			}
			if opts.ErrOut != io.Discard {
				t.Errorf("ErrOut = %v, want the factory's error stream", opts.ErrOut)
			}
		})
	}
}

func TestServerRejects(t *testing.T) {
	t.Parallel()
	// Every configuration error exits with code 2 and prints the error with the usage on
	// standard error. In these cases the run function never runs, so nothing reaches the code
	// that binds.
	const choices = "choose from api, webhook, dashboard, worker, all"
	invalid := func(value, flag, reason string) string {
		return clapError("invalid value '"+value+"' for '"+flag+"': "+reason, serverUsageLine)
	}
	target := func(value, reason string) string { return invalid(value, "--target <TARGET>", reason) }
	unavailableIn := func(value, name string) string {
		return target(value, "target is not available yet: '"+name+"'; only 'api' and 'webhook' can run in this build")
	}
	unavailable := func(name string) string { return unavailableIn(name, name) }
	listen := func(value, reason string) string { return invalid(value, "--listen <ADDR>", reason) }
	timeout := func(value, reason string) string {
		return invalid(value, "--shutdown-timeout <DURATION>", reason)
	}
	const (
		listenForm   = "expected HOST:PORT, such as 127.0.0.1:8080"
		listenHost   = "the host is missing; give one, such as 127.0.0.1:8080"
		listenPort   = "the port must be a number from 0 to 65535"
		durationForm = "expected a Go duration, such as 30s or 1m30s"
		durationPos  = "the duration must be positive"
	)
	missingValue := func(flag string) string {
		return clapError("a value is required for '"+flag+"' but none was supplied", serverUsageLine)
	}
	cases := []struct {
		name   string
		args   []string
		stderr string
	}{
		{
			name: "no target",
			stderr: clapError(
				"the following required arguments were not provided:\n  --target <TARGET>", serverUsageLine,
			),
		},
		{
			name: "no target with other flags",
			args: []string{"--listen=127.0.0.1:9000"},
			stderr: clapError(
				"the following required arguments were not provided:\n  --target <TARGET>", serverUsageLine,
			),
		},
		{name: "empty target", args: []string{"--target="}, stderr: target("", "no target given; "+choices)},
		{
			name:   "empty separate target",
			args:   []string{"--target", ""},
			stderr: target("", "no target given; "+choices),
		},
		{name: "target without value", args: []string{"--target"}, stderr: missingValue("--target <TARGET>")},
		{
			// The flag parser takes the next argument as the value, even when it starts with a dash.
			name:   "target followed by flag",
			args:   []string{"--target", "--listen=127.0.0.1:9000"},
			stderr: target("--listen=127.0.0.1:9000", "unknown target '--listen=127.0.0.1:9000'; "+choices),
		},
		{
			name:   "listen without value",
			args:   []string{"--target=api", "--listen"},
			stderr: missingValue("--listen <ADDR>"),
		},
		{
			name:   "timeout without value",
			args:   []string{"--target=api", "--shutdown-timeout"},
			stderr: missingValue("--shutdown-timeout <DURATION>"),
		},
		{
			name:   "negative separate timeout",
			args:   []string{"--target=api", "--shutdown-timeout", "-1s"},
			stderr: timeout("-1s", durationPos),
		},
		{
			name: "repeated target",
			args: []string{"--target=api", "--target", "api"},
			stderr: clapError(
				"the argument '--target <TARGET>' cannot be used multiple times", serverUsageLine,
			),
		},
		{
			name: "repeated listen",
			args: []string{"--listen=127.0.0.1:1", "--target=api", "--listen=127.0.0.1:2"},
			stderr: clapError(
				"the argument '--listen <ADDR>' cannot be used multiple times", serverUsageLine,
			),
		},
		{
			name:   "empty entry",
			args:   []string{"--target=api,"},
			stderr: target("api,", "empty entry in the target list"),
		},
		{
			name:   "unknown target",
			args:   []string{"--target=bogus"},
			stderr: target("bogus", "unknown target 'bogus'; "+choices),
		},
		{
			name:   "duplicate target",
			args:   []string{"--target=api,api"},
			stderr: target("api,api", "duplicate target 'api'"),
		},
		{
			name:   "duplicate webhook",
			args:   []string{"--target=webhook,api,webhook"},
			stderr: target("webhook,api,webhook", "duplicate target 'webhook'"),
		},
		{
			name:   "webhook and dashboard",
			args:   []string{"--target=webhook,dashboard"},
			stderr: unavailableIn("webhook,dashboard", "dashboard"),
		},
		{
			name:   "webhook without port",
			args:   []string{"--target=webhook", "--listen", "127.0.0.1"},
			stderr: listen("127.0.0.1", listenForm),
		},
		{
			name:   "webhook without host",
			args:   []string{"--target", "webhook", "--listen=:9000"},
			stderr: listen(":9000", listenHost),
		},
		{
			name:   "webhook zero timeout",
			args:   []string{"--target=webhook", "--shutdown-timeout", "0s"},
			stderr: timeout("0s", durationPos),
		},
		{
			name:   "webhook timeout not a duration",
			args:   []string{"--target", "api,webhook", "--shutdown-timeout=later"},
			stderr: timeout("later", durationForm),
		},
		{name: "dashboard", args: []string{"--target=dashboard"}, stderr: unavailable("dashboard")},
		{name: "worker", args: []string{"--target=worker"}, stderr: unavailable("worker")},
		{name: "all", args: []string{"--target=all"}, stderr: unavailable("all")},
		{name: "api and all", args: []string{"--target=api,all"}, stderr: unavailableIn("api,all", "all")},
		{
			name:   "no port",
			args:   []string{"--target=api", "--listen=127.0.0.1"},
			stderr: listen("127.0.0.1", listenForm),
		},
		{name: "empty listen", args: []string{"--target=api", "--listen="}, stderr: listen("", listenForm)},
		{name: "no host", args: []string{"--target=api", "--listen=:8080"}, stderr: listen(":8080", listenHost)},
		{
			name:   "empty port",
			args:   []string{"--target=api", "--listen=127.0.0.1:"},
			stderr: listen("127.0.0.1:", listenPort),
		},
		{name: "named port", args: []string{"--target=api", "--listen=127.0.0.1:http"}, stderr: listen(
			"127.0.0.1:http", listenPort,
		)},
		{name: "port too large", args: []string{"--target=api", "--listen=127.0.0.1:65536"}, stderr: listen(
			"127.0.0.1:65536", listenPort,
		)},
		{name: "negative port", args: []string{"--target=api", "--listen=127.0.0.1:-1"}, stderr: listen(
			"127.0.0.1:-1", listenPort,
		)},
		{
			name:   "zero timeout",
			args:   []string{"--target=api", "--shutdown-timeout=0s"},
			stderr: timeout("0s", durationPos),
		},
		{
			name:   "bare zero timeout",
			args:   []string{"--target=api", "--shutdown-timeout=0"},
			stderr: timeout("0", durationPos),
		},
		{
			name:   "negative timeout",
			args:   []string{"--target=api", "--shutdown-timeout=-5s"},
			stderr: timeout("-5s", durationPos),
		},
		{
			name:   "timeout without unit",
			args:   []string{"--target=api", "--shutdown-timeout=30"},
			stderr: timeout("30", durationForm),
		},
		{
			name:   "timeout not a duration",
			args:   []string{"--target=api", "--shutdown-timeout=soon"},
			stderr: timeout("soon", durationForm),
		},
		{
			name:   "unknown flag",
			args:   []string{"--target=api", "--port=80"},
			stderr: clapError("unexpected argument '--port' found", serverUsageLine),
		},
		{
			name:   "unknown short flag",
			args:   []string{"-x", "--target=api"},
			stderr: clapError("unexpected argument '-x' found", serverUsageLine),
		},
		{
			name:   "version flag",
			args:   []string{"--version"},
			stderr: clapError("unexpected argument '--version' found", serverUsageLine),
		},
		{
			name:   "positional argument",
			args:   []string{"--target=api", "extra"},
			stderr: clapError("unexpected argument 'extra' found", serverUsageLine),
		},
		{
			name:   "after terminator",
			args:   []string{"--target=api", "--", "--listen=127.0.0.1:9000"},
			stderr: clapError("unexpected argument '--listen=127.0.0.1:9000' found", serverUsageLine),
		},
		{
			name: "help flag with value",
			args: []string{"--help=1"},
			stderr: clapError(
				"unexpected value '1' for '--help' found; no more were expected",
				"afakt server --help [OPTIONS] --target <TARGET>",
			),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if opts, err := parseServer(tc.args); err == nil || opts != nil {
				t.Fatalf("NewCmdServer accepted %q: options %+v, error %v", tc.args, opts, err)
			}
			args := append([]string{"server"}, tc.args...)
			var stdout, stderr bytes.Buffer
			if got := root.Execute(t.Context(), newFactory(&stdout, &stderr), args); got != usageExit {
				t.Errorf("exit code = %d, want %d", got, usageExit)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if stderr.String() != tc.stderr {
				t.Errorf("stderr = %q\nwant     %q", stderr.String(), tc.stderr)
			}
		})
	}
}

// logBuffer guards a buffer with a lock, because the server logs into it while the test
// reads it.
type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *logBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Write(p)
	return len(p), nil
}

func (b *logBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// waitAddress waits until the log shows the address that the server listens on, and reports
// false if the wait times out. Listening on port 0 keeps parallel tests from colliding on a
// port.
func waitAddress(logs *logBuffer) (string, bool) {
	deadline := time.Now().Add(serveTimeout)
	for time.Now().Before(deadline) {
		if _, rest, found := strings.Cut(logs.String(), "msg=listening address="); found {
			address, _, _ := strings.Cut(rest, "\n")
			return address, true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "", false
}

// waitLive polls path at address until it answers 200, and reports false if the wait times
// out.
func waitLive(t *testing.T, address, path string) bool {
	t.Helper()
	deadline := time.Now().Add(serveTimeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+path, http.NoBody)
		if err != nil {
			cancel()
			t.Fatalf("NewRequest() error = %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_, err = io.Copy(io.Discard, resp.Body)
			if closeErr := resp.Body.Close(); err == nil {
				err = closeErr
			}
		}
		cancel()
		if err == nil && resp.StatusCode == http.StatusOK {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestServerRunShutsDownCleanly(t *testing.T) {
	t.Parallel()
	cases := []struct {
		target string
		// path answers 200 once the server is up.
		path string
	}{
		{target: "api", path: "/api/v1/"},
		{target: "webhook", path: "/readyz"},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			t.Parallel()
			stderr := runUntilCanceled(t, tc.target, tc.path)
			for _, want := range []string{
				"msg=\"server ready\"", "msg=\"shutting down\"", "msg=\"service terminated\" service=" + tc.target,
			} {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, stderr)
				}
			}
			if strings.Contains(stderr, "Error:") {
				t.Errorf("stderr reports an error:\n%s", stderr)
			}
		})
	}
}

// runUntilCanceled runs the real `serverRun` for target until path answers 200, then cancels
// its context as a signal does in main. It returns the logs once the command returns.
func runUntilCanceled(t *testing.T, target, path string) string {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var stderr logBuffer
	cmd := servercmd.NewCmdServer(newFactory(io.Discard, &stderr), nil)
	cmd.SetArgs([]string{"--target=" + target, "--listen=127.0.0.1:0", "--shutdown-timeout=" + serveTimeout.String()})
	done := make(chan error, 1)
	go func() { done <- cmd.ExecuteContext(ctx) }()
	address, ok := waitAddress(&stderr)
	if !ok || !strings.HasPrefix(address, "127.0.0.1:") || !waitLive(t, address, path) {
		cancel()
		t.Fatalf("the server never answered %s on %q; error %v, stderr:\n%s", path, address, <-done, stderr.String())
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("ExecuteContext() error = %v; stderr:\n%s", err, stderr.String())
		}
	case <-time.After(serveTimeout):
		t.Fatal("the server did not return")
	}
	return stderr.String()
}

func TestServerRunFailure(t *testing.T) {
	t.Parallel()
	// An address already in use is a runtime failure, which exits with code 1.
	var config net.ListenConfig
	taken, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer taken.Close()
	for _, target := range []string{"api", "webhook", "api,webhook"} {
		var stdout, stderr bytes.Buffer
		args := []string{"server", "--target", target, "--listen", taken.Addr().String(), "--shutdown-timeout=5s"}
		if got := root.Execute(t.Context(), newFactory(&stdout, &stderr), args); got != runtimeExit {
			t.Errorf("%s: exit code = %d, want %d", target, got, runtimeExit)
		}
		if want := "Error: start \"http-listener\": listen: "; !strings.Contains(stderr.String(), want) {
			t.Errorf("%s: stderr lacks %q:\n%s", target, want, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Errorf("%s: stdout = %q, want empty", target, stdout.String())
		}
	}
}
