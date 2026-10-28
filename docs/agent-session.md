---
type: Design
title: Agent session
description: The agent's durable conversation, retained transcripts, and session workflow.
tags: [afakt, agents, sessions, acp]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent session

An agent session is the conversation of one durable [agent](agent.md). The session transcript remains after each agent launch ends, and the agent retains its current session and past session transcripts. An [agent launch](agent-launch.md) is one instantiation of that agent in an agent sandbox on a workload host. Interactive conversations use the [Agent Client Protocol (ACP)](agent-acp.md). A session uses the same agent profile, agent workspace, sandbox, credential broker, and hooks as an [agent launch](agent-launch.md). Clients use [session attachment](agent-session-attachment.md) for retained history and signed messages.

[Agent launches over ACP](agent-acp.md#Agent%20launches%20over%20ACP).

[Workspace disk deltas and portable revisions](agent-workspace-lifecycle.md#Workspace%20disk%20deltas%20and%20portable%20revisions).

## Session workflow

For each session, an entity workflow uses the session ID as its workflow ID and holds its control state: lifecycle state, writer, attached clients, and idle timer. Ownership of the prompt queue remains [`OQ-SES-14`](agent-open-questions.md#%60OQ-SES-14%60). It never contains the session's message stream. Which component stores the [session message log](agent-acp.md#Session%20message%20log) and enforces writer and reader roles remains [open](agent-open-questions.md#%60OQ-SES-18%60), along with the replay mechanism.

The [agent session lifecycle](agent-lifecycle.md#Session%20lifecycle), [session states](agent-lifecycle.md#Session%20states), [suspend and resume](workload-lifecycle.md#Suspend%20and%20resume), and [session limits](agent-lifecycle.md#Session%20limits) describe how the conversation's active transport starts, pauses, resumes, and ends. [Relay behavior](agent-acp.md#Relay%20behavior), [ACP message policy](agent-acp.md#ACP%20message%20policy), and [MCP tools and approvals](agent-acp.md#MCP%20tools%20and%20approvals) bound its traffic.

[Session previews](workload-previews.md#Session%20previews%20and%20suspension) follow session read access and suspension.
