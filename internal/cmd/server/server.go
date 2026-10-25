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
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/tbhb-dev/afakt/internal/cmdutil"
	"github.com/tbhb-dev/afakt/internal/server"
	"github.com/tbhb-dev/afakt/internal/temporal"
)

const (
	defaultListen          = "127.0.0.1:8080"
	defaultShutdownTimeout = "30s"

	targetFlag  = "target"
	listenFlag  = "listen"
	timeoutFlag = "shutdown-timeout"

	temporalModeFlag      = "temporal-mode"
	temporalAddressFlag   = "temporal-address"
	temporalNamespaceFlag = "temporal-namespace"
	temporalDirFlag       = "temporal-embedded-dir"
)

var (
	errListenForm     = errors.New("expected HOST:PORT, such as " + defaultListen)
	errListenHost     = errors.New("the host is missing; give one, such as " + defaultListen)
	errListenPort     = errors.New("the port must be a number from 0 to 65535")
	errDurationForm   = errors.New("expected a Go duration, such as 30s or 1m30s")
	errDurationNotPos = errors.New("the duration must be positive")

	errTemporalMode = errors.New(
		"expected " + temporal.ModeEmbedded + " or " + temporal.ModeExternal,
	)
	errTemporalAddressForm = errors.New("expected HOST:PORT of the Temporal frontend, such as 127.0.0.1:7233")
	errTemporalAddressSet  = errors.New(
		"only external mode connects to an address; embedded mode runs its own server, so leave this flag out",
	)
	errTemporalDirSet = errors.New(
		"only embedded mode keeps data in a directory; external mode stores nothing, so leave this flag out",
	)
	errTemporalDirEmpty  = errors.New("the directory is empty; leave the flag out to keep the data in memory")
	errTemporalNamespace = errors.New("the namespace must not be empty")
	errTemporalNoWorker  = errors.New(
		"only the worker target uses Temporal; add worker to --target or leave this flag out",
	)
)

// flagValues holds the raw values of the server flags before `serverConfig` checks them.
type flagValues struct {
	target, listen, timeout                  string
	temporalMode, temporalAddress, namespace string
	temporalDir                              string
}

// Options holds the checked configuration and the stream for the logs.
type Options struct {
	ErrOut io.Writer
	Config server.Config
}

// NewCmdServer returns the server command. It checks every flag before `runF` runs, so a bad
// command line never binds a listener. A nil `runF` runs `serverRun`.
func NewCmdServer(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{ErrOut: f.ErrOut}
	var v flagValues
	cmd := &cobra.Command{
		Use:   "server [OPTIONS] --target <TARGET>",
		Short: "Run the selected targets until a signal stops them",
		Args:  cmdutil.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			cfg, err := serverConfig(c, v)
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
	cmdutil.StringFlag(
		cmd,
		&v.target,
		targetFlag,
		"<TARGET>",
		"",
		"Comma-separated targets: api, webhook, dashboard, worker, or all; api, webhook, dashboard and worker can run, and webhook has no receivers yet, only health endpoints",
	)
	cmdutil.StringFlag(cmd, &v.listen, listenFlag, "<ADDR>", defaultListen, "Host and port the HTTP server listens on")
	cmdutil.StringFlag(cmd, &v.timeout, timeoutFlag, "<DURATION>", defaultShutdownTimeout,
		"Time limit for the whole shutdown, as a Go duration")
	cmdutil.StringFlag(cmd, &v.temporalMode, temporalModeFlag, "<MODE>", "",
		"How the worker reaches Temporal: embedded runs a server in the process, external connects to one")
	cmdutil.StringFlag(cmd, &v.temporalAddress, temporalAddressFlag, "<ADDR>", "",
		"Host and port of the Temporal frontend, in external mode")
	cmdutil.StringFlag(cmd, &v.namespace, temporalNamespaceFlag, "<NAMESPACE>", temporal.DefaultNamespace,
		"Temporal namespace the worker uses")
	cmdutil.StringFlag(cmd, &v.temporalDir, temporalDirFlag, "<DIR>", "",
		"Directory for the data of the embedded server; without it, the data stays in memory")
	cmdutil.AddHelpFlag(cmd)
	return cmd
}

// serverConfig checks the flag values in the order clap reports them.
func serverConfig(cmd *cobra.Command, v flagValues) (server.Config, error) {
	if !cmd.Flags().Changed(targetFlag) {
		return server.Config{}, cmdutil.MissingFlag(cmd, targetFlag)
	}
	targets, err := server.ParseTargets(v.target)
	if err != nil {
		return server.Config{}, cmdutil.InvalidValue(cmd, targetFlag, v.target, err)
	}
	if err = checkListen(v.listen); err != nil {
		return server.Config{}, cmdutil.InvalidValue(cmd, listenFlag, v.listen, err)
	}
	timeout, err := parseTimeout(v.timeout)
	if err != nil {
		return server.Config{}, cmdutil.InvalidValue(cmd, timeoutFlag, v.timeout, err)
	}
	tc, err := temporalConfig(cmd, slices.Contains(targets, server.TargetWorker), v)
	if err != nil {
		return server.Config{}, err
	}
	return server.Config{Targets: targets, Listen: v.listen, ShutdownTimeout: timeout, Temporal: tc}, nil
}

// temporalConfig checks the Temporal flags. The worker target requires a mode, and each
// mode takes only its own flags. Without the worker target, every Temporal flag is an error.
func temporalConfig(cmd *cobra.Command, worker bool, v flagValues) (temporal.Config, error) {
	if !worker {
		return temporal.Config{}, rejectTemporalFlags(cmd)
	}
	if !cmd.Flags().Changed(temporalModeFlag) {
		return temporal.Config{}, cmdutil.MissingFlag(cmd, temporalModeFlag)
	}
	var err error
	switch v.temporalMode {
	case temporal.ModeExternal:
		err = checkExternal(cmd, v)
	case temporal.ModeEmbedded:
		err = checkEmbedded(cmd, v)
	default:
		err = invalidFlag(cmd, temporalModeFlag, errTemporalMode)
	}
	if err == nil && v.namespace == "" {
		err = invalidFlag(cmd, temporalNamespaceFlag, errTemporalNamespace)
	}
	if err != nil {
		return temporal.Config{}, err
	}
	return temporal.Config{
		Mode: v.temporalMode, Address: v.temporalAddress, Namespace: v.namespace, EmbeddedDir: v.temporalDir,
	}, nil
}

// rejectTemporalFlags reports the first Temporal flag the command line sets, for a run
// without the worker target.
func rejectTemporalFlags(cmd *cobra.Command) error {
	for _, name := range []string{temporalModeFlag, temporalAddressFlag, temporalNamespaceFlag, temporalDirFlag} {
		if cmd.Flags().Changed(name) {
			return invalidFlag(cmd, name, errTemporalNoWorker)
		}
	}
	return nil
}

// checkExternal checks the flags of external mode, which needs an address and no directory.
func checkExternal(cmd *cobra.Command, v flagValues) error {
	flags := cmd.Flags()
	switch {
	case !flags.Changed(temporalAddressFlag):
		return cmdutil.MissingFlag(cmd, temporalAddressFlag)
	case checkListen(v.temporalAddress) != nil:
		return invalidFlag(cmd, temporalAddressFlag, errTemporalAddressForm)
	case flags.Changed(temporalDirFlag):
		return invalidFlag(cmd, temporalDirFlag, errTemporalDirSet)
	}
	return nil
}

// checkEmbedded checks the flags of embedded mode, which takes no address and, if it's set,
// a directory that isn't empty.
func checkEmbedded(cmd *cobra.Command, v flagValues) error {
	flags := cmd.Flags()
	switch {
	case flags.Changed(temporalAddressFlag):
		return invalidFlag(cmd, temporalAddressFlag, errTemporalAddressSet)
	case flags.Changed(temporalDirFlag) && v.temporalDir == "":
		return invalidFlag(cmd, temporalDirFlag, errTemporalDirEmpty)
	}
	return nil
}

// invalidFlag returns the invalid value error for the flag called name, with its current
// value and reason.
func invalidFlag(cmd *cobra.Command, name string, reason error) error {
	return cmdutil.InvalidValue(cmd, name, cmd.Flags().Lookup(name).Value.String(), reason)
}

// checkListen checks that address has a host and a port from 0 to 65535. It returns one of
// the listen address errors, which become the reason in the invalid value message.
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

// parseTimeout parses a positive Go duration. It returns one of the duration errors,
// which become the reason in the invalid value message.
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
