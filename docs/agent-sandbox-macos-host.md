---
type: Design
title: "Agent sandbox provider for macOS host"
description: Isolation, mediation, storage, and lifecycle for the macos-host agent sandbox provider.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent sandbox provider for macOS host

[Public key infrastructure](pki.md).

## Platforms and privileges

The macos-host provider runs the harness directly on macOS without a virtual machine (VM). A dedicated non-root user runs the agent host worker. Setup privileges and service lifetime remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60). Supported platforms and distribution requirements remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60). Transparent capture requires an approved, signed Network Extension system extension. The host app activates and deactivates its transparent proxy and name proxy. Installation and distribution requirements remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60).

## Isolation boundary

The agent harness shares the macOS kernel with the agent host. Seatbelt confines filesystem and process access under [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox). It doesn't create a private guest kernel or a private loopback network. Per-launch user separation and the host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). Public host services and inherited interprocess communication remain under [`OQ-SBX-28`](agent-open-questions.md#%60OQ-SBX-28%60).

## Process placement and startup

The agent host starts afakt-agent as a host process. afakt-agent starts afakt-sandbox, which applies Seatbelt before starting the harness. The transparent proxy and name proxy run as system extension providers outside harness confinement. Their lifecycle, registration authority, and mapping to an agent mediator remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60).

Commands started by `afakt agent exec` use standalone afakt-sandbox without the runner. afakt-agent is absent. Command placement, inherited environment, launch binding, capture ordering, and descendant lifetime remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). That choice also determines command transport. Authority for an unsandboxed command remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

Local helper authentication and binding to the intended launch remain under [`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60) and [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

Each workload's agent mediator and agent proxy follow [Launch boundary](agent-workload-proxy.md#Launch%20boundary) and the authenticated [overlay connections](agent-workload-proxy.md#Overlay%20connections).

The workload proxy is the only external traffic path, including Domain Name System (DNS) queries. It requests a mediator decision for each inbound and outbound connection. The proxy forwards DNS queries to the mediator, with no host resolver fallback. The private link to the proxy doesn't require encryption. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft. Provider enforcement, including alternate interfaces, inherited descriptors, and indirect host services, remains under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60). Host-mode capture sends traffic through the same proxy enforcement point. Containment applies against the agent at its sandbox privileges, with the mechanism held under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

[Workload identity](agent-host.md#Workload%20identity) keeps agent-proxy credentials outside this provider’s confinement. Kernel-attested local transport remains the leading option in [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). Current-launch binding and sender attribution remain unresolved under [`OQ-SES-6`](agent-open-questions.md#%60OQ-SES-6%60). Control capacity remains under [`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60), while mediator placement and confinement remain under [`OQ-SBX-12`](agent-open-questions.md#%60OQ-SBX-12%60), [`OQ-SBX-19`](agent-open-questions.md#%60OQ-SBX-19%60), [`OQ-SBX-20`](agent-open-questions.md#%60OQ-SBX-20%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60).

Capture registration, stale flow mappings, original-destination integrity, and privileged component containment remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60). Shared host loopback doesn't become launch-private: permitted listener exposure and rejected sandbox settings remain under [`OQ-SBX-16`](agent-open-questions.md#%60OQ-SBX-16%60). [Egress gateway routing](agent-workload-mediator.md#Agent%20egress%20gateway) governs outbound forwarding. Upstream name policy and command routing remain under [`OQ-SBX-21`](agent-open-questions.md#%60OQ-SBX-21%60), [`OQ-SBX-22`](agent-open-questions.md#%60OQ-SBX-22%60), [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Workspace, scratch, and caches

Host filesystem storage holds the agent workspace and scratch instead of a guest disk. Seatbelt grants the harness the permitted paths. A mount namespace doesn't hide other macOS paths. Host directory placement, ownership, capture access, and alias handling remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60).

Provider storage mappings and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandbox variables](agent-sandbox.md#Sandbox%20variables) defines separate scratch and host-shared named tool caches. Cache authorization, provenance, confidentiality, capacity, and invalidation remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Host-side processing of hostile repository metadata needs the isolation in [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). Worker capture rights and native history path stability remain under [`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60) and [`OQ-WSP-16`](agent-open-questions.md#%60OQ-WSP-16%60).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

[Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup) governs verified extraction and independent writable materialization. Image caches remain immutable and outside launch-writable storage, while materialization creates the writable tree separately. Agent workspace images contain directory trees. Their format is Open Container Initiative (OCI). Materialization creates a macOS-compatible directory tree for the confined host process. Native cloning or copying, ownership transfer, filesystem metadata, and a stable path for native harness state remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). A clone doesn't protect later host-side repository processing under [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60).

Native filesystem compatibility, including name aliases, links, and permission translation, remains under [`OQ-IMG-14`](agent-open-questions.md#%60OQ-IMG-14%60).

## Agent snapshots and agent checkpoints

This provider has no agent snapshot: it doesn't capture a private guest kernel or process memory. A native harness checkpoint is distinct from workspace content. Checkpoint and workspace pairing remain unresolved. Resume at a changed workspace path, restore authorization, retention, and old authority remain under [`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60). [Agent checkpoints](agent.md#Agent%20checkpoints) supplies the shared checkpoint model. A checkpoint can't undo an external effect.

## Cancellation and teardown

[Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) requires observed termination before a terminal state or slot release. A stop request or process-group exit alone doesn't prove every descendant has stopped.

Independent termination, worker and helper failure, runtime restart, and recovery remain under [`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60). Reusing launch identities while descendants survive remains under [`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60). Standalone command termination and capture ordering remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). macOS process-group termination doesn't cover descendants that detach from the group. Seatbelt and flow interception don't themselves provide an independent process-tree supervisor. Capture component shutdown and stale registration cleanup remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60).

## Host observations

The agent host controls process lifecycle, workspace backing, retained captures, and the agent mediator. Host administrators can inspect or change these resources. This provider supplies no confidentiality against the agent host. Independent evidence of a compromised host remains under [`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60). Same-user process access and the broader host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) limits exported content. It doesn't prevent local host access. Command attribution remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Network Extension providers observe flow origin and original destination. Audit tokens leave the per-launch user policy unresolved, and mediator telemetry remains subject to the shared content exclusions.

## Platform dependencies

This provider relies on Seatbelt, Network Extension transparent and name proxy providers, operating-system audit tokens, signed system extensions, and host service supervision. Supported entitlements, approvals, and distribution requirements remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60), [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60). Extension registration and recovery remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60). Reliable all-process teardown remains under [`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60).
