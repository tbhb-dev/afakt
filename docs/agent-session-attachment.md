---
type: Design
title: Agent session attachment
description: Agent session attachment, history replay, and signed messages between agents.
tags: [afakt, agents, sessions, acp, messaging]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent session attachment

[Public key infrastructure](pki.md).

Clients attach to [agent sessions](agent-session.md) through the [Agent Client Protocol (ACP) endpoint](agent-acp.md#The%20protocol%20endpoint). Attachment follows the [writer and reader rules](agent-acp.md#Writer%20and%20readers). Which component accepts connections and reaches the workload remains [open](agent-open-questions.md#%60OQ-SES-17%60).

[Agent launch attachment](agent-acp.md#Agent%20launch%20attachment) admits readers. Interactive [requests from the agent](agent-acp.md#Requests%20from%20the%20agent) require the writer.

## History and replay

- The [session message log](agent-acp.md#Session%20message%20log) is append-only and keyed by ACP message ID. Its owner and replay mechanism remain [open](agent-open-questions.md#%60OQ-SES-18%60).
- A client that attaches with `replayFrom: {type: "start"}` receives the log as `session/update` notifications before the attach response.
- Each client receives only the update kinds that its `initialize` capabilities accept.

## Agent messages

- An agent with the `sessions.message` capability sends a message to another session through the session Application Programming Interface (API). The message enters the target session's prompt queue with the sender in `_meta`.
- A profile limits which sessions its agents may message.
- Each message includes a hop count. afakt drops a message whose hop count passes the deployment's limit.
- The receiving agent treats a message's content as untrusted input, whoever sent it.

## Message signatures

Agent message signatures identify session messages. How agentmail's `afakt agent mail verify` relates to this envelope remains an [open question](agent-open-questions.md#Agents).

- The API server signs each agent message as a JavaScript Object Notation (JSON) Web Signature (JWS) with its message signing key. The signed envelope holds the sender's session ID, agent launch ID, profile name, version, and digest, and the user the session acts for, all derived from the authenticated sender context supplied through its agent mediator. How that context binds the authenticated agent launch identity and sending user remains [open](agent-open-questions.md#Agent%20session). It also holds the target session as its audience, the digest of the message body, the time of signing, and a unique message ID.
- No signing key enters an agent sandbox. An agent sends messages only as itself.
- The API server's verify endpoint checks a signed message's signature against the API server's published keys. It also checks the body digest and the message ID against replay. It reports whether the signature was valid at its signing time.
- The API server keeps retired public keys for as long as it retains messages. These keys allow verification of a signature after the sender's authentication expires.
- A signed message keeps its signature in any channel, such as a pull request comment or a chat thread.
