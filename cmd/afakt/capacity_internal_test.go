// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

const (
	claudeSource    = "claude_usage"
	sourceFlag      = "--source"
	capacityCommand = "capacity"
	unknownStatus   = "unknown"
	parseCommand    = "parse"
)

func TestCapacityParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		status string
		vendor string
	}{
		{"Claude", claudeSource, "ok", "anthropic"},
		{"Codex", "codex_status", unknownStatus, "openai"},
		{"agy", "agy_usage", unknownStatus, "google"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer

			cmd := newRootCmd()
			cmd.SetOut(&out)
			cmd.SetIn(strings.NewReader("Current week (all models)\n1% used\n" +
				"Resets Nov 14 at 10pm (America/New_York)\n"))
			cmd.SetArgs(
				[]string{capacityCommand, parseCommand, sourceFlag, test.source, "--now", "2026-11-08T08:50:00-05:00"},
			)

			err := cmd.Execute()
			if err != nil {
				t.Fatal(err)
			}

			window := "other"
			values := ""

			if test.status == "ok" {
				window = "7d"
				values = `"utilization":0.01,"resetAt":1794711600,`
			}

			want := `[{"vendor":"` + test.vendor + `","window":"` + window + `",` + values +
				`"source":"` + test.source + `","freshness":"fresh","status":"` + test.status +
				`","capturedAt":1794145800}]` + "\n"

			if out.String() != want {
				t.Errorf("stdout = %s, want %s", &out, want)
			}
		})
	}
}

func TestCapacityFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   io.Reader
		out  io.Writer
		name string
		want string
		args []string
	}{
		{
			in:   strings.NewReader(""),
			out:  io.Discard,
			name: "time",
			args: []string{sourceFlag, claudeSource, "--now", "private input"},
			want: "invalid capacity time: expected RFC3339",
		},
		{
			in:   strings.NewReader(""),
			out:  io.Discard,
			name: "source",
			args: []string{sourceFlag, "private input"},
			want: "parse capacity: unsupported capacity source",
		},
		{
			in:   strings.NewReader(""),
			out:  io.Discard,
			name: "missing source",
			args: nil,
			want: "parse capacity: unsupported capacity source",
		},
		{
			in:   failingIO{},
			out:  io.Discard,
			name: "read",
			args: []string{sourceFlag, claudeSource},
			want: "cannot read capacity input",
		},
		{
			in:   strings.NewReader(""),
			out:  failingIO{},
			name: "write",
			args: []string{sourceFlag, claudeSource},
			want: "cannot write capacity JSON",
		},
		{
			in:   strings.NewReader(""),
			out:  io.Discard,
			name: "arguments",
			args: []string{sourceFlag, claudeSource, "extra"},
			want: "unknown command \"extra\" for \"afakt capacity parse\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := newRootCmd()
			cmd.SetArgs(append([]string{capacityCommand, parseCommand}, test.args...))
			cmd.SetIn(test.in)
			cmd.SetOut(test.out)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()

			if err == nil || err.Error() != test.want {
				t.Errorf("error = %v, want %s", err, test.want)
			}

			if test.name == "source" && errors.Unwrap(err) == nil {
				t.Error("source error lost its cause")
			}
		})
	}
}

type failingIO struct{}

func (failingIO) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func (failingIO) Write([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestCapacityDefaultTime(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{capacityCommand, parseCommand, sourceFlag, claudeSource})
	cmd.SetIn(strings.NewReader("private pane text"))
	cmd.SetOut(&out)

	before := time.Now().Unix()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var readings []map[string]any

	decodeErr := json.Unmarshal(out.Bytes(), &readings)
	if decodeErr != nil || len(readings) != 1 {
		t.Fatalf("JSON = %s, error = %v", &out, decodeErr)
	}

	captured, ok := readings[0]["capturedAt"].(float64)
	if !ok || int64(captured) < before || int64(captured) > time.Now().Unix() {
		t.Errorf("capture time = %v, want current time", readings[0]["capturedAt"])
	}

	if readings[0]["status"] != unknownStatus || strings.Contains(out.String(), "private pane text") {
		t.Errorf("unexpected output: %s", &out)
	}
}
