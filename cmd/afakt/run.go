// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"errors"
	"fmt"
	"io"
	"syscall"
)

const (
	ioExit    = 1
	usageExit = 2

	statusLine = "status is not implemented yet."
)

// Run prints help, the version or the status line and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	out := parseRoot(args)
	if out.status {
		return runStatus(stdout, stderr)
	}
	// Help and version output ignore write errors and exit 0, as clap does.
	_, _ = fmt.Fprint(stdout, out.stdout)
	_, _ = fmt.Fprint(stderr, out.stderr)
	return out.code
}

// runStatus writes the status line. A failed write prints the error and exits 1,
// except for a closed standard output, which counts as success.
func runStatus(stdout, stderr io.Writer) int {
	err := status(stdout)
	switch {
	case err == nil, errors.Is(err, syscall.EBADF):
		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", describeWriteError(err))
		return ioExit
	}
}

func status(output io.Writer) error {
	return writeLine(output, statusLine)
}

func writeLine(output io.Writer, line string) error {
	_, err := fmt.Fprintln(output, line)
	//nolint:wrapcheck // The caller reports the writer's own error unchanged.
	return err
}
