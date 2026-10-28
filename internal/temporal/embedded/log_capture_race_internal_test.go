// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package embedded

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestLogCaptureRaceDetector builds isolated checks from the capture's own test source.
func TestLogCaptureRaceDetector(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("log_capture_internal_test.go")
	if err != nil {
		t.Fatal(err)
	}
	old := bytes.Replace(source, []byte("var buf logCapture"), []byte("var buf bytes.Buffer"), 1)
	if bytes.Equal(old, source) {
		t.Fatal("old capture substitution did not change the concurrency check")
	}
	old = bytes.Replace(old, []byte("package embedded"), []byte("package embedded\n\nimport \"bytes\""), 1)
	for _, tc := range []struct {
		name   string
		source []byte
		race   bool
	}{
		{"old", old, true},
		{"corrected", source, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkLogCaptureRace(t, tc.source, tc.race)
		})
	}
}

// checkLogCaptureRace builds and runs a generated capture check under the race detector.
func checkLogCaptureRace(t *testing.T, source []byte, wantRace bool) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module capture\n\ngo 1.27\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "capture_test.go"), source, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "go", "test", "-race", "-count=1", "-v", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if !wantRace {
		if err != nil {
			t.Fatalf("corrected capture: %v\n%s", err, output)
		}
		t.Log("corrected capture passed under the race detector")
		return
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 ||
		!strings.Contains(string(output), "WARNING: DATA RACE") ||
		!strings.Contains(string(output), "bytes.(*Buffer).String") ||
		!strings.Contains(string(output), "bytes.(*Buffer).Write") {
		t.Fatalf("old capture: %v, want a race detector failure\n%s", err, output)
	}
	t.Log("old bytes.Buffer capture reported a data race")
}
