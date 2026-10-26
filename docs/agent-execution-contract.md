---
type: Design
title: Agent execution contract
description: The execution specification an execution provider receives, the execution result it returns, and the launch, status, cancel, and result operations between them.
tags: [afakt, agents, execution, workflows]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent execution contract

Each execution is one harness run owned by a durable [agent](agent.md). Across harnesses, models, and model providers, the agent executor and an execution provider share one contract. To launch, the executor passes an execution specification. It reads back an execution result. Between the two, it uses the four provider operations: launch, status, cancel, and result. [Agent execution](agent-execution.md) describes the lifecycle that these operations drive.

The representation of agent ownership and the mapping from generations to runs of a harness remain [open](agent-open-questions.md#agents).

## Execution specification

The execution specification lists what an execution provider needs to start an agent. The provider can't change any of it. Its canonical encoding has one digest, the specification digest, and the provider compares that digest when a launch repeats. [Agent execution resource](agent-execution-resource.md) gives the key of each field in the resource form.

| Field                  | Required | Meaning                                                                                                                                                                                                                                                                                                                                                                   |
| ---------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Execution ID           | Yes      | The stable identifier of the logical execution, a universally unique identifier (UUID) of version 7 (UUIDv7) as Request for Comments (RFC) 9562 defines it. afakt assigns it when it creates the execution, before the first launch. It doesn't change across retries, restarts, or generations. See [Execution ID format](agent-execution-model.md#execution-id-format). |
| Generation             | Yes      | Starts at 1. A caller raises this positive integer to run the same logical execution again on purpose. Retries and restarts never raise it.                                                                                                                                                                                                                               |
| Agent profile          | Yes      | Profile source, name, exact version, and profile digest. The pinned version fixes the agent's instruction template, capabilities, tools, and credential references, and it bounds the requested fields. See [Agent profile](agent-profile.md).                                                                                                                            |
| Credential references | No | References to named credentials in afakt's credential store for [credential brokering](agent-sandbox.md#credential-brokering). The profile also contains references. Whether execution references must stay within those the profile permits remains [on hold](agent-open-questions.md#agent-sandbox). |
| Workspace access       | No       | `read-only` or `read-write`, never wider than the profile grants. A read-only execution can read the workspace and publish evidence but can't change the workspace. Without it, the profile's workspace access applies.                                                                                                                                                   |
| Harness                | No       | A harness name, plus an exact version when the caller requires one.                                                                                                                                                                                                                                                                                                       |
| Model provider         | No       | Which service should serve the requests.                                                                                                                                                                                                                                                                                                                                  |
| Model                  | No       | A full model ID in the model provider's own form.                                                                                                                                                                                                                                                                                                                         |
| Model effort           | No       | A reasoning effort level, when the harness and model accept one. The execution provider maps it to the harness's own form.                                                                                                                                                                                                                                                |
| Inputs                 | Yes      | References to the prompt and any attachments. Each reference includes a digest.                                                                                                                                                                                                                                                                                           |
| Workspace sources      | Yes      | What the provider builds the workspace from, each placed at a target path. See [Workspace sources](agent-workspace.md#workspace-sources).                                                                                                                                                                                                                                 |
| Environment            | No       | Environment variables for the agent and its hooks, combined with the profile's entries. See [Agent environment](agent-workspace.md#agent-environment).                                                                                                                                                                                                                    |
| Hooks                  | No       | Commands to run on hook events. They add to the profile's hooks and never replace them. See [Agent hooks](agent-hooks.md).                                                                                                                                                                                                                                                |
| Instruction inputs     | Yes      | A JSON object of values for the profile's instruction template, valid against the profile's input schema. It never contains a secret value. See [Instruction templates](agent-profile.md#instruction-templates).                                                                                                                                                          |
| Input digest           | Yes      | A digest over the canonical list of input references and their digests, together with the instruction inputs.                                                                                                                                                                                                                                                             |
| Time limit             | No       | The longest the hooks and the agent may run together. It starts at the observed start of the first `before-agent` hook or, without one, of the agent.                                                                                                                                                                                                                     |
| Stop grace period      | No       | How long the agent has to exit, after a cancellation or a reached limit, before the provider stops it by force.                                                                                                                                                                                                                                                           |
| Resource limits        | No       | Named limits with values, such as a token budget, CPU, memory, or disk. A provider that can't enforce a requested limit rejects the launch.                                                                                                                                                                                                                               |
| Output schema          | No       | A JSON Schema for the structured payload, or a statement that the payload is plain text, when the profile's output schema policy is `per-execution`. Otherwise the profile's output schema applies.                                                                                                                                                                       |
| Capabilities           | No       | A narrower set of the profile's capabilities or scopes for this execution. Without it, the profile's capabilities apply. See [Capabilities](agent-profile.md#capabilities).                                                                                                                                                                                               |
| Model Context Protocol (MCP) access | No | Narrower server and tool allowlists, with extra deny entries, under the profile’s [MCP access rules](agent-profile.md#mcp-access). Without them, the profile’s lists apply. Approval rules come from the profile alone. |
| Sandbox                | No       | A sandbox configuration that the provider combines with the profile's. See [Agent sandbox](agent-sandbox.md).                                                                                                                                                                                                                                                             |
| Placement requirements | Yes      | Capabilities an agent host must have, such as operating system, architecture, and installed harness. Placement adds the profile's and execution's credential references, which the agent host must be able to resolve.                                                                                                                                                                    |
| Retention              | No       | How long the workspace and evidence remain after the execution ends. Without it, the deployment default applies.                                                                                                                                                                                                                                                          |
| Labels                 | No       | Task class, variant, and role, for reporting. They don't change how the agent runs.                                                                                                                                                                                                                                                                                       |

Requests for a harness, model provider, model, and model effort must match one entry in the profile's models, and a matching entry supplies any of them the specification omits. That entry also fixes the [harness transport](agent-profile.md#harness-transport). Limits the specification omits come from the profile's default limits. Requested limits and retention can't exceed the profile's ceilings. [Overrides](agent-profile.md#overrides) gives each bound.

The configuration digest isn't part of the specification. It covers the harness settings, which only the placed agent host can resolve. At launch, the provider computes it and records it with the [generation](agent-execution-model.md#execution-generation) and in the [provenance record](agent-execution-model.md#provenance-record).

Credential references are part of the specification and the profile. Secret values aren't. The [credential broker](agent-sandbox.md#credential-brokering) resolves the references through afakt's credential store and adds credentials only to permitted requests for the duration of one execution. Reference bounds remain on hold alongside repository profile self-grant and credential domain bounds.

## Execution result

The agent executor returns the execution result to the agent runner, which returns it to the parent. A provider produces it when the execution enters a terminal state, and it never changes afterward.

| Field                       | Meaning                                                                                                                                                                                                                                                                   |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Execution ID and generation | The execution this result belongs to.                                                                                                                                                                                                                                     |
| Specification digest        | The digest of the specification that ran.                                                                                                                                                                                                                                 |
| Terminal state              | One of `succeeded`, `failed`, `canceled`, or `terminated`. See [Execution states](agent-execution.md#execution-states).                                                                                                                                                   |
| Reason                      | One of `agent-exit`, `agent-refusal`, `launch-failure`, `workspace-failure`, `hook-failure`, `invalid-output`, `worker-restart`, `time-limit`, `resource-limit`, `sandbox-violation` from a hook's `terminate` decision, `cancellation` for a cancellation request, or `supersession` by a higher generation.                                      |
| Exit status                 | The harness's exit status as observed on a headless run, when the harness reports one. Absent on ACP runs.                                                                                                                                                                                                      |
| Stop reason | The ACP stop reason on an ACP run. Absent on headless runs. See [Harness transport](agent-execution.md#harness-transport) for its state and reason mapping. |
| Payload                     | The structured payload, valid against the output schema, or the plain text final answer when the payload is plain text, or absent. If the agent completes with success but its final answer doesn't match the schema, the result has no payload and ends `failed` with `invalid-output` as its reason. |
| Payload digest              | The digest of the payload when present.                                                                                                                                                                                                                                   |
| Evidence references         | References to the evidence the execution produced. See [Evidence, retention, and cleanup](agent-execution-model.md#evidence-retention-and-cleanup).                                                                                                                       |
| Provenance record           | The provenance record at the moment the execution ended. See [Provenance record](agent-execution-model.md#provenance-record).                                                                                                                                             |
| Observed times              | When the provider accepted the launch, when it observed the agent start and end, and when the execution ended.                                                                                                                                                            |
| Observed usage              | Tokens, cost, or other usage, when the harness or model provider reports them.                                                                                                                                                                                            |

The payload comes from the headless harness's final-answer channel or the ACP turn's last agent message, never from the agent's log stream. Logs remain evidence alongside the transcript and raw final answer.

Normal completion on either transport uses `agent-exit`. ACP `refusal` uses `agent-refusal`, while `max_tokens` and `max_turn_requests` use the existing `resource-limit` reason. An agent failure keeps its reason even if its final answer is invalid.

## Provider operations

Each operation identifies an execution by execution ID and generation, and callers can repeat any operation with the same inputs safely. Each operation runs as the body of an activity on the agent host. Its spans join the trace under the activity's span, and the request contains no trace context, as [Trace topology](agent-observability.md#trace-topology) describes. Callers can tell the following errors apart.

| Error                          | Meaning                                                                                                                                                                               |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `invalid-specification`        | The provider found the specification or request malformed, such as an execution ID that isn't a canonical UUIDv7 or a time limit of zero. The provider didn't record anything.        |
| `workspace-source-unavailable` | An `execution` source names an execution that the provider doesn't hold, that hasn't ended, or whose workspace cleanup removed. An earlier generation of the same execution that the launch supersedes counts as ended, because the provider builds the workspace only after that generation becomes terminal. The provider left the execution unrecorded. |
| `unsupported`                  | The provider can't run the requested harness, model provider, model, model effort, or limits, and it left the execution unrecorded.                                                         |
| Profile validation errors      | The specification doesn't match its pinned profile version, and the provider left the execution unrecorded. [Launch validation](agent-profile.md#launch-validation) lists each error. |
| `stale-generation`             | The request gives a generation lower than the execution's current generation.                                                                                                         |
| `specification-conflict`       | The execution ID and generation exist with a different specification digest.                                                                                                          |
| `not-found`                    | The provider has no record of the execution ID, or of that generation of it.                                                                                                          |
| `not-terminal`                 | The execution hasn't entered a terminal state.                                                                                                                                        |
| `expired`                      | The execution ended, but cleanup has removed its result and evidence after their retention.                                                                                           |
| `unavailable`                  | The provider couldn't answer. This error never means that the execution doesn't exist.                                                                                                |

### Launch

Launch records an execution and starts its agent.

- Inputs: the execution specification.
- Outputs: the execution ID and generation with the specification digest, plus the state after launch. The first launch returns `pending`, `running`, `launch-unknown`, or `failed`, and a repeated launch returns whatever state the execution has at that moment.
- Errors: `invalid-specification`, the profile validation errors, `workspace-source-unavailable`, `unsupported`, `stale-generation`, `specification-conflict`, and `unavailable`.

The provider validates the whole specification, including its fit with the pinned profile version, and records the execution before it starts anything. When validation fails, the provider doesn't record the execution or act on it. When the provider tried to start the agent and can't confirm whether it started, launch returns `launch-unknown` rather than an error. [Execution identity](agent-execution-model.md#execution-identity) gives the outcome of a repeated launch.

### Status

Status reports what the provider knows about an execution.

- Inputs: the execution ID and an optional generation. Without a generation, status reports the current one.
- Outputs:
  - The generation and the current generation
  - The state
  - The phase, as [Execution generation](agent-execution-model.md#execution-generation) defines it
  - The time the provider last observed the hooks or the agent
  - The cancellation acknowledgement, if the provider has issued one
  - The evidence references published so far
  - The provenance record
- Errors: `invalid-specification`, `not-found`, and `unavailable`.

Status reports an older generation's state too, so a caller can confirm that a superseded generation stopped.

### Cancel

Cancel asks the provider to stop an execution's agent.

- Inputs: the execution ID and generation, plus a reason.
- Outputs: a cancellation acknowledgement with the time the provider accepted the request and the state at that time.
- Errors: `invalid-specification`, `not-found`, and `unavailable`.

An acknowledgement means the provider accepted the request and began to stop the agent. It doesn't mean the agent stopped. The `canceled` state shows that. For an execution that's already terminal, cancel returns the acknowledgement with the terminal state and doesn't change anything.

### Result

Result returns the execution result.

- Inputs: the execution ID and generation.
- Outputs: the execution result.
- Errors: `invalid-specification`, `not-found`, `not-terminal`, `expired`, and `unavailable`.
