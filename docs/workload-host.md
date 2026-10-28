---
type: Design
title: Workload host
description: A machine's role in the workload fleet, its launch services, trust boundary, and recovery behavior.
tags: [afakt, agents, workload-host]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload host

[Public key infrastructure](pki.md).

A workload host is a machine that runs workloads for afakt. Its workload host worker connects it to the control plane and runs the [agent launcher workflow](components.md#Agent%20launcher). The [agent dispatcher workflow](components.md#Agent%20dispatcher) selects a host task queue. It finalizes the agent launch spec and queues the launch there, binding every operation for that launch to the chosen host.

[Workload host caches and startup](agent-workspace-lifecycle.md#Workload%20host%20caches%20and%20startup).

## What runs on a workload host

| Part | Responsibility |
| --- | --- |
| Workload host worker and capacity controller | The worker runs placed agent launcher workflows and their activities. Its [capacity controller](workload-placement.md#Slot%20capacity) adjusts agent launch slots within configured bounds from central processing unit (CPU) use, memory use, and free workspace disk space. |
| Workload provider | The [VM provider](workload-sandbox.md) is the default. It prepares and confines agent launches and supplies provision, restore, release, status, cancel, and result operations under the [provider operations](agent-launch-contract.md#Workload%20provider%20operations). |
| Workload host supervisor | The supervisor controls virtual machine monitors independently of the worker, guest, and Agent Client Protocol (ACP). It starts, suspends, resumes, stops, and snapshots VMs and verifies forced stops. See [Virtual machine lifecycle](workload-lifecycle.md#Virtual%20machine%20lifecycle). |
| Workload guest | Each agent has a guest kernel, a read-only operating system disk, and an independent writable workspace disk. The guest contains the relay, agent harness, adapter when needed, hooks, and helpers. See [Guest layout and lifecycle control](workload-sandbox.md#Guest%20layout%20and%20lifecycle%20control) and [Image contents](agent-workspace-image.md#Image%20contents). |
| Workload proxy | One per workload transmits guest traffic and opens permitted connections with the launch's identity. See [Per-launch services](workload-proxy.md#Launch%20boundary). |
| Workload proxy mediator | Each workload's mediator supplies DNS resolution and connection decisions to its proxy. The proxy transmits all guest traffic. The workload proxy ends Transport Layer Security (TLS) for inspected destinations. Traffic to passthrough destinations remains mediated without decryption. See [Workload proxy mediator](workload-proxy-mediator.md#Launch%20boundary). |
| Credential broker | A host-side component that injects credentials at the workload proxy into permitted requests outside the guest. Its placement relative to the workload proxy and workload proxy mediator remains [open](agent-open-questions.md#%60OQ-SBX-23%60). |
| Guest relay's host side | The proxy transmits guest relay traffic under the mediator's ACP message checks. It stamps identity on control requests. The relay itself runs inside the guest. See [Control requests](workload-proxy-mediator.md#Control%20requests) and [ACP message policy](agent-acp.md#ACP%20message%20policy). |
| Telemetry endpoint | The endpoint receives harness telemetry through the workload proxy under mediator decisions and adds agent launch identity. It filters content and forwards it to the deployment's collector. See [Trace context propagation](agent-observability.md#Trace%20context%20propagation). Whether it belongs to the worker or a separate component remains [open](agent-open-questions.md#Agent%20observability). |
| Port forwarder | The [path from the workload proxy to guest listeners](agent-open-questions.md#%60OQ-PORT-14%60) remains open. Each stream requires publication and agent launch ID checks. See [Port forwarder](components.md#Port%20forwarder) and [HTTP previews](workload-previews.md#HTTP%20previews). |
| Layer and image caches | The worker keeps verified layer blobs, unpacked layers, and reusable workspace base disks outside writable agent workspaces. Each agent gets an independent writable disk. See [Workload host caches and startup](agent-workspace-image.md#Workload%20host%20caches%20and%20startup). |

Workspace-local Model Context Protocol (MCP) servers also run inside the workload guest when a configured server needs the workspace. The [MCP gateway](agent-mcp-gateway.md#Server%20hosting%20and%20lifecycle) connects to them through the workload proxy under mediator decisions. Central servers run on the control plane.

## Fleet membership and capacity

At startup, the worker registers its workload host, task queue, metadata, and slot capacity bounds with the [workload host registry](workload-placement.md#Workload%20host%20registry). It polls only its own host task queue. When renewing its registration lease, the worker reports metadata, capacity, the most recent capacity change reason, and the agent launch IDs holding slots. Metadata describes platform, harnesses, agent sandbox runtime, labels, and resolvable credential references.

The [reservation Application Programming Interface (API)](workload-placement.md#Reservation%20API) selects an eligible host with a free slot. Reserved and claimed slots both count against capacity. Lowering capacity preserves running agent launches and existing reservations, blocking new reservations until capacity permits them. [Image locality](workload-placement.md#Image%20locality) is advisory and never overrides requirements, host binding, or slot accounting. [Preheating](agent-workspace-image.md#Placement%20and%20preheating) uses host task queues and respects disk pressure without reserving future slots.

Each workload's proxy accepts direct authenticated [overlay connections](workload-proxy.md#Overlay%20connections) from permitted peers. It also opens permitted API and service connections using that launch's identity. The mediator enforces policy and has no network of its own. User-facing traffic enters through the ingress gateway. Traffic between launches on the same host uses the overlay, with no worker shortcut. Every host also supplies [host-level DNS resolution and authorized connectivity](workload-overlay.md#Workload%20host%20access). The mechanism remains undecided.

## Runtime privileges and platforms

The workload host worker and host-mode agents use separate dedicated users under [Workload host access](workload-overlay.md#Workload%20host%20access). Provider privileges and platform-specific capture services follow [Enforcement](workload-sandbox.md#Enforcement), without settling the held [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60).

afakt permits all the modes in this table: [workload providers](agent-launch-contract.md#Workload%20provider%20operations) use host modes or container managers, or they use virtual machine monitors.

| Mode | Provider kind | Platform |
| --- | --- | --- |
| `docker-desktop` | Container manager | The Docker virtual machine (VM) on macOS and Windows. Windows uses Windows Subsystem for Linux (WSL2) or Hyper-V. |
| `docker-engine` | Container manager | Docker Engine on amd64 Linux bare metal and VMs, and arm64 Linux VMs. |
| `macos-vm` | Virtual machine | The macOS microVM design using Virtualization.framework. |
| `windows-vm` | Virtual machine | The Linux microVM inside WSL2 by default. A Windows-native utility VM for an agent needing a Windows guest. |
| `linux-vm` | Virtual machine | A Kernel-based Virtual Machine (KVM) microVM driven by Cloud Hypervisor. |
| `macos-host` | Host mode | macOS. |
| `windows-host` | Host mode | Windows. |
| `linux-host` | Host mode | Linux. |
| [Claude Managed Agents](workload-provider-claude-managed-agents.md) | Remote provider | Managed harness and cloud execution. |
| [OpenAI Agents API](workload-provider-openai-agents-api.md) | Remote provider | Managed harness and cloud execution. |
| [Modal Sandboxes](workload-provider-modal.md) | Remote provider | Remote Linux sandbox execution. |
| [Cloudflare Sandbox SDK](workload-provider-cloudflare.md) | Remote provider | Remote Linux sandbox execution. |

Host modes share their host kernel. The selected provider's isolation boundary and [afakt-sandbox enforcement](workload-sandbox.md#Enforcement) determine their confinement, while managed harness integration remains [`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60).

## Trust boundary

Trust in the host covers VM control and the connection policy of the workload proxy mediator. It also covers the workload proxy's Hypertext Transfer Protocol (HTTP) policy and the credential broker's [credential custody](workload-sandbox.md#Credential%20brokering). Host configuration defines the workload proxy's certificate authority custody. The credential broker receives credentials according to their types. Brokered secrets and signing material remain outside the agent sandbox, and the guest receives neither direct brokered-secret delivery nor SPIFFE Verifiable Identity Documents (SVIDs). For the agent, the mediator stamps identity and its workload proxy authenticates control requests with the launch's SVID. Receiving services handle workload mutual TLS termination over the overlay. User-facing traffic enters through the ingress gateway. A guest can't choose another agent's identity.

Profiles select brokered credentials, and network rules bound each injection. Bounds on repository profile self-grants and credential domains remain an on-hold [operator decision](agent-open-questions.md#Agent%20profile). Protection of provider-installed hooks and the scope of repository project configuration also remain [undecided](agent-open-questions.md#Agent%20hooks), as sandbox confinement doesn't settle either question. The session-end wait and SVID lifetime remain [on hold together](agent-open-questions.md#%60OQ-HOOK-18%60).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

## Workload identity

afakt identifies its workloads with SPIFFE. Each control plane has one trust domain, and every workload identity is a SPIFFE ID of the form `spiffe://<trust-domain>/...`. A trust domain doesn't federate with other control planes.

Connections from a workload host to the control plane, including workload proxy connections, use mutual Transport Layer Security (TLS) with X.509 SPIFFE Verifiable Identity Documents (SVIDs). SVIDs come from a Workload API endpoint on every workload host. On behalf of the agents they serve, host-side processes request SVIDs from that endpoint, which agents can't connect to. Neither an SVID nor its private key enters the guest.

Inside a SPIFFE ID, the path scheme remains [open](agent-open-questions.md#%60OQ-MOD-9%60), as does the [SVID form for other connections](agent-open-questions.md#%60OQ-SBX-17%60).

## Restart and recovery

A [worker restart](agent-launch.md#Reconciliation%20after%20restart%20or%20retry) leaves the supervisor, virtual machine monitors, workload proxy mediators, and running agents alive. Agent launch records persist on the workload host. The restarted worker loads them and observes each agent before answering status. Before provision, the agent launcher reads status and adopts an existing agent launch. It provisions only on `not-found` and retries `unavailable` on the same host. Interrupted preparation fails after the workload provider observes that every process from the interrupted step has exited. Interrupted hooks follow the hook failure rules. Output validation and result construction can restart.

A missed watch heartbeat or lost host contact never proves termination. The agent launcher retains the [host binding](agent-launch.md#Placement%20on%20workload%20hosts).

An expired registration lease stops new placement until the worker registers again. Agent launches and reservations are unchanged. Unclaimed reservations still expire under their own bounds. Terminal agent launches release their slots, and lease renewals reconcile slots no longer reported. See [Workload host registry](workload-placement.md#Workload%20host%20registry) and [Slot reservations](workload-placement.md#Slot%20reservations).

Recovery after a whole-machine restart remains [open](agent-open-questions.md#Workload%20host), separately from recovery after a worker restart. Handling a host that never returns also remains [open](agent-open-questions.md#Agent%20launch). Neither lost contact nor an expired lease authorizes running the same agent launch on another host.

Capture eligibility and restoration follow [Suspend and resume](workload-lifecycle.md#Suspend%20and%20resume), with the provider-specific limits in [Provider lifecycle contracts](workload-lifecycle.md#Workload%20provider%20lifecycle%20contracts).
