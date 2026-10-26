---
type: Design
title: Agent execution model
description: The fields of a logical agent execution and of each execution generation, who sets each one, which never change, and how they relate to profiles, workspaces, results, evidence, and provenance.
tags: [afakt, agents, execution, model]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent execution model

An [execution](agent-execution.md) is one harness run of one durable [agent](agent.md). Executions belong to the agent, which retains the conversation transcript across runs. The provider record has two levels: logical execution and generation, whose mapping to run boundaries remains an [open question](agent-open-questions.md#agents). From creation to cleanup, the logical execution keeps one execution ID. Every deliberate attempt to run it makes an execution generation, with its own specification, workspace, state, and result. On the placed agent host, the execution provider keeps the execution record. Status and result read from it.

## Logical execution

| Field | Type and format | Meaning | Set by | Changes |
| --- | --- | --- | --- | --- |
| Execution ID | UUIDv7 in canonical text form. See [Execution ID format](#execution-id-format). | Identifies the logical execution. | afakt, when it creates the execution | Never |
| Created at | Coordinated Universal Time (UTC) timestamp in the Request for Comments (RFC) 3339 form | When afakt created the execution. | afakt | Never |
| Current generation | Positive integer | The highest generation the provider has accepted. | The provider, when it accepts a launch | Only upward |
| Generations | List of execution generations | Every generation the provider has accepted, in generation order. | The provider | A generation joins the list at each accepted launch. None leaves it before cleanup. |

## Execution generation

| Field | Type and format | Meaning | Set by | Changes |
| --- | --- | --- | --- | --- |
| Execution ID | UUIDv7 | The logical execution this generation belongs to. | afakt | Never |
| Generation | Positive integer | Which deliberate attempt this is. | The parent workflow | Never |
| Specification | [Execution specification](agent-execution-contract.md#execution-specification) | What the provider runs. | The parent workflow | Never |
| Specification digest | Digest, as the algorithm name, `:`, and lowercase hexadecimal | Identifies the specification for repeated launches. | The provider, at launch | Never |
| Pinned profile | Copy of the profile version's content with its referenced files | The profile the provider validates and runs against. | The provider, at launch | Never |
| Placement | Agent host and host task queue | Where every operation for this generation runs. | The agent runner workflow, from the placement activity | Never, once set |
| State | One of the [execution states](agent-execution.md#execution-states) | Where the generation stands. | The provider, from its observations | Through the allowed transitions, until terminal |
| State history | List of transitions with the time of each and the observation behind it | How the generation reached its state. | The provider | Append-only |
| Phase | `before-agent`, `agent`, `after-agent`, `before-teardown`, or none | Which part of the generation is running. | The provider | With each part. None before the first `before-agent` hook or the agent start, and none once terminal |
| Last observed | UTC timestamp | When the provider last observed the hooks or the agent. | The provider | With each observation |
| Cancellation acknowledgement | Time and state at acceptance, plus the reason | The provider's confirmation of a cancellation. | The provider, on cancel | Set once |
| Superseded by | Generation | The higher generation that requested this one's cancellation. | The provider | Set once |
| Workspace | Absolute path inside the guest, with each source resolved | The directory tree the agent works in. See [Agent workspace](agent-workspace.md). | The provider, when it builds the workspace | Never, once built, except for the agent's own changes |
| Combined configuration | Combined sandbox configuration, agent environment, hooks, harness transport, and harness settings | What confines and steers the agent. | The provider, at launch | Never |
| Configuration digest | Digest, in the same form as the specification digest | The digest over the pinned profile digest, the rendered instructions digest, the combined sandbox configuration digest, the agent environment, the combined hooks, the specification's overrides, the selected harness transport, and the harness settings the provider resolves from them. | The provider, at launch on the placed agent host | Never |
| Evidence references | List of [evidence references](#evidence-retention-and-cleanup) | Evidence published so far. | The provider | Append-only, until cleanup |
| Provenance record | [Provenance record](#provenance-record) | What the generation requested and what the provider observed. | The provider | Requested fields never change. The provider sets each observed field once. The hook counts grow until the generation becomes terminal. |
| Observed times | UTC timestamps for launch accepted, agent start, agent end, and execution end | When each happened. | The provider | Each set once |
| Result | [Execution result](agent-execution-contract.md#execution-result) | The outcome. | The provider, on entering a terminal state | Set once |
| Retention | UTC timestamp of the end of retention, and any retention holds | How long the workspace and evidence remain. | The provider, from the specification's retention, on entering a terminal state. An operator places and releases holds. | Holds come and go |
| Cleanup | `none`, `complete`, or `incomplete`, with any items left over | Whether cleanup removed the workspace and evidence. | The provider, during cleanup | Until `complete` |

Everything the specification contains is immutable, and so is everything the provider derives from it at launch. Fields that track progress change only in the ways the table gives. Once the provider sets the result, it never changes.

## Relations

- **Agent.** Each execution belongs to one durable agent. The representation of this ownership in provider records remains [open](agent-open-questions.md#agents).
- **Session.** The conversation belongs to the agent. A generation publishes its run transcript as evidence. Retaining the conversation beyond evidence cleanup remains [open](agent-open-questions.md#agents).
- **Agent profile.** Each generation pins one exact profile version by source, name, version, and digest, and the provider stores a copy of that version with the generation. Many generations and many executions can pin the same version. A higher generation may pin a different version. See [Pinning a profile version](agent-profile.md#pinning-a-profile-version).
- **Workspace sources.** Each generation builds its own workspace from its specification's sources. An `execution` source refers to an ended generation of any execution on the same agent host, including an earlier generation of the same execution. A generation that the new one supersedes qualifies too, because the provider builds the new workspace only after that generation becomes terminal. Cleanup skips a workspace that a `pending` generation lists as a source.
- **Workspace images.** An `image` source records its resolved manifest digest with the generation. Retries use that digest, and a later generation can request it explicitly. A non-terminal or retained execution protects its image manifest, config, and layers from collection. See [Image sources and digest pinning](agent-workspace-image.md#image-sources-and-digest-pinning) and [Retention and garbage collection](agent-workspace-image.md#retention-and-garbage-collection).
- **Generations.** One logical execution has one or more generations. At most one generation of an execution has a running agent at any time, and a higher generation supersedes a lower one. See [Generations and supersession](agent-execution.md#generations-and-supersession).
- **Result.** Each generation has at most one result, produced when it enters a terminal state.
- **Evidence.** Each generation has its own evidence. Each item belongs to exactly one generation.
- **Provenance.** Each generation has exactly one provenance record. Status returns it while the generation runs, and the result includes it at the end.

## Execution identity

An execution ID identifies one logical execution. Its generation counts the deliberate attempts to run it. For each execution ID, the provider records a current generation, the highest generation it has accepted.

| Launch with | Outcome |
| --- | --- |
| A new execution ID with generation 1 | Accepted, and generation 1 becomes current. |
| A new execution ID with a higher generation | Rejected with `invalid-specification`, because each execution starts at generation 1. |
| The same ID and generation, the same specification digest | Returns the existing execution and its current state without starting a second agent. |
| The same ID and generation, a different specification digest | Rejected with `specification-conflict`, and the existing execution keeps its specification. |
| The same ID, a generation lower than the current generation | Rejected with `stale-generation`. |
| The same ID, a generation higher than the current generation | Accepted, and it becomes the current generation. |

The specification digest includes all fields except the generation. A repeated launch that differs only in its labels is still a conflict.

### Execution ID format

An execution ID is a UUIDv7 in the canonical text form of RFC 9562: 32 lowercase hexadecimal digits in groups of 8, 4, 4, 4, and 12, separated by hyphens, such as `019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b`. Its version digit is `7`, and its variant digit is one of `8`, `9`, `a`, or `b`.

afakt assigns the execution ID when it creates the logical execution, and the parent workflow records it in workflow history before the first launch. Each later attempt and each higher generation reuses that ID. Callers never choose an execution ID, and providers never assign one.

Every place that records or passes an execution ID uses the canonical form, including the specification, provider operations, the execution result, the provenance record, evidence, the execution-ended event, the [execution environment](agent-workspace.md#execution-environment), the `${execution-id}` sandbox variable, and the [execution token](agent-sandbox.md#execution-token). A provider rejects an execution ID in any other form, such as with uppercase digits or without hyphens, with `invalid-specification`.

## Provenance record

Each execution generation has one provenance record. Status and result both return it.

| Field | Requested | Observed |
| --- | --- | --- |
| Harness | Name and version from the specification | Name and version the harness reports |
| Harness transport | The transport of the selected profile model entry | The transport the provider reports applying |
| Model provider | From the specification | The model provider that served the requests, as reported |
| Model | From the specification | The model ID that served the requests, as reported |
| Model effort | From the specification | The effort level the harness reports applying |
| Agent profile | Source, name, version, and profile digest from the specification, plus the rendered instructions digest | Not observed, because the provider resolves and renders the profile before launch |
| Workspace | Each source and target path from the specification | For each `repository` source, the checked-out commit and any new branch. For each `image` source, the resolved manifest digest. For each `execution` source, the digest of the copied workspace. |
| Sandbox | The combined sandbox configuration digest and the capabilities in the execution token | The sandbox settings the provider reports applying |
| Environment | Each agent environment name and value, with its source in the profile or the specification | Not observed, because the provider sets the environment itself |
| Hooks | Each hook's name, event, source, matcher, decisions, and command from the combined list | Each hook's number of runs and failures so far, or that the provider skipped it, each [support warning](agent-hooks.md#harness-support) for an event or decision it uses, and the [workspace repository hooks](agent-hooks.md#workspace-repository-hooks) |
| Agent host | Placement requirements from the specification | The agent host that ran the agent, and its afakt worker version |

The record also contains the execution ID in its canonical UUIDv7 form and the generation, plus the specification digest, the input digest, and the configuration digest. It also gives the trace ID and span ID of the generation's span in the [execution trace](agent-observability.md#trace-topology).

Requested values and observed values are separate fields. The provider never copies one into the other. An observed field is empty until the execution provider reports a value it read from the harness, the model provider, or the agent host. Empty means that no source reported the value, never that the observed value matches the requested one. Once the provider records an observed value for a generation, that value doesn't change. The hook run and failure counts are the exception: they count the runs so far, and they stop changing when the generation becomes terminal. A later report that disagrees goes into the evidence, and the record marks the observations as conflicting.

## Evidence, retention, and cleanup

From the agent host, the provider publishes each item of evidence as an evidence reference with the item's media type and digest. Evidence includes these items:

- The agent's transcript and its standard error log
- The rendered instructions
- The combined sandbox configuration and the sandbox violations it reported
- The agent environment
- The combined hook list, and the record and logs of each hook run, as [Hook evidence](agent-hooks.md#hook-evidence) describes
- The project configuration files with [workspace repository hooks](agent-hooks.md#workspace-repository-hooks)
- The raw final answer
- The workspace changes
- Any conflicting observations from the provenance record

Status returns the references published so far. Callers read evidence through those references and never from the agent host's disk.

Execution cleanup concerns a generation's retained copy of the workspace and evidence. How that retention protects the agent's current workspace and conversation remains an [open question](agent-open-questions.md#agents).

After the execution ends, the provider retains its workspace and evidence for the specification's retention period. Then cleanup removes them, under these rules:

- A retention hold on an execution protects its workspace and evidence past the retention period until someone releases the hold.
- Cleanup runs on the placed agent host. It skips an execution that's non-terminal, held, still has running processes, or has a workspace that a `pending` execution on the same host lists as a source, and it checks those conditions again right before it removes anything.
- Cleanup removes only the items that belong to that execution, and it reports any item it couldn't remove. With items left over, cleanup isn't complete.
- After cleanup, result returns `expired` for that execution, while status still reports its terminal state.
- When an execution ends, the provider records one execution-ended event with the execution ID and generation, the terminal state and reason, the duration, and the labels. A repeated observation of the same end doesn't record another event.
