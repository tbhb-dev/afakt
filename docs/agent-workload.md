---
type: Design
title: Agent workload
description: The guest, mediator, proxy, firewall, and launch boundary of an agent workload.
tags: [afakt, agents, workload, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent workload

A workload consists of the guest, the agent mediator, the agent proxy, and `afakt-agent`. The guest is a virtual machine (VM), or the host or container context in non-VM modes. A workload can exist without an agent launch and without a workspace. In afakt, workload means this boundary. Mediators, gateways, and workers are workloads in the Secure Production Identity Framework for Everyone (SPIFFE) sense, a distinct use of the term.

## Components

- Within the [guest](agent-workload-guest.md), the harness and workspace commands run alongside routing components that don't make policy decisions.
- The [agent proxy](agent-workload-proxy.md) transmits traffic between the guest and network in both directions, including DNS, TCP, User Datagram Protocol (UDP), and QUIC.
- The [agent mediator](agent-workload-mediator.md) decides each inbound and outbound connection. It never transmits traffic.
- The [`afakt-agent` command](agent-acp.md#Running%20the%20harness) runs inside the guest context and interacts with the harness.

## Firewall

Each workload has a firewall with ingress and egress rules. Egress defaults to a cluster-managed deny-all policy. Network restrictions, including allowed domains, belong to the firewall. Other sandbox restrictions belong to the sandbox policy. Host-mode workloads keep the firewall: host-mode capture sends their traffic through the same proxy enforcement point. The [firewall resource format](agent-open-questions.md#%60OQ-SBX-67%60) is open.

For every connection in either direction, the proxy requests the mediator's decision. A rule change ends existing connections the new rules deny. Revocation of persistent guest-initiated egress and UDP flows remains [open](agent-open-questions.md#%60OQ-SBX-69%60).

## Proxy and mediator

Traffic follows guest ⇄ agent proxy ⇄ network. The mediator supplies decisions outside that path. For outbound connections whose server name matches a domain with an HTTP-level rule or credential injection, the mediator can instruct the proxy to end Transport Layer Security (TLS).

For its own connections to secrets, policy bundles, and the API server, the mediator uses its proxy under a fixed rule set built into the workload. Neither the launch's policy nor the guest can change those rules. Control plane pushes arrive on a long-lived stream opened by the mediator. A control plane connection to a workload always starts at the mediator.

Over the [overlay](agent-overlay.md), workload proxies connect directly, each with its own node. Both ends authenticate with SPIFFE mutual TLS, and both mediators check a connection between workloads.

## Lifecycle

A workload runs exactly one agent launch, and afakt destroys it afterwards. A follow-up or relaunch is a new launch on a new workload. The [lifecycle owner and states](agent-open-questions.md#%60OQ-SBX-71%60) remain open.
