---
type: Design
title: Agents
description: Durable agents, their lifecycle, owned resources, and operations across agent launches.
tags: [afakt, agents, lifecycle]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agents

[Public key infrastructure](pki.md).

An agent is a durable entity that represents the identity, configuration, and history of an artificial intelligence (AI) model capable of reasoning, planning, and interacting with its environment. Its lifecycle has one or more [agent launches](agent-launch.md), with at most one agent launch running at a time.

An agent launch instantiates the agent in an agent sandbox on an agent host. Retry, resume, migration, and crash recovery create a new agent launch linked to its predecessor. The agent's conversation forms a [session](agent-session.md). Its transcript remains after an agent launch ends, and the agent retains its current session and past session transcripts.

[Forks, archives, and transfer](agent-workspace-lifecycle.md#Forks,%20archives,%20and%20transfer).

## Agent spec

An agent spec is the agent's mutable desired configuration, typically created from an [agent profile](agent-profile.md). It configures the agent harness, model, credential authorizations, Model Context Protocol (MCP) servers, agent harness plugins, agent skills, agent sandbox, and system instructions. The control plane increments its generation on every change. Generation is the revision counter of the agent spec.

Each agent launch runs a frozen copy of the configuration in its [agent launch spec](agent-launch-contract.md#Agent%20launch%20spec). Any agent spec change creates a new generation. It ends the active agent launch and starts a new one. The change takes effect at a turn boundary. Handling work in flight and continuing the session remain [open](agent-open-questions.md#%60OQ-AGT-27%60).

## Agent change request

An agent never edits its own agent spec. To propose a change, an agent or a human submits an agent change request. Another agent reviews the request from an agent. The control plane applies it within a ceiling the operator sets. Repository profile self-grant and credential domain bounds remain [on hold](agent-open-questions.md#%60OQ-PRF-5%60).

## Agent launch request

An agent launch request asks to use an agent to perform a task. An agent, human, or automation may request an existing agent or an ephemeral agent that afakt deletes when the task completes. The [agent dispatcher workflow](components.md#Agent%20dispatcher) authorizes the request and finalizes the agent launch spec.

## Owned resources

| Resource | Relationship to the agent |
| --- | --- |
| Agent spec and profile binding | The mutable desired configuration and the [agent profile](agent-profile.md) used as its blueprint. Each agent launch records its frozen configuration and profile content. |
| Agent workspace and workspace revisions | Hold the agent's files and named points in their history. [Workspace sources](agent-workspace.md#Workspace%20sources) supply the directory tree for an agent launch. |
| Current session and past session transcripts | Hold the conversation and its retained history, independent of agent harness process lifetime. |
| Agent checkpoints | Mark positions to which the current session can rewind. |
| Agent launches | Record agent launches, their outcomes, evidence, and provenance. |
| Credentials loaded into its broker | Bind authorized credential references to [brokered use](agent-sandbox.md#Credential%20types%20and%20custody), with reflected response content unresolved. |
| Hooks | Add commands at [hook events](agent-hooks.md#Hook%20events), subject to profile rules. |
| MCP servers | Attach Model Context Protocol (MCP) services to the agent within its permitted access. |
| Skills | Supply reusable instructions and supporting resources to the harness. |
| Memory | Retains knowledge that the agent can query across sessions. |
| Artifacts | Associate stored inputs and outputs with the agent. |
| Mail | Holds the agent's incoming and outgoing agentmail messages. |
| Group memberships | Associate the agent with groups for announcements and coordination. |
| Subscriptions | Queue user messages upon delivery of events from a source. |
| Budget | Bounds allowed usage across the agent's work. |
| Labels and annotations | Store metadata for selection and reporting or descriptive context. Neither grants authority. |

Ownership of a reference doesn't grant access to its target. Profile capabilities, agent sandbox rules, and service authorization still bound each agent launch. The API server identifies the agent launch from the agent proxy's peer certificate and reads capabilities and scopes from the agent launch spec it already holds. See [Agent launch authentication](agent-proxy.md#Agent%20launch%20authentication). The agent's durable identity doesn't extend that authority past a terminal state.

[Agent states](agent-lifecycle.md#Agent%20states).

[Create, start, and run](agent-lifecycle.md#Create,%20start,%20and%20run).

## Attach and messages

`afakt agent attach` connects a terminal client to the agent's active conversation under the [writer and reader rules](agent-acp.md#Writer%20and%20readers). A workflow-controlled Agent Client Protocol (ACP) agent launch admits readers only, and headless agent launches don't offer attachment. `steer` interrupts the current turn with a user message. `queue-message` delivers a user message after the current turn. Both remain subject to session authority and transport support.

`history` reads or follows current and historical transcripts. The `session` commands list, rename, download, and delete retained transcripts. `clear` clears the current session, while `compact` reduces its conversation context. Clearing context doesn't itself delete the agent or its agent launch history.

[Suspend, resume, and restart](agent-lifecycle.md#Suspend,%20resume,%20and%20restart).

[Kill, signal, and quarantine](agent-lifecycle.md#Kill,%20signal,%20and%20quarantine).

## Forks

A fork copies an agent into a separate agent. `fork create` leaves the copy created without starting the agent harness. `fork start` starts the copy, and `fork run` also waits for its session to end and outputs the final message. `fork list` lists an agent's forks. The copy has its own identity. The scope of copied resources and authority remains open.

## Agent snapshots and the snapshot registry

An agent snapshot is a snapshot of the VM an agent sandbox runs in, used to restore it. Agent snapshots exist only where the agent runs in a VM. Host modes and container managers have none. `snapshot create` captures VM state and uploads it to the snapshot registry. The snapshot commands list, inspect, download, upload, tag, label, annotate, delete, and prune agent snapshots. An agent snapshot is separate from an agent checkpoint and a workspace image.

[Archive and restore](agent-lifecycle.md#Archive%20and%20restore).

## Export, import, and migrate

`export` packages an agent and its agent workspace as an Open Container Initiative (OCI) artifact. `import` creates an agent from that artifact. The transfer format doesn't imply permission to copy credentials or identity. Which owned resources travel in the artifact remains open.

`migrate` moves an agent to another agent host. The durable agent remains the same resource. Whether migration restores a compatible agent snapshot or reconstructs the transcript and workspace from portable state, and how it relates to agent launch placement, remain open.

[Delete and prune](agent-lifecycle.md#Delete%20and%20prune).

## Agent checkpoints

An agent checkpoint is a session-level capability of the agent harness: its own checkpoint or rewind of session state. The agent can create, inspect, list, delete, or clear agent checkpoints. `afakt agent checkpoint rewind` starts a new conversation branch from an agent checkpoint, retaining the original branch for audit. Rewind doesn't undo external effects. How it represents those effects remains [open](agent-open-questions.md#%60OQ-AGT-4%60).

[Checkpoint capture](agent-lifecycle.md#Checkpoint%20capture).

## Workspace revisions and the workspace registry

A portable workspace revision records a point in an agent's file history as an OCI diff layer plus a content-hashed file inventory. Content hashes determine correctness. Modification times and sizes never do. For an agent running in a VM, local workspace disk deltas use frozen block deltas instead. The agent workspace commands commit revisions, inspect them, reset to another revision, and tag them. Push publishes the agent's workspace to the workspace registry. Pull fetches and applies another revision. These operations concern the agent's workspace rather than the source repositories' own commit and push operations.

The workspace registry stores workspace content and revision references. It also supplies [workspace images and definitions](agent-workspace-image.md). `afakt workspace` manages those reusable images and definitions, while `afakt agent workspace` addresses one agent's revision history. Their relationship to agent-launch-specific workspace copies remains open.

## Agentmail, groups, and subscriptions

Agentmail is messaging addressed to durable agents. An agent has an inbox, outbox, and sent messages. Its mail commands draft, post, schedule, receive, acknowledge, inspect, undo, and verify messages. [Signed session messages](agent-session-attachment.md#Message%20signatures) supply a separate conversation-level mechanism. How agentmail uses that mechanism remains open.

An agent group is a named set of agents with membership, metadata, and events. Group broadcasts send announcements to members. A subscription connects an event source to an agent and queues user messages as deliveries arrive. Delivery ordering, retries, deduplication, and behavior while the recipient is inactive remain open for both mail and subscriptions.

## Memory, skills, and artifacts

Memory stores knowledge for reuse across the agent's sessions. Memory commands create, read, update, delete, copy, compact, and query entries. Storage, query semantics, and access boundaries remain open. Skills provide reusable instructions and resources. Adding a skill changes the agent spec and takes effect through a new agent launch, within the operator ceiling.

Stored content that can serve as input or output forms an artifact. Agent artifact commands manage associations with an agent, while `afakt artifact` manages the content resources and their metadata. The artifact proxy serves an artifact when `afakt artifact serve` publishes it through that service, but access checks, content lifetime, and the scope of serving remain open.

## Policies, roles, and organizational scope

A policy defines authorization rules. It can attach to roles, credentials, or agent sandbox configurations. A policy proposal represents a proposed change with explicit approval or rejection. Neither a proposal nor an agent's request for a change grants authority by itself.

A role groups associations with policies and authorizations for credentials. Within an organization, managed work shares an access scope. Projects group work within that scope. Requests record work to perform, complete or incomplete closure, and a reason for that closure. Agents work under these scopes without replacing their own profile or agent launch identity. Scope inheritance and the relationship between roles and the agent launch label named `role` remain open.

## Secrets, credentials, and trust material

A secret stores a sensitive value. Using secret material and authorization rules, a credential describes authenticated access. Credential references are fields on both the [agent profile](agent-profile.md#Profile%20fields) and the [agent launch](agent-launch-contract.md#Agent%20launch%20spec). Each reference specifies a credential in afakt's credential store. Attaching a credential to an agent permits brokered use without exposing its value to the agent harness. [Credential brokering](agent-sandbox.md#Credential%20brokering) takes references from the profile and agent launch. Attachments change the agent spec and take effect through a new agent launch at a turn boundary. Reference bounds remain [open](agent-open-questions.md#Agents).

On their own workstation, the operator runs `afakt credential add TYPE CREDENTIAL_NAME` from the command-line interface (CLI), with the credential type selecting the authentication flow. Acting as a client of the Application Programming Interface (API) server, which may run on another machine, the CLI runs any harness authentication command and sends its result to the server for storage. The server never runs that command. Interactive authentication requires the operator and never runs headless, and disabling interaction doesn't turn an interactive flow into a headless one. The verb `add` registers an object that originates outside afakt, as with credentials. With `create`, afakt makes an object that it originates, as with `afakt agent create`.

| Credential type | Authentication flow |
| --- | --- |
| `claude-code-oauth-token` | The CLI points Claude Code's configuration directory at a private temporary directory and runs `claude setup-token` interactively. After the operator finishes, the CLI reads the token from the command's output and sends it to the API server. It then removes the directory. The operator's own Claude Code sign-in remains unread and unchanged by the CLI. |
| `anthropic-api-key` | The CLI sends the supplied key to the API server for storage in the credential store. It doesn't run a harness command or perform another authentication flow. |
| `codex-login` | With Codex's configuration directory pointing at a private temporary directory, the CLI runs the device-code flow of `codex login` interactively. Once the operator finishes, the CLI reads the resulting auth file from that directory and sends it to the API server. It then removes the directory, never reading or changing the operator's own Codex sign-in. afakt's credential store contains one central refresh token on which the credential broker relies. |
| `openai-api-key` | For storage in the credential store, the CLI sends the API server the supplied key without running a harness command or performing another authentication flow. |

Each credential type defines what the control plane keeps and what reaches the host-side credential broker. Exchangeable credentials stay central, and the broker receives short-lived credentials issued in exchange. This applies to Open Authorization (OAuth) refresh tokens such as the Codex login. GitHub App keys and cloud role credentials follow the same pattern. For one agent launch, a static secret that the control plane can't exchange stays in the credential broker's memory outside the guest.

For login credentials, afakt's credential store is the only refresh owner. The broker refreshes through that store and never passes the token to the agent or shares a refresh lock with the operator's own CLI sign-in. Profiles, agent launch specs, agent sandbox configurations, records, and evidence contain references, never secret values.

Secure Shell (SSH) keys authenticate SSH connections or sign content, while SSH known hosts record trusted remote host keys. GNU (GNU's Not Unix) Privacy Guard (GPG) keys supply signing or encryption identities. Trusted Transport Layer Security (TLS) certificates establish certificate trust. Sandbox authorizations control which keys, keychain items, and credentials a sandbox may use. Key storage, signing delegation, rotation, and revocation remain open.

## Sandboxes

The [agent sandbox](agent-sandbox.md) confines the agent harness or its tools under the selected provider. Managed-loop placement remains [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60). `afakt agent sandbox` views that configuration. `afakt sandbox` manages agent sandbox configurations and standalone sandboxes, including command execution, network rules, and credential authorizations. A standalone sandbox runs commands without requiring an agent conversation. Its identity and lifecycle remain open, as does its relationship to agent launch authority.

## Budgets and usage

An agent budget bounds allowed spending or consumption. Usage records observed cost, tokens, and subscription usage across its work. `budget` sets or updates the budget, `limits` shows resource and usage limits, and `usage` reads consumption. Accumulation across agent launches, usage forecasts, and enforcement at exhaustion remain open alongside the [agent launch limit contract](agent-launch-contract.md#Agent%20launch%20spec).
