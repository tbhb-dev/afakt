---
type: Design
title: Workload template resource
description: WorkloadTemplate creation settings, version references, and unresolved preparation observations.
tags: [afakt, workload, resources]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:14:00Z" }
---

# Workload template resource

`WorkloadTemplate` is the resource for reusable [workload creation settings](workload-template.md). Its version supplies provenance for the workload's immutable creation record. Exact field spelling, version identifiers, and required fields remain [`OQ-RES-5`](agent-open-questions.md#%60OQ-RES-5%60).

The same envelope applies to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`.

## Envelope

| Key | Meaning | Who sets it |
| --- | --- | --- |
| `apiVersion` | `afakt/v1alpha1`. | Caller |
| `kind` | `WorkloadTemplate`. | Caller |
| `metadata` | Template identity and descriptive metadata. Version representation remains open. | Caller and control plane |
| `spec` | Workload creation settings. | Authorized caller |
| `status` | Observations about preparation, pending `OQ-RES-5` and `OQ-WLD-5`. | Control plane under the open preparation contract |

## Specification keys

These proposed keys represent the [Workload resource](workload-resource.md#Specification%20keys) creation settings. Their wire names and the representation of workload type and version metadata remain [`OQ-RES-5`](agent-open-questions.md#%60OQ-RES-5%60).

| Proposed key | Meaning |
| --- | --- |
| `type` | `empty` without a workload process, `job` for one process to completion, `agent` for one launch, `app` for a release process, or `worker` for a Temporal worker. |
| `role` | The workload's Identity and Access Management (IAM) role. |
| `workload_image` | The [container image selected for the workload process](workload-image-resource.md). Absent for `empty`. |
| `execution_mode` | Workload provider target. |
| `shape` | CPU, memory, and disk allocation. |
| `host` | Workload host selection. |
| `namespace` | Workload namespace. |
| `host_constraints` | Proposed constraints on workload host selection. |
| `preparation` | Proposed requirements for prepared state. |

Parameter rules and host constraint resolution remain [`OQ-WLD-5`](agent-open-questions.md#%60OQ-WLD-5%60). At creation, the workload records its selected template version alongside its resolved immutable fields. Pool keys and claim timing remain [`OQ-WLD-1`](agent-open-questions.md#%60OQ-WLD-1%60).

## Status keys

Observations of prepared snapshots by workload host, including fields for preparation level and the freshness or invalidation of each observation, remain [`OQ-RES-5`](agent-open-questions.md#%60OQ-RES-5%60) and [`OQ-WLD-5`](agent-open-questions.md#%60OQ-WLD-5%60).

## Example

A template selects an agent workload on `macos-vm` for `darwin/arm64`. The key spellings and omitted version representation remain [`OQ-RES-5`](agent-open-questions.md#%60OQ-RES-5%60).

```yaml
apiVersion: afakt/v1alpha1
kind: WorkloadTemplate
metadata:
  name: build
spec:
  type: agent
  role: { name: build, revision: "7" }
  workload_image: example-agent-image
  execution_mode: macos-vm
  shape: { cpu: 2, memory: 4Gi, disk: 20Gi }
  host: example-workload-host
  namespace: example-project
status: {}
```
