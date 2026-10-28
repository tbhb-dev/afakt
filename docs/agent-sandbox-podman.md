---
type: Design
title: "Agent sandbox provider for Podman"
description: Isolation, mediation, storage, and lifecycle for the Podman agent sandbox provider.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent sandbox provider for Podman

[Public key infrastructure](pki.md).

## Platforms and privileges

The Podman provider uses rootless Podman containers, with a dedicated non-root agent host worker. Supported host operating systems and architectures, rootless setup, privileges, and required services remain under [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60). A Linux rootless deployment doesn't choose the macOS or Windows deployment. Setup and service survival remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60).

## Isolation boundary

The agent harness shares the container kernel. It has no private per-agent guest kernel. Whether a Podman virtual machine (VM) supplies that kernel on macOS or Windows remains under [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60). Namespace, mount, capability, host integration, and runtime socket restrictions are unresolved under [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60), [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60). If there is an outer virtual machine, its host-side containment also needs the boundary in [`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60).

## Process placement and startup

The agent host worker controls Podman lifecycle operations. afakt-agent runs in the container and starts afakt-sandbox and the harness. The management service, local process placement, peer authentication, and permissions depend on the deployment in [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60). A runtime socket is management authority, even through a read-only filesystem mount.

Commands started by `afakt agent exec` use standalone afakt-sandbox without the runner. afakt-agent is absent. Command placement, inherited environment, launch binding, capture ordering, and descendant lifetime remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). That choice also determines command transport. Authority for an unsandboxed command remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

Local helper authentication and binding to the intended launch remain under [`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60) and [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

Each workload's agent mediator and agent proxy follow [Launch boundary](agent-workload-proxy.md#Launch%20boundary) and the authenticated [overlay connections](agent-workload-proxy.md#Overlay%20connections).

All external traffic uses the workload proxy. This includes Domain Name System (DNS) queries, which the proxy forwards to the mediator without a host resolver fallback. Each inbound and outbound connection requires a mediator decision. Encryption isn't required on the private link to the proxy. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft. Provider enforcement, including alternate interfaces, inherited descriptors, and indirect host services, remains under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60).

[Workload identity](agent-host.md#Workload%20identity) keeps agent-proxy credentials outside this provider’s confinement. Kernel-attested local transport remains the leading option in [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). Current-launch binding and sender attribution remain unresolved under [`OQ-SES-6`](agent-open-questions.md#%60OQ-SES-6%60). Control capacity remains under [`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60), while mediator placement and confinement remain under [`OQ-SBX-12`](agent-open-questions.md#%60OQ-SBX-12%60), [`OQ-SBX-19`](agent-open-questions.md#%60OQ-SBX-19%60), [`OQ-SBX-20`](agent-open-questions.md#%60OQ-SBX-20%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60).

[Egress gateway routing](agent-workload-mediator.md#Agent%20egress%20gateway) governs host or configured site forwarding after mediator admission. Upstream resolution and workspace command routing remain under [`OQ-SBX-21`](agent-open-questions.md#%60OQ-SBX-21%60), [`OQ-SBX-22`](agent-open-questions.md#%60OQ-SBX-22%60), [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Workspace, scratch, and caches

Podman manages container storage, but the host directory, volume, or outer virtual-machine backing for the agent workspace remains undecided under [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60), [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). Verified image content remains separate from launch-writable state.

Provider storage mappings and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandbox variables](agent-sandbox.md#Sandbox%20variables) defines separate scratch and host-shared named tool caches. Cache authorization, provenance, confidentiality, capacity, and invalidation remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Host-side processing of hostile repository metadata needs the isolation in [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). Worker capture rights and native history path stability remain under [`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60) and [`OQ-WSP-16`](agent-open-questions.md#%60OQ-WSP-16%60).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

[Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup) governs verified extraction and independent writable materialization. Immutable image caches remain outside launch-writable storage. Agent workspace images contain directory trees. Their format is Open Container Initiative (OCI). A Podman runtime image starts the container. The agent workspace image separately supplies the writable workspace tree. Podman’s image cache isn't the shared named tool cache or the verified workspace-image cache. Backing storage, extraction ownership, materialization, and capture access remain under [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60), [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60).

## Agent snapshots and agent checkpoints

This provider has no agent snapshot: it doesn't capture a private guest kernel or process memory. A native harness checkpoint is distinct from workspace content. Checkpoint and workspace pairing remain unresolved. Resume at a changed workspace path, restore authorization, retention, and old authority remain under [`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60). [Agent checkpoints](agent.md#Agent%20checkpoints) supplies the shared checkpoint model. A checkpoint can't undo an external effect. A possible outer Podman virtual machine is an undecided per-agent capture boundary under [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60).

## Cancellation and teardown

[Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) requires observed termination before a terminal state or slot release. A stop request or process-group exit alone doesn't prove every descendant has stopped.

Independent termination, worker and helper failure, runtime restart, and recovery remain under [`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60). Reusing launch identities while descendants survive remains under [`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60). Standalone command termination and capture ordering remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). After container process exit, the container object, workspace volume, and network allocation remain. A worker exit doesn't prove a container exit. Recovery must distinguish these states. Its authority and provider observations remain under [`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60), [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60).

## Host observations

The agent host controls process lifecycle, workspace backing, retained captures, and the agent mediator. Host administrators can inspect or change these resources. This provider supplies no confidentiality against the agent host. Independent evidence of a compromised host remains under [`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60). Same-user process access and the broader host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) limits exported content. It doesn't prevent local host access. Command attribution remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). A Podman service or outer virtual-machine manager adds management custody according to the deployment in [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60).

## Platform dependencies

This provider relies on Podman lifecycle control and container facilities on the selected host. Podman must run without root privileges. Prerequisites, subordinate identity mappings, service requirements, network backend, and any outer virtual machine remain under [`OQ-SBX-58`](agent-open-questions.md#%60OQ-SBX-58%60). Container restrictions and mediation remain under [`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60), [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60).
