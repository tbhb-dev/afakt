// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package embedded

import (
	"fmt"
	"net"

	"go.temporal.io/server/chasm"
	"go.temporal.io/server/common"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/persistence/serialization"
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

// newOwnedServer follows temporal.TopLevelModule in the pinned v1.32.1, replacing only
// its service providers so their RPC factories receive listeners acquired by the adapter.
// Keep this wiring aligned with that module when upgrading the server dependency.
func newOwnedServer(listeners []net.Listener, opts ...temporal.ServerOption) (temporal.Server, error) {
	module := fx.Options(
		fx.Provide(
			temporal.NewServerFxImpl,
			temporal.ServerOptionsProvider,
			resource.ArchivalMetadataProvider,
			temporal.TaskCategoryRegistryProvider,
			temporal.ApplyClusterMetadataConfigProvider,
			func(params temporal.ServiceProviderParamsCommon) (ownedServices, error) {
				return buildServices(params, listeners)
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
	server, err := temporal.NewServerFx(module, opts...)
	if err != nil {
		return nil, fmt.Errorf("construct owned-listener services: %w", err)
	}
	return server, nil
}

// ownedServices supplies the same service group consumed by Temporal's server lifecycle.
type ownedServices struct {
	fx.Out

	Services []*temporal.ServicesMetadata `group:"services,flatten"`
}

// buildServices follows the pinned service providers. The adapter always runs the default
// services with static membership and the frontend's configured claim mapper and logger.
func buildServices(params temporal.ServiceProviderParamsCommon, listeners []net.Listener) (ownedServices, error) {
	modules := []fx.Option{
		fx.Options(frontend.Module, fx.Supply(params.CustomFrontendInterceptors)),
		fx.Options(history.QueueModule, history.Module, replication.Module),
		matching.Module,
		worker.Module,
	}
	var services ownedServices
	for i, name := range serviceNames() {
		app := fx.New(
			params.GetCommonServiceOptions(name),
			fx.Decorate(func(factory common.RPCFactory) common.RPCFactory {
				return &listenerFactory{RPCFactory: factory, listener: listeners[i]}
			}),
			modules[i],
		)
		if err := app.Err(); err != nil {
			return services, fmt.Errorf("construct %s service: %w", name, err)
		}
		services.Services = append(services.Services, temporal.NewService(app, name, params.Logger).Services)
	}
	return services, nil
}
