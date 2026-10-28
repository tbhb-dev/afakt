---
type: Design
title: "Agent sandbox provider for Claude Managed Agents"
description: Remote isolation, mediation, storage, and lifecycle for Claude Managed Agents.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent sandbox provider for Claude Managed Agents

[Public key infrastructure](pki.md).

## Platforms and privileges

Through its Application Programming Interface (API), the Claude Managed Agents provider uses Anthropic-hosted cloud sandboxes. Anthropic specifies isolated Linux containers on amd64 infrastructure with preinstalled development tools. Anthropic limits memory to eight gigabytes and disk to ten gigabytes. Execution user privileges, host privileges, and customer control over them aren't stated in the cloud reference. Those requirements remain under [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60). [cloud sandbox reference](https://platform.claude.com/docs/en/managed-agents/cloud-sandboxes-reference).

The agent host worker retains provider management authority outside the untrusted workspace under [Runtime privileges and platforms](agent-host.md#Runtime%20privileges%20and%20platforms). afakt rejects settings it can't enforce under [Enforcement](agent-sandbox.md#Enforcement). The host isolation baseline remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Isolation boundary

Claude Managed Agents creates a fresh isolated container for each session. Sessions don't share a filesystem. Anthropic runs the agent loop, session history, and vault secrets outside that container. When attached, shared memory stores can connect sessions. The documentation doesn't state the kernel or hypervisor mechanism, containment after a runtime compromise, or protection from Anthropic administrators. These gaps remain under [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60), [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60). [environments](https://platform.claude.com/docs/en/managed-agents/environments). [security boundaries and risks](https://platform.claude.com/docs/en/managed-agents/secure-deployment-risks).

## Process placement and startup

Claude Managed Agents runs bash and file tools in the cloud sandbox, web search and web fetch on Anthropic servers, and custom tools in the integrating app. The managed agent loop stays on Anthropic servers, so the cloud sandbox isn't the location of the whole harness. [security boundaries and risks](https://platform.claude.com/docs/en/managed-agents/secure-deployment-risks).

Installing afakt-agent in its tool container doesn't place the afakt runner around that managed loop. The locally controlled harness path uses [afakt-agent and afakt-sandbox](agent-sandbox.md#Host%20process%20sandbox). Mapping that ownership, Agent Client Protocol (ACP), hook dispatch, model calls, custom tools, and vendor session identity to an agent launch remains under [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60). afakt hasn't selected replacements for these components.

Commands requested through `afakt agent exec` require standalone afakt-sandbox, with placement, inherited environment, launch binding, and descendant lifetime unresolved under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60), [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

## Mediator connection and name resolution

Each launch requires [agent mediation](agent-workload-mediator.md#Launch%20boundary) for control, session, preview, and network traffic.

Remote launches require the [agent proxy](agent-workload-proxy.md#Launch%20boundary) and [global egress path](components.md#Agent%20egress%20gateway), with their placement and carrier subject to the provider mapping below.

Remote placement, mandatory transport, Domain Name System (DNS) mediation, and prevention of direct connections remain under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). Provider assertions for launch identity remain under [`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60).

Claude Managed Agents defaults to unrestricted networking for API-created environments and limited networking in Claude Studio. Limited networking accepts exact or wildcard hosts and separate package-manager and Model Context Protocol (MCP) allowances. With package-manager and MCP allowances off, an empty host list blocks host access, but file, memory, and GitHub resource attachments remain available. Limited host rules also apply to server-side web tools, while package-manager and MCP additions don't extend those tools' host list. Permitted hosts can receive uploads. A host allowlist doesn't constrain operations there. [cloud sandbox reference](https://platform.claude.com/docs/en/managed-agents/cloud-sandboxes-reference), [environments](https://platform.claude.com/docs/en/managed-agents/environments).

The Claude Managed Agents documentation doesn't state a way to force every cloud sandbox packet, name query, managed model request, server-side tool call, and preview stream through afakt's mediator and agent proxy. The unresolved transport includes the global egress path and authenticated overlay connections under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). Managed tool dispatch also remains under [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60). [environments](https://platform.claude.com/docs/en/managed-agents/environments). [security boundaries and risks](https://platform.claude.com/docs/en/managed-agents/secure-deployment-risks).

Claude Managed Agents vaults keep real environment credentials outside the sandbox and substitute placeholders on allowed destinations. Credential host permission is separate from network permission. Substitution can affect headers or bodies. Omitting injection locations at creation enables both. [Vaults](https://platform.claude.com/docs/en/managed-agents/vaults).

Vaults are workspace-scoped, and running sessions periodically resolve updated credential state. MCP credentials may use vendor-managed refresh. These mechanisms don't select afakt's [Credential brokering](agent-sandbox.md#Credential%20brokering) or its header-only injection and central refresh ownership. Their integration remains under [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60), [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60), [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60). Repository profile self-grants and credential domain bounds remain on hold under [`OQ-SBX-2`](agent-open-questions.md#%60OQ-SBX-2%60). [Vaults](https://platform.claude.com/docs/en/managed-agents/vaults).

## Workspace, scratch, and caches

The agent workspace needs an independent writable tree, separate scratch, and a verified image cache outside launch-writable content. Provider backing, permissions, capture access, and cleanup remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). Shared named tool cache access and poisoning remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Hostile repository metadata during capture remains under [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). See [Sandbox variables](agent-sandbox.md#Sandbox%20variables) and [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture).

Claude Managed Agents mounts uploaded files as read-only copies. Changed files use new sandbox paths. Session output files become downloadable through the Files API after upload completes. Sessions using the same environment can reuse cached package installations, but the environment configuration isn't versioned. Mappings from cached package installations and session storage to afakt scratch and named tool caches remain undocumented. Their relationship to verified workspace caches also remains undocumented. [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60) and [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60) cover those gaps. [adding files](https://platform.claude.com/docs/en/managed-agents/files), [environments](https://platform.claude.com/docs/en/managed-agents/environments).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

Agent workspace images contain directory trees in Open Container Initiative (OCI) form. [Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup) requires content verification, confined extraction, and independent writable materialization before hooks or the harness start. The remote storage and transfer mapping remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). A vendor runtime image or saved environment configuration doesn't establish this workspace contract.

Claude Managed Agents accepts environment package configuration and session file attachments. Its cited cloud environment and file interfaces don't state support for importing an afakt workspace image with verified layer ordering, ownership, and independent writable backing. This mapping remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [environments](https://platform.claude.com/docs/en/managed-agents/environments), [adding files](https://platform.claude.com/docs/en/managed-agents/files).

## Agent snapshots and agent checkpoints

An agent snapshot captures virtual-machine state, while an agent checkpoint captures native harness session state. Using both requires pairing at an idle turn boundary under [agent checkpoints](agent.md#Agent%20checkpoints) and [suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Provider capture scope and complete session restoration remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60). Exclusion of active authority and capture authorization remain under [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60), [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). By itself, vendor persistence doesn't establish an afakt agent snapshot or agent checkpoint.

Claude Managed Agents stores conversation history, sandbox state, and outputs server-side, but resuming a vendor session doesn't establish that capture contract. No customer operation for exporting a complete machine snapshot or pairing a native harness checkpoint with frozen workspace state appears in the cited session operations. [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60) and [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60) cover the gap. [Managed Agents overview](https://platform.claude.com/docs/en/managed-agents/overview), [session operations](https://platform.claude.com/docs/en/managed-agents/session-operations).

## Cancellation and teardown

The agent sandbox provider retains launch ownership until it observes termination of every process started for the launch. That observation must precede terminal status or slot release. [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) and [Reconciliation after restart or retry](agent-launch.md#Reconciliation%20after%20restart%20or%20retry) distinguish a stop request, an unavailable provider, and observed exit. Remote termination evidence, retries, detached work, and replacement safety remain under [`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60). Credential retirement remains under [`OQ-SBX-40`](agent-open-questions.md#%60OQ-SBX-40%60). Session-end wait and identity lifetime remain on hold under [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Claude Managed Agents distinguishes sessions in the `running`, `idle`, `rescheduled`, and `terminated` states. Archiving requires an idle session. Interrupt a running session first. Active workflow runs can prevent archiving or deletion even while the session is idle. Deletion removes session history, output files, and sandbox state, but separate uploads and attached resources require separate handling. The documentation doesn't state independent proof that every descendant or remote tool operation has stopped. [session operations](https://platform.claude.com/docs/en/managed-agents/session-operations), [reference](https://platform.claude.com/docs/en/managed-agents/reference).

## Host observations

The agent host worker needs launch-bound status, resource use, termination evidence, and capture results under [Provenance record](agent-launch-model.md#Provenance%20record). Vendor observations aren't independent proof against vendor compromise. [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) excludes secret values and workspace file contents from afakt telemetry. It doesn't constrain vendor custody. Vendor personnel access, internal copies, deletion completion, and independently protected evidence remain under [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60).

Anthropic holds managed conversation history, tool results, cloud workspace state, outputs, and attached vault secrets. Workspace API keys and console Developer or Admin roles can access session history until deletion. Managed Agents isn't eligible for Zero Data Retention (ZDR). In the cited pages, Anthropic doesn't state internal operator access bounds, physical erasure timing, or protections that keep cloud workspace or vault content confidential from the vendor. [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60) covers these gaps. [Managed Agents overview](https://platform.claude.com/docs/en/managed-agents/overview). [security boundaries and risks](https://platform.claude.com/docs/en/managed-agents/secure-deployment-risks), [vaults](https://platform.claude.com/docs/en/managed-agents/vaults).

Claude Managed Agents keeps sandbox state for thirty days after creation. Session history lasts until deletion, while archiving keeps that history. Uploaded files, memory stores, and vaults need separate deletion. [Session retention](https://platform.claude.com/docs/en/managed-agents/secure-deployment-monitoring).

## Platform dependencies

Claude Managed Agents depends on Anthropic's managed loop, session service, cloud containers, file service, and any attached vault or remote tool service. The documented organization limits include 300 create requests and 1,200 read requests per minute, alongside usage-tier limits. The cited cloud reference and service overview don't state a selectable cloud region list or geographic guarantees for each retained data class. Region, retention, account capacity, and placement eligibility remain under [`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60). Verification of provider claims remains under [`OQ-PLC-9`](agent-open-questions.md#%60OQ-PLC-9%60). [reference](https://platform.claude.com/docs/en/managed-agents/reference), [cloud sandbox reference](https://platform.claude.com/docs/en/managed-agents/cloud-sandboxes-reference), [Managed Agents overview](https://platform.claude.com/docs/en/managed-agents/overview).
