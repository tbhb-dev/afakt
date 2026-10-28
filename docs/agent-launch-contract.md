---
type: Design
title: Agent launch contract
description: The agent launch spec a workload provider receives, the agent launch result it returns, and the provision, restore, release, status, cancel, and result operations between them.
tags: [afakt, agents, agent-launch, workflows]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent launch contract

[Public key infrastructure](pki.md).

The agent launcher workflow takes a finalized agent launch spec and obtains an agent sandbox on the workload host through the workload provider specified in that spec. After starting the agent harness, it waits and tears down. Provision, restore, and release manage the sandbox alongside status, cancel, and result operations for the launch. See [Agent launch](agent-launch.md) for the lifecycle of this instantiation of a durable [agent](agent.md).

The representation of durable agent and session ownership remains [open](agent-open-questions.md#%60OQ-AGT-10%60).

[Attachment and provider storage](agent-workspace-lifecycle.md#Attachment%20and%20workload%20provider%20storage).

## Agent launch spec

The agent launch spec is the finalized, immutable configuration of an agent instantiation on a workload host. The agent dispatcher workflow finalizes it from an agent spec and records that spec's generation. It specifies the workload provider and never changes during the agent launch. The provider can't change any of it. Its canonical encoding has one digest, the specification digest, and the provider compares that digest when provision repeats. [Agent launch resource](agent-launch-resource.md) gives the key of each field in the resource form.

| Field                  | Required | Meaning                                                                                                                                                                                                                                                                                                                                                                   |
| ---------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Agent launch ID | Yes | The universally unique identifier (UUID) in canonical UUIDv7 form identifying this instantiation. A replacement launch gets a new ID. See [Agent launch ID format](agent-launch-model.md#Agent%20launch%20ID%20format). |
| Predecessor | On replacement | The preceding agent launch ID on retry, resume, migration, or crash recovery. |
| Workload provider | Yes | The provider through which the agent launcher acquires the agent sandbox. |
| Generation | Yes | The agent spec revision from which the agent dispatcher finalized this agent launch spec. Only the control plane increments the agent spec generation, on every change. |
| Agent profile          | Yes      | Profile source, name, exact version, and profile digest. The pinned version fixes the agent's instruction template, capabilities, tools, and credential references, and it bounds the requested fields. See [Agent profile](agent-profile.md).                                                                                                                            |
| Credential references | No | References to named credentials in afakt's credential store for [credential brokering](workload-sandbox.md#Credential%20brokering). The profile also contains references. Whether agent launch references must stay within those the profile permits remains [on hold](agent-open-questions.md#Agent%20sandbox). |
| Workspace access       | No       | `read-only` or `read-write`, never wider than the profile grants. A read-only agent launch can read the agent workspace and publish evidence but can't change the workspace. Without it, the profile's workspace access applies.                                                                                                                                                   |
| Agent harness                | No       | A harness name, plus an exact version when the caller requires one.                                                                                                                                                                                                                                                                                                       |
| Model provider         | No       | Which service should serve the requests.                                                                                                                                                                                                                                                                                                                                  |
| Model                  | No       | A full model ID in the model provider's own form.                                                                                                                                                                                                                                                                                                                         |
| Model effort           | No       | A reasoning effort level, when the harness and model accept one. The workload provider maps it to the harness's own form.                                                                                                                                                                                                                                                |
| Inputs                 | Yes      | References to the prompt and any attachments. Each reference includes a digest.                                                                                                                                                                                                                                                                                           |
| Workspace sources      | Yes      | What the provider builds the workspace from, each placed at a target path. See [Workspace sources](agent-workspace.md#Workspace%20sources).                                                                                                                                                                                                                                 |
| Environment            | No       | Environment variables for the agent and its hooks, combined with the profile's entries. See [Workload environment](agent-workspace.md#Workload%20environment).                                                                                                                                                                                                                    |
| Hooks                  | No       | Commands to run on hook events. They add to the profile's hooks and never replace them. See [Agent hooks](agent-hooks.md).                                                                                                                                                                                                                                                |
| Instruction inputs     | Yes      | A JavaScript Object Notation (JSON) object of values for the profile's instruction template, valid against the profile's input schema. It never contains a secret value. See [Instruction templates](agent-profile.md#Instruction%20templates).                                                                                                                                                          |
| Input digest           | Yes      | A digest over the canonical list of input references and their digests, together with the instruction inputs.                                                                                                                                                                                                                                                             |
| Time limit             | No       | The longest the hooks and the agent may run together. It starts at the observed start of the first `before-agent` hook or, without one, of the agent.                                                                                                                                                                                                                     |
| Stop grace period      | No       | How long the agent has to exit, after a cancellation or a reached limit, before the provider stops it by force.                                                                                                                                                                                                                                                           |
| Resource limits        | No       | Named limits with values, such as a token budget, central processing unit (CPU), memory, or disk. A provider that can't enforce a requested limit rejects the launch.                                                                                                                                                                                                                               |
| Output schema          | No       | A JSON Schema for the structured payload, or a statement that the payload is plain text, when the profile's output schema policy is `per-agent-launch`. Otherwise the profile's output schema applies.                                                                                                                                                                       |
| Capabilities           | No       | A narrower set of the profile's capabilities or scopes for this agent launch. Without it, the profile's capabilities apply. See [Capabilities](agent-profile.md#Capabilities).                                                                                                                                                                                               |
| Model Context Protocol (MCP) access | No | Narrower server and tool allowlists, with extra deny entries, under the profile’s [MCP access rules](agent-profile.md#MCP%20access). Without them, the profile’s lists apply. Approval rules come from the profile alone. |
| Sandbox                | No       | A sandbox configuration that the provider combines with the profile's. See [Workload sandbox](workload-sandbox.md).                                                                                                                                                                                                                                                             |
| Placement requirements | Yes      | Capabilities a workload host must have, such as operating system, architecture, and installed harness. Placement adds the profile's and agent launch's credential references, which the workload host must be able to resolve.                                                                                                                                                                    |
| Retention              | No       | How long the workspace and evidence remain after the agent launch ends. Without it, the deployment default applies.                                                                                                                                                                                                                                                          |
| Labels                 | No       | Task class, variant, and role, for reporting. They don't change how the agent runs.                                                                                                                                                                                                                                                                                       |

An agent launch request must match one entry in the profile's models. Before finalization, the agent dispatcher resolves the requested agent harness, model provider, model, and model effort from that entry. That entry also fixes the [harness transport](agent-profile.md#Harness%20transport). Request omissions use the profile's default limits before finalization. Requested limits and retention can't exceed the profile's ceilings. [Overrides](agent-profile.md#Overrides) gives each bound.

The configuration digest isn't part of the specification. It covers the harness settings, which only the placed workload host can resolve. At launch, the provider computes it and records it with the [agent launch record](agent-launch-model.md#Agent%20launch%20record) and in the [provenance record](agent-launch-model.md#Provenance%20record).

Credential references are part of the specification and the profile. Secret values aren't. The [credential broker](workload-sandbox.md#Credential%20brokering) resolves the references through afakt's credential store and adds credentials only to permitted requests for the duration of one agent launch. Reference bounds remain on hold alongside repository profile self-grant and credential domain bounds.

## Agent launch result

Headless transport defaults and its integration with agent launch validation, hooks, results, and evidence remain [`OQ-SES-16`](agent-open-questions.md#%60OQ-SES-16%60).

The agent launcher returns the agent launch result to the agent dispatcher, which returns it to the parent. A workload provider produces it when the agent launch enters a terminal state, and it never changes afterward.

| Field                       | Meaning                                                                                                                                                                                                                                                                   |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Agent launch ID | The agent launch this result belongs to.                                                                                                                                                                                                                                     |
| Specification digest        | The digest of the specification that ran.                                                                                                                                                                                                                                 |
| Terminal state              | One of `succeeded`, `failed`, `canceled`, or `terminated`. See [Agent launch states](workload-lifecycle.md#Agent%20launch%20states).                                                                                                                                                   |
| Reason                      | One of `agent-exit`, `agent-refusal`, `launch-failure`, `workspace-failure`, `hook-failure`, `invalid-output`, `worker-restart`, `time-limit`, `resource-limit`, `sandbox-violation` from a hook's `terminate` decision, `cancellation` for a cancellation request, or `supersession` by a successor agent launch.                                      |
| Exit status                 | The agent harness's exit status as observed on a headless agent launch, when the harness reports one. Absent on Agent Client Protocol (ACP) agent launches.                                                                                                                                                                                                      |
| Stop reason | The ACP stop reason on an ACP agent launch. Absent on headless agent launches. See [Harness transport](agent-launch.md#Harness%20transport) for its state and reason mapping. |
| Payload                     | The structured payload, valid against the output schema, or the plain text final answer when the payload is plain text, or absent. If the agent completes with success but its final answer doesn't match the schema, the result has no payload and ends `failed` with `invalid-output` as its reason. |
| Payload digest              | The digest of the payload when present.                                                                                                                                                                                                                                   |
| Evidence references         | References to the evidence the agent launch produced. See [Evidence, retention, and cleanup](agent-launch-model.md#Evidence,%20retention,%20and%20cleanup).                                                                                                                       |
| Provenance record           | The provenance record at the moment the agent launch ended. See [Provenance record](agent-launch-model.md#Provenance%20record).                                                                                                                                             |
| Observed times              | When the provider accepted the launch, when it observed the agent start and end, and when the agent launch ended.                                                                                                                                                            |
| Observed usage              | Tokens, cost, or other usage, when the harness or model provider reports them.                                                                                                                                                                                            |

The payload comes from the headless harness's final-answer channel or the ACP turn's last agent message, never from the agent's log stream. Logs remain evidence alongside the transcript and raw final answer.

Normal completion on either transport uses `agent-exit`. ACP `refusal` uses `agent-refusal`, while `max_tokens` and `max_turn_requests` use the existing `resource-limit` reason. An agent failure keeps its reason even if its final answer is invalid.

## Workload provider operations

Each operation identifies an agent launch by agent launch ID, and callers can repeat any operation with the same inputs safely. Each operation runs as the body of an activity on the workload host. Its spans join the trace under the activity's span, and the request contains no trace context, as [Trace topology](agent-observability.md#Trace%20topology) describes. Callers can tell the following errors apart.

| Error                          | Meaning                                                                                                                                                                               |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `invalid-specification`        | The provider found the specification or request malformed, such as an agent launch ID that isn't a canonical UUIDv7 or a time limit of zero. The provider didn't record anything.        |
| `workspace-source-unavailable` | An `agent-launch` source names an agent launch that the provider doesn't hold, that hasn't ended, or whose workspace cleanup removed. A predecessor used as a workspace source must become terminal before the provider builds the new agent workspace. The provider left the agent launch unrecorded. |
| `unsupported`                  | The provider can't run the requested agent harness, model provider, model, model effort, or limits, and it left the agent launch unrecorded.                                                         |
| Profile validation errors      | The specification doesn't match its pinned profile version, and the provider left the agent launch unrecorded. [Launch validation](agent-profile.md#Launch%20validation) lists each error. |
| `specification-conflict`       | The agent launch ID exists with a different specification digest.                                                                                                          |
| `not-found`                    | The provider has no record for the agent launch ID.                                                                                                          |
| `not-terminal`                 | The agent launch hasn't entered a terminal state.                                                                                                                                        |
| `expired`                      | The agent launch ended, but cleanup has removed its result and evidence after their retention.                                                                                           |
| `unavailable`                  | The provider couldn't answer. This error never means that the agent launch doesn't exist.                                                                                                |

### Provision

Provision acquires an agent sandbox for the agent launch spec named by the request and records the validated launch before any process starts. A repeated provision request with the same ID and spec returns the recorded state, under [Agent launch identity](agent-launch-model.md#Agent%20launch%20identity).

- Inputs: the finalized agent launch spec.
- Outputs: the agent launch ID with the specification digest and observed state.
- Errors: `invalid-specification`, profile validation errors, `workspace-source-unavailable`, `unsupported`, `specification-conflict`, and `unavailable`.

If validation fails, provision doesn't record the launch. The workload provider reports `launch-unknown` when agent harness startup is uncertain. Whether provision returns before or after the agent harness is ready remains [open](agent-open-questions.md#%60OQ-CON-9%60).

### Restore

For an agent running in a VM, restore acquires an agent sandbox from a workload snapshot for a new agent launch linked to its predecessor. Workload snapshots capture VM state. They exist only where the agent runs in a VM. Harness session state belongs to a separate agent checkpoint. What rewind and resume restore in host modes and container managers remains [open](agent-open-questions.md#%60OQ-AGT-28%60). The choice between restore and resume as the abstract verb remains [open](agent-open-questions.md#%60OQ-CON-10%60). Snapshot compatibility and the exact restored state follow the [session restoration questions](agent-open-questions.md#Agents).

### Release

Release frees the agent sandbox and its resources. The workload provider observes that every process it started for the launch has exited before treating the launch as ended. Release doesn't remove retained workspace content or evidence before their [retention rules](agent-launch-model.md#Evidence,%20retention,%20and%20cleanup) permit it.

### Status

Status reports what the workload provider knows about an agent launch.

- Inputs: the agent launch ID.
- Outputs:
  - The agent launch ID and predecessor link
  - The state
  - The phase, as [Agent launch record](agent-launch-model.md#Agent%20launch%20record) defines it
  - The time the provider last observed the hooks or the agent
  - The cancellation acknowledgement, if the provider has issued one
  - The evidence references published so far
  - The provenance record
- Errors: `invalid-specification`, `not-found`, and `unavailable`.

Status reports an earlier agent launch's state too, so a caller can confirm that a superseded agent launch stopped.

### Cancel

Cancel asks the workload provider to stop an agent launch's agent.

- Inputs: the agent launch ID, plus a reason.
- Outputs: a cancellation acknowledgement with the time the provider accepted the request and the state at that time.
- Errors: `invalid-specification`, `not-found`, and `unavailable`.

An acknowledgement means the provider accepted the request and began to stop the agent. It doesn't mean the agent stopped. The `canceled` state shows that. For an agent launch that's already terminal, cancel returns the acknowledgement with the terminal state and doesn't change anything.

### Result

Result returns the agent launch result.

- Inputs: the agent launch ID.
- Outputs: the agent launch result.
- Errors: `invalid-specification`, `not-found`, `not-terminal`, `expired`, and `unavailable`.
