// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Ignoring the broken pipe signal turns a write to a closed pipe into an error and exit code 1.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
