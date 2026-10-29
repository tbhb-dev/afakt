---
type: Design
title: Workload image
description: Execution images across virtual machines, container managers, and host providers, separate from workspace content.
tags: [afakt, workload, images]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:14:00Z" }
---

# Workload image

A workload image supplies the execution environment selected at workload creation. On a virtual machine (VM) provider, it supplies the disk and kernel the [workload guest](workload-guest.md) boots. Container managers use container images, while host providers use sandbox roots without guest kernels. The workload fixes its image reference at creation. Its [workload template](workload-template.md) supplies that setting.

An [agent workspace image](agent-workspace-image.md) supplies the prepared workspace directory tree. Workspace sources and harness state occupy different layers from the execution environment. A workload can exist without a workspace. Workload snapshots capture VM memory plus disk. See [Suspend and resume](workload-lifecycle.md#Suspend%20and%20resume).

## Provider forms

| Tier 1 cell | Image form and boundary |
| --- | --- |
| `linux-vm` on `linux/amd64` and `linux/arm64`, `macos-vm` on `darwin/arm64`, and `windows-vm` on `windows/amd64` | Guest disk and kernel with compatibility for the guest operating system as well as the workload provider's VM format. |
| `docker-engine` on `linux/amd64` and `linux/arm64` | Container image for the workload sandbox container, distinct from the workload proxy container. |
| `linux-host` on `linux/amd64` and `linux/arm64` | Sandbox root compatible with the host's process environment. |
| `macos-host` on `darwin/arm64` | Sandbox root compatible with the host's process environment. |
| `windows-host` on `windows/amd64` | Sandbox root compatible with the host's process environment. |

A workload image doesn't supply the proxy's identity or select its process placement. Workload provider isolation and image compatibility are separate contracts.

## Compatibility and content

The guest root, writable workspace, and harness configuration have distinct contents. Payload keys remain open. [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60) covers image format, kernel and disk binding, compatibility declarations, digest resolution, admission, and retention for the [WorkloadImage resource](workload-image-resource.md).

An app release image is distinct from the standard guest that runs it. Whether a container or host workload uses the release image directly remains [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60). Apps don't bind an agent workspace.

## Boot and restoration

Running a harness in a confined root filesystem doesn't establish VM boot or restoration compatibility. The guest image format and boot and restoration compatibility across Tier 1 cells remain [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60).
