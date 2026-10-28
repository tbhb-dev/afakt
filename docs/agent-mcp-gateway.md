---
type: Design
title: Agent MCP gateway
description: One MCP endpoint for each afakt agent, with profile tool access, upstream credentials, server hosting, approvals, and audit.
tags: [afakt, agents, agent-launch, sessions, mcp]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent MCP gateway

[Public key infrastructure](pki.md).

afakt's Model Context Protocol (MCP) gateway serves durable [agents](agent.md) during their agent launches. A session retains the conversation across agent launches. It presents one MCP endpoint to each agent harness and routes permitted tool calls to upstream MCP servers or the built-in afakt MCP server. Outside MCP clients have no access to the gateway.

## Endpoint and harness integration

Headless agent harness integration remains [`OQ-SES-16`](agent-open-questions.md#%60OQ-SES-16%60).

The MCP gateway serves `/mcp` with MCP's Streamable Hypertext Transfer Protocol (HTTP) transport. From inside the sandbox, the agent harness connects through the guest relay, [agent mediator](agent-mediator.md#Launch%20boundary), and agent proxy over the overlay. After the mediator stamps launch identity, the agent proxy authenticates with the launch's X.509 SPIFFE Verifiable Identity Document (SVID) over mutual Transport Layer Security (TLS). At the receiving service, the peer certificate identifies the launch. Its immutable spec supplies the pinned profile, capabilities, and scopes. On each request, the MCP gateway checks those rights and refuses terminal launch identities. Neither the sandbox nor an upstream MCP server receives the SVID or private key. Built-in tools call the Application Programming Interface (API) server under that same launch authority. See [Per-launch services](agent-proxy.md#Launch%20boundary).

For ACP and headless agent launches, the harness integration points the harness at this one endpoint and turns off its own MCP server configuration. Other MCP server lists, including repository settings, can't add servers alongside the gateway. This restriction covers MCP server configuration alone. The scope of other repository project settings remains an [open question](agent-open-questions.md#Agent%20hooks). Before starting the harness, the integration applies the gateway settings and rejects a launch when it can't enforce them.

Each connection and request binds to the authenticated agent launch. A caller can't select another agent launch by supplying a session ID, tool argument, or protocol identifier. When the agent launch becomes terminal, the gateway rejects further requests under the same rules for agent launch identities as the API. SVID lifetime and the session-end wait remain [on hold](agent-open-questions.md#%60OQ-HOOK-18%60).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

## Server and tool access

`afakt agent mcp` manages servers attached to one agent. `afakt mcp-gateway` manages the deployment-configured servers. Whether an agent may attach a server outside that set remains [open](agent-open-questions.md#Agents).

The pinned [agent profile](agent-profile.md#MCP%20access) selects MCP servers from the deployment's server configuration and allowlists their tools. Servers absent from the server allowlist and tools absent from their server's tool allowlist are unavailable. Deny entries win over allow entries. An empty allowlist grants nothing.

Server and tool access may narrow under the agent launch spec. An omitted allowlist inherits the profile's list. Any supplied allowlist must be a subset of it. Deny lists from the profile and specification combine. A specification can't change upstream locations, credentials, hosting, or the profile's approval rules. Before dispatching each call, the gateway checks the effective lists again after filtering tool discovery. If a call specifies a hidden or unknown tool, the gateway denies it without contacting an upstream server.

Each exposed tool name identifies both the configured server and its upstream tool name. Through a reversible namespace mapping, the gateway keeps distinct server and tool pairs distinct, even when their names contain separators. A namespace belongs to a configured server identity, never its display label. If two servers offer a tool with the same name, their exposed names can't collide. The built-in server reserves the `afakt` server identity.

Agent harness-native tools remain subject to the profile's tools and harness permissions, while gateway enforcement doesn't depend on a harness firing a [tool hook](agent-hooks.md#Hook%20events). The boundary between harness-native tools and MCP tools is separate from the open question about routing native tools through the gateway.

## Upstream credentials

The MCP gateway holds upstream credentials outside the agent sandbox and uses a separate credential for each upstream service. It doesn't supply an upstream Application Programming Interface (API) key or Open Authorization (OAuth) access or refresh token directly to the harness. Preventing credentials from returning through upstream errors or results remains [`OQ-SBX-26`](agent-open-questions.md#%60OQ-SBX-26%60). Configuration and audit records contain credential references rather than secret values.

For remote servers that use [MCP authorization](https://modelcontextprotocol.io/docs/tutorials/security/authorization), afakt acts as the OAuth client. It completes the authorization flow once for the operator, or for each user when the upstream server requires per-user identity. afakt stores and refreshes the resulting tokens. Across agent launches, it reuses that authorization without another interactive sign-in for each tool call. Revoked or unusable authorization makes the upstream unavailable until afakt obtains valid authorization again.

Static API keys use the same [credential references](agent-sandbox.md#Credential%20brokering) as sandbox credential brokering. Where the upstream connection runs, the gateway resolves them. Credentials for an agent-host server remain outside the agent launch's sandbox. Credentialed network requests from that server use the host-side credential broker outside the guest, whose [placement remains open](agent-open-questions.md#%60OQ-SBX-23%60), with custody defined by each [credential type](agent-sandbox.md#Credential%20types%20and%20custody). Central server credentials stay on the control plane. Credential reference syntax and the fields of an `oauth` entry remain [open questions](agent-open-questions.md#Agent%20sandbox), as do the [bounds on repository profile self-grants and credential domains](agent-open-questions.md#Agent%20profile).

## Server hosting and lifecycle

Each server's deployment configuration selects its transport and hosting location. Profiles select configured servers rather than supplying commands or upstream endpoints.

| Server kind | Location and access | Lifecycle |
| --- | --- | --- |
| Remote HTTP | The gateway connects to the remote server over HTTP, with Transport Layer Security (TLS) for credentialed requests. | The remote service manages its processes. afakt maintains its client connections and authorization. |
| Agent-host stdio | A local server that needs the agent workspace runs inside the agent sandbox VM under the agent user, with that agent launch's workspace access and network restrictions. The gateway exchanges MCP messages over standard input and output. | The provider starts it with the agent launch and stops it with the agent launch, including verifying termination of its processes. |
| Central stdio | A shared server, such as an issue tracker integration, runs once in its own sandbox on the control plane. MCP messages pass between this server and the gateway through standard input and output. | afakt supervises the server and waits for readiness. It restarts failures with backoff and stops the server during shutdown. Agent launches share it. |

For agent-host servers, the guest relay and agent mediator forward gateway requests and responses over the launch's authenticated agent-proxy connection to the agent hub. The control plane never opens an inbound connection to an agent host. The provider manages the server's process lifecycle and keeps its standard input and output separate from the agent harness's streams. Session suspension stops the VM containing these servers. Resume restores or restarts them against the restored workspace and rechecks their channels for the new agent launch linked to its predecessor.

For a shared central server, the gateway scopes each request to its agent launch identity and effective tool policy along with its upstream authorization identity. Which central servers may share a process and how their internal state and credentials remain isolated is [`OQ-MCP-1`](agent-open-questions.md#%60OQ-MCP-1%60). The gateway keeps concurrent responses associated with their originating calls. By default, central servers don't receive an agent launch workspace. afakt permits all [agent sandbox provider modes](agent-host.md#Runtime%20privileges%20and%20platforms), including the weaker host modes. Central server confinement remains separate from the agent launch's sandbox.

## Tool approvals

An agent profile may mark selected tools as needing approval in `mcp.approvals`. After enforcing the allowlists and before dispatching an approved-only tool, the gateway raises an Agent Client Protocol (ACP) `session/request_permission` on the session or agent launch responsible for the call. The call waits for the answer. This request comes from the gateway independently of the agent harness transport and native command approval prompts.

The agent hub presents the request and its outcome in every attached client. This includes the web client and chat bridges. Under the [writer and reader rules](agent-acp.md#Writer%20and%20readers), only an attached writer can answer. Readers see updates but never gain approval authority by seeing the request.

For an interactive session without a writer, the request remains pending under [Requests from the agent](agent-acp.md#Requests%20from%20the%20agent). An unattended agent launch, including a headless agent launch without an attached writer, uses the pinned profile's `mcp.unattended_approval` rule: `deny` or `allow`, with `deny` as the default. ACP agent launches whose only writer is the agent launcher workflow use that same rule. Their attached readers can't answer. afakt records the request and the policy answer on the agent launch even without an attached client. These gateway requests are separate from the relay's handling of harness permission requests.

Approval permits only the pending call. It never overrides a denied tool, an API capability, a scope, or agent sandbox access. The gateway dispatches only after an allow answer for that call and rejects a denied or canceled request. Cancellation or termination ends a pending approval without dispatching the call. An answer to that ended request can't revive it.

## Built-in afakt MCP server

The gateway exposes a built-in afakt MCP server under the reserved `afakt` identity. Profiles allowlist its tools and mark them for approval just as they do for any other server.

| Tool operation | Authority check |
| --- | --- |
| Start a child agent launch | The calling agent launch must have permission to create agents from the requested template or role. The child's authority comes from its own role, pinned profile and specification, not from the caller's capabilities and scopes, so a caller without web access can start a child that has it. |
| Start a session | The same caller authority bounds the requested session and its profile. |
| Send a message to another session | The caller's `sessions.message` capability and target scopes apply, including the [agent message rules](agent-session-attachment.md#Agent%20messages). |
| Read another agent launch's status and results | The caller must have read access to that agent launch and its result or evidence. |
| Declare a forwarded port | The exposure belongs to the calling agent and agent launch. The agent mediator stamps the identity and evaluates policy for that port. The guest relay checks listener ownership and guest loopback scope under [Port discovery and records](agent-ports.md#Port%20discovery%20and%20records). |

Each call acts as the agent and agent launch authenticated by its mediator, with that agent launch's capabilities and scopes. An upstream credential or a gateway service identity never adds afakt authority. The API applies the same checks as for the [afakt command-line interface (CLI)](agent-profile.md#Capabilities). A caller can start a child only from a template or role it's permitted to use, so selecting a different child profile can't grant a child more authority than that template or role allows. Neither can tool approval.

## Audit and telemetry

Every tool call, including a denied call and a built-in afakt call, emits an `afakt.mcp.tool.call` span in the trace of the agent launch. It records the configured server, the namespaced tool, the outcome, and elapsed duration, including any approval wait. The gateway uses the authenticated agent launch's trace context rather than trusting a caller to choose another agent launch's trace. [MCP gateway telemetry](agent-observability.md#MCP%20gateway) defines span attributes, metrics, units, and cardinality bounds.

Each call also produces an audit record with the agent launch ID, session ID when present, call ID, server, tool, start and end times, duration, outcome, and any approval request and answer with its source. Audit records never contain arguments, results, or authentication material. Authentication material includes SVIDs and their private keys, which telemetry also excludes. Telemetry omits arguments and results by default under the profile's [telemetry content rule](agent-observability.md#What%20telemetry%20never%20records). Approval presentation to an authorized client doesn't permit content logging.

Metrics count tool calls by server, tool, and outcome, measure call duration in seconds, count approvals requested and resolved by outcome, and report upstream server health. Server and tool values come from deployment-bounded sets. Unknown values become `other`. Agent launch and session IDs, upstream uniform resource locators (URLs), arguments, results, and credential references never become metric attributes.
