// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package cmdutil_test

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

var errOther = errors.New("other failure")

// newTool returns a small command tree that uses every helper, as the root, a run command,
// and a command without flags. The root has help and version flags, and the run command
// has a help flag.
func newTool() (*cobra.Command, *cobra.Command, *cobra.Command) {
	tool := &cobra.Command{Use: "tool <COMMAND>", Short: "A test tool"}
	cmdutil.AddHelpFlag(tool)
	cmdutil.AddVersionFlag(tool)
	run := &cobra.Command{Use: "run", Short: "Run something"}
	cmdutil.AddHelpFlag(run)
	about := &cobra.Command{Use: "about [TOPIC]...", Short: "Describe a topic", DisableFlagParsing: true}
	cmdutil.SetArguments(about, "[TOPIC]...", "Topics to describe")
	cmdutil.AddCommands(tool, run, about)
	return tool, run, about
}

func TestTranslateFlagError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		root    bool
		args    []string
		flag    string
		message string
	}{
		{
			name:    "unknown long flag",
			root:    true,
			args:    []string{"--bogus"},
			message: "unexpected argument '--bogus' found",
		},
		{
			name:    "unknown long flag with value",
			root:    true,
			args:    []string{"--bogus=x"},
			message: "unexpected argument '--bogus' found",
		},
		{name: "unknown short flag", root: true, args: []string{"-x"}, message: "unexpected argument '-x' found"},
		{
			name:    "unknown short flag before help",
			root:    true,
			args:    []string{"-xh"},
			message: "unexpected argument '-x' found",
		},
		{name: "version on a command without it", args: []string{"-V"}, message: "unexpected argument '-V' found"},
		{name: "three dashes", root: true, args: []string{"---x"}, message: "unexpected argument '---x' found"},
		{name: "no flag name", root: true, args: []string{"--=x"}, message: "unexpected argument '--' found"},
		{
			name:    "help with value",
			root:    true,
			args:    []string{"--help=1"},
			flag:    "--help",
			message: "unexpected value '1' for '--help' found; no more were expected",
		},
		{
			name:    "version with empty value",
			root:    true,
			args:    []string{"--version="},
			flag:    "--version",
			message: "unexpected value '' for '--version' found; no more were expected",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tool, run, _ := newTool()
			cmd := run
			if tc.root {
				cmd = tool
			}
			var flagErr *cmdutil.FlagError
			if err := cmdutil.TranslateFlagError(cmd, cmd.Flags().Parse(tc.args)); !errors.As(err, &flagErr) {
				t.Fatalf("parse(%q) error = %v, want a FlagError", tc.args, err)
			}
			if flagErr.Message != tc.message || flagErr.Flag != tc.flag || flagErr.Cmd != nil {
				t.Errorf("FlagError = %+v, want message %q and flag %q", *flagErr, tc.message, tc.flag)
			}
			if flagErr.Error() != tc.message {
				t.Errorf("Error() = %q, want %q", flagErr.Error(), tc.message)
			}
		})
	}
}

func TestTranslateFlagErrorPassesOtherErrors(t *testing.T) {
	t.Parallel()
	tool, _, _ := newTool()
	if err := cmdutil.TranslateFlagError(tool, nil); err != nil {
		t.Errorf("TranslateFlagError(nil) = %v, want nil", err)
	}
	if err := cmdutil.TranslateFlagError(tool, errOther); !errors.Is(err, errOther) {
		t.Errorf("TranslateFlagError(errOther) = %v, want errOther", err)
	}
}

func TestActionFlagsEndParsing(t *testing.T) {
	t.Parallel()
	// The first help or version flag ends parsing, so a later flag in the same group or
	// argument list never counts. A short flag acts on its first letter alone, as in clap.
	cases := []struct {
		args []string
		want error
	}{
		{args: []string{"-h"}, want: cmdutil.ErrHelp},
		{args: []string{"--help"}, want: cmdutil.ErrHelp},
		{args: []string{"-hV"}, want: cmdutil.ErrHelp},
		{args: []string{"-hx"}, want: cmdutil.ErrHelp},
		{args: []string{"-h=1"}, want: cmdutil.ErrHelp},
		{args: []string{"-V"}, want: cmdutil.ErrVersion},
		{args: []string{"--version", "--bogus"}, want: cmdutil.ErrVersion},
		{args: []string{"-Vh"}, want: cmdutil.ErrVersion},
		{args: []string{"-V=1"}, want: cmdutil.ErrVersion},
	}
	for _, tc := range cases {
		tool, _, _ := newTool()
		if err := cmdutil.TranslateFlagError(tool, tool.Flags().Parse(tc.args)); !errors.Is(err, tc.want) {
			t.Errorf("parse(%q) error = %v, want %v", tc.args, err, tc.want)
		}
	}
}

func TestNoArgs(t *testing.T) {
	t.Parallel()
	if err := cmdutil.NoArgs(nil, nil); err != nil {
		t.Errorf("NoArgs() with no arguments = %v, want nil", err)
	}
	var flagErr *cmdutil.FlagError
	if err := cmdutil.NoArgs(nil, []string{"extra", "more"}); !errors.As(err, &flagErr) ||
		flagErr.Message != "unexpected argument 'extra' found" {
		t.Errorf("NoArgs() with arguments = %v, want an unexpected argument error for 'extra'", err)
	}
}
