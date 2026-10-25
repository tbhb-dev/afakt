// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package temporal_test

import (
	"bytes"
	"context"
	"log/slog"
	"net"
	"strings"
	"testing"

	"github.com/tbhb-dev/afakt/internal/temporal"
	"github.com/tbhb-dev/afakt/internal/temporal/temporaltest"
)

// listen returns a listener on a free loopback port.
func listen(t *testing.T) net.Listener {
	t.Helper()
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	return ln
}

// newClient returns a client service for namespace at address that logs into the buffer.
func newClient(address, namespace string, buf *bytes.Buffer) *temporal.Client {
	return temporal.NewClient(func() string { return address }, namespace, slog.New(slog.NewTextHandler(buf, nil)))
}

func TestClientConnects(t *testing.T) {
	t.Parallel()
	for _, namespace := range []string{temporal.DefaultNamespace, "orders"} {
		t.Run(namespace, func(t *testing.T) {
			t.Parallel()
			connect(t, namespace)
		})
	}
}

// connect runs a client service for namespace through its life against a stand-in frontend.
// Its check passes while the frontend serves and fails once the frontend has stopped.
func connect(t *testing.T, namespace string) {
	t.Helper()
	frontend := temporaltest.Serve(t, listen(t), namespace)
	var buf bytes.Buffer
	c := newClient(frontend.Address(), namespace, &buf)
	if err := c.Start(t.Context()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	want := "msg=\"temporal connected\" address=" + frontend.Address() + " namespace=" + namespace
	if !strings.Contains(buf.String(), want) {
		t.Errorf("log = %q, want it to contain %q", buf.String(), want)
	}
	if err := c.Check(t.Context()); err != nil {
		t.Errorf("Check() error = %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := c.Run(ctx); err != nil {
		t.Errorf("Run() error = %v", err)
	}
	frontend.Stop()
	if err := c.Check(t.Context()); err == nil {
		t.Error("Check() after the frontend stopped error = nil, want an error")
	}
	if err := c.Stop(t.Context()); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestClientRejects(t *testing.T) {
	t.Parallel()
	refused := listen(t)
	address := refused.Addr().String()
	if err := refused.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	frontend := temporaltest.Serve(t, listen(t), temporal.DefaultNamespace)
	for _, tc := range []struct {
		name, address, namespace, want string
	}{
		{"unreachable", address, temporal.DefaultNamespace, "connect to Temporal at " + address + ": "},
		{"unknown namespace", frontend.Address(), "orders", `describe Temporal namespace "orders": `},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := newClient(tc.address, tc.namespace, &buf).Start(t.Context())
			if err == nil || !strings.HasPrefix(err.Error(), tc.want) {
				t.Errorf("Start() error = %v, want prefix %q", err, tc.want)
			}
		})
	}
}
