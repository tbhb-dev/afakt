---
type: Design
title: "Agent sandbox provider for Docker Desktop on macOS"
description: Isolation, mediation, storage, and lifecycle for the docker-desktop-macos agent sandbox provider.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent sandbox provider for Docker Desktop on macOS

[Public key infrastructure](pki.md).

## Platforms and privileges

The docker-desktop-macos provider runs Linux containers in Docker Desktop’s shared Linux virtual machine (VM) on macOS. A non-root macOS user runs the agent host worker. Access to the desktop engine is management authority, not proof that the inner engine runs without root. Setup and service survival remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60). Supported platforms remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60). Container management must run without root privileges, with desktop backend enforcement unresolved under [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60).

## Isolation boundary

The container shares the desktop Linux kernel with other containers. An outer virtual machine separates that shared kernel from macOS. Agents share this machine. Container and desktop file-sharing restrictions remain under [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60). Containment after exploitation of the outer machine or host virtual devices remains under [`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60). The host isolation baseline remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Process placement and startup

The agent host worker on macOS controls the desktop engine. afakt-agent runs inside the Linux container and starts afakt-sandbox and the harness. Desktop backend services and the outer virtual machine remain outside the container. Container startup, management caller authentication, inherited descriptors, and restrictions on runtime access remain under [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60).

Commands started by `afakt agent exec` use standalone afakt-sandbox without the runner. afakt-agent is absent. Command placement, inherited environment, launch binding, capture ordering, and descendant lifetime remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). That choice also determines command transport. Authority for an unsandboxed command remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

Local helper authentication and binding to the intended launch remain under [`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60) and [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

Each workload's agent mediator and agent proxy follow [Launch boundary](agent-workload-proxy.md#Launch%20boundary) and the authenticated [overlay connections](agent-workload-proxy.md#Overlay%20connections).

All external traffic uses the workload proxy. This includes Domain Name System (DNS) queries, which the proxy forwards to the mediator without a host resolver fallback. Each inbound and outbound connection requires a mediator decision. Encryption isn't required on the private link to the proxy. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft. Provider enforcement, including alternate interfaces, inherited descriptors, and indirect host services, remains under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60).

[Workload identity](agent-host.md#Workload%20identity) keeps agent-proxy credentials outside this provider’s confinement. Kernel-attested local transport remains the leading option in [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). Current-launch binding and sender attribution remain unresolved under [`OQ-SES-6`](agent-open-questions.md#%60OQ-SES-6%60). Control capacity remains under [`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60), while mediator placement and confinement remain under [`OQ-SBX-12`](agent-open-questions.md#%60OQ-SBX-12%60), [`OQ-SBX-19`](agent-open-questions.md#%60OQ-SBX-19%60), [`OQ-SBX-20`](agent-open-questions.md#%60OQ-SBX-20%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60).

[Egress gateway routing](agent-workload-mediator.md#Agent%20egress%20gateway) governs host or configured site forwarding after mediator admission. Upstream resolution, loopback publication, and standalone command routing remain under [`OQ-SBX-21`](agent-open-questions.md#%60OQ-SBX-21%60), [`OQ-SBX-22`](agent-open-questions.md#%60OQ-SBX-22%60), [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60), [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Workspace, scratch, and caches

Desktop virtual-machine storage or a physical host share could back the workspace. The mapping remains undecided. Shared host folders are extra paths to macOS files, not inherently isolated workspace storage. Verified image caches remain outside launch-writable content.

Provider storage mappings and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandbox variables](agent-sandbox.md#Sandbox%20variables) defines separate scratch and host-shared named tool caches. Cache authorization, provenance, confidentiality, capacity, and invalidation remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Host-side processing of hostile repository metadata needs the isolation in [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). Worker capture rights and native history path stability remain under [`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60) and [`OQ-WSP-16`](agent-open-questions.md#%60OQ-WSP-16%60).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

[Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup) governs verified extraction and independent writable materialization. Immutable image caches remain outside launch-writable storage. Agent workspace images contain directory trees. Their format is Open Container Initiative (OCI). A container runtime image starts the container. Workspace files come from the separate agent workspace image, whose verified cache remains distinct from engine image-layer caches. Engine storage, workspace attachments, cloning or copying, and ownership mappings remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60).

## Agent snapshots and agent checkpoints

This provider has no agent snapshot: it doesn't capture a private guest kernel or process memory. A native harness checkpoint is distinct from workspace content. Checkpoint and workspace pairing remain unresolved. Resume at a changed workspace path, restore authorization, retention, and old authority remain under [`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60). [Agent checkpoints](agent.md#Agent%20checkpoints) supplies the shared checkpoint model. A checkpoint can't undo an external effect. The shared desktop virtual machine is outside the per-agent snapshot boundary. Saving it would also capture unrelated containers.

## Cancellation and teardown

[Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) requires observed termination before a terminal state or slot release. A stop request or process-group exit alone doesn't prove every descendant has stopped.

Independent termination, worker and helper failure, runtime restart, and recovery remain under [`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60). Reusing launch identities while descendants survive remains under [`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60). Standalone command termination and capture ordering remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Container exit proves process state, while losing the worker proves neither container exit nor object removal. Workspace cleanup and named cache retention need separate observations. Launches share the desktop backend. Stopping it affects unrelated launches, so that operation isn't a selected way to stop one launch.

## Host observations

The agent host controls process lifecycle, workspace backing, retained captures, and the agent mediator. Host administrators can inspect or change these resources. This provider supplies no confidentiality against the agent host. Independent evidence of a compromised host remains under [`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60). Same-user process access and the broader host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) limits exported content. It doesn't prevent local host access. Command attribution remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). The desktop backend also observes Linux container state and any physical host file shares.

## Platform dependencies

This provider relies on Docker Desktop, a shared Linux virtual machine, container isolation, desktop backend management, and controlled physical host integration. Exact runtime restrictions and backend support remain under [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60). Setup and service lifetime remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60). Network enforcement remains unresolved under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60).
