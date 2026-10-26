// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package temporaltest_test

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/tbhb-dev/afakt/internal/temporal/temporaltest"
)

func TestFrontend(t *testing.T) {
	t.Parallel()
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	f := temporaltest.Serve(t, ln, "known")
	conn, err := grpc.NewClient(f.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	svc := workflowservice.NewWorkflowServiceClient(conn)
	if _, err = svc.GetSystemInfo(t.Context(), &workflowservice.GetSystemInfoRequest{}); err != nil {
		t.Errorf("GetSystemInfo() error = %v", err)
	}
	resp, err := svc.DescribeNamespace(t.Context(), &workflowservice.DescribeNamespaceRequest{Namespace: "known"})
	if err != nil || resp.GetNamespaceInfo().GetName() != "known" {
		t.Errorf("DescribeNamespace(known) = %v, %v; want known", resp, err)
	}
	_, err = svc.DescribeNamespace(t.Context(), &workflowservice.DescribeNamespaceRequest{Namespace: "other"})
	notFound, ok := errors.AsType[*serviceerror.NamespaceNotFound](serviceerror.FromStatus(status.Convert(err)))
	if !ok || notFound.Namespace != "other" {
		t.Errorf("DescribeNamespace(other) error = %v, want namespace not found", err)
	}
	f.Stop()
	if _, err = svc.GetSystemInfo(t.Context(), &workflowservice.GetSystemInfoRequest{}); err == nil {
		t.Error("GetSystemInfo() after Stop error = nil, want an error")
	}
}

// cleanupTB records the cleanups and errors of a `Serve` call, which the test then runs and
// checks itself.
type cleanupTB struct {
	testing.TB

	cleanups []func()
	errors   []string
}

func (c *cleanupTB) Cleanup(f func()) {
	c.cleanups = append(c.cleanups, f)
}

func (c *cleanupTB) Errorf(format string, args ...any) {
	c.errors = append(c.errors, fmt.Sprintf(format, args...))
}

func TestFrontendReportsServeError(t *testing.T) {
	t.Parallel()
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	// A closed listener makes serving fail at once.
	if err = ln.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	tb := &cleanupTB{TB: t}
	temporaltest.Serve(tb, ln)
	for _, f := range tb.cleanups {
		f()
	}
	if len(tb.errors) != 1 || !strings.HasPrefix(tb.errors[0], "serving the stand-in frontend: ") {
		t.Errorf("errors = %q, want one serving error", tb.errors)
	}
}
