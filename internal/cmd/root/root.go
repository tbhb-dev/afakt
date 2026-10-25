// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package root assembles the command tree and maps the result of a run to an exit code.
package root

import (
	"context"
	"errors"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/tbhb-dev/afakt/internal/cmd/help"
	"github.com/tbhb-dev/afakt/internal/cmd/status"
	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

const (
	runtimeExit = 1
	usageExit   = 2
)

// errNoCommand reports a command line without a command. The root help goes to standard
// error, and the exit code is 2.
var errNoCommand = errors.New("no command given")

// errCompletion stops Cobra's completion command. `exitCode` replaces it with the usage
// error for an unknown command.
var errCompletion = errors.New("completion requested")

// NewCmdRoot returns the root command with all its child commands.
func NewCmdRoot(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "afakt <COMMAND>",
		Short: "A proof of concept CLI",
		Args:  rootArgs,
		// Cobra adds its hidden completion command whenever the first command name is
		// `__complete`, and no option turns that off. This hook stops it before it prints
		// completions. `exitCode` reports the name as an unknown command.
		PersistentPreRunE: func(c *cobra.Command, _ []string) error {
			if isCompletion(c) {
				return errCompletion
			}
			return nil
		},
		RunE: func(*cobra.Command, []string) error {
			return errNoCommand
		},
		// `Execute` prints every error and the usage itself, through `exitCode`.
		SilenceErrors: true,
		SilenceUsage:  true,
		// The root parses its own flags before it looks for the command, so the version flag
		// prints the version, as clap does.
		TraverseChildren:  true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	cmd.SetIn(f.In)
	cmd.SetOut(f.Out)
	cmd.SetErr(f.ErrOut)
	// Every command inherits the translation. Without it, Cobra would print its own help for
	// the help error from pflag before `exitCode` sees it.
	cmd.SetFlagErrorFunc(cmdutil.TranslateFlagError)
	cmdutil.AddHelpFlag(cmd)
	cmdutil.AddVersionFlag(cmd)
	helpCmd := help.NewCmdHelp(f, nil)
	cmdutil.AddCommands(cmd, status.NewCmdStatus(f, nil), helpCmd)
	cmd.SetHelpCommand(helpCmd)
	return cmd
}

// rootArgs rejects any argument that isn't a command name. Cobra has already matched the
// command names, so an argument here is unknown or came after the end of the flags.
func rootArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if cmd.ArgsLenAtDash() == 0 && cmdutil.Subcommand(cmd, args[0]) != nil {
		return cmdutil.UnexpectedArgument(args[0])
	}
	return unrecognized(args[0])
}

func unrecognized(name string) *cmdutil.FlagError {
	return &cmdutil.FlagError{Message: fmt.Sprintf("unrecognized subcommand '%s'", name)}
}

func isCompletion(cmd *cobra.Command) bool {
	return cmd.Name() == cobra.ShellCompRequestCmd
}

// Execute runs the command line in `args` and returns the process exit code.
func Execute(ctx context.Context, f *cmdutil.Factory, args []string) int {
	cmd := NewCmdRoot(f)
	// Cobra reads the process arguments when the arguments are nil.
	if args == nil {
		args = []string{}
	}
	// Traversal reads a leading terminator as a flag that takes a value and would dispatch
	// past it. The root parses the whole command line instead, so every argument after the
	// terminator counts as positional.
	if len(args) > 0 && args[0] == "--" {
		cmd.TraverseChildren = false
	}
	cmd.SetArgs(args)
	failed, err := cmd.ExecuteContextC(ctx)
	return exitCode(f, failed, err)
}

// exitCode prints what err calls for and returns the exit code. Help and version output
// and usage errors ignore write errors, as clap does. A failed write of the status line is
// a runtime failure, except on a closed standard output, which counts as success.
func exitCode(f *cmdutil.Factory, failed *cobra.Command, err error) int {
	var (
		flagErr  *cmdutil.FlagError
		writeErr *cmdutil.WriteError
	)
	// The root's own flags fail before Cobra reaches the flag error function.
	err = cmdutil.TranslateFlagError(failed, err)
	// Either the completion command's own argument check or the pre-run hook stopped it.
	if isCompletion(failed) {
		failed, err = failed.Root(), unrecognized(failed.CalledAs())
	}
	switch {
	case err == nil:
		return 0
	case errors.Is(err, cmdutil.ErrHelp):
		_, _ = fmt.Fprint(f.Out, cmdutil.HelpText(failed))
		return 0
	case errors.Is(err, cmdutil.ErrVersion):
		_, _ = fmt.Fprintf(f.Out, "%s %s\n", failed.Root().Name(), f.Version)
		return 0
	case errors.Is(err, errNoCommand):
		_, _ = fmt.Fprint(f.ErrOut, cmdutil.HelpText(failed))
		return usageExit
	case errors.As(err, &flagErr):
		_, _ = fmt.Fprint(f.ErrOut, cmdutil.UsageText(failed, flagErr))
		return usageExit
	case errors.As(err, &writeErr) && errors.Is(writeErr.Err, syscall.EBADF):
		return 0
	}
	_, _ = fmt.Fprintf(f.ErrOut, "Error: %s\n", err)
	return runtimeExit
}
