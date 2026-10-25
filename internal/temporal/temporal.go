// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package temporal connects the worker target to a Temporal service. Its client service
// checks that the namespace exists when it connects, then reports for readiness whether the
// namespace is still reachable.
package temporal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	sdklog "go.temporal.io/sdk/log"
)

// The modes of the worker target's Temporal service.
const (
	// ModeEmbedded runs a Temporal server inside the process.
	ModeEmbedded = "embedded"
	// ModeExternal connects to a Temporal service at an address.
	ModeExternal = "external"
)

// DefaultNamespace is the namespace the worker target uses unless the configuration names
// another.
const DefaultNamespace = "default"

// checkTimeout bounds one readiness check, so a frontend that stops answering makes the check
// fail rather than hang.
const checkTimeout = 2 * time.Second

// Config is a checked Temporal configuration. Address is set only in external mode, and
// EmbeddedDir only in embedded mode, where an empty value keeps the data in memory.
type Config struct {
	Mode        string
	Address     string
	Namespace   string
	EmbeddedDir string
}

// Client is the service that owns the connection to Temporal. Start dials the address and checks
// the namespace, Check repeats that check for readiness, and Stop closes the connection.
type Client struct {
	address   func() string
	namespace string
	logger    *slog.Logger
	client    client.Client
}

// NewClient returns a client service for namespace. Start calls address for the address of
// the frontend, so an embedded server that starts first can supply the address it bound.
func NewClient(address func() string, namespace string, logger *slog.Logger) *Client {
	return &Client{address: address, namespace: namespace, logger: logger}
}

// Start dials the frontend and checks that the namespace exists. On a failed check it closes
// the connection before it returns.
func (c *Client) Start(ctx context.Context) error {
	address := c.address()
	conn, err := client.DialContext(ctx, client.Options{
		HostPort:  address,
		Namespace: c.namespace,
		Logger:    sdklog.NewStructuredLogger(c.logger.With("component", "temporal-client")),
	})
	if err != nil {
		return fmt.Errorf("connect to Temporal at %s: %w", address, err)
	}
	if err = describe(ctx, conn, c.namespace); err != nil {
		conn.Close()
		return err
	}
	c.client = conn
	c.logger.InfoContext(ctx, "temporal connected", "address", address, "namespace", c.namespace)
	return nil
}

// Run has no work of its own. It waits until the manager cancels it.
func (c *Client) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Stop closes the connection.
func (c *Client) Stop(context.Context) error {
	c.client.Close()
	return nil
}

// Check reports whether the namespace is reachable now. Call it only after Start succeeded.
func (c *Client) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	return describe(ctx, c.client, c.namespace)
}

// describe asks the frontend to describe namespace, which fails unless the frontend answers
// and the namespace exists.
func describe(ctx context.Context, conn client.Client, namespace string) error {
	req := &workflowservice.DescribeNamespaceRequest{Namespace: namespace}
	if _, err := conn.WorkflowService().DescribeNamespace(ctx, req); err != nil {
		return fmt.Errorf("describe Temporal namespace %q: %w", namespace, err)
	}
	return nil
}
