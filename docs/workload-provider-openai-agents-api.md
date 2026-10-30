---
type: Design
title: "Workload provider for OpenAI Agents API"
description: Remote isolation, mediation, storage, and lifecycle for OpenAI Agents API.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload provider for OpenAI Agents API

[Public key infrastructure](pki.md).

## Platforms and privileges

The OpenAI Agents Application Programming Interface (API) workload provider uses an OpenAI-hosted Linux workspace with preinstalled development tools. OpenAI provisions and connects it. Task and environment configuration come from the app. Documented sizes pair one virtual processor with one gigabyte of memory, two processors with four gigabytes, or four processors with sixteen gigabytes. Missing details include the guest architecture and the execution user's privileges as well as the kernel boundary and host privilege restrictions. Runtime qualification remains under [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60). [OpenAI-hosted environments](https://developers.openai.com/api/docs/guides/agents-api/environments/openai-hosted).

The workload host worker keeps provider management credentials outside the sandbox under [Runtime privileges and platforms](workload-host.md#Runtime%20privileges%20and%20platforms). afakt rejects settings it can't enforce under [Enforcement](workload-sandbox.md#Enforcement). The isolation baseline for workload hosts remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Isolation boundary

OpenAI runs the managed harness separately from the execution environment. A session-created hosted environment has a separate workspace, while more than one session can attach to a standalone pre-warmed environment and share its files and processes. Conversation histories remain separate. A vendor session doesn't always identify a distinct sandbox boundary. Selection and binding of an environment to one afakt launch remain under [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60), [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). The documentation doesn't state containment after a kernel or runtime compromise. [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60) covers that gap. [Agents API architecture](https://developers.openai.com/api/docs/guides/agents-api/architecture), [OpenAI-hosted environments](https://developers.openai.com/api/docs/guides/agents-api/environments/openai-hosted), [Pre-warmed environments](https://developers.openai.com/api/docs/guides/agents-api/environments/prewarmed).

## Process placement and startup

OpenAI runs the Codex harness and session orchestration on its servers. Shell and filesystem tools run in the environment, remote Model Context Protocol (MCP) tools connect from OpenAI by default, and function tools run in the app. Installing afakt-agent in the hosted workspace doesn't place the managed harness inside afakt-sandbox. [Agents API architecture](https://developers.openai.com/api/docs/guides/agents-api/architecture).

The agent launcher and runner ownership in [Host process sandbox](workload-sandbox.md#Host%20process%20sandbox) needs a managed-harness mapping under [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60). That question covers Agent Client Protocol (ACP), hooks, model requests, tool execution outside the sandbox, and vendor session identity. Replacement of those components for a managed harness remains undecided. Commands requested through `afakt workload exec` require standalone afakt-sandbox, with placement, inherited environment, launch binding, and descendant lifetime unresolved under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Protection of hooks installed by the workload provider and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60), [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

OpenAI prepares packages and files before setup commands, and a nonzero setup exit prevents agent startup. Environment templates save configuration, not a running workspace, and a session override can't broaden the template network policy. Mapping afakt's immutable launch spec and startup checks to that configuration remains under [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60). [OpenAI-hosted environments](https://developers.openai.com/api/docs/guides/agents-api/environments/openai-hosted).

## Mediator connection and name resolution

Each launch requires [agent mediation](workload-proxy-mediator.md#Launch%20boundary) for control, session, preview, and network traffic.

Remote launches require the [workload proxy](workload-proxy.md#Launch%20boundary) for egress, with its placement and carrier open under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60) and the workload provider mapping below.

Remote placement, mandatory transport, Domain Name System (DNS) mediation, and prevention of direct connections remain under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). Provider assertions for launch identity remain under [`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60).

OpenAI-hosted environments default to enabled outbound access unless a template supplies another policy. Disabled access blocks outbound traffic. Restricted access accepts one to 100 exact host names, without wildcards, schemes, paths, or ports. Subdomains and redirect destinations need separate entries. Hosted standard-input-and-output MCP servers require enabled access. Mandatory custom DNS resolution has no documented interface on the cited page, which also omits an afakt mediator carrier and routing of all managed service traffic through the workload proxy. These gaps remain under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60), [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60). [OpenAI-hosted environments](https://developers.openai.com/api/docs/guides/agents-api/environments/openai-hosted).

OpenAI supplies placeholders for vault credentials in the environment and replace unchanged request headers outside the sandbox for approved exact hosts over Hypertext Transfer Protocol Secure (HTTPS) on ports 443 or 8443. Credential hosts and network hosts are separate checks. Updating a vault credential requires a new session to change an existing hosted sandbox's credential. Vault deletion neither revokes upstream tokens nor stops a running session. MCP credentials can use vendor-managed Open Authorization (OAuth) refresh. Mapping these services to afakt's [Credential brokering](workload-sandbox.md#Credential%20brokering) and central refresh ownership remains under [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60), [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60), [`OQ-SBX-40`](agent-open-questions.md#%60OQ-SBX-40%60). Repository profile self-grants and credential domain bounds remain on hold under [`OQ-SBX-2`](agent-open-questions.md#%60OQ-SBX-2%60). [Vault credentials](https://developers.openai.com/api/docs/guides/agents-api/tools/vaults).

OpenAI also supports MCP connections that originate in the environment, but vault-backed MCP authentication applies only to service-origin connections. Environment-origin connections need inline authentication or a trusted proxy, leaving their afakt gateway mapping under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). [MCP tools](https://developers.openai.com/api/docs/guides/agents-api/tools/mcp).

## Workspace, scratch, and caches

The agent workspace needs an independent writable tree, separate scratch, and a verified workspace image cache outside launch-writable content. Workload provider backing, permissions, capture access, and cleanup remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). Shared named tool cache access and poisoning remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Hostile repository metadata during capture remains under [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). See [Sandbox variables](workload-sandbox.md#Sandbox%20variables) and [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture).

OpenAI-hosted files persist across turns while the environment exists. Published output artifacts are immutable copies that survive environment expiry. Deleting an artifact leaves its workspace file intact. Pre-warmed shared environments keep files, processes, packages, and environment configuration visible to every attached session. The documentation doesn't state afakt scratch or cache separation. The storage mapping of the workload provider must distinguish these shared environments from separate session-created workspaces under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60), [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). [Files and published artifacts](https://developers.openai.com/api/docs/guides/agents-api/environments/files), [Pre-warmed environments](https://developers.openai.com/api/docs/guides/agents-api/environments/prewarmed).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20workload%20provider%20storage) connects materialization to capture and cleanup.

Agent workspace images contain directory trees in Open Container Initiative (OCI) form. [Workload host caches and startup](agent-workspace-image.md#Workload%20host%20caches%20and%20startup) requires content verification, confined extraction, and independent writable materialization before hooks or the harness start. The remote storage and transfer mapping remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). A vendor's container image or saved environment configuration doesn't establish this workspace contract.

OpenAI-hosted environments accept packages, uploaded files, and setup commands. Templates keep that configuration. An afakt workspace-image materialization operation remains undocumented, including its layer verification and independent disk backing as well as external cache ownership. The mapping remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [OpenAI-hosted environments](https://developers.openai.com/api/docs/guides/agents-api/environments/openai-hosted).

## Workload snapshots and agent checkpoints

A workload snapshot captures virtual-machine state, while an agent checkpoint captures native harness session state. Using both requires pairing at an idle turn boundary under [agent checkpoints](agent.md#Agent%20checkpoints) and [suspend and resume](workload-lifecycle.md#Suspend%20and%20resume). Workload provider capture scope and complete session restoration remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60). Exclusion of active authority and capture authorization remain under [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60), [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). By itself, vendor persistence doesn't establish an afakt workload snapshot or agent checkpoint.

OpenAI retains session state and supports preparing a running environment before attaching sessions. The cited hosted environment and pre-warming interfaces don't state an exportable machine-memory snapshot or an atomic native-session checkpoint paired with workspace state. Published artifacts are file copies, not that paired capture. Those gaps remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60). [Agents API overview](https://developers.openai.com/api/docs/guides/agents-api/overview), [Pre-warmed environments](https://developers.openai.com/api/docs/guides/agents-api/environments/prewarmed), and [Files and published artifacts](https://developers.openai.com/api/docs/guides/agents-api/environments/files).

## Cancellation and teardown

The workload provider retains launch ownership until it observes termination of every process started for the launch. That observation must precede terminal status or slot release. [Cancellation and limits](workload-lifecycle.md#Cancellation%20and%20limits) and [Reconciliation after restart or retry](agent-launch.md#Reconciliation%20after%20restart%20or%20retry) distinguish a stop request, an unavailable provider, and observed exit. Remote termination evidence, retries, detached work, and replacement safety remain under [`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60). Credential retirement remains under [`OQ-SBX-40`](agent-open-questions.md#%60OQ-SBX-40%60). Session-end wait and identity lifetime remain on hold under [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

OpenAI session-created sandboxes receive keepalive messages between turns and OpenAI can delete them after an hour without activity or keepalive messages. Closing a stream doesn't cancel work. Session deletion requests cleanup, which may continue asynchronously. Shared pre-warmed environments instead have a twenty-minute idle lifetime after attachment, extended by activity, and support up to 128 attached sessions. Deleting a session only detaches it, even for the last session. The interface has no endpoint for deleting a standalone environment. Unattached environments expire after five minutes. These are distinct lifecycle contracts, and neither session deletion nor idle status proves afakt launch termination. [OpenAI-hosted environments](https://developers.openai.com/api/docs/guides/agents-api/environments/openai-hosted), [Manage sessions](https://developers.openai.com/api/docs/guides/agents-api/sessions/manage), [Pre-warmed environments](https://developers.openai.com/api/docs/guides/agents-api/environments/prewarmed).

## Host observations

The workload host worker needs launch-bound status, resource use, termination evidence, and capture results under [Provenance record](agent-launch-model.md#Provenance%20record). Vendor observations aren't independent proof against vendor compromise. [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) excludes secret values and workspace file contents from afakt telemetry. It doesn't constrain vendor custody. Vendor personnel access, internal copies, deletion completion, and independently protected evidence remain under [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60).

OpenAI retains session state and receives tool results through its managed harness. OpenAI-hosted workspace content and vault secrets also enter vendor custody. The Agents API supports data residency only in the United States. Zero Data Retention (ZDR) isn't available. In the cited pages, internal operator access bounds and physical erasure timing remain unstated, as does protection of workspace and vault secrets from vendor access. [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60) covers those gaps. [Agents API overview](https://developers.openai.com/api/docs/guides/agents-api/overview), [Agents API architecture](https://developers.openai.com/api/docs/guides/agents-api/architecture), [Vault credentials](https://developers.openai.com/api/docs/guides/agents-api/tools/vaults).

## Platform dependencies

The OpenAI Agents API workload provider depends on managed sessions, hosted environments, file and artifact operations, and any attached vault or tool service. Restricted app keys require separate agent and vault permissions. File limits include 50 inputs per creation request, five mebibytes per inline file, and 200 mebibytes per published artifact. The documented United States residency scope doesn't identify a selectable compute-region inventory or separate guarantees for every retained data class. Eligibility and capacity remain under [`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60). Verification of provider claims remains under [`OQ-PLC-9`](agent-open-questions.md#%60OQ-PLC-9%60). [Environment security](https://developers.openai.com/api/docs/guides/agents-api/environments/security), [Files and published artifacts](https://developers.openai.com/api/docs/guides/agents-api/environments/files), [Agents API overview](https://developers.openai.com/api/docs/guides/agents-api/overview).
