---
type: Component Reference
title: Components
description: The parts of afakt and what each one does.
tags: [afakt, components]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Components

## ACP adapter

The ACP adapter is the harness's Agent Client Protocol (ACP) agent inside the sandbox. Claude Code uses `claude-agent-acp`, and Codex uses `codex-acp`. The [harness transport](agent-profile.md#harness-transport) selects ACP or headless for an execution. An interactive [agent session](agent-session.md) uses ACP.

## ACP gate

The ACP gate runs in the per-agent mediator. Cedar checks every `JSON-RPC` message in both directions with default deny, using the sending user for messages toward the agent and the agent for requests from it. Update notifications have size and rate bounds. Denials return errors and enter the agent's decision log. See [ACP gate](agent-session.md#acp-gate).

## ACP relay

The ACP relay is the [guest relay](#guest-relay) inside the sandbox VM. It acts as the ACP client to the adapter over standard input and output. For executions, it answers permission requests from profile grants and cancels elicitation requests. Its session channel ends at the mediator's [ACP gate](#acp-gate), which checks each message before forwarding. See [Executions over ACP](agent-session.md#executions-over-acp).

## afakt MCP server

Through the built-in afakt Model Context Protocol (MCP) server, an agent can start a sub-execution or session and message another session. It can also read execution status and results or declare a forwarded port. Each call uses the calling execution's capabilities and scopes, and profiles allowlist these tools like any upstream server. See [Built-in afakt MCP server](agent-mcp-gateway.md#built-in-afakt-mcp-server).

## Agentmail

Agentmail manages messages addressed to durable agents. These include incoming and outgoing mail. Delivery and signing semantics remain [open](agent-open-questions.md#agents). See [Agentmail, groups, and subscriptions](agent.md#agentmail-groups-and-subscriptions).

## API server

The API server serves afakt's REST API, which clients like the CLI and dashboard consume, as well as its coordination workflows. It keeps the [agent host registry](agent-placement.md#agent-host-registry) and serves the reservation API that places executions on agent hosts.

## Artifact proxy

The artifact proxy serves stored artifacts associated with agents or managed independently. Its access and serving contract remains [open](agent-open-questions.md#agents). See [Memory, skills, and artifacts](agent.md#memory-skills-and-artifacts).

## Capacity controller

The capacity controller runs in each agent host's afakt worker. It measures host CPU use, memory use, and free workspace disk space at intervals and adjusts the capacity for execution slots within the host's configured lower and upper bounds. The worker reports capacity to the agent host registry in each lease renewal. See [Slot capacity](agent-placement.md#slot-capacity).

### Health endpoints

The server answers `GET /livez`, `GET /readyz`, and `GET /healthz` at its root, outside `/api/v1/`, without authentication. Each returns `application/json`, and none reports on workflows or agents.

- `/livez` returns 200 with `{"status":"pass"}` whenever the server can answer. That includes startup and draining.
- `/readyz` returns 200 with `{"status":"pass"}` only while every required service runs and, for the worker target, the Temporal namespace is reachable. It returns 503 with `{"status":"fail"}` while the server starts, while it drains after SIGINT or SIGTERM, after a required service fails, and while the namespace is unreachable.
- `/healthz` returns `{"status":S,"checks":{"livez":{"status":L},"readyz":{"status":R}}}`, where `L` and `R` match what `/livez` and `/readyz` report at that moment. `S` and the status code are `"pass"` and 200 when both pass, and `"fail"` and 503 otherwise.

## Command-line interface (CLI)

The `afakt` command provides one command-line entrypoint for both running afakt's server-side components and communicating with them. See [Command-line reference](cli.md).

## GitHub webhook receiver

The GitHub webhook receiver accepts webhook deliveries from GitHub.

## Guest relay

The guest relay runs inside the sandbox VM under its own user, which alone has access to the vsock device. It handles ACP over standard input and output to the adapter, CLI and MCP control requests, listener discovery, and preview connections to guest loopback listeners. It reconnects and buffers session messages through the mediator. See [Guest layout and lifecycle control](agent-sandbox.md#guest-layout-and-lifecycle-control).

## Harness transport

Harness transport selects ACP or headless from the profile's model entry. ACP uses the relay and adapter, while headless runs the harness directly. Validation, hooks, results, and evidence apply to both paths. See [Harness transport](agent-execution.md#harness-transport).

## MCP gateway

The MCP gateway serves afakt agents only, with one Streamable HTTP endpoint per harness through the guest relay and mediator. On each call, it authenticates the mediator-stamped execution identity and token and enforces server and tool allowlists. It holds upstream credentials outside the execution sandbox. ACP requests govern tool approvals, and every call produces audit and telemetry records. Calls route to remote HTTP servers or stdio servers inside the sandbox VM or sandboxed on the control plane. The built-in afakt server supplies afakt API tools. See [Agent MCP gateway](agent-mcp-gateway.md).

## Mediator

Each agent's mediator runs as a confined process on the agent host, outside the sandbox VM. Its user-space TCP/IP stack serves the guest's only network device. The mediator resolves DNS and records name-to-address bindings. It applies Cedar connection and HTTP policy and inspects Transport Layer Security (TLS) with a per-agent certificate authority. Outside the guest, it brokers credentials. Alongside stamping control request identity, it enforces the ACP gate. It maintains the outbound connection to the preview broker. All decisions enter one log for the agent. See [Per-agent mediator](agent-sandbox.md#per-agent-mediator).

## Object store

The object store supplies the S3 API compatible endpoint used by the registry and every other afakt component that stores blobs. afakt includes RustFS as a supervised child process for single-node and small installs, and a deployment can select an external endpoint that meets the same contract. Conditional writes, lifecycle expiry, and path-style addressing are part of that contract. See [Object store](agent-workspace-image.md#object-store).

## Port forwarder

The port forwarder connects authorized publication streams through the mediator and guest relay to guest loopback listeners. It checks the current publication and generation before each stream. It never requires an inbound listener on the agent host. See [Port discovery and records](agent-ports.md#port-discovery-and-records).

## Preview broker

The preview broker runs beside the preview proxy. Each mediator dials out to it and maintains that connection. Authorized preview and client tunnel requests return as streams over the connection. See [HTTP previews](agent-ports.md#http-previews).

## Preview listener

The preview listener belongs to the [preview proxy](#preview-proxy) inside the API server process and remains separate from the API and dashboard listener. See [HTTP previews](agent-ports.md#http-previews).

## Preview proxy

The preview proxy runs inside the API server process. Cedar checks each viewer request before it routes a publication through the preview broker. The proxy strips afakt cookies and Authorization headers and rewrites the Host header or configures allowed hosts for the development server. See [Access and share links](agent-ports.md#access-and-share-links).

## Publication

A publication records an authorized port exposure with a random preview label, agent, port, generation, and agent host. The mediator validates it for each stream so stale routes can't select another listener. See [Port discovery and records](agent-ports.md#port-discovery-and-records).

## Registry

The registry stores workspace image manifests, config blobs, and layers in the object store. afakt runs Zot as a supervised child process with only the extensions it uses, behind an afakt-owned registry interface. The registry verifies control plane bearer tokens against published signing keys. Their scopes limit repositories and actions. It also caches upstream base images and collects unreferenced content online. See [Registry](agent-workspace-image.md#registry).

## Sandbox `VM`

A sandbox VM is the per-agent isolation boundary, with a guest kernel, read-only operating system disk, and writable workspace disk. The agent host supervisor controls its `VMM` independently of ACP and guest policy. See [Enforcement](agent-sandbox.md#enforcement).

## Session hub

The session hub is part of the API server. It joins mediator session connections to attached clients and enforces writer and reader roles. Its session message log supports history replay. An ACP execution keeps the agent executor workflow as its only writer and allows people with read access to attach as readers. See [Agent session attachment](agent-session-attachment.md).

## Snapshot registry

The snapshot registry stores and serves agent state archives for restoration and transfer. Snapshot scope and retention remain [open](agent-open-questions.md#agents). See [Snapshots and the snapshot registry](agent.md#snapshots-and-the-snapshot-registry).

## Tunnel providers

The afakt binary includes Cloudflare Tunnel and Tailscale Funnel. Their connectors run in the API server process and publish the preview listener through outbound connections, with optional separate publication of the API and dashboard. Provider credentials stay on the control plane. See [Public tunnel providers](agent-ports.md#public-tunnel-providers).

## Web dashboard

The web dashboard provides afakt's web UI.

The dashboard shares an origin and HTTP listener with the API, including when every runtime role runs together. Its routes belong under `/dashboard/`. Readiness requires every selected service to run and fails during startup and shutdown.

## Workflow orchestrator

The workflow orchestrator ([Temporal](https://temporal.io)) provides durable execution for afakt's workflows. A worker connects to an external Temporal service or runs an embedded Temporal server inside its own process. See [Workflows](architecture.md#workflows).

The Temporal Web UI shares the HTTP listener with the other runtime roles and gives developers and testers access to the embedded server's namespaces, workflows, and histories. Embedded Temporal operates independently of afakt's operational store.

## Workflows

afakt coordinates work using Temporal workflows.

### Agent executor

The agent executor workflow runs one generation of an agent execution on the agent host where the agent runner workflow placed it. For ACP executions, it's the only writer: it sends the rendered instructions and the prompt together as one prompt before waiting for the turn to end. It calls the operations of the host's execution provider and returns the execution result to the agent runner workflow. A conforming execution provider meets each item of [Agent provider requirements](agent-provider-requirements.md).

### Agent runner

The agent runner workflow runs harnesses for durable [agents](agent.md), placing each execution on an agent host and starting the [agent executor workflow](#agent-executor) there. [Agent execution](agent-execution.md) describes the lifecycle and states of each execution. [Agent execution model](agent-execution-model.md) defines its fields, which [Agent execution resource](agent-execution-resource.md) writes out as a resource, and [Agent execution contract](agent-execution-contract.md) defines the operations between the agent executor workflow and an execution provider. [Agent workspace](agent-workspace.md) and [Agent hooks](agent-hooks.md) define what the agent works in and the commands that run around and inside it, while [Agent profile](agent-profile.md) defines the profiles that bound each execution. [Agent sandbox](agent-sandbox.md) defines the sandbox that confines each execution and the broker that supplies its credentials, and [Agent observability](agent-observability.md) defines the trace and metrics each execution produces.

### Agent session

The agent session workflow holds an interactive session's control state, including its lifecycle, writer, attached clients, prompt queue, idle timer, and relay connection location. It doesn't hold the message stream. See [Agent session](agent-session.md).

## Workspace image builder

The workspace image builder resolves image definitions and coordinates builds triggered by schedules, repository branch pushes, or manual requests. Each build runs as an execution-like job on an agent host, with its commands inside the pinned agent profile's sandbox. The worker snapshots and pushes the completed tree, and publication moves the output tag only after success. See [Building workspace images](agent-workspace-image.md#image-builds).

## Workspace registry

The workspace registry stores workspace content and revision references for durable agents and reusable workspace images. The Open Container Initiative (OCI) [registry](#registry) distributes images. The relation between agent revisions and image publication remains [open](agent-open-questions.md#agents). See [Workspace revisions and the workspace registry](agent.md#workspace-revisions-and-the-workspace-registry).
