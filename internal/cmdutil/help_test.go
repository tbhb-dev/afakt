// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package cmdutil_test

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

func TestHelpText(t *testing.T) {
	t.Parallel()
	tool, run, about := newTool()
	cases := []struct {
		name string
		cmd  *cobra.Command
		want string
	}{
		{
			name: "root lists commands in the order added",
			cmd:  tool,
			want: `A test tool

Usage: tool <COMMAND>

Commands:
  run    Run something
  about  Describe a topic

Options:
  -h, --help     Print help
  -V, --version  Print version
`,
		},
		{
			name: "command with a help flag",
			cmd:  run,
			want: "Run something\n\nUsage: tool run\n\nOptions:\n  -h, --help  Print help\n",
		},
		{
			name: "command without flags lists its arguments",
			cmd:  about,
			want: `Describe a topic

Usage: tool about [TOPIC]...

Arguments:
  [TOPIC]...  Topics to describe
`,
		},
	}
	for _, tc := range cases {
		if got := cmdutil.HelpText(tc.cmd); got != tc.want {
			t.Errorf("%s: HelpText() =\n%s\nwant\n%s", tc.name, got, tc.want)
		}
	}
}

func TestUsageText(t *testing.T) {
	t.Parallel()
	tool, run, about := newTool()
	cases := []struct {
		name   string
		failed *cobra.Command
		err    *cmdutil.FlagError
		want   string
	}{
		{
			name:   "failed command",
			failed: run,
			err:    &cmdutil.FlagError{Message: "bad"},
			want:   "error: bad\n\nUsage: tool run\n\nFor more information, try '--help'.\n",
		},
		{
			name:   "cited flag joins the usage line",
			failed: tool,
			err:    &cmdutil.FlagError{Flag: "--help", Message: "bad"},
			want:   "error: bad\n\nUsage: tool --help <COMMAND>\n\nFor more information, try '--help'.\n",
		},
		{
			name:   "named command",
			failed: about,
			err:    &cmdutil.FlagError{Cmd: run, Message: "bad"},
			want:   "error: bad\n\nUsage: tool run\n\nFor more information, try '--help'.\n",
		},
		{
			name:   "command without flags has no hint",
			failed: tool,
			err:    &cmdutil.FlagError{Cmd: about, Message: "bad"},
			want:   "error: bad\n\nUsage: tool about [TOPIC]...\n",
		},
	}
	for _, tc := range cases {
		if got := cmdutil.UsageText(tc.failed, tc.err); got != tc.want {
			t.Errorf("%s: UsageText() = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestSubcommand(t *testing.T) {
	t.Parallel()
	tool, run, _ := newTool()
	tool.AddCommand(&cobra.Command{Use: "secret", Hidden: true})
	if got := cmdutil.Subcommand(tool, "run"); got != run {
		t.Errorf("Subcommand(run) = %v, want the run command", got)
	}
	for _, name := range []string{"secret", "nosuch"} {
		if got := cmdutil.Subcommand(tool, name); got != nil {
			t.Errorf("Subcommand(%s) = %v, want nil", name, got)
		}
	}
}
