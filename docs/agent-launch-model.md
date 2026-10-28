---
type: Design
title: Agent launch model
description: The fields and identity of each agent launch, who sets each one, which never change, and how they relate to profiles, workspaces, results, evidence, and provenance.
tags: [afakt, agents, agent-launch, model]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent launch model

[Public key infrastructure](pki.md).

An [agent launch](agent-launch.md) instantiates one durable [agent](agent.md) in an agent sandbox on a workload host. Its agent launch ID identifies that instantiation. Retry, resume, migration, and crash recovery create another agent launch with a predecessor link. The workload provider keeps each launch's record on its placed workload host. Status and result read that record.

[Deletion and teardown](agent-workspace-lifecycle.md#Deletion%20and%20teardown).

## Agent launch record

| Field | Type and format | Meaning | Set by | Changes |
| --- | --- | --- | --- | --- |
| Agent launch ID | Canonical universally unique identifier (UUID) | Identifies this agent launch. See [Agent launch ID format](#Agent%20launch%20ID%20format). | afakt | Never |
| Generation | Positive integer | The agent spec generation recorded at finalization. | The control plane increments the revision. The agent dispatcher records it | Never during this agent launch |
| Predecessor | Agent launch ID, absent for the first launch | The preceding launch on retry, resume, migration, or crash recovery. | The agent dispatcher | Never |
| Created at | Coordinated Universal Time (UTC) timestamp | When afakt created the agent launch. | afakt | Never |
| Specification | [Agent launch spec](agent-launch-contract.md#Agent%20launch%20spec) | What the workload provider runs. | The agent dispatcher | Never |
| Specification digest | Digest, as the algorithm name, `:`, and lowercase hexadecimal | Identifies the specification for repeated provision requests. | The provider, at launch | Never |
| Pinned profile | Copy of the profile version's content with its referenced files | The profile the provider validates and runs against. | The provider, at launch | Never |
| Placement | Workload host and host task queue | Where every operation for this agent launch runs. | The agent dispatcher workflow, from the placement activity | Never, once set |
| State | One of the [agent launch states](workload-lifecycle.md#Agent%20launch%20states) | Where the agent launch stands. | The provider, from its observations | Through the allowed transitions, until terminal |
| State history | List of transitions with the time of each and the observation behind it | How the agent launch reached its state. | The provider | Append-only |
| Phase | `before-agent`, `agent`, `after-agent`, `before-teardown`, or none | Which part of the agent launch is running. | The provider | With each part. None before the first `before-agent` hook or the agent start, and none once terminal |
| Last observed | UTC timestamp | When the provider last observed the hooks or the agent. | The provider | With each observation |
| Cancellation acknowledgement | Time and state at acceptance, plus the reason | The provider's confirmation of a cancellation. | The provider, on cancel | Set once |
| Superseded by | Agent launch ID | The successor agent launch that requested this launch's cancellation. | The provider | Set once |
| Agent workspace | Absolute path in the execution environment, with each source resolved | The directory tree the agent works in. See [Agent workspace](agent-workspace.md). | The provider, when it builds the workspace | Never, once built, except for the agent's own changes |
| Combined configuration | Combined agent sandbox configuration, workload environment, hooks, agent harness transport, and harness settings | What confines and steers the agent. | The provider, at launch | Never |
| Configuration digest | Digest, in the same form as the specification digest | The digest over the pinned profile digest, the rendered instructions digest, the digest of the sandbox configuration built from the profile and launch spec, the workload environment, the combined hooks, the specification's overrides, the selected harness transport, and the harness settings the provider resolves from them. | The provider, at launch on the placed workload host | Never |
| Evidence references | List of [evidence references](#Evidence,%20retention,%20and%20cleanup) | Evidence published so far. | The provider | Append-only, until cleanup |
| Provenance record | [Provenance record](#Provenance%20record) | What the agent launch requested and what the provider observed. | The provider | Requested fields never change. The provider sets each observed field once. The hook counts grow until the agent launch becomes terminal. |
| Observed times | UTC timestamps for launch accepted, agent start, agent end, and agent launch end | When each happened. | The provider | Each set once |
| Result | [Agent launch result](agent-launch-contract.md#Agent%20launch%20result) | The outcome. | The provider, on entering a terminal state | Set once |
| Retention | UTC timestamp of the end of retention, and any retention holds | How long the workspace and evidence remain. | The provider, from the specification's retention, on entering a terminal state. An operator places and releases holds. | Holds come and go |
| Cleanup | `none`, `complete`, or `incomplete`, with any items left over | Whether cleanup removed the workspace and evidence. | The provider, during cleanup | Until `complete` |

Everything the specification contains is immutable, and so is everything the provider derives from it at launch. Fields that track progress change only in the ways the table gives. Once the provider sets the result, it never changes.

## Relations

- **Agent.** Each agent launch belongs to one durable agent. The representation of this ownership in workload provider records remains [open](agent-open-questions.md#Agents).
- **Session.** The conversation belongs to the agent. An agent launch publishes its transcript as evidence. Retaining the conversation beyond evidence cleanup remains [open](agent-open-questions.md#Agents).
- **Agent profile.** Each agent launch pins one exact profile version by source, name, version, and digest, and the workload provider stores a copy of that version with the agent launch. Many agent launches can pin the same version. A successor agent launch may pin a different version. See [Pinning a profile version](agent-profile.md#Pinning%20a%20profile%20version).
- **Workspace sources.** Each agent launch builds its own agent workspace from its specification's sources. An `agent-launch` source refers to an ended agent launch on the same workload host, including its predecessor. An agent launch that the new one supersedes qualifies too, because the provider builds the new workspace only after that agent launch becomes terminal. Cleanup skips a workspace that a `pending` agent launch lists as a source.
- **Workspace images.** An `image` source records its resolved manifest digest with the agent launch. Retries use that digest, and a later agent launch can request it explicitly. A non-terminal or retained agent launch protects its image manifest, config, and layers from collection. See [Image sources and digest pinning](agent-workspace-image.md#Image%20sources%20and%20digest%20pinning) and [Retention and garbage collection](agent-workspace-image.md#Retention%20and%20garbage%20collection).
- **Agent launches.** An agent has one or more launches, with at most one running at a time. Each replacement links to its predecessor. See [Agent launches and supersession](agent-lifecycle.md#Agent%20launches%20and%20supersession).
- **Result.** Each agent launch has at most one result, produced when it enters a terminal state.
- **Evidence.** Each agent launch has its own evidence. Each item belongs to exactly one agent launch.
- **Provenance.** Each agent launch has exactly one provenance record. Status returns it while the agent launch runs, and the result includes it at the end.

## Agent launch identity

An agent launch ID identifies one instantiation. Each replacement receives a new ID linked to its predecessor, while repeated workload provider operations for an existing ID reconcile that launch without creating another instantiation.

| Provision with | Outcome |
| --- | --- |
| An unrecorded agent launch ID and a valid agent launch spec | Records the agent launch. |
| The same ID and the same specification digest | Returns the existing agent launch and its state without starting another agent harness. |
| The same ID and a different specification digest | Rejects with `specification-conflict`. The existing spec remains unchanged. |

The specification digest covers the fields of the finalized agent launch spec, including the agent spec generation. Repeating provision with different labels is still a conflict. Admission of competing successor requests and rejection of stale requests remain [open](agent-open-questions.md#%60OQ-EXE-8%60).

### Agent launch ID format

An agent launch ID is a UUIDv7 in the canonical text form of Request for Comments (RFC) 9562: 32 lowercase hexadecimal digits in groups of 8, 4, 4, 4, and 12, separated by hyphens. Its version digit is `7`, and its variant digit is one of `8`, `9`, `a`, or `b`. afakt assigns it when it creates the agent launch and records it in workflow history before provision. Callers never choose an agent launch ID, and workload providers never assign one. The exact afakt component that assigns it remains [open](agent-open-questions.md#%60OQ-MOD-1%60).

Every place that passes or records an agent launch ID uses that form, including provider operations, results, provenance, evidence, the agent-launch-ended event, the [agent launch environment](agent-workspace.md#Agent%20launch%20environment), and `${agent-launch-id}`. A workload provider rejects an ID outside the canonical form with `invalid-specification`.

## Provenance record

Each agent launch has one provenance record. Status and result both return it.

| Field | Requested | Observed |
| --- | --- | --- |
| Agent harness | Name and version from the specification | Name and version the harness reports |
| Harness transport | The transport of the selected profile model entry | The transport the workload provider reports applying |
| Model provider | From the specification | The model provider that served the requests, as reported |
| Model | From the specification | The model ID that served the requests, as reported |
| Model effort | From the specification | The effort level the harness reports applying |
| Agent profile | Source, name, version, and profile digest from the specification, plus the rendered instructions digest | Not observed, because the provider resolves and renders the profile before launch |
| Agent workspace | Each source and target path from the specification | For each `repository` source, the checked-out commit and any new branch. For each `image` source, the resolved manifest digest. For each `agent-launch` source, the digest of the copied workspace. |
| Agent sandbox | The capabilities in the agent launch spec and the digest of the sandbox configuration built from the profile and launch spec | The sandbox settings the provider reports applying |
| Environment | Each workload environment name and value, with its source in the profile or the specification | Not observed, because the provider sets the environment itself |
| Hooks | Each hook's name, event, source, matcher, decisions, and command from the combined list | Each hook's number of runs and failures so far, or that the provider skipped it, each [support warning](agent-hooks.md#Harness%20support) for an event or decision it uses, and the [workspace repository hooks](agent-hooks.md#Workspace%20repository%20hooks) |
| Workload host | Placement requirements from the specification | The workload host that ran the agent, and the version of its workload host worker |

The record also contains the canonical agent launch ID, its predecessor link, and the agent spec generation recorded at finalization, plus the specification digest, the input digest, and the configuration digest. It also gives the trace ID and span ID of the agent launch's span in the [trace of the agent launch](agent-observability.md#Trace%20topology).

Requested values and observed values are separate fields. The provider never copies one into the other. An observed field is empty until the workload provider reports a value it read from the harness, the model provider, or the workload host. Empty means that no source reported the value, never that the observed value matches the requested one. Once the provider records an observed value for an agent launch, that value doesn't change. The hook run and failure counts are the exception: they count the runs so far, and they stop changing when the agent launch becomes terminal. A later report that disagrees goes into the evidence, and the record marks the observations as conflicting.

## Evidence, retention, and cleanup

From the workload host, the workload provider publishes each item of evidence as an evidence reference with the item's media type and digest. Evidence includes these items:

- The agent's transcript and its standard error log
- The rendered instructions
- The combined agent sandbox configuration and the sandbox violations it reported
- The workload environment
- The combined hook list, and the record and logs of each hook run, as [Hook evidence](agent-hooks.md#Hook%20evidence) describes
- The project configuration files with [workspace repository hooks](agent-hooks.md#Workspace%20repository%20hooks)
- The raw final answer
- The workspace changes
- Any conflicting observations from the provenance record

Status returns the references published so far. Callers read evidence through those references and never from the workload host's disk.

[Workspace retention and cleanup](agent-workspace-lifecycle.md#Deletion%20and%20teardown).

When an agent launch ends, the provider records one agent-launch-ended event with the agent launch ID, the terminal state and reason, the duration, and the labels. A repeated observation of the same end doesn't record another event.
