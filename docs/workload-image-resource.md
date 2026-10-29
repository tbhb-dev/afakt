---
type: Design
title: Workload image resource
description: WorkloadImage identifies a selected container image, with payload and observation fields still open.
tags: [afakt, workload, resources]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:14:00Z" }
---

# Workload image resource

`WorkloadImage` identifies the [container image selected for a workload process](workload-image.md). A workload template supplies that selection, and the workload fixes its workload image reference at creation. An `empty` workload has no workload image. Payload fields and their wire names remain [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60).

## Envelope

| Key | Meaning | Who sets it |
| --- | --- | --- |
| `apiVersion` | Proposed envelope version, `afakt/v1alpha1`. | Caller |
| `kind` | `WorkloadImage`. | Caller |
| `metadata` | Workload image resource identity. | Caller and control plane |
| `spec` | Container image content and compatibility contract, pending `OQ-WLD-6`. | Container image publisher |
| `status` | Workload image observations, pending `OQ-WLD-6`. | Control plane |

## Specification keys

The workload image supplies a root filesystem and a default command. Artifact references, digest resolution, compatibility declarations, required fields, and validation rules remain [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60). That question includes how each workload provider runs or materializes the container image.

A guest image has a separate boot contract. afakt supplies its Linux kernel, init, and root disk, and the workload records each component's digest. Users don't select a guest image through `WorkloadImage`. Guest format and boot and restoration compatibility remain under `OQ-WLD-6`.

## Status keys

Workload image admission, availability, verification, and retention observations remain [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60). Building or running a container image doesn't establish those separate facts. Status keys remain open under that question.

## Example

A workload image identity uses the common envelope. Empty payload objects below leave the unresolved fields unspecified, and don't establish a valid creation request.

```yaml
apiVersion: afakt/v1alpha1
kind: WorkloadImage
metadata:
  name: example-agent-image
spec: {}
status: {}
```
