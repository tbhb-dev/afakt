---
type: Design
title: "Workload provider for Linux host"
description: Isolation, mediation, storage, and lifecycle for the linux-host workload provider.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload provider for Linux host

[Public key infrastructure](pki.md).

## Platforms and privileges

The linux-host provider runs the harness directly on Linux without a virtual machine (VM). A dedicated non-root user runs the workload host worker. Setup privileges and service lifetime remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60). Supported platforms and distribution requirements remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60). Launch confinement uses unprivileged namespaces with a network tunnel device. Systems that can't enforce the selected configuration can't silently fall back to unrestricted host networking.

## Isolation boundary

The agent harness shares the Linux kernel with the workload host. A private network namespace and a tunnel device constrain external traffic. bubblewrap and Landlock both constrain filesystem access, with seccomp restricting system calls. Landlock doesn't replace the private network namespace. Namespace and mount construction must exclude host service sockets and inherited host-network descriptors. Complete handle and service coverage remains under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60), [`OQ-SBX-28`](agent-open-questions.md#%60OQ-SBX-28%60). The isolation baseline for workload hosts and per-launch users remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Process placement and startup

The workload host starts afakt-agent on Linux. afakt-agent starts afakt-sandbox, which establishes the network and filesystem confinement before the harness runs. Outside harness control, the tunnel endpoint sends packets to the workload proxy mediator. Authority for endpoint construction, registration, and retirement remains under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60).

`afakt workload exec` is a workload provider capability. When it's available, the supervisor starts the command inside a fresh afakt-sandbox by default. Command placement, inherited environment, launch binding, capture ordering, and descendant lifetime remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). That choice also determines command transport. Authority for an unsandboxed command remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

Local helper authentication and binding to the intended launch remain under [`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60). Protection of hooks installed by the workload provider and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60) and [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

For each workload, its proxy mediator and workload proxy follow [Launch boundary](workload-proxy.md#Launch%20boundary) and the authenticated [overlay connections](workload-proxy.md#Overlay%20connections).

All external traffic passes through the workload proxy, which requests a mediator decision for each inbound and outbound connection and forwards Domain Name System (DNS) queries to the mediator. Queries never fall back to the host resolver. Encryption isn't required on the private link to the proxy. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft. Workload provider enforcement, including alternate interfaces, inherited descriptors, and indirect host services, remains under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60). Host-mode capture sends traffic through the same proxy enforcement point. Containment applies against the agent at its sandbox privileges, with the mechanism held under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

[Workload identity](workload-host.md#Workload%20identity) keeps workload proxy credentials outside this workload provider’s confinement. Kernel-attested local transport remains the leading option in [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). Current-launch binding and sender attribution remain unresolved under [`OQ-SES-6`](agent-open-questions.md#%60OQ-SES-6%60). Control capacity remains under [`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60), while mediator placement and confinement remain under [`OQ-SBX-12`](agent-open-questions.md#%60OQ-SBX-12%60), [`OQ-SBX-19`](agent-open-questions.md#%60OQ-SBX-19%60), [`OQ-SBX-20`](agent-open-questions.md#%60OQ-SBX-20%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60).

Complete descriptor and address-family coverage, indirect host services, and attachment of workspace commands remain under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60), [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Mediator loss cuts off external routing, but binding within the namespace still needs the listener policy in [`OQ-SBX-16`](agent-open-questions.md#%60OQ-SBX-16%60). A private namespace alone doesn't prohibit wildcard binds. Tunnel registration and reuse remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60). Upstream name policy remains under [`OQ-SBX-21`](agent-open-questions.md#%60OQ-SBX-21%60), [`OQ-SBX-22`](agent-open-questions.md#%60OQ-SBX-22%60), and [Workload proxy egress](workload-proxy.md#Internet%20connections) carries permitted outbound traffic under mediator decisions.

## Workspace, scratch, and caches

Host directory storage holds the agent workspace. The confined process mount view exposes it to the harness. Scratch remains a separate allocation, and shared named tool caches require explicit attachment to that view. Ownership, filesystem backing, and worker capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60).

Workload provider storage mappings and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandbox variables](workload-sandbox.md#Sandbox%20variables) defines separate scratch and host-shared named tool caches. Cache authorization, provenance, confidentiality, capacity, and invalidation remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Host-side processing of hostile repository metadata needs the isolation in [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). Worker capture rights and native history path stability remain under [`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60) and [`OQ-WSP-16`](agent-open-questions.md#%60OQ-WSP-16%60).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20workload%20provider%20storage) connects materialization to capture and cleanup.

[Workload host caches and startup](agent-workspace-image.md#Workload%20host%20caches%20and%20startup) governs verified extraction and independent writable materialization. Immutable workspace image caches remain outside launch-writable storage. Agent workspace images contain directory trees. Their format is Open Container Initiative (OCI). Materialization creates an independent Linux directory tree, which the process mount view exposes to the harness. Native filesystem cloning or full copying and the mapping between worker and launch ownership remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). Read-only mounts don't prevent connections to a visible host service socket. Materialization and socket exclusion have separate responsibilities.

## Workload snapshots and agent checkpoints

This workload provider has no workload snapshot: it doesn't capture a private guest kernel or process memory. A native harness checkpoint is distinct from workspace content. Checkpoint and workspace pairing remain unresolved. Resume at a changed workspace path, restore authorization, retention, and old authority remain under [`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60). [Agent checkpoints](agent.md#Agent%20checkpoints) supplies the shared checkpoint model. A checkpoint can't undo an external effect.

## Cancellation and teardown

[Cancellation and limits](workload-lifecycle.md#Cancellation%20and%20limits) requires observed termination before a terminal state or slot release. A stop request or process-group exit alone doesn't prove every descendant has stopped.

Independent termination, worker and helper failure, runtime restart, and recovery remain under [`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60). Reusing launch identities while descendants survive remains under [`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60). Standalone command termination and capture ordering remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). A private process namespace can bound descendants, but the authority that owns and verifies the complete namespace lifetime remains unresolved. Harness or parent exit is insufficient. Tunnel retirement and stale registration cleanup remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60).

## Host observations

The workload host controls process lifecycle, workspace backing, retained captures, and the workload proxy mediator. Host administrators can inspect or change these resources. This workload provider supplies no confidentiality against the workload host. Independent evidence of a compromised host remains under [`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60). Same-user process access and the broader host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) limits exported content. It doesn't prevent local host access. Command attribution remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Administrators can inspect the process tree despite private namespace membership. The workload host also controls the namespace and tunnel endpoint and can observe captured packet destinations.

## Platform dependencies

This workload provider relies on Linux user, network, mount, and process namespaces, tunnel devices, bubblewrap, Landlock, and seccomp. Required kernel support and device permissions remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60), [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60). Landlock availability without a private network namespace doesn't enforce this provider’s selected confinement.
