---
type: Design
title: Workload guest
description: Guest traffic routing, containment, and the private link to the workload proxy.
tags: [afakt, agents, workload, sandbox, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload guest

A workload guest is the Linux virtual machine (VM) of a [workload](workload.md) on a VM provider. Every workload guest runs Linux, including `macos-vm` and `windows-vm`. It boots a [guest image](workload-image.md#Guest%20images) built and published by afakt. The workload records each guest component's digest. Users don't choose the guest image. Its format and boot and restoration compatibility remain [`OQ-WLD-6`](agent-open-questions.md#%60OQ-WLD-6%60). Guest components only route traffic. They don't make policy decisions. The [workload proxy](workload-proxy.md) transmits traffic, and the [workload proxy mediator](workload-proxy-mediator.md) allows or denies each connection.

## Containment

External traffic, including Domain Name System (DNS) queries, exits the guest only through its proxy. For DNS, the mediator resolves queries forwarded by the proxy without falling back to the workload host resolver. Containment holds against a guest compromised to root.

Each workload's link to its proxy is private to that workload and doesn't require encryption. Platforms may use different links. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft.

## Host modes

Host modes have no separate guest. Containment applies against the agent at its sandbox privileges. The enforcement mechanism belongs to the held [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60).
