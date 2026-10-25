// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package help holds the help command, which prints the help of the root or of a named
// command.
package help

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

// Options holds what the help command needs.
type Options struct {
	Out   io.Writer
	Root  *cobra.Command
	Names []string
}

// NewCmdHelp returns the help command, which prints the help of the root or of the named
// command. A nil `runF` runs `helpRun`.
func NewCmdHelp(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{Out: f.Out}
	cmd := &cobra.Command{
		Use:   "help [COMMAND]...",
		Short: "Print this message or the help of the given subcommand(s)",
		// The help command takes command names as its arguments, so it has no flags at all.
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			opts.Root = c.Root()
			opts.Names = args
			if runF != nil {
				return runF(opts)
			}
			return helpRun(opts)
		},
	}
	cmdutil.SetArguments(cmd, "[COMMAND]...", "Print help for the subcommand(s)")
	return cmd
}

// helpRun looks up the command in its first argument and prints that help. No command
// has children of its own, so a second name is always unrecognized.
func helpRun(opts *Options) error {
	current := opts.Root
	for i, name := range opts.Names {
		next := cmdutil.Subcommand(opts.Root, name)
		if next == nil || i > 0 {
			return &cmdutil.FlagError{Cmd: current, Message: fmt.Sprintf("unrecognized subcommand '%s'", name)}
		}
		current = next
	}
	_, _ = fmt.Fprint(opts.Out, cmdutil.HelpText(current))
	return nil
}
