// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package status_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/tbhb-dev/afakt/internal/cmd/status"
	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

const wantStatus = "status is not implemented yet.\n"

var errWrite = errors.New("write failed")

// failWriter fails every write with its error.
type failWriter struct {
	err error
}

func (w failWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func newFactory(out, errOut io.Writer) *cmdutil.Factory {
	return &cmdutil.Factory{In: strings.NewReader(""), Out: out, ErrOut: errOut, Version: "1.2.3-test"}
}

func TestNewCmdStatus(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var got *status.StatusOptions
	cmd := status.NewCmdStatus(newFactory(&stdout, &bytes.Buffer{}), func(opts *status.StatusOptions) error {
		got = opts
		return nil
	})
	cmd.SetArgs([]string{"--"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got == nil || got.Out != &stdout {
		t.Errorf("options = %+v, want the factory's output stream", got)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty: the run function never ran", stdout.String())
	}
}

func TestNewCmdStatusRejectsArguments(t *testing.T) {
	t.Parallel()
	cmd := status.NewCmdStatus(newFactory(&bytes.Buffer{}, &bytes.Buffer{}), func(*status.StatusOptions) error {
		t.Error("the run function ran")
		return nil
	})
	cmd.SetArgs([]string{"extra"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	var flagErr *cmdutil.FlagError
	if err := cmd.Execute(); !errors.As(err, &flagErr) || flagErr.Message != "unexpected argument 'extra' found" {
		t.Errorf("Execute() error = %v, want an unexpected argument error", err)
	}
}

func TestStatusRun(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	cmd := status.NewCmdStatus(newFactory(&stdout, &bytes.Buffer{}), nil)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.String() != wantStatus {
		t.Errorf("stdout = %q, want %q", stdout.String(), wantStatus)
	}
}

func TestStatusRunWriteError(t *testing.T) {
	t.Parallel()
	cmd := status.NewCmdStatus(newFactory(failWriter{err: errWrite}, &bytes.Buffer{}), nil)
	cmd.SetArgs([]string{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	var writeErr *cmdutil.WriteError
	if err := cmd.Execute(); !errors.As(err, &writeErr) || !errors.Is(err, errWrite) {
		t.Errorf("Execute() error = %v, want a WriteError around %v", err, errWrite)
	}
}
