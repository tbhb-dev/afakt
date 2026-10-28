---
type: Design
title: "Workload provider for Windows VM"
description: Isolation, mediation, storage, and lifecycle for the windows-vm workload provider.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload provider for Windows VM

[Public key infrastructure](pki.md).

## Platforms and privileges

The windows-vm provider has two guest forms on Windows. Linux is the default guest. A Linux micro virtual machine (VM) runs inside Windows Subsystem for Linux (WSL2), while a Windows-native utility VM runs a Windows harness. Cloud Hypervisor uses nested Kernel-based Virtual Machine (KVM) support inside WSL2 for the Linux guest. The workload host worker runs under a dedicated non-administrator account. Setup, supported Windows editions and hardware, and service lifetime remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60).

## Isolation boundary

A private Linux guest kernel separates the Linux harness from the outer WSL2 environment. Windows-native execution uses a private Windows utility VM kernel. Both separate guest and host kernels. The outer virtualization environment and host-side virtual devices add trust boundaries whose containment remains under [`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60). Native guest process, filesystem, handle, and network confinement remain under [`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60).

## Process placement and startup

The workload host worker on Windows controls the selected guest. Linux execution starts Cloud Hypervisor inside WSL2, while Windows-native execution requires a utility VM adapter. The management channel, delegated permissions, startup authority, and independent supervision for both paths remain under [`OQ-SBX-60`](agent-open-questions.md#%60OQ-SBX-60%60). Inside either guest, afakt-agent starts afakt-sandbox and the harness. Linux guest init and its separate relay user follow [Guest layout and lifecycle control](workload-sandbox.md#Guest%20layout%20and%20lifecycle%20control). Windows guest startup and relay confinement remain under [`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60).

Commands started by `afakt agent exec` use standalone afakt-sandbox without the runner. afakt-agent is absent. Command placement, inherited environment, launch binding, capture ordering, and descendant lifetime remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). That choice also determines command transport. Authority for an unsandboxed command remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

Local helper authentication and binding to the intended launch remain under [`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60) and [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

For each workload, its proxy mediator and workload proxy follow [Launch boundary](workload-proxy.md#Launch%20boundary) and the authenticated [overlay connections](workload-proxy.md#Overlay%20connections).

The workload proxy is the only external traffic path, including Domain Name System (DNS) queries. It requests a mediator decision for each inbound and outbound connection. The proxy forwards DNS queries to the mediator, with no host resolver fallback. Encryption isn't required on the private link to the proxy. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft. Provider enforcement, including alternate interfaces, inherited descriptors, and indirect host services, remains under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60).

[Workload identity](workload-host.md#Workload%20identity) keeps workload proxy credentials outside this workload provider’s confinement. Kernel-attested local transport remains the leading option in [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). Current-launch binding and sender attribution remain unresolved under [`OQ-SES-6`](agent-open-questions.md#%60OQ-SES-6%60). Control capacity remains under [`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60), while mediator placement and confinement remain under [`OQ-SBX-12`](agent-open-questions.md#%60OQ-SBX-12%60), [`OQ-SBX-19`](agent-open-questions.md#%60OQ-SBX-19%60), [`OQ-SBX-20`](agent-open-questions.md#%60OQ-SBX-20%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60).

[Workload proxy egress](workload-proxy.md#Internet%20connections) carries permitted outbound traffic under mediator decisions. Upstream resolution and synthetic answers remain under [`OQ-SBX-21`](agent-open-questions.md#%60OQ-SBX-21%60), [`OQ-SBX-22`](agent-open-questions.md#%60OQ-SBX-22%60). Exact listener binding remains under [`OQ-PORT-6`](agent-open-questions.md#%60OQ-PORT-6%60). Standalone command routing remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Workspace, scratch, and caches

The Linux guest has separate read-only boot and writable workspace disks. Windows-native guest boot and workspace storage need a separate choice of Windows filesystem and access mapping because the Linux disk layout doesn't choose them. Both forms keep launch-writable state separate from verified image caches owned by the workload host. Windows, WSL2, and guest storage mappings remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60).

Provider storage mappings and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandbox variables](workload-sandbox.md#Sandbox%20variables) defines separate scratch and host-shared named tool caches. Cache authorization, provenance, confidentiality, capacity, and invalidation remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Host-side processing of hostile repository metadata needs the isolation in [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). Worker capture rights and native history path stability remain under [`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60) and [`OQ-WSP-16`](agent-open-questions.md#%60OQ-WSP-16%60).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

[Workload host caches and startup](agent-workspace-image.md#Workload%20host%20caches%20and%20startup) governs verified extraction and independent writable materialization. Immutable image caches remain outside launch-writable storage. Agent workspace images contain directory trees. Their format is Open Container Initiative (OCI). Linux materialization places workspace content in a guest disk. Windows-native execution needs compatible paths, metadata, ownership, and dependency artifacts. Mapping the image tree into native guest storage, including name collisions and link handling, remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). A Linux filesystem tree doesn't establish Windows execution compatibility.

Native filesystem compatibility, including name aliases, links, and permission translation, remains under [`OQ-IMG-14`](agent-open-questions.md#%60OQ-IMG-14%60).

## Workload snapshots and agent checkpoints

[Workload snapshots and the snapshot registry](agent.md#Workload%20snapshots%20and%20the%20snapshot%20registry) defines workload snapshots, while [Agent checkpoints](agent.md#Agent%20checkpoints) separates native harness state from the captured machine. Capture pairs an idle agent with frozen workspace state. Snapshot state, disk and memory consistency, rewind effects, private data and restored authority, retention, and owner binding remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60), [`OQ-AGT-4`](agent-open-questions.md#%60OQ-AGT-4%60), [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60), [`OQ-AGT-11`](agent-open-questions.md#%60OQ-AGT-11%60), [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). Linux capture targets the inner microVM rather than the entire WSL2 environment. Windows-native capture requires its own machine capture and restore adapter under [`OQ-SBX-60`](agent-open-questions.md#%60OQ-SBX-60%60). A disk checkpoint and a live-memory snapshot are different states. Both guest forms keep that distinction. Capture capability and restore semantics remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60).

## Cancellation and teardown

[Cancellation and limits](workload-lifecycle.md#Cancellation%20and%20limits) requires independent machine termination and observed exit before release. The Windows workload host needs an authority that can act after loss of the Linux management connection or refusal by the native utility VM. A successful management response alone doesn't establish stopped guest processes. The adapter and independent stop observations for both guest forms remain under [`OQ-SBX-60`](agent-open-questions.md#%60OQ-SBX-60%60). Worker and whole-host restart reconciliation follow [Restart and recovery](workload-host.md#Restart%20and%20recovery). Identity reuse and standalone commands remain under [`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60), [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Host observations

The workload host controls process lifecycle, workspace backing, retained captures, and the workload proxy mediator. Host administrators can inspect or change these resources. This workload provider supplies no confidentiality against the workload host. Independent evidence of a compromised host remains under [`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60). Same-user process access and the broader host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) limits exported content. It doesn't prevent local host access. Command attribution remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). The Linux path also exposes inner guest backing and runtime state to the outer WSL2 environment. The native path exposes them to Windows virtualization management.

## Platform dependencies

The Linux path needs nested virtualization, WSL2, Cloud Hypervisor, and a Linux guest. The native path needs Windows virtualization and a Windows utility VM image. Supported editions, architectures, image distribution, service permissions, and management interfaces remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60), [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-60`](agent-open-questions.md#%60OQ-SBX-60%60). Native guest confinement remains under [`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60).
