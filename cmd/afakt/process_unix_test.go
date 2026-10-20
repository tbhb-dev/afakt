// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

//go:build unix

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	// runMainEnv tells a re-executed test binary to run the real main function instead of the tests.
	runMainEnv = "AFAKT_TEST_RUN_MAIN"

	// childTimeout bounds one child process. The child writes one line and exits, so the
	// limit only matters when something hangs.
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

// childResult is what a process test observes of a finished child.
type childResult struct {
	exitCode int
	state    string
	stderr   string
}

// runStatusChild runs `afakt status` as a child process whose standard output is output.
// The caller closes its own copy of output after the call. The call returns once the child
// is reaped.
func runStatusChild(t *testing.T, output *os.File) childResult {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), childTimeout)
	defer cancel()
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, exe, "status")
	cmd.Env = append(os.Environ(), runMainEnv+"=1")
	cmd.Stdout = output
	cmd.Stderr = &stderr
	cmd.WaitDelay = childWaitDelay
	runErr := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("child did not finish within %s: %v", childTimeout, runErr)
	}
	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		t.Fatalf("running the child: %v", runErr)
	}
	return childResult{
		exitCode: cmd.ProcessState.ExitCode(),
		state:    cmd.ProcessState.String(),
		stderr:   stderr.String(),
	}
}

// TestStatusOnClosedPipe checks that a write to a pipe with no reader ends the program with
// exit status 1 and the broken pipe error, not with death by SIGPIPE. The read end closes
// before the child starts, so the write always fails.
func TestStatusOnClosedPipe(t *testing.T) {
	t.Parallel()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	if err = reader.Close(); err != nil {
		t.Fatalf("closing the read end: %v", err)
	}
	got := runStatusChild(t, writer)
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
