// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/tbhb-dev/afakt/internal/buildmeta"
)

const (
	// runMainEnv tells a re-executed test binary to run the real main function instead of the tests.
	runMainEnv = "AFAKT_TEST_RUN_MAIN"

	// childTimeout bounds one child process. It only matters when something hangs.
	childTimeout = 30 * time.Second

	// childWaitDelay bounds the wait for the child's output pipes after the child is killed.
	childWaitDelay = 5 * time.Second
)

// TestMain re-executes the test binary as the afakt program. A child started with runMainEnv
// set runs main and never reaches the tests. The process tests then see the exit status and
// the signal handling of the real program, and the package doesn't need a separate build.
func TestMain(m *testing.M) {
	if os.Getenv(runMainEnv) != "" {
		main()
	}
	os.Exit(m.Run())
}

func skipUnlessUnix(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the process tests need Unix signals")
	}
}

// childResult is what a process test observes of a finished child.
type childResult struct {
	exitCode int
	state    string
	stdout   string
	stderr   string
}

func newChild(ctx context.Context, t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Env = append(os.Environ(), runMainEnv+"=1")
	cmd.WaitDelay = childWaitDelay
	return cmd
}

func result(ctx context.Context, t *testing.T, cmd *exec.Cmd, stdout, stderr string, waitErr error) childResult {
	t.Helper()
	if ctx.Err() != nil {
		t.Fatalf("child did not finish within %s: %v", childTimeout, waitErr)
	}
	var exitErr *exec.ExitError
	if waitErr != nil && !errors.As(waitErr, &exitErr) {
		t.Fatalf("running the child: %v", waitErr)
	}
	return childResult{
		exitCode: cmd.ProcessState.ExitCode(),
		state:    cmd.ProcessState.String(),
		stdout:   stdout,
		stderr:   stderr,
	}
}

// runChild runs afakt with `args` as a child process whose standard output is output, or a
// buffer when output is nil. The caller closes its own copy of output after the call.
func runChild(t *testing.T, output *os.File, args ...string) childResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), childTimeout)
	defer cancel()
	cmd := newChild(ctx, t, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	if output != nil {
		cmd.Stdout = output
	}
	cmd.Stderr = &stderr
	err := cmd.Run()
	return result(ctx, t, cmd, stdout.String(), stderr.String(), err)
}

func TestVersionComesFromBuildMetadata(t *testing.T) {
	t.Parallel()
	// The version comes from the build metadata that main reads.
	if buildmeta.Get().Version == "" {
		t.Error("buildmeta.Get().Version is empty")
	}
	got := runChild(t, nil, "--version")
	if want := "afakt " + buildmeta.Get().Version + "\n"; got.exitCode != 0 || got.stdout != want {
		t.Errorf("exit code %d, stdout %q; want 0 and %q", got.exitCode, got.stdout, want)
	}
}

func TestStatusOnClosedPipe(t *testing.T) {
	t.Parallel()
	skipUnlessUnix(t)
	// A write to a pipe with no reader must end the program with exit status 1 and the broken
	// pipe error, not with death by SIGPIPE. The read end closes before the child starts, so
	// the write always fails.
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	if err = reader.Close(); err != nil {
		t.Fatalf("closing the read end: %v", err)
	}
	got := runChild(t, writer, "status")
	if err = writer.Close(); err != nil {
		t.Fatalf("closing the write end: %v", err)
	}
	// ExitCode is -1 for a child that a signal ended.
	if got.exitCode != 1 {
		t.Errorf("exit code = %d (%s), want 1", got.exitCode, got.state)
	}
	want := "Error: Os { code: 32, kind: BrokenPipe, message: \"Broken pipe\" }\n"
	if got.stderr != want {
		t.Errorf("stderr = %q, want %q", got.stderr, want)
	}
}

func TestStatusOnFullDevice(t *testing.T) {
	t.Parallel()
	// This is the failure that `afakt status > /dev/full` reproduces.
	if runtime.GOOS != "linux" {
		t.Skip("/dev/full exists only on Linux")
	}
	full, err := os.OpenFile("/dev/full", os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("opening /dev/full: %v", err)
	}
	got := runChild(t, full, "status")
	if err = full.Close(); err != nil {
		t.Fatalf("closing /dev/full: %v", err)
	}
	if got.exitCode != 1 {
		t.Errorf("exit code = %d (%s), want 1", got.exitCode, got.state)
	}
	want := "Error: Os { code: 28, kind: StorageFull, message: \"No space left on device\" }\n"
	if got.stderr != want {
		t.Errorf("stderr = %q, want %q", got.stderr, want)
	}
}

// logBuffer guards a buffer with a lock, because the child logs into it while the test
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
	deadline := time.Now().Add(childTimeout)
	for time.Now().Before(deadline) {
		if _, rest, found := strings.Cut(logs.String(), "msg=listening address="); found {
			address, _, _ := strings.Cut(rest, "\n")
			return address, true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "", false
}

func waitLive(t *testing.T, address string) bool {
	t.Helper()
	deadline := time.Now().Add(childTimeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+"/livez", http.NoBody)
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

// runServerChild starts the server command as a child process with the given arguments.
// Once the logged address answers /livez, it sends the signal and returns after the child
// is reaped.
func runServerChild(t *testing.T, sig syscall.Signal, args ...string) childResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), childTimeout)
	defer cancel()
	cmd := newChild(ctx, t, append([]string{"server"}, args...)...)
	var stderr logBuffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting the child: %v", err)
	}
	address, ok := waitAddress(&stderr)
	if !ok || !strings.HasPrefix(address, "127.0.0.1:") || !waitLive(t, address) {
		killErr := cmd.Process.Kill()
		t.Fatalf("the child never answered /livez on %q (kill: %v, wait: %v); stderr:\n%s",
			address, killErr, cmd.Wait(), stderr.String())
	}
	if err := cmd.Process.Signal(sig); err != nil {
		t.Fatalf("signaling the child: %v", err)
	}
	waitErr := cmd.Wait()
	return result(ctx, t, cmd, "", stderr.String(), waitErr)
}

func TestServerStopsOnSignal(t *testing.T) {
	t.Parallel()
	skipUnlessUnix(t)
	// Each flag form starts the server, and the signal shuts it down cleanly with exit status 0.
	cases := []struct {
		name string
		sig  syscall.Signal
		args []string
	}{
		{
			name: "SIGTERM with equals form",
			sig:  syscall.SIGTERM,
			args: []string{"--target=api", "--listen=127.0.0.1:0", "--shutdown-timeout=10s"},
		},
		{
			name: "SIGINT with separate form",
			sig:  syscall.SIGINT,
			args: []string{"--target", "api", "--listen", "127.0.0.1:0", "--shutdown-timeout", "10s"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := runServerChild(t, tc.sig, tc.args...)
			if got.exitCode != 0 {
				t.Errorf("exit code = %d (%s), want 0; stderr:\n%s", got.exitCode, got.state, got.stderr)
			}
			for _, want := range []string{
				`msg="shutting down"`,
				`msg="service terminated" service=api`,
				`msg="service terminated" service=api-listener`,
			} {
				if !strings.Contains(got.stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, got.stderr)
				}
			}
		})
	}
}
