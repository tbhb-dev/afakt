// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package server holds the server command, which checks its flags and runs the selected
// targets until a signal stops them.
package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/tbhb-dev/afakt/internal/cmdutil"
	"github.com/tbhb-dev/afakt/internal/server"
)

const (
	defaultListen          = "127.0.0.1:8080"
	defaultShutdownTimeout = "30s"

	targetFlag  = "target"
	listenFlag  = "listen"
	timeoutFlag = "shutdown-timeout"
)

var (
	errListenForm     = errors.New("expected HOST:PORT, such as " + defaultListen)
	errListenHost     = errors.New("the host is missing; give one, such as " + defaultListen)
	errListenPort     = errors.New("the port must be a number from 0 to 65535")
	errDurationForm   = errors.New("expected a Go duration, such as 30s or 1m30s")
	errDurationNotPos = errors.New("the duration must be positive")
)

// Options holds the checked configuration and the stream for the logs.
type Options struct {
	ErrOut io.Writer
	Config server.Config
}

// NewCmdServer returns the server command. It checks every flag before `runF` runs, so a bad
// command line never binds a listener. A nil `runF` runs `serverRun`.
func NewCmdServer(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{ErrOut: f.ErrOut}
	var target, listen, timeout string
	cmd := &cobra.Command{
		Use:   "server [OPTIONS] --target <TARGET>",
		Short: "Run the selected targets until a signal stops them",
		Args:  cmdutil.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			cfg, err := serverConfig(c, target, listen, timeout)
			if err != nil {
				return err
			}
			opts.Config = cfg
			if runF != nil {
				return runF(opts)
			}
			return serverRun(c.Context(), opts)
		},
	}
	cmdutil.StringFlag(cmd, &target, targetFlag, "<TARGET>", "",
		"Comma-separated targets: api, webhook, dashboard, worker, or all; only api can run")
	cmdutil.StringFlag(cmd, &listen, listenFlag, "<ADDR>", defaultListen, "Host and port the API listens on")
	cmdutil.StringFlag(cmd, &timeout, timeoutFlag, "<DURATION>", defaultShutdownTimeout,
		"Time limit for the whole shutdown, as a Go duration")
	cmdutil.AddHelpFlag(cmd)
	return cmd
}

// serverConfig checks the flag values in the order clap reports them.
func serverConfig(cmd *cobra.Command, target, listen, timeoutText string) (server.Config, error) {
	if !cmd.Flags().Changed(targetFlag) {
		return server.Config{}, cmdutil.MissingFlag(cmd, targetFlag)
	}
	targets, err := server.ParseTargets(target)
	if err != nil {
		return server.Config{}, cmdutil.InvalidValue(cmd, targetFlag, target, err)
	}
	if err = checkListen(listen); err != nil {
		return server.Config{}, cmdutil.InvalidValue(cmd, listenFlag, listen, err)
	}
	timeout, err := parseTimeout(timeoutText)
	if err != nil {
		return server.Config{}, cmdutil.InvalidValue(cmd, timeoutFlag, timeoutText, err)
	}
	return server.Config{Targets: targets, Listen: listen, ShutdownTimeout: timeout}, nil
}

func checkListen(address string) error {
	// A host name passes unresolved, because binding resolves it later.
	host, port, err := net.SplitHostPort(address)
	switch {
	case err != nil:
		return errListenForm
	case host == "":
		return errListenHost
	}
	if _, err = strconv.ParseUint(port, 10, 16); err != nil {
		return errListenPort
	}
	return nil
}

func parseTimeout(text string) (time.Duration, error) {
	timeout, err := time.ParseDuration(text)
	switch {
	case err != nil:
		return 0, errDurationForm
	case timeout <= 0:
		return 0, errDurationNotPos
	}
	return timeout, nil
}

// serverRun runs the server until SIGINT, SIGTERM, or the end of the context, with logs on
// the error stream. Only this command handles those signals, so `status` keeps the default
// action, which ends the process.
func serverRun(ctx context.Context, opts *Options) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger := slog.New(slog.NewTextHandler(opts.ErrOut, nil))
	//nolint:wrapcheck // The root prints the runtime error as it is, after Error:.
	return server.Run(ctx, opts.Config, logger)
}
