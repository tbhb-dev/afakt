// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package help_test

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tbhb-dev/afakt/internal/cmd/help"
	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

// newTool returns a root with one child command and the help command, whose output goes to
// `out`.
func newTool(out io.Writer) (*cobra.Command, *cobra.Command) {
	tool := &cobra.Command{Use: "tool <COMMAND>", Short: "A test tool"}
	cmdutil.AddHelpFlag(tool)
	run := &cobra.Command{Use: "run", Short: "Run something"}
	cmdutil.AddHelpFlag(run)
	cmd := help.NewCmdHelp(&cmdutil.Factory{Out: out}, nil)
	cmdutil.AddCommands(tool, run, cmd)
	tool.SetHelpCommand(cmd)
	tool.SetOut(io.Discard)
	tool.SetErr(io.Discard)
	tool.SilenceErrors = true
	tool.SilenceUsage = true
	return tool, run
}

// TestNewCmdHelp checks that the help command takes every argument as a command name,
// flags included, and hands the names and the root to the run function.
func TestNewCmdHelp(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var got *help.Options
	cmd := help.NewCmdHelp(&cmdutil.Factory{Out: &stdout}, func(opts *help.Options) error {
		got = opts
		return nil
	})
	parent := &cobra.Command{Use: "afakt"}
	parent.AddCommand(cmd)
	parent.SetHelpCommand(cmd)
	parent.SetArgs([]string{"help", "status", "--help", "-x"})
	if err := parent.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got == nil || got.Out != &stdout || got.Root != parent ||
		!slices.Equal(got.Names, []string{"status", "--help", "-x"}) {
		t.Errorf("options = %+v, want the names, the root and the factory's output stream", got)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty: the run function never ran", stdout.String())
	}
}

type pick func(tool, run *cobra.Command) *cobra.Command

func pickTool(tool, _ *cobra.Command) *cobra.Command { return tool }

func pickRun(_, run *cobra.Command) *cobra.Command { return run }

type helpResult struct {
	tool, run *cobra.Command
	stdout    string
	err       error
}

// Each run of the help command gets a new tree, so the cases can run in parallel.
func runHelp(args []string) helpResult {
	var stdout bytes.Buffer
	tool, run := newTool(&stdout)
	tool.SetArgs(append([]string{"help"}, args...))
	err := tool.Execute()
	return helpResult{tool: tool, run: run, stdout: stdout.String(), err: err}
}

func TestHelpRunPrintsHelp(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want pick
	}{
		{name: "root", want: pickTool},
		{name: "command", args: []string{"run"}, want: pickRun},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := runHelp(tc.args)
			if got.err != nil {
				t.Fatalf("Execute() error = %v", got.err)
			}
			if want := cmdutil.HelpText(tc.want(got.tool, got.run)); got.stdout != want {
				t.Errorf("stdout = %q, want %q", got.stdout, want)
			}
		})
	}
}

func TestHelpRunRejectsNames(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		args    []string
		cmd     pick
		message string
	}{
		{name: "unknown command", args: []string{"nosuch"}, cmd: pickTool, message: "unrecognized subcommand 'nosuch'"},
		{name: "second name", args: []string{"run", "run"}, cmd: pickRun, message: "unrecognized subcommand 'run'"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := runHelp(tc.args)
			var flagErr *cmdutil.FlagError
			if !errors.As(got.err, &flagErr) || flagErr.Cmd != tc.cmd(got.tool, got.run) ||
				flagErr.Message != tc.message {
				t.Errorf("Execute() error = %#v, want a FlagError %q", got.err, tc.message)
			}
			if got.stdout != "" {
				t.Errorf("stdout = %q, want empty", got.stdout)
			}
		})
	}
}
