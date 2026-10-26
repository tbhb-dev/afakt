// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package temporaltest provides a stand-in Temporal frontend for tests of the external mode.
// It answers the calls the client service makes, for the namespaces it's given, on a
// listener the test supplies.
package temporaltest

import (
	"context"
	"fmt"
	"net"
	"slices"
	"testing"

	"go.temporal.io/api/namespace/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/grpc"
)

// Frontend is a stand-in Temporal frontend. It knows the namespaces that Serve names and
// rejects every other one.
type Frontend struct {
	workflowservice.UnimplementedWorkflowServiceServer

	namespaces []string
	server     *grpc.Server
	address    string
}

// Serve serves a frontend on ln for namespaces until the test ends. The test fails if serving
// ends with an error rather than through Stop.
func Serve(tb testing.TB, ln net.Listener, namespaces ...string) *Frontend {
	tb.Helper()
	f := &Frontend{namespaces: namespaces, server: grpc.NewServer(), address: ln.Addr().String()}
	workflowservice.RegisterWorkflowServiceServer(f.server, f)
	served := make(chan error, 1)
	go func() { served <- f.server.Serve(ln) }()
	tb.Cleanup(func() {
		f.Stop()
		if err := <-served; err != nil {
			tb.Errorf("serving the stand-in frontend: %v", err)
		}
	})
	return f
}

// Address returns the host and port the frontend serves on.
func (f *Frontend) Address() string {
	return f.address
}

// Stop closes the frontend and every connection to it, as a lost Temporal service would.
func (f *Frontend) Stop() {
	f.server.Stop()
}

// GetSystemInfo answers the call a client makes when it connects.
func (*Frontend) GetSystemInfo(
	context.Context, *workflowservice.GetSystemInfoRequest,
) (*workflowservice.GetSystemInfoResponse, error) {
	return &workflowservice.GetSystemInfoResponse{}, nil
}

// DescribeNamespace describes a namespace the frontend knows and rejects any other one as not
// found.
func (f *Frontend) DescribeNamespace(
	_ context.Context, req *workflowservice.DescribeNamespaceRequest,
) (*workflowservice.DescribeNamespaceResponse, error) {
	if !slices.Contains(f.namespaces, req.GetNamespace()) {
		return nil, fmt.Errorf(
			"describe namespace: %w",
			serviceerror.ToStatus(serviceerror.NewNamespaceNotFound(req.GetNamespace())).Err(),
		)
	}
	return &workflowservice.DescribeNamespaceResponse{
		NamespaceInfo: &namespace.NamespaceInfo{Name: req.GetNamespace()},
	}, nil
}
