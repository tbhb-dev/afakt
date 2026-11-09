// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"go.tbhb.dev/afakt/internal/capacity"
)

var (
	errCapacityTime  = errors.New("invalid capacity time: expected RFC3339")
	errCapacityRead  = errors.New("cannot read capacity input")
	errCapacityWrite = errors.New("cannot write capacity JSON")
)

func newCapacityCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "capacity", Short: "Inspect plan capacity"}
	cmd.AddCommand(newCapacityParseCmd())

	return cmd
}

func newCapacityParseCmd() *cobra.Command {
	var source, captureAt string

	cmd := &cobra.Command{
		Use: "parse", Short: "Parse a captured usage pane from stdin as JSON", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			now := time.Now()

			if captureAt != "" {
				parsed, err := time.Parse(time.RFC3339, captureAt)
				if err != nil {
					return errCapacityTime
				}

				now = parsed
			}

			pane, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return errCapacityRead
			}

			readings, err := capacity.Parse(source, string(pane), now)
			if err != nil {
				return fmt.Errorf("parse capacity: %w", err)
			}

			encodeErr := json.NewEncoder(cmd.OutOrStdout()).Encode(readings)
			if encodeErr != nil {
				return errCapacityWrite
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&source, "source", "", "Source: claude_usage, codex_status, agy_usage")
	cmd.Flags().StringVar(&captureAt, "now", "", "Capture time (RFC3339; defaults to current time)")

	return cmd
}
