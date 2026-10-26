---
type: Design
title: Agent MCP gateway
description: One MCP endpoint for each afakt agent, with profile tool access, upstream credentials, server hosting, approvals, and audit.
tags: [afakt, agents, execution, sessions, mcp]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent MCP gateway

afakt's Model Context Protocol (MCP) gateway serves durable [agents](agent.md) during their harness runs. A session retains the conversation across runs. It presents one MCP endpoint to each harness and routes permitted tool calls to upstream MCP servers or the built-in afakt MCP server. Outside MCP clients have no access to the gateway.

## Endpoint and harness integration

The gateway serves `/mcp` through the afakt API with MCP's Streamable HTTP transport, and the harness connects from inside the sandbox VM through the guest relay and [mediator](agent-sandbox.md#per-agent-mediator). The mediator stamps the agent's identity and supplies its execution token outside the guest. Then the gateway validates its execution ID, generation, pinned profile, capabilities, and scopes on every request. This endpoint belongs to the afakt API audience of the [execution token](agent-sandbox.md#execution-token). Neither the sandbox nor an upstream MCP server receives the token.

For ACP and headless runs, the harness integration points the harness at this one endpoint and turns off its own MCP server configuration. Other MCP server lists, including repository settings, can't add servers alongside the gateway. This restriction covers MCP server configuration alone. The scope of other repository project settings remains an [open question](agent-open-questions.md#agent-hooks). Before starting the harness, the integration applies the gateway settings and rejects a launch when it can't enforce them.

Each connection and request binds to the authenticated execution and generation. A caller can't select another execution by supplying a session ID, tool argument, or protocol identifier. When the execution becomes terminal, the gateway rejects further requests under the same token rules as the API. Token lifetime and the session-end wait remain governed by the [open question in Agent hooks](agent-open-questions.md#agent-hooks).

## Server and tool access

`afakt agent mcp` manages servers attached to one agent. `afakt mcp-gateway` manages the deployment-configured servers. Whether an agent may attach a server outside that set remains [open](agent-open-questions.md#agents).

The pinned [agent profile](agent-profile.md#mcp-access) selects MCP servers from the deployment's server configuration and allowlists their tools. Servers absent from the server allowlist and tools absent from their server's tool allowlist are unavailable. Deny entries win over allow entries. An empty allowlist grants nothing.

Server and tool access may narrow under the execution specification. An omitted allowlist inherits the profile's list. Any supplied allowlist must be a subset of it. Deny lists from the profile and specification combine. A specification can't change upstream locations, credentials, hosting, or the profile's approval rules. Before dispatching each call, the gateway checks the effective lists again after filtering tool discovery. If a call specifies a hidden or unknown tool, the gateway denies it without contacting an upstream server.

Each exposed tool name identifies both the configured server and its upstream tool name. Through a reversible namespace mapping, the gateway keeps distinct server and tool pairs distinct, even when their names contain separators. A namespace belongs to a configured server identity, never its display label. If two servers offer a tool with the same name, their exposed names can't collide. The built-in server reserves the `afakt` server identity.

Harness-native tools remain subject to the profile's tools and harness permissions. Gateway enforcement doesn't depend on a harness firing a [tool hook](agent-hooks.md#hook-events). The boundary between harness-native tools and MCP tools is separate from the open question about routing native tools through the gateway.

## Upstream credentials

The gateway holds upstream credentials outside the execution's sandbox. Neither an upstream API key nor an OAuth access or refresh token goes to the agent, which can't retrieve them through tool discovery, errors, results, configuration, evidence, or telemetry. The execution token authenticates the caller to afakt alone. For each upstream service, the gateway uses a separate credential.

For remote servers that use [MCP authorization](https://modelcontextprotocol.io/docs/tutorials/security/authorization), afakt acts as the OAuth client. It completes the authorization flow once for the operator, or for each user when the upstream server requires per-user identity. afakt stores and refreshes the resulting tokens. Across executions, it reuses that authorization. An execution doesn't trigger a new interactive sign-in for each tool call. Revoked or unusable authorization makes the upstream unavailable until afakt obtains valid authorization again.

Static API keys use the same [credential references](agent-sandbox.md#credential-brokering) as sandbox credential brokering. Where the upstream connection runs, the gateway resolves them. Credentials for an agent-host server remain outside the execution's sandbox. Credentialed network requests from that server use the broker. Central server credentials stay on the control plane. Credential reference syntax and the fields of an `oauth` entry remain [open questions](agent-open-questions.md#agent-sandbox), as do the [bounds on repository profile self-grants and credential domains](agent-open-questions.md#agent-profile).

## Server hosting and lifecycle

Each server's deployment configuration selects its transport and hosting location. Profiles select configured servers rather than supplying commands or upstream endpoints.

| Server kind | Location and access | Lifecycle |
| --- | --- | --- |
| Remote HTTP | The gateway connects to the remote server over HTTP, with Transport Layer Security (TLS) for credentialed requests. | The remote service manages its processes. afakt maintains its client connections and authorization. |
| Agent-host stdio | A local server that needs the workspace runs inside the sandbox VM under the agent user, with that execution's workspace access and network restrictions. The gateway exchanges MCP messages over standard input and output. | The provider starts it with the execution and stops it with the execution, including verifying termination of its processes. |
| Central stdio | A shared server, such as an issue tracker integration, runs once in its own sandbox on the control plane. MCP messages pass between this server and the gateway through standard input and output. | afakt supervises the server and waits for readiness. It restarts failures with backoff and stops the server during shutdown. Executions share it. |

For agent-host servers, the guest relay and mediator forward gateway requests and responses over the agent host's authenticated outbound connection to the API server. The control plane never opens an inbound connection to an agent host. The provider manages the server's process lifecycle and keeps its standard input and output separate from the harness's streams. Session suspension stops the VM containing these servers. Resume restores or restarts them against the restored workspace and rechecks their channels without choosing how resume affects generation identity.

For a shared central server, the gateway scopes each request to its execution identity and effective tool policy along with its upstream authorization identity. Sharing the process can't grant an execution access to another execution's requests, responses, or credentials. The gateway keeps concurrent responses associated with their originating calls. By default, central servers don't receive an execution workspace, and sandboxing these processes doesn't settle the execution's [host isolation baseline](agent-open-questions.md#agent-sandbox).

## Tool approvals

A profile may mark selected tools as needing approval in `mcp.approvals`. After enforcing the allowlists and before dispatching an approved-only tool, the gateway raises an Agent Client Protocol (ACP) `session/request_permission` on the session or execution responsible for the call. The call waits for the answer. This request comes from the gateway independently of the harness transport and native command approval prompts.

The session hub presents the request and its outcome in every attached client. This includes the web client and chat bridges. Under the [writer and reader rules](agent-session-attachment.md#writer-and-readers), only an attached writer can answer. Readers see updates but never gain approval authority by seeing the request.

For an interactive session without a writer, the request remains pending under [Requests from the agent](agent-session-attachment.md#requests-from-the-agent). An unattended execution, including a headless execution without an attached writer, uses the pinned profile's `mcp.unattended_approval` rule: `deny` or `allow`, with `deny` as the default. ACP executions whose only writer is the agent executor workflow use that same rule. Their attached readers can't answer. afakt records the request and the policy answer on the execution even without an attached client. These gateway requests are separate from the relay's handling of harness permission requests.

Approval permits only the pending call. It never overrides a denied tool, an API capability, a scope, or sandbox access. The gateway dispatches only after an allow answer for that call and rejects a denied or canceled request. Cancellation or termination ends a pending approval without dispatching the call. An answer to that ended request can't revive it.

## Built-in afakt MCP server

The gateway exposes a built-in afakt MCP server under the reserved `afakt` identity. Profiles allowlist its tools and mark them for approval just as they do for any other server.

| Tool operation | Authority check |
| --- | --- |
| Start a sub-execution | The calling execution must have permission to start the requested child, whose authority can't exceed the caller's capabilities and scopes and may narrow further under its own pinned profile and specification. |
| Start a session | The same caller authority bounds the requested session and its profile. |
| Send a message to another session | The caller's `sessions.message` capability and target scopes apply, including the [agent message rules](agent-session-attachment.md#agent-messages). |
| Read another execution's status and results | The caller must have read access to that execution and its result or evidence. |
| Declare a forwarded port | The exposure belongs to the calling agent and execution. The mediator stamps the identity and asks Cedar to allow the port. The guest relay checks listener ownership and guest loopback scope under [Port discovery and records](agent-ports.md#port-discovery-and-records). |

Each call acts as the agent and execution authenticated by its mediator, with that execution's capabilities and scopes. An upstream credential or a gateway service identity never adds afakt authority. The API applies the same checks as for the [afakt CLI](agent-profile.md#capabilities). Selecting a more permissive child profile can't grant more authority to a child. Neither can tool approval.

## Audit and telemetry

Every tool call, including a denied call and a built-in afakt call, emits an `afakt.mcp.tool.call` span in the execution trace. It records the configured server, the namespaced tool, the outcome, and elapsed duration, including any approval wait. The gateway uses the authenticated execution's trace context rather than trusting a caller to choose another execution's trace. [MCP gateway telemetry](agent-observability.md#mcp-gateway) defines span attributes, metrics, units, and cardinality bounds.

Each call also produces an audit record with the execution ID, generation, session ID when present, call ID, server, tool, start and end times, duration, outcome, and any approval request and answer with its source. Audit records never contain credentials, execution tokens, arguments, or results. Telemetry omits arguments and results by default under the profile's [telemetry content rule](agent-observability.md#what-telemetry-never-records). Approval presentation to an authorized client doesn't permit content logging.

Metrics count tool calls by server, tool, and outcome, measure call duration in seconds, count approvals requested and resolved by outcome, and report upstream server health. Server and tool values come from deployment-bounded sets. Unknown values become `other`. Execution and session IDs, upstream URLs, arguments, results, and credential references never become metric attributes.
