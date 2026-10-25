// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

//go:build linux

package main

import (
	"os"
	"testing"
)

// TestStatusOnFullDevice checks the failure that
// `afakt status > /dev/full` reproduces: exit status 1 and the storage full error on standard error.
func TestStatusOnFullDevice(t *testing.T) {
	t.Parallel()
	full, err := os.OpenFile("/dev/full", os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("opening /dev/full: %v", err)
	}
	got := runStatusChild(t, full)
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
