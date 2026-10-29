---
type: Design
title: Workload image resource
description: WorkloadImage identity and the unresolved payload and observation contract for provider-specific execution images.
tags: [afakt, workload, resources]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:14:00Z" }
---

# Workload image resource

`WorkloadImage` identifies a [workload execution image](workload-image.md). A workload template selects it, and a workload fixes its image reference at creation. Payload fields and their wire names remain [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60).

This boundary applies to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`.

## Envelope

| Key | Meaning | Who sets it |
| --- | --- | --- |
| `apiVersion` | Proposed envelope version, `afakt/v1alpha1`. | Caller |
| `kind` | `WorkloadImage`. | Caller |
| `metadata` | Image resource identity. | Caller and control plane |
| `spec` | Image content and compatibility contract, pending `OQ-WLD-6`. | Image publisher |
| `status` | Image observations, pending `OQ-WLD-6`. | Control plane |

## Specification keys

Each image form needs content whose field keys remain open.

| Workload provider kind | Content | Open field contract |
| --- | --- | --- |
| VM provider | Guest disk and kernel. | Artifact references, digests, boot compatibility, and the binding between disk and kernel. |
| Container manager provider | Container image for the workload sandbox container. | Image reference, resolution, and compatibility. |
| Host provider | Sandbox root for the host process environment. | Root representation, materialization, and compatibility. |

[`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60) holds these keys, required fields, and validation rules. The [agent workspace image](agent-workspace-image.md#Image%20contents) contract doesn't establish an execution image's payload format.

## Status keys

Image admission, availability, verification, and retention observations remain [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60). Building or booting an image doesn't establish those separate facts. Status keys remain open under that question.

## Example

An image identity uses the common envelope. Empty payload objects below leave the unresolved fields unspecified, and don't establish a valid creation request.

```yaml
apiVersion: afakt/v1alpha1
kind: WorkloadImage
metadata:
  name: example-guest-image
spec: {}
status: {}
```
