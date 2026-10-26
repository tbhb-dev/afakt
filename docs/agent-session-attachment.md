---
type: Design
title: Agent session attachment
description: How clients attach to agent sessions through the API server's Agent Client Protocol (ACP) endpoint, the writer and reader rules, history replay, and signed messages between agents.
tags: [afakt, agents, sessions, acp, messaging]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent session attachment

Clients attach to the conversations of durable [agents](agent.md) in [agent sessions](agent-session.md) through the API server's `/acp` endpoint, where afakt acts as one Agent Client Protocol (ACP) agent that hosts every session the caller may see. One writer and any number of readers attach to a session at once. Session traffic between the hub and guest relay crosses the mediator's [ACP gate](agent-session.md#acp-gate). Writer authority never bypasses it. Agents use the same hub to send each other signed messages.

## The protocol endpoint

The `/acp` endpoint serves the ACP Streamable HTTP and WebSocket transport. A client that speaks ACP over standard input and output attaches through `afakt agent attach --acp`, which bridges it to the endpoint.

| ACP method | What afakt does |
| --- | --- |
| `initialize` | Opens a connection for the authenticated caller. |
| `session/list` | Lists the sessions the caller may see. |
| `session/new` | Creates a session from a profile in the session configuration options. |
| `session/resume` | Attaches to a session as a reader. With `replayFrom`, it replays the session's history first. |
| `session/prompt` | Sends a prompt as the writer. |
| `session/cancel` | Cancels the current turn as the writer. |
| `session/close` | Closes the session. |
| `_afakt/session/take_writer` | Makes the caller the session's writer. |
| `_afakt/session/release_writer` | Gives up the writer role. |

The API server also offers the same session operations as REST resources, for clients that don't speak ACP.

## Writer and readers

- Every attached client receives every permitted `session/update` notification of the session, within the ACP gate's size and rate bounds.
- Only the writer sends prompts and cancels turns. Answering permission and elicitation requests and changing session modes and configuration options also require that role. The hub rejects these messages from a reader.
- A client becomes the writer by taking the role. Taking it from another client requires the `sessions.write` capability and moves the role at once.
- Without an attached writer, prompts from people with `sessions.write` still queue, and the hub sends each one when the agent goes idle.
- Each prompt specifies its author in `_meta`, and commits and pull requests from the session credit that author.

## Execution attachment

People with read access may attach to a running [ACP execution](agent-session.md#executions-over-acp) as readers through the same endpoint. Its only writer is the agent executor workflow. The hub rejects attempts by an attached client to take the writer role, submit prompts, cancel turns, close the session, answer agent requests, or change modes or configuration options. A headless execution doesn't offer attachment.

For executions, the relay answers `session/request_permission` from the profile's harness permissions and cancels each `elicitation/create`. No person answers either request. The writer transfer, prompt queue, and request forwarding rules for interactive sessions don't apply to executions.

## Requests from the agent

- For interactive sessions, the hub sends `session/request_permission` and `elicitation/create` only to the writer. Readers see each request as a session update.
- When the writer detaches with a request pending, the hub keeps the request and sends it again to the next writer.
- Without an attached writer, the session enters `requires-action` and afakt notifies the session's owner.

## History and replay

- The hub keeps an append-only log of every message in the session, keyed by ACP message ID.
- A client that attaches with `replayFrom: {type: "start"}` receives the log as `session/update` notifications before the attach response.
- The hub sends each client only the update kinds that its `initialize` capabilities accept.

## Agent messages

- An agent with the `sessions.message` capability sends a message to another session through the session API. The message enters the target session's prompt queue with the sender in `_meta`.
- A profile limits which sessions its agents may message.
- Each message includes a hop count. The hub drops a message whose hop count passes the deployment's limit.
- The receiving agent treats a message's content as untrusted input, whoever sent it.

## Message signatures

These signatures identify session messages. How agentmail's `afakt agent mail verify` relates to this envelope remains an [open question](agent-open-questions.md#agents).

- The API server signs each agent message as a JSON Web Signature (JWS) with its message signing key. The signed envelope holds the sender's session ID, execution ID, generation, profile name, version, and digest, and the user the session acts for, all derived from the authenticated sender context supplied through its mediator. How that context binds the token and sending user remains [open](agent-open-questions.md#agent-session). It also holds the target session as its audience, the digest of the message body, the time of signing, and a unique message ID.
- No signing key enters a sandbox. An agent sends messages only as itself.
- The API server's verify endpoint checks a signed message's signature against the API server's published keys. It also checks the body digest and the message ID against replay. It reports whether the signature was valid at its signing time.
- The API server keeps retired public keys for as long as it retains messages. These keys allow verification of a signature after its token expires.
- A signed message keeps its signature in any channel, such as a pull request comment or a chat thread.
