---
type: Design
title: Agent session
description: The agent's durable conversation and its interactive ACP transport, session workflow, control states, and suspend and resume.
tags: [afakt, agents, sessions, acp]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent session

An agent session is the conversation of one durable [agent](agent.md). The session transcript remains after each harness run ends, and the agent retains its current session and past session transcripts. An [execution](agent-execution.md) is one harness run of that agent. Interactive conversations use the Agent Client Protocol (ACP). People and other agents use any client that speaks ACP to talk to it while it works. These clients include the web dashboard, a native app, an editor, or a chat bot. A session uses the same agent profile, workspace, sandbox, credential broker, and hooks as an [agent execution](agent-execution.md). The execution provider starts the guest relay inside the sandbox VM. It runs the harness through its ACP adapter and ends its ACP channel at the per-agent mediator, which forwards the session to the API server. Clients then attach to the session through the [session attachment](agent-session-attachment.md) API.

## Executions over ACP

Each entry in a profile's models selects `acp` or `headless` as its [harness transport](agent-profile.md#harness-transport). It defaults to ACP when the harness has an adapter and to headless otherwise. An entry can select headless even when an adapter is available, such as for a feature available only in that form. Selecting ACP without an adapter fails launch validation. Interactive agent sessions use ACP.

For an ACP execution, the provider starts the guest relay inside the sandbox VM, and the relay starts the adapter under the separate agent user. The agent executor workflow is the only writer. It sends the rendered instructions and the prompt together as one prompt and waits for the turn to end. The turn's last agent message is the final answer. The relay then closes the ACP session. The provider verifies termination of the relay, adapter, harness, and helpers. Output schema validation, hooks, results, and evidence apply equally to ACP and headless executions.

People with read access may attach to a running ACP execution as readers and never become writers. A headless execution doesn't offer attachment. The relay answers `session/request_permission` from the profile's harness permissions. It allows what the profile grants and rejects the rest. It cancels each `elicitation/create`. No person answers either request in an execution. ACP stop reasons follow the [execution outcome mapping](agent-execution.md#harness-transport) and replace the exit status in the execution result.

| Harness | Default execution transport |
| --- | --- |
| Claude Code | ACP, through the `claude-agent-acp` adapter |
| Codex | ACP, through the `codex-acp` adapter |
| Other harnesses | Headless |

## Parts

- **Agent session workflow**: an entity workflow for each session with the session ID as its workflow ID. It holds the session's control state. This includes its lifecycle state, the writer, the attached clients, the prompt queue, the idle timer, and the API server replica that holds the mediator's session connection. It never contains the session's message stream.
- **Guest relay**: a process inside the sandbox VM under its own user. It acts as the ACP client to the adapter over standard input and output and connects to the mediator.
- **Mediator and ACP gate**: the per-agent mediator ends the relay channel and checks ACP messages. It forwards permitted messages to the session hub over the agent host's authenticated outbound connection.
- **ACP adapter**: the harness's ACP agent, which runs inside the sandbox.
- **Session hub**: the part of the API server that joins the mediator's session connection to the attached clients. It enforces the writer rules and keeps the session's message log.
- **Client**: any program that attaches to the session through the API server's `/acp` endpoint.

## Session lifecycle

The lifecycle below controls an interactive conversation and its active transport. Closing a transport or ending an execution leaves the retained transcript with its agent. How these control states map to [agent states](agent.md#lifecycle) and run boundaries remains [open](agent-open-questions.md#agents).

1. **Creation.** When a client creates a session for an agent through the API server with a profile, workspace sources, and an optional first prompt, the API server starts the agent session workflow with the new session ID.
2. **Placement.** The agent runner workflow places the session on an agent host, as for an execution.
3. **Start.** The provider builds the workspace and starts the guest relay before the `before-agent` hooks. Their CLI control requests use the mediator. After the hooks, it starts the adapter in the sandbox and the relay sends `initialize` and `session/new`. It registers the session through the mediator and the agent host's authenticated outbound connection to the API server.
4. **Turns.** The agent receives the writer's prompts one turn at a time. Prompts that arrive during a turn stay in the hub's queue until the agent goes idle, when the hub sends each one.
5. **Suspend.** After the idle timeout, the agent session workflow suspends the session. The supervisor saves a checkpoint at an idle turn boundary, including the VM snapshot, frozen workspace disk delta, and harness session state. After the `before-teardown` hooks, it stops the VM and frees the slot.
6. **Resume.** A prompt or an attach request for a suspended session resumes it. The supervisor restores the checkpoint under the same guest paths and reacquires credentials outside the guest. The guest relay reconnects through the mediator. A restarted adapter receives `session/resume`. Which process and transport state survives restoration remains [open](agent-open-questions.md#agents).
7. **Close.** A close request, the session's lifetime limit, or a failure ends the session. After running the `session-end` and `before-teardown` hooks, the provider records the workspace changes and publishes the transcript as evidence.

## Session states

These are conversation control states, separate from the durable agent lifecycle and provider execution states.

| State | Meaning |
| --- | --- |
| `starting` | The provider is building the workspace or starting the relay. |
| `running` | The agent is working on a turn. |
| `requires-action` | The agent waits on a permission or elicitation answer from the writer. |
| `idle` | The agent has no turn in progress and waits for a prompt. |
| `suspended` | A checkpoint holds the session after its VM stops. |
| `closed` | The session ended on a close request or at its lifetime limit. |
| `failed` | The session ended on a failure. |

## Relay behavior

- The relay advertises no file system or terminal capabilities to the adapter. Without those capabilities, the agent uses its own tools inside the sandbox and never accesses a client's machine.
- For an interactive session, the relay forwards messages between the adapter and the mediator without changing them, apart from sender details in `_meta`. The mediator applies the ACP gate before forwarding a message to the hub or adapter. For an execution, it handles permission and elicitation requests itself under [Executions over ACP](#executions-over-acp).
- When the channel drops, the relay reconnects with backoff and buffers the adapter's messages until the channel returns. Replayed messages still pass the ACP gate. The workflow for the agent session records which API server replica holds the mediator's session connection.
- The mediator stamps the session's identity and authenticates through the agent host's outbound connection. Any [execution token](agent-sandbox.md#execution-token) stays outside the guest.

## Suspend and resume

A checkpoint pairs a VM snapshot, the workspace disk delta, and the harness's session state at one idle turn boundary. The guest filesystem is frozen for capture so the transcript and workspace agree. Credentials and broker state stay outside the guest and checkpoint.

The provider stores checkpoint data in the object store. Resume restores the same paths inside the guest and reacquires credentials. It must recheck port publications and channels. Which process state, compatibility constraints, and placement rules allow restoration on another agent host remain [open](agent-open-questions.md#agents).

## ACP gate

The relay's ACP channel ends at the agent's mediator. Cedar decides on every `JSON-RPC` message in both directions, with default deny. Messages toward the agent, including prompts and cancel, have the sending user as their principal. For workflow prompts, the authenticated workflow supplies the user it acts for. Requests from the agent, including file reads and writes, terminal creation, and permission requests, have the agent as their principal. The relay's lack of file system and terminal capabilities still applies. A policy permit doesn't advertise those capabilities.

Update notifications have size and rate bounds. A denied message gets a `JSON-RPC` error and an entry in the agent's decision log. ACP never bypasses the mediator's policy or network rules. During executions, the relay submits permission and elicitation requests and the answers it produces locally to the gate too. Handling a request locally never exempts it from the message check.

The mediator forwards the session to the session hub over the agent host's authenticated outbound connection to the API server. Its transport and the relay-to-mediator carrier remain [open](agent-open-questions.md#agent-session). Lifecycle control through the supervisor and `VMM` is independent of ACP cancellation. The gate can't block it.

## MCP tools and approvals

The session harness uses the same single [Model Context Protocol (MCP) gateway endpoint](agent-mcp-gateway.md#endpoint-and-harness-integration) as an execution. Its pinned profile selects the servers and tools, and any specification restrictions only narrow that access. The gateway keeps upstream credentials outside the sandbox.

For tools marked as needing approval, the gateway raises an ACP permission request on the session before dispatch. Every attached client shows the request and outcome. This includes the web client and chat bridges. Only the writer may answer. Without an attached writer, an interactive session follows [Requests from the agent](agent-session-attachment.md#requests-from-the-agent). Unattended executions, including headless runs, instead use the profile's `mcp.unattended_approval` rule. See [Tool approvals](agent-mcp-gateway.md#tool-approvals).

## Forwarded ports and previews

The web client and chat bridges show the session's forwarded ports and preview links as listeners appear, change, and close. A preview opens in a browser tab beside the session. People with session read access can use these links without taking the writer role. Previews use afakt sign-in and the session's access checks, with expiring, revocable share links for one port. Raw TCP access requires an authenticated afakt client tunnel. See [Agent ports and previews](agent-ports.md).

Suspension stops the VM. Whether a preview request wakes it or fails remains [open](agent-open-questions.md#agent-ports-and-previews). Resume rechecks listener ownership and publication generations before routing traffic. Closed records and their share links never attach to another listener.

## Limits

- The idle timeout bounds how long an idle session holds a slot before it suspends.
- The lifetime limit bounds a session from creation to close across every suspend and resume.
