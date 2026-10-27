---
type: Design
title: "Agent sandbox provider for Docker Engine"
description: Isolation, mediation, storage, and lifecycle for the docker-engine agent sandbox provider.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent sandbox provider for Docker Engine

[Public key infrastructure](pki.md).

## Platforms and privileges

The docker-engine provider runs Linux containers through a rootless Docker Engine on an amd64 Linux agent host, either a physical machine or a virtual machine (VM), and on an arm64 Linux VM. The worker and engine run without root privileges. Setup and service survival remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60). Supported platforms remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60).

## Isolation boundary

The agent harness shares the Linux agent host kernel with other containers. The container has no private guest kernel. Container namespace, mount, capability, engine socket, and host service restrictions remain under [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60). The host isolation baseline remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Process placement and startup

The agent host worker asks the rootless engine to start a container. afakt-agent runs in the container and starts afakt-sandbox and the harness. The worker retains engine management authority outside the container. Container startup, management caller authentication, inherited descriptors, and restrictions on runtime access remain under [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60).

Commands started by `afakt agent exec` use standalone afakt-sandbox without the runner. afakt-agent is absent. Command placement, inherited environment, launch binding, capture ordering, and descendant lifetime remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). That choice also determines command transport. Authority for an unsandboxed command remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

Local helper authentication and binding to the intended launch remain under [`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60) and [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

Each agent launch's agent mediator and agent proxy follow [Per-launch services](agent-proxy.md#Launch%20boundary) and the authenticated [hub connection](agent-proxy.md#Hub%20connection).

The container needs its own agent mediator for Domain Name System (DNS) resolution and external traffic. Control requests and preview streams also need that mediator. The mediator’s placement relative to the engine and its mandatory transport remain under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60). An internal bridge can expose its gateway and other attached services. Removing ordinary network interfaces instead requires an explicit control and data carrier. The topology remains undecided. Engine sockets, host aliases, embedded resolvers, and desktop service paths remain unresolved confinement surfaces.

[Workload identity](agent-host.md#Workload%20identity) keeps agent-proxy credentials outside this provider’s confinement. Kernel-attested local transport remains the leading option in [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). Current-launch binding and sender attribution remain unresolved under [`OQ-SES-6`](agent-open-questions.md#%60OQ-SES-6%60). Control capacity remains under [`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60), while mediator placement and confinement remain under [`OQ-SBX-12`](agent-open-questions.md#%60OQ-SBX-12%60), [`OQ-SBX-19`](agent-open-questions.md#%60OQ-SBX-19%60), [`OQ-SBX-20`](agent-open-questions.md#%60OQ-SBX-20%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60).

[Egress gateway routing](agent-mediator.md#Egress%20gateway%20routing) governs host or configured site forwarding after mediator admission. Upstream resolution, loopback publication, and standalone command routing remain under [`OQ-SBX-21`](agent-open-questions.md#%60OQ-SBX-21%60), [`OQ-SBX-22`](agent-open-questions.md#%60OQ-SBX-22%60), [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60), [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Workspace, scratch, and caches

The rootless engine controls container storage on the Linux agent host. The agent workspace is launch-specific writable content attached to the container. Its directory or volume mapping remains undecided. Verified image caches remain outside launch-writable content.

Provider storage mappings and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandbox variables](agent-sandbox.md#Sandbox%20variables) defines separate scratch and host-shared named tool caches. Cache authorization, provenance, confidentiality, capacity, and invalidation remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Host-side processing of hostile repository metadata needs the isolation in [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). Worker capture rights and native history path stability remain under [`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60) and [`OQ-WSP-16`](agent-open-questions.md#%60OQ-WSP-16%60).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

[Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup) governs verified extraction and independent writable materialization. Immutable image caches remain outside launch-writable storage. Agent workspace images contain directory trees. Their format is Open Container Initiative (OCI). A container runtime image starts the container. Workspace files come from the separate agent workspace image, whose verified cache remains distinct from engine image-layer caches. Engine storage, workspace attachments, cloning or copying, and ownership mappings remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60).

## Agent snapshots and agent checkpoints

This provider has no agent snapshot: it doesn't capture a private guest kernel or process memory. A native harness checkpoint is distinct from workspace content. Checkpoint and workspace pairing remain unresolved. Resume at a changed workspace path, restore authorization, retention, and old authority remain under [`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60). [Agent checkpoints](agent.md#Agent%20checkpoints) supplies the shared checkpoint model. A checkpoint can't undo an external effect.

## Cancellation and teardown

[Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) requires observed termination before a terminal state or slot release. A stop request or process-group exit alone doesn't prove every descendant has stopped.

Independent termination, worker and helper failure, runtime restart, and recovery remain under [`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60). Reusing launch identities while descendants survive remains under [`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60). Standalone command termination and capture ordering remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Container exit proves process state, while losing the worker proves neither container exit nor object removal. Workspace cleanup and named cache retention need separate observations.

## Host observations

The agent host controls process lifecycle, workspace backing, retained captures, and the agent mediator. Host administrators can inspect or change these resources. This provider supplies no confidentiality against the agent host. Independent evidence of a compromised host remains under [`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60). Same-user process access and the broader host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) limits exported content. It doesn't prevent local host access. Command attribution remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Platform dependencies

This provider relies on Linux namespaces, rootless engine operation, user and group mappings, and container lifecycle control. Exact runtime restrictions and backend support remain under [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60). Setup and service lifetime remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60). Network enforcement remains unresolved under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60).
