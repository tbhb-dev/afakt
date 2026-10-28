---
type: Design
title: Agent ACP
description: Agent Client Protocol connections, message policy, turns, session logs, and writer and reader roles.
tags: [afakt, agents, acp, sessions]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent ACP

[Public key infrastructure](pki.md).

Interactive conversations use the Agent Client Protocol (ACP). People and other agents use any client that speaks ACP to talk to it while it works. These clients include the web dashboard, a native app, an editor, or a chat bot.

Clients attach to the conversations of durable [agents](agent.md) in [agent sessions](agent-session.md) through the [ingress gateway](components.md#Ingress%20gateway) and its `/acp` route. afakt acts as one ACP agent that hosts every session the caller may see. One writer and any number of readers attach at once. The mediator applies its [default-deny ACP message policy](#ACP%20message%20policy). Writer authority never bypasses that policy. The [component accepting client connections](agent-open-questions.md#%60OQ-SES-17%60) and [session state owner](agent-open-questions.md#%60OQ-SES-18%60) remain open.

## Running the harness

One component inside the guest's `afakt-agent` command runs and talks to the harness program. It supports interaction through ACP and a lighter mode in which the harness runs headless and prints its output. The [component and mode names](agent-open-questions.md#%60OQ-SES-12%60) are undecided.

## Session connection

For a locally controlled harness, the agent sandbox provider starts the guest relay inside the sandbox virtual machine (VM). Managed-loop placement remains [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60). The [component inside `afakt-agent`](agent-acp.md#Running%20the%20harness) runs the harness through its ACP adapter. ACP traffic from the guest relay passes through the agent proxy. The mediator authorizes which session messages the proxy may forward. Clients attach through the ingress gateway using the [session attachment](agent-session-attachment.md) Application Programming Interface (API).

## Parts

- **Agent session workflow**: holds the [session control state](agent-session.md#Session%20workflow), separately from the message stream.
- **Guest relay**: a process inside the agent sandbox VM under its own user. It acts as the ACP client to the adapter over standard input and output and sends session traffic through the proxy under mediator decisions.
- **Agent mediator**: the agent mediator applies default-deny policy to ACP messages. It supplies message decisions to the agent proxy without transmitting network traffic.
- **ACP adapter**: the agent harness's ACP agent, which runs inside the sandbox.
- **Agent proxy**: transmits permitted traffic and keeps the launch credentials outside the guest.
- **Ingress gateway**: the public entry point for attached clients, with [endpoint ownership](agent-open-questions.md#%60OQ-SES-17%60) open. Internal launch connections use the overlay.
- **Client**: any program that attaches to the session through the ingress gateway's `/acp` endpoint.

## Agent launches over ACP

Agent profiles select the [harness transport](agent-profile.md#Harness%20transport). Interactive sessions use ACP. For workflow-controlled ACP launches, stop reasons follow the [launch outcome mapping](agent-launch.md#Harness%20transport). Headless transport defaults and integration remain [`OQ-SES-16`](agent-open-questions.md#%60OQ-SES-16%60).

## Launch connection

For a workflow-controlled ACP agent launch, the agent sandbox provider starts the [guest relay](agent-acp.md#Parts) inside the agent sandbox VM, and the adapter runs under the separate agent user. Startup ownership and ordering remain [`OQ-SES-15`](agent-open-questions.md#%60OQ-SES-15%60). The agent mediator checks each ACP message under default deny while the proxy transmits permitted traffic. The [client connection path](agent-open-questions.md#%60OQ-SES-17%60) remains open. After initializing the adapter, the relay opens an ACP session. Only the agent launcher workflow writes to it. It sends the rendered instructions and the prompt together as one prompt and waits for the turn to end. The turn's last agent message is the final answer. Then the relay closes the ACP session. Before running `after-agent` hooks, the provider verifies that the adapter, harness, and their helpers have exited. Control services remain available to hooks under [Running hooks](agent-hooks.md#Running%20hooks). Before the launch becomes terminal, every launch-owned process must exit under [Agent launch states](agent-sandbox-lifecycle.md#Agent%20launch%20states).

The guest relay answers `session/request_permission` from the profile's harness permissions. It allows what the profile grants and rejects the rest. It cancels every `elicitation/create`. No person answers either request in a workflow-controlled agent launch. People with read access may [attach](agent-acp.md#Agent%20launch%20attachment) to a running ACP agent launch as readers and never become writers. Headless attachment remains [`OQ-SES-16`](agent-open-questions.md#%60OQ-SES-16%60).

## The protocol endpoint

Clients connect at `/acp`. Through that endpoint, the ingress gateway accepts writers and readers using the ACP Streamable Hypertext Transfer Protocol (HTTP) and WebSocket transport. A client that speaks ACP over standard input and output attaches through `afakt agent attach --acp`, which bridges it to the endpoint.

| ACP method                      | What afakt does                                                                               |
| ------------------------------- | --------------------------------------------------------------------------------------------- |
| `initialize`                    | Opens a connection for the authenticated caller.                                              |
| `session/list`                  | Lists the sessions the caller may see.                                                        |
| `session/new`                   | Creates a session from a profile in the session configuration options.                        |
| `session/resume`                | Attaches to a session as a reader. With `replayFrom`, it replays the session's history first. |
| `session/prompt`                | Sends a prompt as the writer.                                                                 |
| `session/cancel`                | Cancels the current turn as the writer.                                                       |
| `session/close`                 | Closes the session.                                                                           |
| `_afakt/session/take_writer`    | Makes the caller the session's writer.                                                        |
| `_afakt/session/release_writer` | Gives up the writer role.                                                                     |

For clients that don't speak ACP, the API server also offers the same session operations as Representational State Transfer (REST) resources through the ingress gateway.

The API server exposes per-agent remote-control and ACP routes through the ingress gateway. Their relationship to the shared `/acp` endpoint remains [undecided](agent-open-questions.md#%60OQ-SES-11%60), as does [connection ownership](agent-open-questions.md#%60OQ-SES-17%60).

## Writer and readers

The enforcement owner remains [open](agent-open-questions.md#%60OQ-SES-18%60). The role requirements apply independently of that owner.

How concurrent clients map to one launch's ACP stream while preserving these roles and request correlation remains [open](agent-open-questions.md#%60OQ-SES-20%60).

- Every client attached to the agent session receives every permitted `session/update` notification of the session, within the agent mediator's size and rate bounds for ACP updates.
- Only the writer sends prompts and cancels turns. Answering permission and elicitation requests and changing session modes and configuration options also require that role. afakt rejects these messages from a reader.
- A client becomes the writer by taking the role. Taking it from another client requires the `sessions.write` capability and moves the role at once.
- Without an attached writer, prompts from people with `sessions.write` still queue for delivery when the agent goes idle. Queue ownership remains [`OQ-SES-14`](agent-open-questions.md#%60OQ-SES-14%60).
- Each prompt specifies its author in `_meta`, and commits and pull requests from the session credit that author.

## Agent launch attachment

People with read access may attach to a running workflow-controlled [ACP agent launch](agent-acp.md#Agent%20launches%20over%20ACP) as readers through the same endpoint. Its only writer is the agent launcher workflow. afakt rejects attempts by an attached client to take the writer role, submit prompts, cancel turns, close the session, answer agent requests, or change modes or configuration options. Headless attachment remains [`OQ-SES-16`](agent-open-questions.md#%60OQ-SES-16%60).

For workflow-controlled agent launches, the relay answers `session/request_permission` from the profile's agent harness permissions and cancels each `elicitation/create`. No person answers either request. The writer transfer, prompt queue, and request forwarding rules for interactive sessions don't apply to workflow-controlled agent launches.

## Requests from the agent

- For interactive agent sessions, afakt sends `session/request_permission` and `elicitation/create` only to the writer. Readers see each request as a session update.
- When the writer detaches with a request pending, afakt keeps the request and sends it again to the next writer.
- Without an attached writer, the session enters `requires-action` and afakt notifies the session's owner.

## Remote control and sessions

The API server exposes ACP at `/api/v1/agents/{agent_id}/acp` and remote control at `/api/v1/agents/{agent_id}/remote`. Which component accepts connections from clients and reaches the workload remains [open](agent-open-questions.md#%60OQ-SES-17%60). [Endpoint relationships](agent-open-questions.md#%60OQ-SES-11%60) leave compatibility with the shared `/acp` endpoint undecided.

The outcome of [requests without an active launch](agent-open-questions.md#%60OQ-SES-19%60) and [protocol compatibility negotiation](agent-open-questions.md#%60OQ-SES-21%60) remain undecided.

For each ACP message, the agent mediator applies its policy. Both directions use default deny. A workflow-controlled ACP launch keeps the agent launcher workflow as its only writer. [Session-end wait and SVID lifetime](agent-open-questions.md#%60OQ-HOOK-18%60) remain on hold.

## Session message log

Session history requires a retained message log and replay for attached clients. Which component retains that log and enforces [writer and reader roles](#Writer%20and%20readers), and how replay works, remains [open](agent-open-questions.md#%60OQ-SES-18%60). [Replay cursor recovery](agent-open-questions.md#%60OQ-ATT-5%60) remains open independently of the owner. See [History and replay](agent-session-attachment.md#History%20and%20replay).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

## Relay behavior

- The guest relay advertises no file system or terminal capabilities to the adapter. Without those capabilities, the agent uses its own tools inside the agent sandbox and never accesses a client's machine.
- For an interactive session, the relay forwards messages between the adapter and the agent proxy without changing them, apart from sender details in `_meta`. The mediator applies its default-deny policy to each ACP message before the proxy transmits it toward the client or adapter. For a workflow-controlled agent launch, it handles permission and elicitation requests itself under [Agent launches over ACP](agent-acp.md#Agent%20launches%20over%20ACP).
- When the channel drops, the relay reconnects with backoff and buffers the adapter's messages until the channel returns. The mediator applies the same default-deny policy to replayed ACP messages. Connection recovery depends on the open [session state ownership](agent-open-questions.md#%60OQ-SES-18%60).
- The agent mediator stamps the session's identity. The agent proxy authenticates with the launch's X.509 SPIFFE Verifiable Identity Document (SVID) over mutual Transport Layer Security (TLS). Neither the SVID nor its private key enters the guest.

## ACP message policy

At the agent mediator, policy evaluation uses default deny for each JavaScript Object Notation remote procedure call (`JSON-RPC`) message in both directions. Messages toward the agent, including prompts and cancel, have the sending user as their principal. For workflow prompts, the authenticated workflow supplies the user it acts for. Requests from the agent, including file reads and writes, terminal creation, and permission requests, have the agent as their principal. The relay's lack of file system and terminal capabilities still applies. A policy permit doesn't advertise those capabilities.

Update notifications have size and rate bounds. A denied message gets a `JSON-RPC` error and an entry in the agent's decision log. ACP never bypasses the mediator's policy or network rules. During workflow-controlled agent launches, the relay submits permission and elicitation requests and the answers it produces locally to the agent mediator for the same policy checks. Handling a request locally never exempts it from the message check.

Attached writers and readers enter through the ingress gateway. [Client connection ownership](agent-open-questions.md#%60OQ-SES-17%60), [session state ownership](agent-open-questions.md#%60OQ-SES-18%60), and the relay-to-proxy carrier remain undecided. Lifecycle control through the supervisor and virtual machine monitors (VMMs) is independent of ACP cancellation. Mediator policy can't block it.

## MCP tools and approvals

The agent session harness uses the same single [Model Context Protocol (MCP) gateway endpoint](agent-mcp-gateway.md#Endpoint%20and%20harness%20integration) as an agent launch. Its pinned profile selects the servers and tools, and any specification restrictions only narrow that access. The gateway keeps upstream credentials outside the agent sandbox.

For tools marked as needing approval, the gateway raises an ACP permission request on the session before dispatch. Every attached client shows the request and outcome. This includes the web client and chat bridges. Only the writer may answer. Without an attached writer, an interactive session follows [Requests from the agent](agent-acp.md#Requests%20from%20the%20agent). Unattended agent launches, including headless agent launches, instead use the profile's `mcp.unattended_approval` rule. See [Tool approvals](agent-mcp-gateway.md#Tool%20approvals).

## Turns

The agent receives the writer's prompts one turn at a time. Prompts arriving during a turn wait until the agent goes idle. Durable prompt queue ownership remains [`OQ-SES-14`](agent-open-questions.md#%60OQ-SES-14%60).

## Running the harness without interaction

The component inside the guest's `afakt-agent` command can run the harness headless and receive its printed output. Examples are `claude --print --output-format text|json|stream-json`, `agy --print --output-format text|json|stream-json`, and `codex exec --json`. These command examples don't establish a harness support matrix. The [component and mode names](agent-open-questions.md#%60OQ-SES-12%60) are undecided.

Headless transport defaults and integration with workflow-controlled or interactive launches remain [`OQ-SES-16`](agent-open-questions.md#%60OQ-SES-16%60), including printed output delivery, approvals, cancellation, attachment, hooks, results, and evidence.
