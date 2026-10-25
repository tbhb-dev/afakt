// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/tbhb-dev/afakt/internal/buildmeta"
	"github.com/tbhb-dev/afakt/internal/cmd/root"
	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

func main() {
	// Ignoring the broken pipe signal turns a write to a closed pipe into an error and exit code 1.
	signal.Ignore(syscall.SIGPIPE)
	f := &cmdutil.Factory{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr, Version: buildmeta.Get().Version}
	os.Exit(root.Execute(context.Background(), f, os.Args[1:]))
}
