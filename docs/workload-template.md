---
type: Design
title: Workload template
description: Versioned workload creation settings, claim selection, and the open relationship to prepared capacity.
tags: [afakt, workload, templates]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:14:00Z" }
---

# Workload template

Workload templates store creation-time settings for a [workload](workload-resource.md): type, Identity and Access Management (IAM) role, workload image, execution mode, shape, workload host, and namespace. Each workload records its template version. For agent launches, the [agent profile](agent-profile.md) supplies launch settings. Commands for workload templates use `afakt workload template`.

The contract applies to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`. The [platform matrix](https://github.com/tbhb-dev/afakt-drafts/blob/a4e663c7217fecfe7e4b2582e0b78e6291b42416/workloads/workload-platform-support.md) defines these Tier 1 cells. A template's execution mode and [workload image](workload-image.md) must suit its selected cell.

## Creation and versioning

The workload fixes its creation-time fields, including its template version. It also pins the role revision, every referenced policy revision, and the policy-language version. A shared policy edit can't widen an existing workload's authority. See [Specification keys](workload-resource.md#Specification%20keys).

Template version identifiers, edits, parameters, defaults, and catalog visibility remain [`OQ-WLD-5`](agent-open-questions.md#%60OQ-WLD-5%60). Callers need rules for selecting versions and supplying parameters, separately from the workload's immutable creation record. See the [template resource](workload-template-resource.md).

## Claims and the pool key

Claims can select by workload template or by workload type, role, and other creation-time fields. A launch's IAM role name expresses the authority it needs. How that role composes with the workload's role remains [`OQ-CON-15`](agent-open-questions.md#%60OQ-CON-15%60). A claim can't change an existing workload's creation-time fields.

The proposal to use the workload template as the pool key competes with preparation keyed by image, shape, execution mode, and workload host. Pool ownership, the creation boundary, and the relationship between a template selection and a workload claim remain [`OQ-WLD-1`](agent-open-questions.md#%60OQ-WLD-1%60). Neither a pool entry nor a template grants credentials under the held [`OQ-PRF-5`](agent-open-questions.md#%60OQ-PRF-5%60) and [`OQ-SBX-2`](agent-open-questions.md#%60OQ-SBX-2%60).

## Prepared state

A template describes configuration. Prepared content resides in images and snapshots: workload snapshots contain virtual machine (VM) memory plus disk, while agent snapshots contain portable workspace and harness layers. The proposal for configuration, disk, and memory preparation levels depends on workload provider capabilities. It doesn't establish memory restoration for every target. Preparation and version invalidation remain [`OQ-WLD-5`](agent-open-questions.md#%60OQ-WLD-5%60), while capture eligibility remains [`OQ-WLD-3`](agent-open-questions.md#%60OQ-WLD-3%60).

The creation contract follows [workload terminology](https://github.com/tbhb-dev/afakt-drafts/blob/a4e663c7217fecfe7e4b2582e0b78e6291b42416/workloads/workload-terminology.md). Versioning and preparation questions follow [resource template design](https://github.com/tbhb-dev/afakt-drafts/blob/a4e663c7217fecfe7e4b2582e0b78e6291b42416/workloads/resource-template-design.md#Workload%20template).
