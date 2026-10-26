// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package embedded runs a Temporal server inside the process, for development and test. The
// server keeps its data in SQLite, in memory or in a file in its own directory, and binds
// its own loopback ports.
package embedded

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/membership/static"
	"go.temporal.io/server/common/primitives"
	"go.temporal.io/server/temporal"

	sqliteplugin "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
	sqliteschema "go.temporal.io/server/schema/sqlite"
)

const (
	// DatabaseFile is the name of the SQLite file inside the data directory.
	DatabaseFile = "temporal.db"

	// clusterName names the server's single cluster.
	clusterName = "active"
	// storeName names the server's single data store.
	storeName = "sqlite"
	// loopback is the host the server binds every port on.
	loopback = "127.0.0.1"
	// archivalDisabled turns off an archival setting.
	archivalDisabled = "disabled"
	// maxJoinDuration bounds the wait for the services to join the cluster.
	maxJoinDuration = 30 * time.Second
	// failoverVersionIncrement is the step between the versions of a cluster. The server's one
	// cluster never hands over to another, and any value greater than the initial version works.
	failoverVersionIncrement = 10
)

var errLifecyclePanic = errors.New("embedded Temporal lifecycle panic")

// runtimeServer exposes context-aware startup and cleanup owned by the adapter.
type runtimeServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// newServerFunc returns the owner even on construction failure, so cleanup can release
// resources created before the error.
type newServerFunc func(context.Context, []net.Listener, ...temporal.ServerOption) (runtimeServer, error)

// Server is the service that runs the embedded Temporal server. Start builds and starts it,
// Address returns the address of its frontend, and Stop stops it.
type Server struct {
	dir              string
	namespace        string
	logger           *slog.Logger
	host             string
	newServer        newServerFunc
	setupSchema      func(*config.SQL) error
	createNamespaces func(*config.SQL, ...*sqliteschema.NamespaceConfig) error
	server           runtimeServer
	address          string
	listeners        []net.Listener
	listen           func(context.Context, string, string) (net.Listener, error)
	shutdownTimeout  time.Duration
}

// New returns a server for namespace. With an empty dir it keeps its data in memory, and
// otherwise in DatabaseFile inside dir, which Start creates if it's missing. Cancellation
// during Start allows shutdownTimeout for cleanup before returning a StartupError.
func New(dir, namespace string, logger *slog.Logger, shutdownTimeout time.Duration) *Server {
	return &Server{
		dir: dir, namespace: namespace, logger: logger, host: loopback,
		newServer: newOwnedServer, setupSchema: sqliteschema.SetupSchema,
		createNamespaces: sqliteschema.CreateNamespaces,
		listen:           (&net.ListenConfig{}).Listen,
		shutdownTimeout:  shutdownTimeout,
	}
}

// Address returns the host and port of the frontend. Call it only after Start succeeded.
func (s *Server) Address() string {
	return s.address
}

// Start prepares the database and the namespace, then starts the server. A failure after the
// server is built stops it again before Start returns.
func (s *Server) Start(ctx context.Context) error {
	// Temporal constructors and some hooks cannot be interrupted through the public API.
	// Keep their goroutine tracked through cleanup, including resources returned late.
	finished := make(chan error, 1)
	started := make(chan error, 1)
	cleanupDeadline := make(chan time.Time, 1)
	go func() {
		err := guarded(func() error { return s.start(ctx) })
		started <- err
		deadline := <-cleanupDeadline
		if deadline.IsZero() {
			return
		}
		stopCtx, cancel := context.WithDeadline(context.WithoutCancel(ctx), deadline)
		defer cancel()
		finished <- errors.Join(err, guarded(func() error { return s.Stop(stopCtx) }))
	}()
	select {
	case err := <-started:
		if err == nil && ctx.Err() == nil {
			cleanupDeadline <- time.Time{}
			return nil
		}
		return s.awaitCleanup(ctx, err, started, finished, cleanupDeadline)
	case <-ctx.Done():
		return s.awaitCleanup(ctx, nil, started, finished, cleanupDeadline)
	}
}

// guarded retains the lifecycle manager's panic protection inside the supervision
// goroutine. Temporal's fatal logger panics, including during constructor or hook calls.
func guarded(call func() error) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = fmt.Errorf("%w: %v", errLifecyclePanic, failure)
		}
	}()
	return call()
}

// StartupError preserves cancellation and any construction or cleanup error.
type StartupError struct {
	Cause error
}

func (e *StartupError) Error() string {
	return "embedded Temporal startup interrupted: " + e.Cause.Error()
}
func (e *StartupError) Unwrap() error { return e.Cause }

func (s *Server) awaitCleanup(
	ctx context.Context,
	cause error,
	started <-chan error,
	finished <-chan error,
	cleanupDeadline chan<- time.Time,
) error {
	deadline := time.Now().Add(s.shutdownTimeout)
	stopCtx, cancel := context.WithDeadline(context.WithoutCancel(ctx), deadline)
	defer cancel()
	cleanupDeadline <- deadline
	var cleanupErr error
	select {
	case err := <-finished:
		cleanupErr = err
	case <-stopCtx.Done():
	}
	if stopCtx.Err() != nil {
		// Cancellation may win the initial select even though startup returned an
		// error before cleanup stalled. Retain that error when it is now available.
		select {
		case err := <-started:
			cause = errors.Join(cause, err)
		default:
		}
		return s.cleanupTimeout(stopCtx, errors.Join(cause, ctx.Err(), cleanupErr))
	}
	if ctx.Err() != nil {
		return &StartupError{Cause: errors.Join(ctx.Err(), cleanupErr)}
	}
	return cleanupErr
}

func (s *Server) cleanupTimeout(ctx context.Context, cause error) error {
	s.logger.ErrorContext(ctx, "embedded Temporal startup cleanup exceeded shutdown deadline")
	return &StartupError{Cause: errors.Join(cause, ctx.Err())}
}

func (s *Server) start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("prepare embedded Temporal server: %w", err)
	}
	sqlConfig, err := s.database()
	if err != nil {
		return err
	}
	names := serviceNames()
	listeners, err := s.listenAll(ctx, len(names)+1)
	if err != nil {
		return err
	}
	s.listeners = listeners
	ports := make([]int, len(listeners))
	for i, ln := range listeners {
		// ListenConfig with network tcp always returns a TCP address.
		if addr, ok := ln.Addr().(*net.TCPAddr); ok {
			ports[i] = addr.Port
		}
	}
	conf := serverConfig(sqlConfig, s.host, ports)
	hosts := make(map[primitives.ServiceName]static.Hosts, len(names))
	for i, name := range names {
		hosts[name] = static.SingleLocalHost(net.JoinHostPort(s.host, strconv.Itoa(ports[i])))
	}
	server, err := s.newServer(ctx, listeners,
		temporal.WithConfig(conf),
		temporal.ForServices(temporal.DefaultServices),
		temporal.WithStaticHosts(hosts),
		temporal.WithLogger(newLogger(s.logger.With("component", "temporal-server"))),
	)
	s.server = server
	if err != nil {
		return fmt.Errorf("build embedded Temporal server: %w", err)
	}
	if err = ctx.Err(); err != nil {
		return fmt.Errorf("construct embedded Temporal server: %w", err)
	}
	if err = server.Start(ctx); err != nil {
		return fmt.Errorf("start embedded Temporal server: %w", err)
	}
	s.address = conf.PublicClient.HostPort
	s.logger.InfoContext(ctx, "temporal embedded server started", "address", s.address, "namespace", s.namespace,
		"dir", s.dir)
	return nil
}

// Run has no work of its own. It waits until the manager cancels it.
func (s *Server) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Stop stops the server. The manager stops waiting for it at the shutdown deadline.
func (s *Server) Stop(ctx context.Context) error {
	defer closeListeners(s.listeners)
	if s.server == nil {
		return nil
	}
	if err := s.server.Stop(ctx); err != nil {
		return fmt.Errorf("stop embedded Temporal server: %w", err)
	}
	return nil
}

// database returns the SQL configuration of the server's store, with the schema and the
// namespace in place. The SQLite plugin sets up the schema of an in-memory database itself.
// A file database gets its schema only when the file is new, and a restart reuses the file
// with its data.
func (s *Server) database() (*config.SQL, error) {
	conf := &config.SQL{PluginName: sqliteplugin.PluginName}
	fresh := false
	if s.dir == "" {
		// A random name keeps two servers in one process from sharing an in-memory database.
		conf.ConnectAttributes = map[string]string{"mode": "memory", "cache": "shared"}
		conf.DatabaseName = "afakt-temporal-" + rand.Text()
	} else {
		if err := os.MkdirAll(s.dir, 0o700); err != nil {
			return nil, fmt.Errorf("create embedded Temporal directory: %w", err)
		}
		conf.ConnectAttributes = map[string]string{"mode": "rwc"}
		conf.DatabaseName = filepath.Join(s.dir, DatabaseFile)
		_, err := os.Lstat(conf.DatabaseName)
		fresh = errors.Is(err, os.ErrNotExist)
	}
	if fresh {
		if err := s.setupSchema(conf); err != nil {
			return nil, fmt.Errorf("set up embedded Temporal schema: %w", err)
		}
	}
	// NewNamespaceConfig fails only for custom search attributes, which this server has none of.
	namespace, err := sqliteschema.NewNamespaceConfig(clusterName, s.namespace, false, nil)
	if err == nil {
		// The pinned SQLite plugin's public Close retains its private pool connection.
		// Its final release remains tracked in https://github.com/tbhb-dev/afakt/issues/73.
		err = s.createNamespaces(conf, namespace)
	}
	if err != nil {
		return nil, fmt.Errorf("create embedded Temporal namespace %q: %w", s.namespace, err)
	}
	return conf, nil
}

// serviceNames lists the Temporal services the server runs, in the order of their ports in
// `serverConfig`.
func serviceNames() []primitives.ServiceName {
	return []primitives.ServiceName{
		primitives.FrontendService,
		primitives.HistoryService,
		primitives.MatchingService,
		primitives.WorkerService,
	}
}

// serverConfig returns the server's configuration. The `ports` slice lists one gRPC port for
// each of `serviceNames`, then the HTTP port of the frontend. The configuration has no
// metrics section. Without one, the server doesn't report metrics or register collectors.
func serverConfig(sqlConfig *config.SQL, host string, ports []int) *config.Config {
	var conf config.Config
	names := serviceNames()
	frontend := net.JoinHostPort(host, strconv.Itoa(ports[0]))
	frontendHTTP := net.JoinHostPort(host, strconv.Itoa(ports[len(names)]))
	conf.Global.Membership.MaxJoinDuration = maxJoinDuration
	conf.Global.Membership.BroadcastAddress = host
	conf.Persistence.DefaultStore = storeName
	conf.Persistence.VisibilityStore = storeName
	conf.Persistence.NumHistoryShards = 1
	conf.Persistence.DataStores = map[string]config.DataStore{storeName: {SQL: sqlConfig}}
	conf.ClusterMetadata = &cluster.Config{
		FailoverVersionIncrement: failoverVersionIncrement,
		MasterClusterName:        clusterName,
		CurrentClusterName:       clusterName,
		ClusterInformation: map[string]cluster.ClusterInformation{clusterName: {
			Enabled:                true,
			InitialFailoverVersion: 1,
			RPCAddress:             frontend,
			HTTPAddress:            frontendHTTP,
		}},
	}
	conf.DCRedirectionPolicy.Policy = "noop"
	conf.Services = make(map[string]config.Service, len(names))
	for i, name := range names {
		var service config.Service
		service.RPC.BindOnIP = host
		service.RPC.GRPCPort = ports[i]
		if name == primitives.FrontendService {
			service.RPC.HTTPPort = ports[len(names)]
		}
		conf.Services[string(name)] = service
	}
	conf.Archival.History.State = archivalDisabled
	conf.Archival.Visibility.State = archivalDisabled
	conf.NamespaceDefaults.Archival.History.State = archivalDisabled
	conf.NamespaceDefaults.Archival.Visibility.State = archivalDisabled
	conf.PublicClient.HostPort = frontend
	return &conf
}

// listenAll acquires listeners before constructing any Temporal service. A bind failure
// retains its original error and releases the listeners already acquired.
func (s *Server) listenAll(ctx context.Context, count int) ([]net.Listener, error) {
	listeners := make([]net.Listener, 0, count)
	for range count {
		ln, err := s.listen(ctx, "tcp", net.JoinHostPort(s.host, "0"))
		if err != nil {
			closeListeners(listeners)
			return nil, fmt.Errorf("pick an embedded Temporal port: %w", err)
		}
		listeners = append(listeners, ln)
	}
	return listeners, nil
}

// closeListeners releases sockets even when construction failed before Temporal could own them.
func closeListeners(listeners []net.Listener) {
	for _, ln := range slices.Backward(listeners) {
		_ = ln.Close()
	}
}
