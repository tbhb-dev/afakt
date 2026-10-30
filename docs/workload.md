---
type: Design
title: Workload
description: The guest, mediator, proxy, firewall, and launch boundary of a workload.
tags: [afakt, agents, workload, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload

A workload is the resource for an execution context together with its workload proxy and workload proxy mediator. Its supervisor starts a workload process when the workload type requires one. For an agent workload, that supervisor is `afakt-agent`. On a virtual machine (VM) provider, the execution context is a workload guest. Other workload providers use a host or container context. A workload can exist without an agent launch and without a workspace. In afakt, workload means this boundary. Mediators, gateways, and workers are workloads in the Secure Production Identity Framework for Everyone (SPIFFE) sense, a distinct use of the term.

## Creation-time fields

At creation, the [Workload resource](workload-resource.md) records its type, role, workload image, execution mode, shape, workload host, namespace, and workload template version. These fields are immutable. Its `type` classifies it as `empty`, `job`, `agent`, `app`, or `worker`. Its workload image selects the container image for the process. An `empty` workload has neither a workload process nor a workload image. VM providers supply the Linux guest image separately.

A role groups Identity and Access Management (IAM) policies under a name and binds to the workload at creation, like an IAM instance profile on an EC2 instance. Workload templates store the creation-time fields, with commands under `afakt workload template`. For an agent launch, an agent profile supplies the configuration.

## Workload types

| Type | Process and lifetime |
| --- | --- |
| `empty` | Runs the supervisor without a workload process until destruction. |
| `job` | Runs one process from its workload image to completion, then afakt destroys the workload. |
| `agent` | Runs one agent launch, then afakt destroys the workload. |
| `app` | Runs one process of an app release, with supervisor restarts, until replacement or scale-down. |
| `worker` | Runs a Temporal worker for one deployment version until that version drains. |

A job runs once. Retrying a failed job requires a new workload, so a later run can't inherit files or secrets from the failed process. Detailed lifecycle rules remain open, and the job specification and release configuration fields remain proposed, under [`OQ-WLD-2`](agent-open-questions.md#%60OQ-WLD-2%60).

## Components

- Within the execution context, including the [workload guest](workload-guest.md) on a VM provider, the harness and workspace commands run alongside routing components that don't make policy decisions.
- The [workload proxy](workload-proxy.md) transmits traffic between the guest and network in both directions, including DNS, TCP, User Datagram Protocol (UDP), and QUIC.
- The [workload proxy mediator](workload-proxy-mediator.md) decides each inbound and outbound connection. It never transmits traffic.
- The [`afakt-agent` command](agent-acp.md#Running%20the%20harness) runs inside the guest context and interacts with the harness.

## Firewall

Each workload has a firewall with ingress and egress rules. Egress defaults to a cluster-managed deny-all policy. Network restrictions, including allowed domains, belong to the firewall. Other sandbox restrictions belong to the sandbox policy. Host-mode workloads keep the firewall: host-mode capture sends their traffic through the same proxy enforcement point. The [firewall resource format](agent-open-questions.md#%60OQ-SBX-67%60) is open.

For every connection in either direction, the proxy requests the mediator's decision. A rule change ends existing connections the new rules deny. Revocation of persistent guest-initiated egress and UDP flows remains [open](agent-open-questions.md#%60OQ-SBX-69%60).

## Proxy and mediator

Traffic follows guest ⇄ workload proxy ⇄ network. The mediator supplies decisions outside that path. For outbound connections whose server name matches a domain with an HTTP-level rule or credential injection, the mediator can instruct the proxy to end Transport Layer Security (TLS).

For its own connections to secrets, policy bundles, and the API server, the mediator uses its proxy under a fixed rule set built into the workload. Neither the launch's policy nor the guest can change those rules. Control plane pushes arrive on a long-lived stream opened by the mediator. A control plane connection to a workload always starts at the mediator.

Over the [overlay](workload-overlay.md), workload proxies connect directly, each with its own node. Both ends authenticate with SPIFFE mutual TLS, and both mediators check a connection between workloads.

## Lifecycle

An agent workload runs exactly one agent launch, and afakt destroys it afterwards. App and user code worker lifetimes remain [open](agent-open-questions.md#%60OQ-WLD-2%60). A follow-up or relaunch is a new launch on a new workload. The [lifecycle owner and states](agent-open-questions.md#%60OQ-SBX-71%60) remain open.
