// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package status holds the status command, which prints a placeholder line.
package status

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

const statusLine = "status is not implemented yet."

// StatusOptions holds what the status command needs.
//
//nolint:revive // Every command names its options after itself, as the GitHub CLI does.
type StatusOptions struct {
	Out io.Writer
}

// NewCmdStatus returns the status command. A nil `runF` runs `statusRun`.
func NewCmdStatus(f *cmdutil.Factory, runF func(*StatusOptions) error) *cobra.Command {
	opts := &StatusOptions{Out: f.Out}
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the placeholder status",
		Args:  cmdutil.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			if runF != nil {
				return runF(opts)
			}
			return statusRun(opts)
		},
	}
	cmdutil.AddHelpFlag(cmd)
	return cmd
}

func statusRun(opts *StatusOptions) error {
	if _, err := fmt.Fprintln(opts.Out, statusLine); err != nil {
		return &cmdutil.WriteError{Err: err}
	}
	return nil
}
