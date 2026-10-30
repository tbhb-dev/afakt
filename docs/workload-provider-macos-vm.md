---
type: Design
title: "Workload provider for macOS VM"
description: Isolation, mediation, storage, and lifecycle for the macos-vm workload provider.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload provider for macOS VM

[Public key infrastructure](pki.md).

## Platforms and privileges

The macos-vm provider runs a Linux virtual machine (VM) on macOS through Apple Virtualization. A dedicated non-root user runs the workload host worker. Setup privileges and service lifetime remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60). Supported platforms and distribution requirements remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60). Direct use of Apple Virtualization is the leading option in [`OQ-SBX-9`](agent-open-questions.md#%60OQ-SBX-9%60). Choosing a wrapper remains unresolved.

## Isolation boundary

Each agent launch has a separate Linux guest kernel. The guest boundary separates the agent harness from macOS, and afakt-sandbox adds process confinement inside the guest under [Host process sandbox](workload-sandbox.md#Host%20process%20sandbox). The workload host controls virtual devices and machine lifecycle. Confinement of the hypervisor and host-side device processing after guest exploitation remains under [`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60).

## Process placement and startup

The workload host starts the virtual machine under an independent host supervisor. Guest init starts afakt-agent, which starts afakt-sandbox and then the harness as the agent user. A guest relay runs as a separate guest user and controls the guest control endpoint under [Guest layout and lifecycle control](workload-sandbox.md#Guest%20layout%20and%20lifecycle%20control).

`afakt workload exec` is a workload provider capability. When it's available, the supervisor starts the command inside a fresh afakt-sandbox by default. Command placement, inherited environment, launch binding, capture ordering, and descendant lifetime remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). That choice also determines command transport. Authority for an unsandboxed command remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

Local helper authentication and binding to the intended launch remain under [`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60). Protection of hooks installed by the workload provider and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60) and [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

For each workload, its proxy mediator and workload proxy follow [Launch boundary](workload-proxy.md#Launch%20boundary) and the authenticated [overlay connections](workload-proxy.md#Overlay%20connections).

The workload proxy is the only external traffic path, including Domain Name System (DNS) queries. It requests a mediator decision for each inbound and outbound connection. The proxy forwards DNS queries to the mediator, with no host resolver fallback. Encryption isn't required on the private link to the proxy. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft. Workload provider enforcement, including alternate interfaces, inherited descriptors, and indirect host services, remains under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60).

[Workload identity](workload-host.md#Workload%20identity) keeps workload proxy credentials outside this workload provider’s confinement. Kernel-attested local transport remains the leading option in [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). Current-launch binding and sender attribution remain unresolved under [`OQ-SES-6`](agent-open-questions.md#%60OQ-SES-6%60). Control capacity remains under [`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60), while mediator placement and confinement remain under [`OQ-SBX-12`](agent-open-questions.md#%60OQ-SBX-12%60), [`OQ-SBX-19`](agent-open-questions.md#%60OQ-SBX-19%60), [`OQ-SBX-20`](agent-open-questions.md#%60OQ-SBX-20%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60).

[Workload proxy egress](workload-proxy.md#Internet%20connections) carries permitted outbound traffic under mediator decisions. Resolver upstreams and synthetic name answers remain under [`OQ-SBX-21`](agent-open-questions.md#%60OQ-SBX-21%60), [`OQ-SBX-22`](agent-open-questions.md#%60OQ-SBX-22%60). Listener and publication binding remain under [`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60), [`OQ-PORT-6`](agent-open-questions.md#%60OQ-PORT-6%60). Workspace command routing remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Workspace, scratch, and caches

The Linux guest has a read-only root disk and a separate writable workspace disk under [Guest layout and lifecycle control](workload-sandbox.md#Guest%20layout%20and%20lifecycle%20control). Through guest confinement, the harness accesses its attached workspace. Backing files and the verified workspace image cache remain under workload host control. macOS backing paths and guest-to-host ownership mappings remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60).

Workload provider storage mappings and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandbox variables](workload-sandbox.md#Sandbox%20variables) defines separate scratch and host-shared named tool caches. Cache authorization, provenance, confidentiality, capacity, and invalidation remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Host-side processing of hostile repository metadata needs the isolation in [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). Worker capture rights and native history path stability remain under [`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60) and [`OQ-WSP-16`](agent-open-questions.md#%60OQ-WSP-16%60).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20workload%20provider%20storage) connects materialization to capture and cleanup.

[Workload host caches and startup](agent-workspace-image.md#Workload%20host%20caches%20and%20startup) governs verified extraction and independent writable materialization. Immutable workspace image caches remain outside launch-writable storage. Agent workspace images contain directory trees. Their format is Open Container Initiative (OCI). Guest filesystem content comes from the workspace directory tree, separate from the Linux guest image. The choice of macOS cloning or copying for guest disk backing, storage accounting, and capture access remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). Linux dependency artifacts remain necessary for Linux execution.

## Workload snapshots and agent checkpoints

[Workload snapshots and the snapshot registry](agent.md#Workload%20snapshots%20and%20the%20snapshot%20registry) defines workload snapshots, while [Agent checkpoints](agent.md#Agent%20checkpoints) separates native harness state from the captured machine. Capture pairs an idle agent with frozen workspace state. Snapshot state, disk and memory consistency, rewind effects, private data and restored authority, retention, and owner binding remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60), [`OQ-AGT-4`](agent-open-questions.md#%60OQ-AGT-4%60), [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60), [`OQ-AGT-11`](agent-open-questions.md#%60OQ-AGT-11%60), [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). Apple Virtualization memory save and restore need a supported deployment. A disk-only substitute remains undecided. Capture remains under [`OQ-SBX-9`](agent-open-questions.md#%60OQ-SBX-9%60), [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60).

## Cancellation and teardown

The workload host supervisor can stop the virtual machine without cooperation from the harness, afakt-agent, or guest kernel under [Guest layout and lifecycle control](workload-sandbox.md#Guest%20layout%20and%20lifecycle%20control). Stopping the machine stops the guest processes together. [Workload provider observations](workload-lifecycle.md#Workload%20provider%20observations) requires the workload provider to observe that stop before release. Worker restart preserves the supervisor and machine. Machine restart reconciliation remains under [`OQ-HST-1`](agent-open-questions.md#%60OQ-HST-1%60). Host helper ownership and safe identity reuse remain under [`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60). Standalone commands require the placement and stop contract in [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Host observations

The workload host controls process lifecycle, workspace backing, retained captures, and the workload proxy mediator. Host administrators can inspect or change these resources. This workload provider supplies no confidentiality against the workload host. Independent evidence of a compromised host remains under [`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60). Same-user process access and the broader host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) limits exported content. It doesn't prevent local host access. Command attribution remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Platform dependencies

This workload provider relies on Apple Virtualization, a Linux guest kernel, host service supervision, and guest process confinement. Host signing, entitlements, supported hardware, and installation requirements remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60), [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60). Memory capture and direct framework use remain under [`OQ-SBX-9`](agent-open-questions.md#%60OQ-SBX-9%60). Host process containment remains under [`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60).
