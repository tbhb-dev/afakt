// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
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
