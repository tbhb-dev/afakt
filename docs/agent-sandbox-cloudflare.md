---
type: Design
title: "Agent sandbox provider for Cloudflare Sandbox SDK"
description: Remote isolation, mediation, storage, and lifecycle for Cloudflare Sandbox SDK.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent sandbox provider for Cloudflare Sandbox SDK

[Public key infrastructure](pki.md).

## Platforms and privileges

The Cloudflare Sandbox Software Development Kit (SDK) provider uses a Cloudflare Worker, a Durable Object, and a Linux container. Calling the SDK from the Worker routes operations through a Durable Object, which owns container lifecycle and routing, while the container executes code. [Sandbox SDK architecture](https://developers.cloudflare.com/sandbox/sdk/concepts/architecture/).

Outside the sandbox, the agent host worker retains provider management authority under [Runtime privileges and platforms](agent-host.md#Runtime%20privileges%20and%20platforms). Execution-user privileges and compatibility with afakt-sandbox's confinement mechanisms lack a complete contract in the SDK documentation. Those gaps remain under [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60). afakt rejects unsupported settings under [Enforcement](agent-sandbox.md#Enforcement), and the host isolation baseline remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [Container runtime](https://developers.cloudflare.com/sandbox/sdk/concepts/containers/).

## Isolation boundary

Cloudflare Sandbox SDK runs each sandbox in a separate virtual machine (VM), with separate filesystem, process, and network state and per-sandbox resource quotas. Shell sessions inside one sandbox share files and processes. A session isn't a security boundary. Inside the container, loading kernel modules and accessing host hardware are unavailable. Containment after a virtual-machine escape isn't stated, and protection from vendor access isn't established. These gaps remain under [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60), [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60). [Security model](https://developers.cloudflare.com/sandbox/sdk/concepts/security/), [Sessions](https://developers.cloudflare.com/sandbox/sdk/concepts/sessions/), [Container runtime](https://developers.cloudflare.com/sandbox/sdk/concepts/containers/).

## Process placement and startup

Cloudflare Sandbox SDK executes foreground and background processes inside the container. It doesn't supply a required managed agent loop. afakt's [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) places afakt-agent, afakt-sandbox, and the harness there. The integrating Worker and Durable Object remain outside the container. Mapping those components to the agent host worker's provider operations remains undecided, as do the local relay and launch identity under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60), [`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60), [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60). [Sandbox SDK architecture](https://developers.cloudflare.com/sandbox/sdk/concepts/architecture/), [Container runtime](https://developers.cloudflare.com/sandbox/sdk/concepts/containers/).

Commands requested through `afakt agent exec` require standalone afakt-sandbox, with placement, inherited environment, launch binding, and descendant lifetime unresolved under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60), [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

Cloudflare Sandbox SDK creates a container lazily on its first operation. Reusing a sandbox identifier routes to the same Durable Object, while a stopped container loses its execution state. Identifiers aren't cryptographic authentication. The app must authenticate callers. Binding the identifier and each container lifetime to the intended launch remains under [`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60), [`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60). [Lifecycle API](https://developers.cloudflare.com/sandbox/sdk/api/lifecycle/), [Security model](https://developers.cloudflare.com/sandbox/sdk/concepts/security/).

## Mediator connection and name resolution

Each launch requires [agent mediation](agent-workload-mediator.md#Launch%20boundary) for control, session, preview, and network traffic.

Remote launches require the [agent proxy](agent-workload-proxy.md#Launch%20boundary) and [global egress path](components.md#Agent%20egress%20gateway), with their placement and carrier subject to the provider mapping below.

Remote placement, mandatory transport, Domain Name System (DNS) mediation, and prevention of direct connections remain under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). Provider assertions for launch identity remain under [`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60).

Cloudflare Sandbox SDK allows internet access by default. With internet access off, only explicitly permitted hosts or outbound handlers allow web traffic, while DNS uses Cloudflare resolvers. Outbound handlers process Hypertext Transfer Protocol (HTTP) and Hypertext Transfer Protocol Secure (HTTPS) on ports 80 and 443, not other protocols. HTTPS interception defaults to off. Host rules and handlers apply only to HTTP without it. HTTPS can reach any host with internet access on and fails with internet access off. These settings don't select afakt's mandatory DNS and proxy path under mediator decisions under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). [Handle outbound traffic](https://developers.cloudflare.com/sandbox/sdk/guides/outbound-traffic/).

Cloudflare Sandbox SDK outbound handlers run in the Workers runtime outside the sandbox, on the same machine, with access to Worker bindings. Handlers can inject credentials kept in the Worker and scope them by container identity. Their rules can change while a sandbox runs. HTTPS interception uses an ephemeral certificate authority (CA) that the runtime tries to add to common trust stores. Using these handlers for afakt mediation, networking, or credential brokering remains undecided. Their identity binding, no-network mediator confinement, immutable launch configuration, and global egress path remain under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60), [`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60). Repository profile self-grants and credential domain bounds remain on hold under [`OQ-SBX-2`](agent-open-questions.md#%60OQ-SBX-2%60). [Handle outbound traffic](https://developers.cloudflare.com/sandbox/sdk/guides/outbound-traffic/).

Cloudflare Sandbox SDK preview uniform resource locators (URLs) grant access to anyone holding them. Quick tunnels use the random host name itself as access control. The SDK requires app authentication for sensitive services, without establishing afakt's per-stream publication checks or authenticated workload proxy path through these vendor routes. Their integration remains under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). See [Port discovery and records](agent-ports.md#Port%20discovery%20and%20records). [Security model](https://developers.cloudflare.com/sandbox/sdk/concepts/security/).

## Workspace, scratch, and caches

The agent workspace needs an independent writable tree, separate scratch, and a verified image cache outside launch-writable content. Provider backing, permissions, capture access, and cleanup remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). Shared named tool cache access and poisoning remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Hostile repository metadata during capture remains under [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). See [Sandbox variables](agent-sandbox.md#Sandbox%20variables) and [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture).

Cloudflare Sandbox SDK preserves files, processes, and shell state only while its container is active. Sleep or destruction removes that state. Shell sessions in one sandbox share its filesystem and process space. Persistent bucket mounts and directory backups are separate storage mechanisms. Their mapping to launch-specific workspace backing, scratch, and shared named tool caches remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60), [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). [Sandbox lifecycle](https://developers.cloudflare.com/sandbox/sdk/concepts/sandboxes/), [Sessions](https://developers.cloudflare.com/sandbox/sdk/concepts/sessions/), [Directory backups](https://developers.cloudflare.com/sandbox/sdk/concepts/backup-restore/).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

Agent workspace images contain directory trees in Open Container Initiative (OCI) form. [Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup) requires content verification, confined extraction, and independent writable materialization before hooks or the harness start. The remote storage and transfer mapping remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). A vendor runtime image or saved environment configuration doesn't establish this workspace contract.

Cloudflare Sandbox SDK supports customized container images. A directory backup restores in production as a read-only squashfs archive under a writable overlay using Filesystem in User Space (FUSE). Restoring the handle again discards the upper layer. Sleep removes the mount. This is different from the disk materialization and no-lazy-pull rules in the workspace topic, and it isn't a selected afakt image path. Transfer, validation, and ownership remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Dockerfile reference](https://developers.cloudflare.com/sandbox/sdk/configuration/dockerfile/), [Directory backups](https://developers.cloudflare.com/sandbox/sdk/concepts/backup-restore/).

## Agent snapshots and agent checkpoints

An agent snapshot captures virtual-machine state, while the agent checkpoint holds the native session state of the harness. Using both requires pairing at an idle turn boundary under [agent checkpoints](agent.md#Agent%20checkpoints) and [suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Provider capture scope and complete session restoration remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60). Exclusion of active authority and capture authorization remain under [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60), [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). By itself, vendor persistence doesn't establish an afakt agent snapshot or agent checkpoint.

Cloudflare Sandbox SDK backs up directory content to an R2 archive, without a documented operation to capture guest kernel state and process memory with the native session history that afakt needs to pair an agent snapshot with its agent checkpoint. Directory restoration is insufficient. Agent snapshot eligibility and checkpoint restoration remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60), with retained content authorization under [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). [Directory backups](https://developers.cloudflare.com/sandbox/sdk/concepts/backup-restore/).

## Cancellation and teardown

The agent sandbox provider retains launch ownership until it observes termination of every process started for the launch. That observation must precede terminal status or slot release. [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) and [Reconciliation after restart or retry](agent-launch.md#Reconciliation%20after%20restart%20or%20retry) distinguish a stop request, an unavailable provider, and observed exit. Remote termination evidence, retries, detached work, and replacement safety remain under [`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60). Credential retirement remains under [`OQ-SBX-40`](agent-open-questions.md#%60OQ-SBX-40%60). Session-end wait and identity lifetime remain on hold under [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Cloudflare Sandbox SDK stops an idle container after ten minutes by default, with configurable sleep duration. Keep-alive mode prevents automatic sleep and requires explicit management. Destruction immediately stops the container, its processes, sessions, network connections, and exposed ports. A later operation can start a fresh container behind the same sandbox identifier. The documentation doesn't state an independent afakt observation of completed destruction or a bound preventing late requests from recreating execution. Those gaps remain under [`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60). [Sandbox lifecycle](https://developers.cloudflare.com/sandbox/sdk/concepts/sandboxes/), [Lifecycle API](https://developers.cloudflare.com/sandbox/sdk/api/lifecycle/).

## Host observations

The agent host worker needs launch-bound status, resource use, termination evidence, and capture results under [Provenance record](agent-launch-model.md#Provenance%20record). Vendor observations aren't independent proof against vendor compromise. [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) excludes secret values and workspace file contents from afakt telemetry. It doesn't constrain vendor custody. Vendor personnel access, internal copies, deletion completion, and independently protected evidence remain under [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60).

Cloudflare Sandbox SDK carries commands and output through the Worker and Durable Object to the container. Worker outbound handlers can hold credentials and inspect web traffic outside that container. Directory backups place workspace content in R2. In the cited SDK pages, internal personnel access limits, log retention, and physical erasure timing remain unstated, as does protection of workspace and Worker secrets from vendor access. Those gaps remain under [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60). [Sandbox SDK architecture](https://developers.cloudflare.com/sandbox/sdk/concepts/architecture/), [Handle outbound traffic](https://developers.cloudflare.com/sandbox/sdk/guides/outbound-traffic/), [Directory backups](https://developers.cloudflare.com/sandbox/sdk/concepts/backup-restore/).

## Platform dependencies

Cloudflare Sandbox SDK depends on Workers, Durable Objects, and Containers. Its limits page delegates container processor, memory, disk, image, and concurrent-instance limits to the Containers platform. HTTP transport consumes one Worker subrequest per operation. Remote Procedure Call (RPC) transport multiplexes operations over one connection. The first request determines sandbox placement, and later requests use that location. An eligible region inventory and per-data-class residency guarantees remain absent from the cited SDK pages. Capacity and geographic eligibility remain under [`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60), and provider verification remains under [`OQ-PLC-9`](agent-open-questions.md#%60OQ-PLC-9%60). [Platform limits](https://developers.cloudflare.com/sandbox/sdk/platform/limits/), [Transport modes](https://developers.cloudflare.com/sandbox/sdk/configuration/transport/), [Sandbox lifecycle](https://developers.cloudflare.com/sandbox/sdk/concepts/sandboxes/).
