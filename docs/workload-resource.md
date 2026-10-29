---
type: Design
title: Workload resource
description: Creation-time fields, lifecycle observations, and the boundary between user workloads and control plane deployments.
tags: [afakt, workload, resources]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T22:38:36Z" }
---

# Workload resource

The `Workload` resource represents the execution context and its [workload](workload.md) boundary. That boundary includes `afakt-agent` together with the workload proxy and its mediator. Its `type` is `agent`, `app`, or `worker`. The control plane's own deployments, including its platform Temporal worker, aren't workloads. Workers that run user code are workloads. This boundary prevents workload creation from depending on a platform worker that itself needs a workload.

afakt destroys an agent workload after its single agent launch, while apps and workers never bind an agent. App replacement and worker deployment lifetimes remain [`OQ-WLD-2`](agent-open-questions.md#%60OQ-WLD-2%60).

The contract applies to every Tier 1 cell: `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`. Provider capabilities determine the available operations, without changing creation-time immutability.

## Envelope

The envelope contains only `apiVersion`, `kind`, `metadata`, `spec`, and `status`. Field spelling, mandatory fields, and resource version concurrency remain [`OQ-RES-4`](agent-open-questions.md#%60OQ-RES-4%60). The keys below are a proposed representation of the creation-time contract.

| Key | Meaning | Who sets it |
| --- | --- | --- |
| `apiVersion` | `afakt/v1alpha1`. | Caller |
| `kind` | `Workload`. | Caller |
| `metadata.name` | Workload identifier. | Control plane |
| `metadata.creationTimestamp` | Creation time. | Control plane |
| `spec` | Immutable creation-time fields. | Control plane at creation |
| `status` | Observed lifecycle state and conditions. | Lifecycle authority, pending [`OQ-SBX-71`](agent-open-questions.md#%60OQ-SBX-71%60) |

## Specification keys

| Key | Meaning |
| --- | --- |
| `type` | Agent, app, or worker classification. |
| `role` | Pinned Identity and Access Management (IAM) role revision. |
| `authorization` | Exact revisions of every policy the role references and the policy-language version. |
| `workload_image` | Workload image reference resolved at creation. |
| `execution_mode` | Selected workload provider target. |
| `shape` | CPU, memory, and disk allocation. |
| `host` | Selected workload host. |
| `namespace` | Namespace fixed at creation. |
| `workload_template` | Workload template identity and version used at creation. |

These fields never change on an existing workload. A shared policy edit can't widen its pinned authority. Changing creation-time fields requires a new workload. Creation settings come from the workload template, while the agent profile supplies agent launch settings. The launch's IAM role name selects its own authority, with composition and binding under [`OQ-CON-15`](agent-open-questions.md#%60OQ-CON-15%60). Credential grants remain held under [`OQ-PRF-5`](agent-open-questions.md#%60OQ-PRF-5%60) and [`OQ-SBX-2`](agent-open-questions.md#%60OQ-SBX-2%60).

How creation and a claim relate to prepared capacity remains [`OQ-WLD-1`](agent-open-questions.md#%60OQ-WLD-1%60). After an agent launch ends, afakt destroys its workload.

## Status keys

| Proposed key | Meaning |
| --- | --- |
| `state` | Workload lifecycle state. The state vocabulary remains [`OQ-SBX-71`](agent-open-questions.md#%60OQ-SBX-71%60). |
| `conditions` | Observed readiness, failure, and reclaim facts. Exact condition keys remain [`OQ-RES-4`](agent-open-questions.md#%60OQ-RES-4%60). |
| `last_observed` | Freshness of the observations. |
| `agent_id` | The bound agent, for an agent workload. |
| `agent_launch_id` | Agent launch, when bound. |
| `workspace_id` | The optional workspace attached by a one-shot claim. |

A stale observation or lost workload host connection isn't proof of termination. Workload status and agent launch status describe different lifetimes. See [workload lifecycle](workload-lifecycle.md).

## Example

An illustrative agent workload pins its creation settings and records one launch. The proposed key contract remains [`OQ-RES-4`](agent-open-questions.md#%60OQ-RES-4%60).

```yaml
apiVersion: afakt/v1alpha1
kind: Workload
metadata:
  name: example-workload
  creationTimestamp: "2026-10-30T12:00:00Z"
spec:
  type: agent
  role: { name: build, revision: "7" }
  authorization:
    policies: [{ name: repository-read, revision: "3" }]
    policy_language_version: "1"
  workload_image: example-guest-image
  execution_mode: macos-vm
  shape: { cpu: 2, memory: 4Gi, disk: 20Gi }
  host: example-workload-host
  namespace: example-project
  workload_template: { name: build, version: "2" }
status:
  state: bound
  last_observed: "2026-10-30T12:01:00Z"
  agent_id: example-agent
  agent_launch_id: example-launch
```
