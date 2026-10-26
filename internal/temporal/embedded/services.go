// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package embedded

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"sync"

	"go.temporal.io/server/chasm"
	"go.temporal.io/server/common"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/persistence/serialization"
	"go.temporal.io/server/common/persistence/visibility/manager"
	"go.temporal.io/server/common/pprof"
	"go.temporal.io/server/common/resource"
	"go.temporal.io/server/service/frontend"
	"go.temporal.io/server/service/history"
	"go.temporal.io/server/service/history/replication"
	"go.temporal.io/server/service/matching"
	"go.temporal.io/server/service/worker"
	"go.temporal.io/server/temporal"
	"go.uber.org/fx"
)

// listenerFactory delegates connections and TLS to Temporal but never asks its factory to
// bind. Temporal v1.32.1's GetGRPCListener fatally logs a bind failure instead of returning it.
type listenerFactory struct {
	common.RPCFactory

	listener net.Listener
}

func (f *listenerFactory) GetGRPCListener() net.Listener { return f.listener }

// httpHandoff retains the HTTP reservation until the pinned frontend HTTP constructor
// asks for its bind address. In v1.32.1 that is the first Addr call on this listener.
// The public constructor then binds privately; the remaining close-and-rebind window
// requires an upstream listener handoff: https://github.com/tbhb-dev/afakt/issues/73.
type httpHandoff struct {
	net.Listener

	reservation net.Listener
	once        sync.Once
}

func (h *httpHandoff) Addr() net.Addr {
	h.once.Do(func() { _ = h.reservation.Close() })
	return h.Listener.Addr()
}

// servingListener acknowledges that gRPC entered its accept loop before cleanup can
// stop it. Otherwise a late Serve sees an already stopped server and Temporal logs fatal.
type servingListener struct {
	net.Listener

	accepted chan struct{}
	once     sync.Once
}

func (l *servingListener) Accept() (net.Conn, error) {
	l.once.Do(func() { close(l.accepted) })
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("accept embedded Temporal connection: %w", err)
	}
	return conn, nil
}

// ownedApp records construction-time stop hooks too: Fx only stops hooks reached by
// Start, so app.Stop alone cannot roll back a failed fx.New. The adapter runs the recorded
// hooks itself and app.Stop releases the Fx app after their cleanup.
type ownedApp struct {
	app         *fx.App
	cleanup     []func(context.Context) error
	hooks       []fx.Hook
	running     bool
	serviceHook bool
	listener    *servingListener
}

type ownedLifecycle struct {
	owner *ownedApp
}

func (l *ownedLifecycle) Append(hook fx.Hook) {
	if l.owner.serviceHook {
		// In the pinned module, ServiceLifetimeHooks appends the service hook
		// immediately after resolving the service. Record its start and wait for the
		// serving goroutine before cleanup can stop its gRPC server.
		start := hook.OnStart
		hook.OnStart = func(ctx context.Context) error {
			// Service.Start has no error result in the pinned version.
			err := start(ctx)
			l.owner.running = true
			<-l.owner.listener.accepted
			return err
		}
		l.owner.serviceHook = false
	}
	if hook.OnStart != nil {
		l.owner.hooks = append(l.owner.hooks, hook)
	} else if hook.OnStop != nil {
		l.owner.cleanup = append(l.owner.cleanup, hook.OnStop)
	}
}

func (a *ownedApp) options() fx.Option {
	return fx.Options(
		fx.Decorate(func(fx.Lifecycle) fx.Lifecycle { return &ownedLifecycle{owner: a} }),
	)
}

func (a *ownedApp) stop(ctx context.Context) error {
	var err error
	for _, cleanup := range slices.Backward(a.cleanup) {
		err = errors.Join(err, guarded(func() error { return cleanup(ctx) }))
	}
	return errors.Join(err, a.app.Stop(ctx))
}

func (a *ownedApp) start(ctx context.Context) error {
	for _, hook := range a.hooks {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("start owned hook: %w", err)
		}
		if hook.OnStop != nil {
			// A failed or panicking start can still acquire resources that need cleanup.
			a.cleanup = append(a.cleanup, hook.OnStop)
		}
		if err := hook.OnStart(ctx); err != nil {
			return fmt.Errorf("start owned app: %w", err)
		}
	}
	return nil
}

type ownedServer struct {
	top      *ownedApp
	services []*ownedApp
	newApp   func(...fx.Option) *fx.App
}

func (s *ownedServer) Start(ctx context.Context) error {
	if err := s.top.start(ctx); err != nil {
		return fmt.Errorf("start owned services: %w", err)
	}
	for _, service := range s.services {
		if err := service.start(ctx); err != nil {
			return fmt.Errorf("start owned service: %w", err)
		}
	}
	return nil
}

func (s *ownedServer) Stop(ctx context.Context) error {
	var err error
	for _, service := range slices.Backward(s.services) {
		err = errors.Join(err, service.stop(ctx))
	}
	return errors.Join(err, s.top.stop(ctx))
}

type ownedVisibility struct {
	manager.VisibilityManager

	closeOnce sync.Once
}

func (v *ownedVisibility) Close() { v.closeOnce.Do(v.VisibilityManager.Close) }

// newOwnedServer follows temporal.TopLevelModule in the pinned v1.32.1, replacing only
// its service providers so their RPC factories receive listeners acquired by the adapter.
// Keep this wiring aligned with that module when upgrading the server dependency.
func newOwnedServer(
	ctx context.Context,
	listeners []net.Listener,
	opts ...temporal.ServerOption,
) (runtimeServer, error) {
	return constructOwnedServer(ctx, listeners, fx.New, opts...)
}

func constructOwnedServer(
	ctx context.Context,
	listeners []net.Listener,
	newApp func(...fx.Option) *fx.App,
	opts ...temporal.ServerOption,
) (runtimeServer, error) {
	s := &ownedServer{top: &ownedApp{}, newApp: newApp}
	module := fx.Options(
		fx.Provide(
			temporal.NewServerFxImpl,
			temporal.ServerOptionsProvider,
			resource.ArchivalMetadataProvider,
			temporal.TaskCategoryRegistryProvider,
			temporal.ApplyClusterMetadataConfigProvider,
			func(params temporal.ServiceProviderParamsCommon) (ownedServices, error) {
				return s.buildServices(ctx, params, listeners)
			},
		),
		dynamicconfig.Module,
		pprof.Module,
		temporal.TraceExportModule,
		chasm.Module,
		serialization.Module,
		temporal.FxLogAdapter,
		fx.Invoke(temporal.ServerLifetimeHooks),
	)
	s.top.app = newApp(module, fx.Supply(opts), s.top.options(), fx.RecoverFromPanics())
	if err := s.top.app.Err(); err != nil {
		return s, fmt.Errorf("construct owned-listener services: %w", err)
	}
	return s, nil
}

// ownedServices supplies the same service group consumed by Temporal's server lifecycle.
type ownedServices struct {
	fx.Out

	Services []*temporal.ServicesMetadata `group:"services,flatten"`
}

// buildServices follows the pinned service providers. Services are constructed and started
// in dependency order, then stopped in reverse. The top-level ServerImpl receives an empty
// group: it initializes system namespaces, while this adapter starts the apps with ctx.
func (s *ownedServer) buildServices(
	ctx context.Context,
	params temporal.ServiceProviderParamsCommon,
	listeners []net.Listener,
) (ownedServices, error) {
	modules := []fx.Option{
		fx.Options(frontend.Module, fx.Supply(params.CustomFrontendInterceptors)),
		fx.Options(history.QueueModule, history.Module, replication.Module),
		matching.Module,
		worker.Module,
	}
	var services ownedServices
	for _, i := range []int{2, 1, 0, 3} {
		if err := ctx.Err(); err != nil {
			return services, fmt.Errorf("construct owned service: %w", err)
		}
		name := serviceNames()[i]
		owner := &ownedApp{}
		s.services = append(s.services, owner)
		owner.listener = &servingListener{Listener: listeners[i], accepted: make(chan struct{})}
		var ln net.Listener = owner.listener
		if i == 0 {
			ln = &httpHandoff{Listener: ln, reservation: listeners[len(serviceNames())]}
		}
		owner.app = s.newApp(
			fx.RecoverFromPanics(),
			params.GetCommonServiceOptions(name),
			fx.Decorate(func(factory common.RPCFactory) common.RPCFactory {
				return &listenerFactory{RPCFactory: factory, listener: ln}
			}),
			fx.Decorate(func(v manager.VisibilityManager) manager.VisibilityManager {
				owned := &ownedVisibility{VisibilityManager: v}
				owner.cleanup = append(owner.cleanup, func(context.Context) error { owned.Close(); return nil })
				return owned
			}),
			modules[i],
			owner.options(),
			fx.Decorate(func(svc *frontend.Service) *frontend.Service {
				owner.serviceHook = true
				return svc
			}),
			fx.Decorate(func(svc *matching.Service) *matching.Service {
				owner.serviceHook = true
				return svc
			}),
			fx.Decorate(func(svc *history.Service) *history.Service {
				owner.serviceHook = true
				return svc
			}),
			fx.Decorate(func(svc *worker.Service) *worker.Service {
				owner.serviceHook = true
				return svc
			}),
			fx.Decorate(func(h *frontend.HTTPAPIServer) *frontend.HTTPAPIServer {
				owner.cleanup = append(owner.cleanup, func(context.Context) error {
					if !owner.running {
						// Shutdown before Serve marks net/http closed. Serve then closes
						// the private listener without accepting any requests.
						h.GracefulStop(0)
						return h.Serve()
					}
					return nil
				})
				return h
			}),
		)
		if err := owner.app.Err(); err != nil {
			return services, fmt.Errorf("construct %s service: %w", name, err)
		}
	}
	return services, nil
}
