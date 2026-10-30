---
type: Component Reference
title: Components
description: The parts of afakt and what each one does.
tags: [afakt, components]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Components

[Public key infrastructure](pki.md).

## ACP adapter

The ACP adapter is the agent harness's Agent Client Protocol (ACP) agent inside the agent sandbox. Claude Code uses `claude-agent-acp`, and Codex uses `codex-acp`. The [harness transport](agent-profile.md#Harness%20transport) selects ACP or headless for an agent launch. An interactive [agent session](agent-session.md) uses ACP.

## ACP relay

The ACP relay is the [guest relay](#Guest%20relay) inside the workload guest. It acts as the ACP client to the adapter over standard input and output. For agent launches, it answers permission requests from profile grants and cancels elicitation requests. Its session traffic passes through the [workload proxy](#Workload%20proxy) under the mediator's default-deny policy for each ACP message. See [Agent launches over ACP](agent-acp.md#Agent%20launches%20over%20ACP).

## afakt MCP server

Through the built-in afakt Model Context Protocol (MCP) server, an agent can start a child agent launch or session and message another session. It can also read agent launch status and results or declare a forwarded port. Each call uses the calling agent launch's capabilities and scopes, and profiles allowlist these tools like any upstream server. See [Built-in afakt MCP server](agent-mcp-gateway.md#Built-in%20afakt%20MCP%20server).

## Agentmail

Agentmail manages messages addressed to durable agents. These include incoming and outgoing mail. Delivery and signing semantics remain [open](agent-open-questions.md#Agents). See [Agentmail, groups, and subscriptions](agent.md#Agentmail,%20groups,%20and%20subscriptions).

## API server

The Application Programming Interface (API) server serves afakt's Representational State Transfer (REST) API to its coordination workflows and, through the [ingress gateway](#Ingress%20gateway), to the dashboard, command-line interface (CLI), and apps. Workload proxies connect over the overlay. For an agent launch call, the API server identifies the launch from its peer certificate at Transport Layer Security (TLS) termination. From the immutable spec, it reads capabilities and scopes and refuses requests after the launch enters a terminal state. For remote control and sessions, it exposes [endpoints whose connection owner remains open](agent-acp.md#Remote%20control%20and%20sessions). It keeps the [workload host registry](workload-placement.md#Workload%20host%20registry), and its reservation API places launches on hosts.

Workload creation uses [workload templates](workload-template.md) and [workload images](workload-image.md). The control plane records the selected template version and immutable creation fields on the [Workload resource](workload-resource.md). The [workload API](workload-api.md) also covers host workers. A workload image supplies the container image for its process. An `empty` workload doesn't select one. afakt supplies Linux guest images separately on VM providers. Workload image and worker operations remain open. [Workload identity](workload-identity.md) separates caller authentication from permission to use those resources.

### Health endpoints

The server answers `GET /livez`, `GET /readyz`, and `GET /healthz` at its root, outside `/api/v1/`, without authentication. Each returns `application/json`, and none reports on workflows or agents.

- `/livez` returns 200 with `{"status":"pass"}` whenever the server can answer. That includes startup and draining.
- `/readyz` returns 200 with `{"status":"pass"}` only while every required service runs and, for the worker target, the Temporal namespace is reachable. It returns 503 with `{"status":"fail"}` while the server starts, while it drains after SIGINT or SIGTERM, after a required service fails, and while the namespace is unreachable.
- `/healthz` returns `{"status":S,"checks":{"livez":{"status":L},"readyz":{"status":R}}}`, where `L` and `R` match what `/livez` and `/readyz` report at that moment. `S` and the status code are `"pass"` and 200 when both pass, and `"fail"` and 503 otherwise.

## Artifact proxy

The artifact proxy serves stored artifacts associated with agents or managed independently. Its access and serving contract remains [open](agent-open-questions.md#Agents). See [Memory, skills, and artifacts](agent.md#Memory,%20skills,%20and%20artifacts).

## Capacity controller

The capacity controller runs in each workload host worker. It measures host central processing unit (CPU) use, memory use, and free workspace disk space at intervals and adjusts the capacity for agent launch slots within the host's configured lower and upper bounds. The worker reports capacity to the workload host registry in each lease renewal. See [Slot capacity](workload-placement.md#Slot%20capacity).

## Command-line interface (CLI)

The `afakt` command provides one command-line entrypoint for both running afakt's server-side components and communicating with them. See [Command-line reference](cli.md).

## Credential broker

The credential broker is a host-side component outside the guest. It injects credentials at the workload proxy into permitted requests outside the guest. Reflected response content remains [`OQ-SBX-26`](agent-open-questions.md#%60OQ-SBX-26%60). Its placement relative to the workload proxy and workload proxy mediator remains [open](agent-open-questions.md#%60OQ-SBX-23%60). See [Credential brokering](workload-sandbox.md#Credential%20brokering).

## GitHub webhook receiver

The GitHub webhook receiver accepts webhook deliveries from GitHub.

## Guest relay

The guest relay handles session traffic, control requests, and authorized listener streams under [Relay behavior](agent-acp.md#Relay%20behavior) and [Port discovery and records](workload-ports.md#Port%20discovery%20and%20records). Its carrier and local authentication follow the selected workload provider, with the shared choice under [`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60).

## Host-level network service

The host-level network service provides connectivity for workload ports, previews, services, and cluster services under host-level authorization. It's separate from launch traffic. Host-level name resolution doesn't grant access. Its process, local caller authentication, and connection mechanism remain [undecided](workload-overlay.md#Workload%20host%20access).

## Ingress gateway

The external ingress gateway uses Traefik with afakt authentication middleware to route user-facing traffic to the API server and preview server, and also to workload proxies. It also routes `/dashboard` on the API origin to the [web dashboard](#Web%20dashboard). Preview origins and listeners are separate from the API and dashboard. It can attach to user tailnets without routing packets into the cluster overlay. See [External ingress gateway](workload-overlay.md#External%20ingress%20gateway) for construction and integration, and [Access and share links](workload-previews.md#Access%20and%20share%20links) for viewer authorization.

## MCP gateway

The MCP gateway serves afakt agents only, with one Streamable HTTP endpoint per agent harness through the guest relay and workload proxy under mediator decisions. On each call, it checks the agent launch identity stamped by the mediator, whose workload proxy authenticates with the launch's SVID, and enforces server and tool allowlists. It holds upstream credentials outside the agent sandbox for the launch. ACP requests govern tool approvals, and every call produces audit and telemetry records. Calls route to remote HTTP servers or stdio servers inside the workload guest or sandboxed on the control plane. The built-in afakt server supplies afakt API tools. See [Agent MCP gateway](agent-mcp-gateway.md).

## Object store

The object store supplies the S3 API compatible endpoint used by the registry and every other afakt component that stores blobs. afakt includes RustFS as a supervised child process for single-node and small installs, and a deployment can select an external endpoint that meets the same contract. Conditional writes, lifecycle expiry, and path-style addressing are part of that contract. See [Object store](agent-workspace-image.md#Object%20store).

## Overlay node

An overlay node represents one workload proxy, workload host, or control plane service instance in the isolated cluster overlay. Service instances share a service tag but keep separate node identities. Guests aren't nodes. A node identity doesn't grant service access. SPIFFE authentication and service authorization are separate checks. See [Overlay node identity](workload-identity.md#Overlay%20node%20identity) and [Overlay transport](workload-overlay.md#Overlay%20transport).

## Port forwarder

Port forwarding requires authorization of the current publication and agent launch before reaching a guest loopback listener. Its [path through the workload proxy](agent-open-questions.md#%60OQ-PORT-14%60) remains open. See [Port discovery and records](workload-ports.md#Port%20discovery%20and%20records).

## Preview server

The preview server serves authorized preview requests as a client of published ports. It dials the workload proxy over the overlay with mutual authentication. The [published-port path](agent-open-questions.md#%60OQ-PORT-14%60), including the binding to the guest listener, remains open. See [Previews](workload-previews.md#Preview%20routing%20and%20names).

## Registry

The registry stores workspace image manifests, config blobs, and layers in the object store. An afakt-owned interface supplies scoped token authentication, upstream caching, and collection of unreferenced content. Registry software and process placement remain [`OQ-IMG-16`](agent-open-questions.md#%60OQ-IMG-16%60). See [Registry](agent-workspace-image.md#Registry) and [registry admission](agent-workspace-lifecycle.md#Registry%20admission).

## Snapshot registry

The snapshot registry stores and serves workload snapshots of VM memory and disk for restoration and transfer. Workload snapshot scope and retention remain [open](agent-open-questions.md#Agents). See [Workload snapshots and the snapshot registry](agent.md#Workload%20snapshots%20and%20the%20snapshot%20registry). [Workspace disk deltas and portable revisions](agent-workspace-lifecycle.md#Workspace%20disk%20deltas%20and%20portable%20revisions).

## Tailnet attachment

A tailnet attachment gives the external ingress gateway, or optionally a host-level node, access to a user's tailnet. It never adds launch nodes or routes packets between the user tailnet and cluster overlay. afakt sign-in remains required. See [User tailnet attachments](workload-overlay.md#User%20tailnet%20attachments).

## Tunnel providers

The afakt binary includes Cloudflare Tunnel and Tailscale Funnel. Their connectors run in the API server process and publish the ingress gateway's listener for previews through outbound connections, with optional separate publication of the API and dashboard. Provider credentials stay on the control plane. See [Public tunnel providers](workload-previews.md#Public%20tunnel%20providers).

## Web dashboard

The web dashboard, afakt's web UI, runs as a platform app on afakt's Platform as a Service (PaaS) rather than inside the binary, so no `afakt server` target includes it.

It combines a static single-page app on the static origin with a backend-for-frontend app workload. The [ingress gateway](#Ingress%20gateway) routes `/dashboard` on the API origin to both: assets and the app shell go to the static origin, and session and API-forwarding routes go to the backend-for-frontend.

As an OpenID Connect confidential client of afakt's provider, the backend-for-frontend authenticates with its own workload identity. It stores the user's tokens on the server behind a `__Host-` cookie, which keeps every access and refresh token out of the browser. On the API, only bearer tokens count, and the API ignores the cookie. Acting for the signed-in user, the backend-for-frontend has no standing authority of its own. See [`KS-04`](pki-signing.md#User%20authentication).

When the PaaS is down, the dashboard is down too, and the CLI, the recovery client, covers that time. A fresh install deploys the dashboard onto the PaaS from a release bundle included with the binary.

## Workflow orchestrator

The workflow orchestrator ([Temporal](https://temporal.io)) provides durable execution for afakt's workflows. A worker connects to an external Temporal service or runs an embedded Temporal server inside its own process. See [Workflows](architecture.md#Workflows).

The Temporal Web UI shares the HTTP listener with the other runtime services and gives developers and testers access to the embedded server's namespaces, workflows, and histories. Embedded Temporal operates independently of afakt's operational store.

## Workflows

afakt coordinates work using Temporal workflows.

### Agent launcher

The agent launcher workflow takes an immutable agent launch spec and manages one agent launch. The workload host worker runs it on the selected host task queue. It acquires an agent sandbox through the workload provider named in the spec, starts the agent harness, waits, and tears down. For workflow-controlled Agent Client Protocol (ACP) launches, it's the only writer and sends the rendered instructions and prompt together. It returns the agent launch result to the agent dispatcher workflow. See [Agent launch](agent-launch.md) and [Provider operations](agent-launch-contract.md#Workload%20provider%20operations).

### Agent dispatcher

The agent dispatcher workflow starts by default as a child of the caller's workflow. It authorizes an [agent launch request](agent.md#Agent%20launch%20request) and picks a host task queue. After finalizing the immutable [agent launch spec](agent-launch-contract.md#Agent%20launch%20spec), it queues the agent launch. Temporal delivers workflow and activity tasks to workers. This app workflow makes the authorization, placement, and configuration choices. See [Workload placement](workload-placement.md) and [Agent launch model](agent-launch-model.md).

### Agent session

The agent session workflow holds an interactive session's control state, including its lifecycle, writer, attached clients, idle timer, and relay connection location. Prompt queue ownership remains [`OQ-SES-14`](agent-open-questions.md#%60OQ-SES-14%60). It doesn't hold the message stream. See [Agent session](agent-session.md).

## Workload guest

The [workload guest](workload-guest.md) is the VM that runs the workload on a VM provider. Guest routing components don't decide policy. External traffic, including DNS, uses only the workload proxy.

A workload guest is the isolation boundary, with a Linux guest kernel, read-only operating system disk, and writable workspace disk. The workload host supervisor controls its virtual machine monitors (VMMs) independently of ACP and guest policy. See [Enforcement](workload-sandbox.md#Enforcement).

## Workload host worker

The [workload host worker](workload-host-worker.md) runs as a dedicated non-root user on a [workload host](workload-host.md). It registers capacity and polls its host task queue. Alongside reconciliation of agent launch records, the worker observes [workload lifecycle](workload-lifecycle.md) state under the [open ownership contract](agent-open-questions.md#%60OQ-SBX-71%60). Control connections originate at the worker. The control plane never opens a connection to a host. Workload data paths exclude this daemon. See [Workload host worker resource](workload-host-worker-resource.md) and [Workload host caches and startup](agent-workspace-lifecycle.md#Workload%20host%20caches%20and%20startup).

## Workload proxy

Each workload's proxy transmits guest traffic in both directions and requests a mediator decision for each connection. It has its own overlay node and holds the launch's X.509 SPIFFE Verifiable Identity Document (SVID) and private key outside the guest. The ingress gateway and permitted peers dial it directly with SPIFFE mutual TLS. For inspected outbound destinations, it ends TLS under mediator decisions using a per-workload certificate authority trusted by the guest. The credential broker injects credentials there, outside the guest. See [Workload proxy](workload-proxy.md).

## Workload proxy mediator

Each workload's proxy mediator decides every inbound and outbound connection without transmitting traffic and stamps launch identity on control requests. The proposed workload Domain Name System (DNS) design puts resolution under mediator control with no host fallback. Its contract remains [open](agent-open-questions.md#%60OQ-SBX-21%60). The mediator's control plane connections pass through its proxy under a fixed rule set that neither the launch's policy nor the guest can change. See [Workload proxy mediator](workload-proxy-mediator.md) and [ACP message policy](agent-acp.md#ACP%20message%20policy).

## Workspace image builder

The workspace image builder resolves image definitions and coordinates builds triggered by schedules, repository branch pushes, or manual requests. Each build runs as a job on a workload host, with its commands inside the pinned agent profile's agent sandbox. The worker snapshots and pushes the completed tree, and publication moves the output tag only after success. See [Building workspace images](agent-workspace-image.md#Image%20builds). [Workspace image build lifecycle](agent-workspace-lifecycle.md#Image%20builds).

## Workspace registry

The workspace registry stores workspace content and revision references for durable agents and reusable workspace images. The Open Container Initiative (OCI) [registry](#Registry) distributes images. The relation between agent revisions and image publication remains [open](agent-open-questions.md#Agents). See [Workspace revisions and the workspace registry](agent.md#Workspace%20revisions%20and%20the%20workspace%20registry). [Retention and garbage collection](agent-workspace-lifecycle.md#Retention%20and%20garbage%20collection).
