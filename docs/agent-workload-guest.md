---
type: Design
title: Agent workload guest
description: Guest traffic routing, containment, and the private link to the workload proxy.
tags: [afakt, agents, workload, sandbox, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent workload guest

The guest is the execution context within an [agent workload](agent-workload.md): a virtual machine (VM), or the host or container context in non-VM modes. Guest components only route traffic. They don't make policy decisions. The [agent proxy](agent-workload-proxy.md) transmits traffic, and the [agent mediator](agent-workload-mediator.md) allows or denies each connection.

## Containment

External traffic, including Domain Name System (DNS) queries, exits the guest only through its proxy. For DNS, the mediator resolves queries forwarded by the proxy without falling back to the host resolver. Containment holds against a guest compromised to root.

Each workload's link to its proxy is private to that workload and doesn't require encryption. Platforms may use different links. The [conditions each platform must meet](agent-open-questions.md#%60OQ-SBX-70%60) remain a working draft.

## Host modes

Host modes have no separate guest. Containment applies against the agent at its sandbox privileges. The enforcement mechanism belongs to the held [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60).
