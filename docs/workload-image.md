---
type: Design
title: Workload image
description: Container images selected for workload processes, separate from Linux guest boot content and workspace trees.
tags: [afakt, workload, images]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:14:00Z" }
---

# Workload image

A workload image is the container image a [workload](workload.md) selects at creation for its workload process. An `empty` workload doesn't run a workload process or select a workload image. A [workload template](workload-template.md) supplies the selection when the workload needs one. The workload fixes that reference at creation.

## Container images

A container image contains a root filesystem and a default command in Open Container Initiative (OCI) format. Jobs, app releases, and agents select container images for their processes, so their selected workload images can also take those type names. OCI is a format. Workspace images and agent snapshots use it too.

A test image is a container image afakt's tests run as a job to probe a workload's network path.

## Guest images

A guest image supplies the Linux kernel, init, and root disk that a [workload guest](workload-guest.md) boots. Every workload guest runs Linux, regardless of the workload host's operating system. Both `macos-vm` and `windows-vm` run Linux guests.

afakt builds and publishes guest images. Users don't choose them. Each workload records the digests of the guest components it uses. A workload proxy VM boots its own guest image.

A guest image and a workload image have separate purposes. The guest boots afakt's Linux environment, while the workload image supplies the process's filesystem and default command. Whether that process runs directly in the guest or in a workload container remains [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60). The guest format, component binding, and boot and restoration compatibility also remain under that question. An outside VM image, such as a Linux distribution's cloud image, can supply inputs for building a guest image.

## Workspace images

A [workspace image](agent-workspace-image.md) contains a prepared workspace directory tree in OCI layers. It never runs and isn't a container image. Workspace content and harness state occupy different layers from the workload process's filesystem. A workload can exist without a workspace.

## Compatibility and admission

The [WorkloadImage resource](workload-image-resource.md) identifies the container image selected for the process. It doesn't select proxy placement or identity. Its payload keys, reference resolution, admission, retention, and compatibility with each workload provider remain [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60). That question includes how host providers materialize the selected content for their native process environment.
