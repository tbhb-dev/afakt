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

The same envelope applies to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`. See the [Tier 1 matrix](https://github.com/tbhb-dev/afakt-drafts/blob/a4e663c7217fecfe7e4b2582e0b78e6291b42416/workloads/workload-platform-support.md).

## Envelope

| Key | Meaning | Who sets it |
| --- | --- | --- |
| `apiVersion` | `afakt/v1alpha1`. | Caller |
| `kind` | `WorkloadTemplate`. | Caller |
| `metadata` | Template identity and descriptive metadata. Version representation remains open. | Caller and control plane |
| `spec` | Workload creation settings. | Authorized caller |
| `status` | Observations about preparation, if the preparation proposal applies. | Control plane under the open preparation contract |

## Specification keys

These proposed keys follow the [Workload resource](workload-resource.md#Specification%20keys). The reference schema uses `guest_image` and `agent_host`. Its fields omit `type` and version metadata. That mismatch remains [`OQ-RES-5`](agent-open-questions.md#%60OQ-RES-5%60).

| Proposed key | Meaning |
| --- | --- |
| `type` | Agent, app, or worker classification. |
| `role` | The workload's Identity and Access Management (IAM) role. |
| `workload_image` | The [workload image](workload-image-resource.md). |
| `execution_mode` | Workload provider target. |
| `shape` | CPU, memory, and disk allocation. |
| `host` | Workload host selection. |
| `namespace` | Workload namespace. |
| `host_constraints` | Proposed constraints on workload host selection. |
| `preparation` | Proposed requirements for prepared state. |

Parameter rules and host constraint resolution remain [`OQ-WLD-5`](agent-open-questions.md#%60OQ-WLD-5%60). At creation, the workload records its selected template version alongside its resolved immutable fields. Pool keys and claim timing remain [`OQ-WLD-1`](agent-open-questions.md#%60OQ-WLD-1%60).

## Status keys

The reference schema declares an empty status object. The preparation proposal calls for observations of prepared snapshots by workload host, but supplies no settled key contract. Preparation level, freshness, and invalidation fields remain [`OQ-RES-5`](agent-open-questions.md#%60OQ-RES-5%60) and [`OQ-WLD-5`](agent-open-questions.md#%60OQ-WLD-5%60).

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
  workload_image: example-guest-image
  execution_mode: macos-vm
  shape: { cpu: 2, memory: 4Gi, disk: 20Gi }
  host: example-workload-host
  namespace: example-project
status: {}
```

The field inventory follows the [reference schema](https://github.com/tbhb-dev/afakt-drafts/blob/a4e663c7217fecfe7e4b2582e0b78e6291b42416/resources/schemas/workload-template.schema.json) and [template design](https://github.com/tbhb-dev/afakt-drafts/blob/a4e663c7217fecfe7e4b2582e0b78e6291b42416/workloads/resource-template-design.md#Workload%20template), with canonical names from [workload terminology](https://github.com/tbhb-dev/afakt-drafts/blob/a4e663c7217fecfe7e4b2582e0b78e6291b42416/workloads/workload-terminology.md).
