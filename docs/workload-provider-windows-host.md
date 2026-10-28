---
type: Design
title: "Workload provider for Windows host"
description: Isolation, mediation, storage, and lifecycle for the windows-host workload provider.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload provider for Windows host

[Public key infrastructure](pki.md).

## Platforms and privileges

The windows-host provider runs a Windows harness directly on Windows without a virtual machine (VM). The workload host worker runs under a dedicated non-administrator account. Windows Filtering Platform (WFP) capture requires a kernel callout driver and a user-mode service that installs filters and registers launches. Driver installation, signing, service privileges, supported editions, and hardware requirements remain under [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60). Service survival remains under [`OQ-SBX-11`](agent-open-questions.md#%60OQ-SBX-11%60).

## Isolation boundary

The agent harness and workload host share the Windows kernel. Network capture uses an AppContainer Security Identifier (SID) per launch, without settling the host isolation baseline or the complete native process sandbox. Filesystem rights, tokens, inherited handles, descendant confinement, and supported sandbox settings remain under [`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60). The host isolation baseline remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Process placement and startup

The workload host starts afakt-agent as a Windows process. afakt-agent starts afakt-sandbox and then the native harness. The kernel callout driver and user-mode registration service run outside harness confinement and map the launch’s AppContainer identity to its mediator endpoint. Their management authority and stale-registration handling remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60). Native process confinement remains under [`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60).

Commands started by `afakt agent exec` use standalone afakt-sandbox without the runner. afakt-agent is absent. Command placement, inherited environment, launch binding, capture ordering, and descendant lifetime remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). That choice also determines command transport. Authority for an unsandboxed command remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

Local helper authentication and binding to the intended launch remain under [`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60) and [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

For each workload, its proxy mediator and workload proxy follow [Launch boundary](workload-proxy.md#Launch%20boundary) and the authenticated [overlay connections](workload-proxy.md#Overlay%20connections).

The workload proxy is the only external traffic path, including Domain Name System (DNS) queries. It requests a mediator decision for each inbound and outbound connection. The proxy forwards DNS queries to the mediator, with no host resolver fallback. Encryption isn't required on the private link to the proxy. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft. Provider enforcement, including alternate interfaces, inherited descriptors, and indirect host services, remains under [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60). Host-mode capture sends traffic through the same proxy enforcement point. Containment applies against the agent at its sandbox privileges, with the mechanism held under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

[Workload identity](workload-host.md#Workload%20identity) keeps workload proxy credentials outside this workload provider’s confinement. Kernel-attested local transport remains the leading option in [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). Current-launch binding and sender attribution remain unresolved under [`OQ-SES-6`](agent-open-questions.md#%60OQ-SES-6%60). Control capacity remains under [`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60), while mediator placement and confinement remain under [`OQ-SBX-12`](agent-open-questions.md#%60OQ-SBX-12%60), [`OQ-SBX-19`](agent-open-questions.md#%60OQ-SBX-19%60), [`OQ-SBX-20`](agent-open-questions.md#%60OQ-SBX-20%60), [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60).

Registration authentication, destination-record integrity, and privileged driver and service containment remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60). The AppContainer network identity doesn't establish private host loopback. Listener exposure remains under [`OQ-SBX-16`](agent-open-questions.md#%60OQ-SBX-16%60). [Workload proxy egress](workload-proxy.md#Internet%20connections) carries permitted outbound traffic under mediator decisions. Upstream name policy and standalone command routing remain under [`OQ-SBX-21`](agent-open-questions.md#%60OQ-SBX-21%60), [`OQ-SBX-22`](agent-open-questions.md#%60OQ-SBX-22%60), [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

## Workspace, scratch, and caches

The agent workspace and scratch use Windows host storage. The launch’s native confinement must grant only the intended filesystem access. Linux mount and ownership rules don't choose Windows access rights. Directory layout, access mappings, link handling, and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60).

Provider storage mappings and capture access remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandbox variables](workload-sandbox.md#Sandbox%20variables) defines separate scratch and host-shared named tool caches. Cache authorization, provenance, confidentiality, capacity, and invalidation remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Host-side processing of hostile repository metadata needs the isolation in [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). Worker capture rights and native history path stability remain under [`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60) and [`OQ-WSP-16`](agent-open-questions.md#%60OQ-WSP-16%60).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

[Workload host caches and startup](agent-workspace-image.md#Workload%20host%20caches%20and%20startup) governs verified extraction and independent writable materialization. Immutable image caches remain outside launch-writable storage. Agent workspace images contain directory trees. Their format is Open Container Initiative (OCI). The directory tree needs Windows-compatible names, metadata, links, permissions, and native dependency artifacts. Native materialization, cloning or copying, and workspace path stability remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). Image verification alone doesn't supply the native process confinement in [`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60).

Native filesystem compatibility, including name aliases, links, and permission translation, remains under [`OQ-IMG-14`](agent-open-questions.md#%60OQ-IMG-14%60).

## Workload snapshots and agent checkpoints

This workload provider has no workload snapshot: it doesn't capture a private guest kernel or process memory. A native harness checkpoint is distinct from workspace content. Checkpoint and workspace pairing remain unresolved. Resume at a changed workspace path, restore authorization, retention, and old authority remain under [`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60). [Agent checkpoints](agent.md#Agent%20checkpoints) supplies the shared checkpoint model. A checkpoint can't undo an external effect.

## Cancellation and teardown

[Cancellation and limits](workload-lifecycle.md#Cancellation%20and%20limits) requires observed termination before a terminal state or slot release. A stop request or process-group exit alone doesn't prove every descendant has stopped.

Independent termination, worker and helper failure, runtime restart, and recovery remain under [`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60). Reusing launch identities while descendants survive remains under [`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60). Standalone command termination and capture ordering remain under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Native descendant tracking and process confinement remain under [`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60). A driver blocking new traffic doesn't stop processes or prove existing connections closed. Filter removal, registration retirement, and capture-service recovery remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60).

## Host observations

The workload host controls process lifecycle, workspace backing, retained captures, and the workload proxy mediator. Host administrators can inspect or change these resources. This workload provider supplies no confidentiality against the workload host. Independent evidence of a compromised host remains under [`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60). Same-user process access and the broader host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) limits exported content. It doesn't prevent local host access. Command attribution remains under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). After kernel or privileged service compromise, observations by the callout driver and service can't establish trustworthy evidence about launch network identities or original destinations.

## Platform dependencies

This workload provider relies on Windows Filtering Platform connect and bind redirection, persistent block filters, AppContainer identities, and a native process sandbox. Driver signing, distribution, memory-integrity support, and installation requirements remain under [`OQ-PRQ-7`](agent-open-questions.md#%60OQ-PRQ-7%60), [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60). The selected capture driver doesn't settle the native sandbox mechanism in [`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60).
