---
type: CLI Reference
title: Command-line reference
description: Commands grouped by noun, with agent and deployment resource boundaries.
tags: [afakt, cli, agents, reference]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Command-line reference

The `afakt` command manages durable [agents](agent.md), their work, and deployment services. Angle brackets mark a required choice or value, and square brackets mark optional arguments.

## Resource boundaries

`afakt agent mcp` manages MCP servers attached to one agent. Use `afakt mcp-gateway` to configure the deployment's servers. Profile server and tool rules still apply. Whether an agent may attach a server outside deployment configuration remains [open](agent-open-questions.md#agents).

To view the agent's sandbox configuration, use `afakt agent sandbox`. `afakt sandbox` manages configurations and standalone sandboxes, including sandboxes that run commands without an agent conversation. See [Sandboxes](agent.md#sandboxes).

Manage revisions of one agent's workspace with `afakt agent workspace`. `afakt workspace` manages reusable workspace images and definitions. The workspace registry stores their content and references. See [Workspace revisions and the workspace registry](agent.md#workspace-revisions-and-the-workspace-registry).

## Configuration and indexing

Client and deployment settings don't override an agent profile's authority bounds. Configuration holds these settings. The explain operation identifies a setting's effective value and source. Use paths to list configuration sources and lint to check their validity.

An index organizes registered content for retrieval by agents. `afakt index add` registers content for indexing. Its storage, source types, and query interface remain [open](agent-open-questions.md#command-line-reference).

## Agent operations

See [Agents](agent.md).

| Command | Description |
| --- | --- |
| `afakt agent` | Manage durable agents and their work history. |
| `afakt agent annotate` | Update an agent's annotations. |
| `afakt agent archive` | Snapshot a stopped agent and move snapshot to archive. |
| `afakt agent attach` | Attach interactively to an agent in the terminal. |
| `afakt agent budget` | Set or update an agent's usage budget. |
| `afakt agent clear` | Clear an agent's session. |
| `afakt agent compact` | Compact the selected agent's session. |
| `afakt agent copy` | Copy files to or from an agent's workspace. |
| `afakt agent create` | Create an agent without starting the harness. |
| `afakt agent debug` | Fetch or follow an agent's debug logs. |
| `afakt agent delete` | Permanently delete a stopped agent (can't be undone). |
| `afakt agent diff` | Inspect changes to an agent's workspace. |
| `afakt agent env` | List an agent's environment variables. |
| `afakt agent events` | List or follow an agent's events. |
| `afakt agent exec` | Run a sandboxed or unsandboxed command in an agent's workspace. |
| `afakt agent export` | Export an agent and its workspace as an Open Container Initiative (OCI) artifact. |
| `afakt agent history` | Fetch or follow an agent's current and historical transcripts. |
| `afakt agent import` | Create an agent from an Open Container Initiative (OCI) artifact. |
| `afakt agent info` | View a summary or detailed information about an agent. |
| `afakt agent kill` | Stop the sandbox VM through the agent host supervisor. |
| `afakt agent label` | Update an agent's labels. |
| `afakt agent limits` | View details about an agent's resource and usage limits. |
| `afakt agent list` | List agents. |
| `afakt agent logs` | Read agent logs or follow new log entries. |
| `afakt agent migrate` | Move an agent to another host. |
| `afakt agent prune` | Permanently delete unused agents. |
| `afakt agent quarantine` | Remove network access from an agent. |
| `afakt agent queue-message` | Send a user message to an agent after its current turn. |
| `afakt agent rename` | Rename an agent. |
| `afakt agent restart` | Restart an agent's workspace and harness. |
| `afakt agent restore` | Start an agent from a snapshot. |
| `afakt agent result` | View an agent's final message. |
| `afakt agent resume` | Start a suspended agent. |
| `afakt agent run` | Start an agent, wait for its session to end, and output its final message. |
| `afakt agent sandbox` | View details about an agent's sandbox configuration (use `afakt sandbox` for managing sandbox configuration). |
| `afakt agent shell` | Start a sandboxed or unsandboxed shell in an agent's workspace. |
| `afakt agent signal` | Send a process signal to an agent. |
| `afakt agent start` | Create an agent and start the harness. |
| `afakt agent stats` | View an agent's statistics. |
| `afakt agent status` | View status information about an agent. |
| `afakt agent steer` | Interrupt an agent's current turn with a user message. |
| `afakt agent suspend` | Save a checkpoint at an idle turn boundary and stop the sandbox VM. |
| `afakt agent top` | Start a real-time view of agent activity. |
| `afakt agent unarchive` | Restore an archived agent. |
| `afakt agent unquarantine` | Restore network connectivity to an agent. |
| `afakt agent update` | Update an agent's attributes, may require restarting the harness. |
| `afakt agent usage` | View cost, token, and subscription usage. |
| `afakt agent wait` | Wait for an agent's session to end. |

## Agent artifact

See [Memory, skills, and artifacts](agent.md#memory-skills-and-artifacts).

| Command | Description |
| --- | --- |
| `afakt agent artifact` | Manage an agent's artifact associations. |
| `afakt agent artifact add` | Add an artifact to an agent. |
| `afakt agent artifact clear` | Remove all artifacts from an agent. |
| `afakt agent artifact list` | List an agent's artifacts. |
| `afakt agent artifact remove` | Remove an artifact from an agent. |

## Agent checkpoint

See [Checkpoints](agent.md#checkpoints).

| Command | Description |
| --- | --- |
| `afakt agent checkpoint` | Manage checkpoints in an agent's current session. |
| `afakt agent checkpoint clear` | Delete all checkpoints in an agent's current session. |
| `afakt agent checkpoint create` | Create a checkpoint in an agent's current session. |
| `afakt agent checkpoint delete` | Delete a checkpoint from the current session. |
| `afakt agent checkpoint info` | View details about a checkpoint. |
| `afakt agent checkpoint list` | List checkpoints for an agent's current session. |
| `afakt agent checkpoint rewind` | Rewind an agent's session to a checkpoint. |

## Agent client

See [Agent session attachment](agent-session-attachment.md).

| Command | Description |
| --- | --- |
| `afakt agent client` | Inspect and disconnect attached clients. |
| `afakt agent client disconnect` | Disconnect a client attached to an agent. |
| `afakt agent client info` | View details about a client attached to an agent. |
| `afakt agent client list` | List clients attached to an agent. |

## Agent credential

See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

| Command | Description |
| --- | --- |
| `afakt agent credential` | Manage credentials loaded into an agent broker. |
| `afakt agent credential attach` | Load a credential into an agent's credential broker. |
| `afakt agent credential clear` | Clear all credentials from an agent's credential broker. |
| `afakt agent credential detach` | Detach a credential from an agent's credential broker. |
| `afakt agent credential list` | List credentials loaded in an agent's credential broker. |

## Agent fork

See [Forks](agent.md#forks).

| Command | Description |
| --- | --- |
| `afakt agent fork` | Copy an agent and inspect its forks. |
| `afakt agent fork create` | Copy an agent without starting the harness. |
| `afakt agent fork list` | List an agent's forks. |
| `afakt agent fork run` | Copy an agent, start the harness, wait for its session to end, and output its final message. |
| `afakt agent fork start` | Copy an agent and start the harness. |

## Agent group

See [Agentmail, groups, and subscriptions](agent.md#agentmail-groups-and-subscriptions).

| Command | Description |
| --- | --- |
| `afakt agent group` | Manage agent groups and their memberships. |
| `afakt agent group add` | Add an agent to a group. |
| `afakt agent group broadcast` | Broadcast an announcement to an agent group. |
| `afakt agent group create` | Create an agent group. |
| `afakt agent group delete` | Delete an agent group. |
| `afakt agent group events` | List or follow an agent group's events. |
| `afakt agent group info` | View details about an agent group. |
| `afakt agent group list` | List agent groups. |
| `afakt agent group members` | List agents in an agent group. |
| `afakt agent group remove` | Remove an agent from a group. |
| `afakt agent group update` | Update an agent group. |

## Agent hook

See [Agent hooks](agent-hooks.md).

| Command | Description |
| --- | --- |
| `afakt agent hook` | Manage an agent's hooks. |
| `afakt agent hook add` | Add a hook to an agent. |
| `afakt agent hook disable` | Disable a hook added to an agent. |
| `afakt agent hook enable` | Turn on a hook attached to the agent. |
| `afakt agent hook events` | Fetch or follow an agent's hook events. |
| `afakt agent hook info` | View details about a hook. |
| `afakt agent hook list` | List hooks added to an agent. |
| `afakt agent hook remove` | Remove an attached hook from the agent. |

## Agent mail

See [Agentmail, groups, and subscriptions](agent.md#agentmail-groups-and-subscriptions).

| Command | Description |
| --- | --- |
| `afakt agent mail` | Exchange agentmail messages. |
| `afakt agent mail ack` | Acknowledge receipt of an agentmail message. |
| `afakt agent mail draft` | Create an agentmail message. |
| `afakt agent mail inbox` | View an agent's incoming agentmail messages. |
| `afakt agent mail message` | View an agentmail message. |
| `afakt agent mail outbox` | View an agent's outgoing agentmail messages. |
| `afakt agent mail post` | Send an agentmail message. |
| `afakt agent mail receive` | Wait for the next agentmail message. |
| `afakt agent mail schedule` | Choose when to send an agentmail message. |
| `afakt agent mail sent` | Show agentmail messages the agent has sent. |
| `afakt agent mail undo` | Unsend an agentmail message. |
| `afakt agent mail verify` | Verify the signature of an agentmail message. |

## Agent MCP

See [Agent MCP gateway](agent-mcp-gateway.md).

| Command | Description |
| --- | --- |
| `afakt agent mcp` | Manage an agent's MCP servers (local and gateway). |
| `afakt agent mcp add` | Add an MCP server to an agent. |
| `afakt agent mcp disable` | Disable an MCP server added to an agent. |
| `afakt agent mcp enable` | Turn on an MCP server attached to the agent. |
| `afakt agent mcp info` | View details about an MCP server added to an agent. |
| `afakt agent mcp list` | List an agent's MCP servers. |
| `afakt agent mcp login` | Authenticate an MCP server added to an agent. |
| `afakt agent mcp reconnect` | Reestablish the connection to an attached MCP server. |
| `afakt agent mcp remove` | Remove an MCP server from an agent. |
| `afakt agent mcp restart` | Stop and start an attached MCP server. |
| `afakt agent mcp status` | View the status of an MCP server added to an agent. |
| `afakt agent mcp update` | Update an attached MCP server. |

## Agent memory

See [Memory, skills, and artifacts](agent.md#memory-skills-and-artifacts).

| Command | Description |
| --- | --- |
| `afakt agent memory` | Manage an agent's retained memories. |
| `afakt agent memory compact` | Compact an agent's memories. |
| `afakt agent memory copy` | Copy agent memories from one target to another. |
| `afakt agent memory create` | Create an agent memory. |
| `afakt agent memory delete` | Delete an agent memory. |
| `afakt agent memory info` | View details about an agent's memory. |
| `afakt agent memory list` | Show the agent's retained memories. |
| `afakt agent memory query` | Find relevant agent memories. |
| `afakt agent memory update` | Update an agent memory. |

## Agent profile

See [Agent profile](agent-profile.md).

| Command | Description |
| --- | --- |
| `afakt agent profile` | Manage agent profiles. |
| `afakt agent profile create` | Create a new agent profile. |
| `afakt agent profile delete` | Delete an agent profile. |
| `afakt agent profile info` | View details about an agent profile. |
| `afakt agent profile list` | List agent profiles. |
| `afakt agent profile update` | Update an agent profile. |

## Agent session

See [Agent session](agent-session.md).

| Command | Description |
| --- | --- |
| `afakt agent session` | Manage retained session transcripts. |
| `afakt agent session delete` | Delete an agent's session transcript. |
| `afakt agent session download` | Download the selected session's transcript. |
| `afakt agent session list` | List an agent's session transcripts. |
| `afakt agent session rename` | Change a session transcript's name. |

## Agent skill

See [Memory, skills, and artifacts](agent.md#memory-skills-and-artifacts).

| Command | Description |
| --- | --- |
| `afakt agent skill` | Manage an agent's skills. |
| `afakt agent skill add` | Add a skill to an agent. |
| `afakt agent skill info` | View details about an agent's skill. |
| `afakt agent skill list` | List an agent's skills. |
| `afakt agent skill remove` | Remove a skill from an agent. |

## Agent snapshot

See [Snapshots and the snapshot registry](agent.md#snapshots-and-the-snapshot-registry).

| Command | Description |
| --- | --- |
| `afakt agent snapshot` | Manage archives of agent state in the snapshot registry. |
| `afakt agent snapshot annotate` | Update an agent snapshot's annotations. |
| `afakt agent snapshot create` | Create an archive of an agent's state and upload to the snapshot registry. |
| `afakt agent snapshot delete` | Delete an agent snapshot. |
| `afakt agent snapshot download` | Download an agent snapshot. |
| `afakt agent snapshot info` | View a summary of an agent snapshot. |
| `afakt agent snapshot label` | Update an agent snapshot's labels. |
| `afakt agent snapshot list` | List agent snapshots. |
| `afakt agent snapshot prune` | Delete old agent snapshots. |
| `afakt agent snapshot tag` | Tag an agent snapshot. |
| `afakt agent snapshot upload` | Upload an agent snapshot. |

## Agent subscription

See [Agentmail, groups, and subscriptions](agent.md#agentmail-groups-and-subscriptions).

| Command | Description |
| --- | --- |
| `afakt agent subscription` | Manage event deliveries to agents. |
| `afakt agent subscription create` | Subscribe an agent to an event source, queuing a user message to the agent when deliveries arrive. |
| `afakt agent subscription delete` | Unsubscribe an agent from an event source. |
| `afakt agent subscription deliveries` | Fetch or follow deliveries from a subscription. |
| `afakt agent subscription info` | View a subscription's details. |
| `afakt agent subscription list` | List subscriptions. |
| `afakt agent subscription update` | Update a subscription. |

## Agent workspace

See [Workspace revisions and the workspace registry](agent.md#workspace-revisions-and-the-workspace-registry).

| Command | Description |
| --- | --- |
| `afakt agent workspace` | Git-like operations against an agent's workspace. |
| `afakt agent workspace commit` | Create a local revision of an agent's workspace. |
| `afakt agent workspace info` | View details about an agent's workspace. |
| `afakt agent workspace pull` | Fetch and apply another revision of an agent's workspace from the workspace registry. |
| `afakt agent workspace push` | Update an agent's workspace in the workspace registry. |
| `afakt agent workspace reset` | Reset the state of an agent's workspace to a different revision. |
| `afakt agent workspace tag` | Create a tag referencing an agent's workspace at a specific revision. |

## Artifact

See [Memory, skills, and artifacts](agent.md#memory-skills-and-artifacts).

| Command | Description |
| --- | --- |
| `afakt artifact` | Manage stored artifacts and serve their content. |
| `afakt artifact annotate` | Update an artifact's annotations. |
| `afakt artifact copy` | Copy an artifact from one target to another. |
| `afakt artifact create` | Create an artifact. |
| `afakt artifact delete` | Delete an artifact. |
| `afakt artifact info` | View details about an artifact. |
| `afakt artifact label` | Update an artifact's labels. |
| `afakt artifact list` | List artifacts. |
| `afakt artifact rename` | Rename an artifact. |
| `afakt artifact serve` | Serve an artifact on the artifact proxy. |
| `afakt artifact update` | Update an artifact. |

## Certificate

See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

| Command | Description |
| --- | --- |
| `afakt certificate` | Manage trusted Transport Layer Security (TLS) certificates. |
| `afakt certificate add` | Add a trusted Transport Layer Security (TLS) certificate. |
| `afakt certificate info` | View details about a trusted Transport Layer Security (TLS) certificate. |
| `afakt certificate list` | List trusted Transport Layer Security (TLS) certificates. |
| `afakt certificate remove` | Remove the selected trusted Transport Layer Security (TLS) certificate. |

## Configuration

See [Configuration and indexing](#configuration-and-indexing).

| Command | Description |
| --- | --- |
| `afakt config` | Inspect and change afakt configuration. |
| `afakt config explain` | Explain the effective value and source of a setting. |
| `afakt config info` | Read a configuration value. |
| `afakt config lint` | Check configuration validity. |
| `afakt config list` | List configuration values. |
| `afakt config paths` | List configuration sources and lookup paths. |
| `afakt config set` | Set a configuration value. |
| `afakt config unset` | Remove a configuration override. |

## Credential

See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

| Command | Description |
| --- | --- |
| `afakt credential` | Manage authenticated service access. |
| `afakt credential add <claude-code-oauth-token\|anthropic-api-key\|codex-login\|openai-api-key> <credential-name> [--no-interactive]` | Add a named credential through the authentication flow its type selects. |
| `afakt credential delete` | Delete a credential. |
| `afakt credential info` | Inspect a credential without revealing its value. |
| `afakt credential list` | List credentials. |
| `afakt credential policy` | Manage credential authorization policies. |
| `afakt credential policy attach` | Attach a policy to a credential. |
| `afakt credential policy create` | Create a credential policy. |
| `afakt credential policy delete` | Delete the selected credential policy. |
| `afakt credential policy detach` | Detach a policy from a credential. |
| `afakt credential policy explain` | Explain a credential authorization decision. |
| `afakt credential policy info` | Inspect a credential policy. |
| `afakt credential policy list` | List credential policies. |
| `afakt credential policy update` | Update a credential policy. |
| `afakt credential read` | Read an authorized credential value. |
| `afakt credential update` | Update credential settings. |

## Deployment and client commands

See [Components](components.md).

| Command | Description |
| --- | --- |
| `afakt doctor` | Diagnose deployment and client configuration problems. |
| `afakt help` | Show command help. |
| `afakt info` | Show deployment details. |
| `afakt init` | Initialize afakt configuration. |
| `afakt logs` | Read or follow deployment logs. |
| `afakt report` | Produce a report of managed work and usage. |
| `afakt status` | Show deployment health and service status. |
| `afakt version` | Show the afakt build version. |
| `afakt whoami` | Show the authenticated caller and its identity. |

## GPG key

See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

| Command | Description |
| --- | --- |
| `afakt gpg-key` | Manage GNU Privacy Guard (GPG) signing and encryption keys. |
| `afakt gpg-key clear` | Delete all GPG keys. |
| `afakt gpg-key generate` | Generate a GPG key and add it to the GPG key store. |
| `afakt gpg-key import` | Import a GPG key. |
| `afakt gpg-key info` | View details about a GPG key. |
| `afakt gpg-key list` | List GPG keys. |
| `afakt gpg-key remove` | Remove a GPG key. |

## Index

See [Configuration and indexing](#configuration-and-indexing).

| Command | Description |
| --- | --- |
| `afakt index` | Register content for indexing. |
| `afakt index add` | Register content for retrieval by agents. |

## MCP gateway

See [Agent MCP gateway](agent-mcp-gateway.md).

| Command | Description |
| --- | --- |
| `afakt mcp-gateway` | Manage deployment-configured MCP servers. |
| `afakt mcp-gateway add` | Configure an upstream MCP server for the deployment. |
| `afakt mcp-gateway info` | Inspect a configured MCP server. |
| `afakt mcp-gateway list` | List deployment-configured MCP servers. |
| `afakt mcp-gateway login` | Authenticate to an upstream MCP server. |
| `afakt mcp-gateway reconnect` | Reconnect an upstream MCP server. |
| `afakt mcp-gateway remove` | Remove a deployment MCP server configuration. |
| `afakt mcp-gateway restart` | Restart a deployment-managed MCP server. |

## Organization

See [Policies, roles, and organizational scope](agent.md#policies-roles-and-organizational-scope).

| Command | Description |
| --- | --- |
| `afakt organization` | Manage organizational work scopes. |
| `afakt organization add` | Register an organization. |
| `afakt organization info` | Inspect an organization. |
| `afakt organization list` | List organizations. |
| `afakt organization remove` | Remove an organization registration. |

## Policy

See [Policies, roles, and organizational scope](agent.md#policies-roles-and-organizational-scope).

| Command | Description |
| --- | --- |
| `afakt policy` | Manage authorization policies. |
| `afakt policy create` | Create an authorization policy. |
| `afakt policy delete` | Delete an authorization policy. |
| `afakt policy explain` | Explain a policy decision. |
| `afakt policy info` | Inspect an authorization policy. |
| `afakt policy list` | List authorization policies. |
| `afakt policy update` | Update an authorization policy. |

## Policy proposal

See [Policies, roles, and organizational scope](agent.md#policies-roles-and-organizational-scope).

| Command | Description |
| --- | --- |
| `afakt policy-proposal` | Propose and review policy changes. |
| `afakt policy-proposal approve` | Approve a proposed policy change. |
| `afakt policy-proposal create` | Create a proposal to change a policy. |
| `afakt policy-proposal delete` | Delete the selected policy change proposal. |
| `afakt policy-proposal info` | View a proposed policy change. |
| `afakt policy-proposal list` | List proposed policy changes. |
| `afakt policy-proposal reject` | Decline the proposed change to a policy. |

## Port

See [Agent ports and previews](agent-ports.md).

| Command | Description |
| --- | --- |
| `afakt port` | Manage port publications and authenticated client tunnels. |
| `afakt port expose <port>` | Request Cedar authorization to publish a loopback listener from inside the agent VM through its guest relay and mediator. |
| `afakt port forward` | Open an authenticated tunnel to a forwarded port. |
| `afakt port list` | List accessible forwarded ports. |
| `afakt port unforward` | Close a client port tunnel. |

## Project

See [Policies, roles, and organizational scope](agent.md#policies-roles-and-organizational-scope).

| Command | Description |
| --- | --- |
| `afakt project` | Manage projects and their work. |
| `afakt project archive` | Archive a project. |
| `afakt project create` | Create a project. |
| `afakt project delete` | Delete a project. |
| `afakt project info` | Inspect a project. |
| `afakt project list` | List projects. |
| `afakt project update` | Update project attributes. |

## Repository

See [Workspace sources](agent-workspace.md#workspace-sources).

| Command | Description |
| --- | --- |
| `afakt repository` | Manage registered source repositories. |
| `afakt repository add` | Register a source repository. |
| `afakt repository info` | Inspect a registered source repository. |
| `afakt repository list` | List registered source repositories. |
| `afakt repository remove` | Remove a source repository registration. |

## Request

See [Policies, roles, and organizational scope](agent.md#policies-roles-and-organizational-scope).

| Command | Description |
| --- | --- |
| `afakt request` | Manage requests for work. |
| `afakt request close <--complete\|--incomplete> --reason=TEXT` | Close a work request with an outcome and reason. |
| `afakt request create` | Create a work request. |
| `afakt request update` | Update a work request. |
| `afakt request view` | View a work request. |

## Role

See [Policies, roles, and organizational scope](agent.md#policies-roles-and-organizational-scope).

| Command | Description |
| --- | --- |
| `afakt role` | Manage authorization roles. |
| `afakt role create` | Create an authorization role. |
| `afakt role credential` | Manage role credential authorizations. |
| `afakt role credential authorize` | Permit a role to use a credential. |
| `afakt role credential clear` | Clear a role's credential authorizations. |
| `afakt role credential list` | Show the role's grants for credential use. |
| `afakt role credential revoke` | Revoke a role's credential authorization. |
| `afakt role delete` | Delete an authorization role. |
| `afakt role explain` | Explain a role's effective authority. |
| `afakt role info` | Inspect an authorization role. |
| `afakt role list` | List authorization roles. |
| `afakt role policy` | Manage role policy associations. |
| `afakt role policy attach` | Attach a policy to a role. |
| `afakt role policy clear` | Clear a role's policy associations. |
| `afakt role policy detach` | Detach a policy from a role. |
| `afakt role policy list` | Show policy associations for the role. |
| `afakt role update` | Update an authorization role. |

## Sandbox

See [Agent sandbox](agent-sandbox.md).

| Command | Description |
| --- | --- |
| `afakt sandbox` | Manage sandbox configurations and standalone sandboxes. |
| `afakt sandbox cap` | Manage sandbox capabilities. |
| `afakt sandbox cap disable` | Disable a capability for a sandbox. |
| `afakt sandbox cap enable` | Turn on a capability in the sandbox. |
| `afakt sandbox cap list` | List a sandbox's capabilities. |
| `afakt sandbox check` | Check sandbox configuration validity. |
| `afakt sandbox copy` | Copy a sandbox configuration. |
| `afakt sandbox create` | Create a standalone sandbox configuration. |
| `afakt sandbox credential` | Manage sandbox credential authorizations. |
| `afakt sandbox credential authorize` | Allow the sandbox to use a credential. |
| `afakt sandbox credential list` | List credentials allowed for a sandbox. |
| `afakt sandbox credential revoke` | Disallow the sandbox from using a credential. |
| `afakt sandbox decisions` | Fetch or follow a sandbox's authorization decisions. |
| `afakt sandbox delete` | Delete a sandbox. |
| `afakt sandbox doctor` | Diagnose sandbox issues. |
| `afakt sandbox exec` | Run a command in a sandbox. |
| `afakt sandbox explain` | Explain a sandbox authorization decision. |
| `afakt sandbox gpg-key` | Manage sandbox GPG key authorizations. |
| `afakt sandbox gpg-key authorize` | Allow a sandbox to use a GPG key. |
| `afakt sandbox gpg-key list` | List GPG keys allowed for a sandbox. |
| `afakt sandbox gpg-key revoke` | Disallow a sandbox from using a GPG key. |
| `afakt sandbox info` | View details about a sandbox. |
| `afakt sandbox keychain` | Manage keychain authorizations for a sandbox. |
| `afakt sandbox keychain authorize` | Permit the sandbox to use a keychain item. |
| `afakt sandbox keychain list` | List keychain items allowed for a sandbox. |
| `afakt sandbox keychain revoke` | Block the sandbox's use of a keychain item. |
| `afakt sandbox list` | List sandboxes. |
| `afakt sandbox logs` | Fetch or follow a sandbox's logs. |
| `afakt sandbox network` | Manage network access for a sandbox. |
| `afakt sandbox network disable` | Disable a sandbox's network access. |
| `afakt sandbox network enable` | Turn on network access for the sandbox. |
| `afakt sandbox network explain` | View how a sandbox's network rules allow or deny an endpoint. |
| `afakt sandbox network filter` | Manage network filters for a sandbox. |
| `afakt sandbox network filter add` | Add a network filter to a sandbox. |
| `afakt sandbox network filter info` | Inspect a network filter's details. |
| `afakt sandbox network filter list` | List network filters in a sandbox. |
| `afakt sandbox network filter remove` | Remove a network filter from a sandbox. |
| `afakt sandbox network rule` | Manage network rules for a sandbox. |
| `afakt sandbox network rule allow` | Add a network allow rule to a sandbox. |
| `afakt sandbox network rule deny` | Add a rule that denies network access in the sandbox. |
| `afakt sandbox network rule info` | Show details of the selected rule. |
| `afakt sandbox network rule list` | List a sandbox's network rules. |
| `afakt sandbox network rule revoke` | Delete a sandbox network rule. |
| `afakt sandbox policy` | Manage sandbox policies. |
| `afakt sandbox policy attach` | Associate a global policy with a sandbox. |
| `afakt sandbox policy create` | Create a local policy in a sandbox. |
| `afakt sandbox policy delete` | Delete a local policy from a sandbox. |
| `afakt sandbox policy detach` | Dissociate a global policy from a sandbox. |
| `afakt sandbox policy list` | List a sandbox's local and attached policies. |
| `afakt sandbox policy save` | Save a sandbox-local policy as a global policy. |
| `afakt sandbox port` | Manage forwarded sandbox ports. |
| `afakt sandbox port close` | Close a port forwarded from the port proxy to the sandbox. |
| `afakt sandbox port expose` | Forward a port from the port proxy to the sandbox. |
| `afakt sandbox port list` | List ports that the port proxy forwards to the sandbox. |
| `afakt sandbox rename` | Rename a sandbox configuration. |
| `afakt sandbox run` | Create a standalone sandbox configuration and execute a command in it. |
| `afakt sandbox ssh-key` | Manage sandbox SSH key authorizations. |
| `afakt sandbox ssh-key authorize` | Allow a sandbox to use an SSH key. |
| `afakt sandbox ssh-key list` | List SSH keys allowed for a sandbox. |
| `afakt sandbox ssh-key revoke` | Disallow a sandbox from using an SSH key. |
| `afakt sandbox test` | Run sandbox configuration tests. |
| `afakt sandbox update` | Update a sandbox. |

## Secret

See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

| Command | Description |
| --- | --- |
| `afakt secret` | Manage stored sensitive values. |
| `afakt secret annotate` | Update a secret's annotations. |
| `afakt secret create` | Store a sensitive value. |
| `afakt secret delete` | Delete a secret. |
| `afakt secret label` | Update a secret's labels. |
| `afakt secret list` | List secret metadata. |
| `afakt secret prune` | Delete unused secrets. |
| `afakt secret rename` | Rename a secret. |
| `afakt secret update` | Update a secret value or attributes. |

## Server

See [Architecture](architecture.md).

| Command | Description |
| --- | --- |
| `afakt server [--target=<all\|api\|dashboard\|webhook\|worker>] [--temporal-mode=<embedded\|external>] [--temporal-address=<address>]` | Run selected API, dashboard, webhook, and worker services with embedded or external Temporal. |

## SSH host

See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

| Command | Description |
| --- | --- |
| `afakt ssh-host` | Manage SSH known hosts. |
| `afakt ssh-host add` | Add an SSH known host. |
| `afakt ssh-host clear` | Delete all SSH known hosts. |
| `afakt ssh-host info` | View details about an SSH known host. |
| `afakt ssh-host list` | List SSH known hosts. |
| `afakt ssh-host remove` | Remove the selected known host. |

## SSH key

See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

| Command | Description |
| --- | --- |
| `afakt ssh-key` | Manage SSH keys. |
| `afakt ssh-key clear` | Delete all SSH keys. |
| `afakt ssh-key generate` | Generate an SSH key and add it to the SSH key store. |
| `afakt ssh-key import` | Add an SSH key. |
| `afakt ssh-key info` | View details about an SSH key. |
| `afakt ssh-key list` | List SSH keys. |
| `afakt ssh-key remove` | Remove an SSH key. |

## Workspace

See [Agent workspace image](agent-workspace-image.md).

| Command | Description |
| --- | --- |
| `afakt workspace` | Manage reusable workspace images and definitions. |
| `afakt workspace annotate` | Update a workspace definition's annotations. |
| `afakt workspace build` | Build a workspace image from its definition. |
| `afakt workspace create` | Create a workspace image definition. |
| `afakt workspace delete` | Delete a workspace image or definition. |
| `afakt workspace export` | Export workspace image content. |
| `afakt workspace fork` | Copy a workspace image definition. |
| `afakt workspace import` | Import content for a workspace image. |
| `afakt workspace label` | Update a workspace definition's labels. |
| `afakt workspace list` | List workspace images and definitions. |
| `afakt workspace manifest` | Inspect a workspace image manifest. |
| `afakt workspace prune` | Delete unused workspace image content. |
| `afakt workspace rename` | Rename a workspace image definition. |
| `afakt workspace stats` | Show workspace image storage and build statistics. |
| `afakt workspace tag` | Tag a workspace image. |
| `afakt workspace update` | Update a workspace image definition. |
