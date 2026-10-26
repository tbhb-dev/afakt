---
type: Design
title: Agent host
description: A machine's role in the agent fleet, its execution services, trust boundary, and recovery behavior.
tags: [afakt, agents, agent-host]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent host

An agent host is a machine that runs agents, with an afakt worker that polls its host task queue for agent work. Fleet hosts supply execution capacity. The [API server](components.md#api-server) keeps their registry and reserves slots. Each execution starts with placement by the [agent runner workflow](components.md#agent-runner), which starts the agent executor workflow on the chosen host. That host's [agent executor workflow](components.md#agent-executor) calls its execution provider and returns the result. Placement keeps later operations on the same host.

## What runs on an agent host

| Part | Responsibility |
| --- | --- |
| afakt worker and capacity controller | The worker runs placed agent executor workflows and their activities. Its [capacity controller](agent-placement.md#slot-capacity) adjusts execution slots within configured bounds from CPU use, memory use, and free workspace disk space. |
| Execution provider | The [host agent provider](agent-sandbox.md) is the default. It prepares and confines executions and supplies launch, status, cancel, and result operations under the [provider requirements](agent-provider-requirements.md#4-provider-operations). |
| Agent host supervisor | The supervisor controls virtual machine monitors independently of the worker, guest, and Agent Client Protocol (ACP). It starts, suspends, resumes, stops, and snapshots VMs and verifies forced stops. See [Virtual machine lifecycle](agent-execution.md#virtual-machine-lifecycle). |
| Sandbox VM | Each agent has a guest kernel, a read-only operating system disk, and an independent writable workspace disk. The guest contains the relay, harness, adapter when needed, hooks, and helpers. See [Guest layout and lifecycle control](agent-sandbox.md#guest-layout-and-lifecycle-control) and [Image contents](agent-workspace-image.md#image-contents). |
| Per-agent mediator | A confined host process outside the guest supplies its only network route and applies network and HTTP policy. It inspects permitted encrypted traffic and records decisions. See [Per-agent mediator](agent-sandbox.md#per-agent-mediator). |
| Credential broker | Inside the mediator, the broker resolves credential references on the host and injects credentials into permitted outgoing requests without exposing them to the agent. See [Credential brokering](agent-sandbox.md#credential-brokering). |
| Guest relay's host side | The mediator ends the guest relay's channel and checks ACP messages. It stamps identity on control requests. The relay itself runs inside the guest. See [Control requests](agent-sandbox.md#control-requests) and [ACP gate](agent-session.md#acp-gate). |
| Telemetry endpoint | The endpoint receives harness telemetry through the mediator and adds execution identity. It filters content and forwards it to the deployment's collector. See [Trace context propagation](agent-observability.md#trace-context-propagation). Whether it belongs to the worker or a separate component remains [open](agent-open-questions.md#agent-observability). |
| Port forwarder | Authorized streams cross the mediator and guest relay to guest loopback listeners. Each stream requires publication and generation checks. See [Port forwarder](components.md#port-forwarder) and [HTTP previews](agent-ports.md#http-previews). |
| Layer and image caches | The worker keeps verified layer blobs, unpacked layers, and reusable workspace base disks outside writable workspaces. Each agent gets an independent writable disk. See [Agent host caches and startup](agent-workspace-image.md#agent-host-caches-and-startup). |

Workspace-local Model Context Protocol (MCP) servers also run inside the sandbox VM when a configured server needs the workspace. The [MCP gateway](agent-mcp-gateway.md#server-hosting-and-lifecycle) connects to them through the guest relay and mediator. Central servers run on the control plane.

## Fleet membership and capacity

At startup, the worker registers its host, task queue, metadata, and slot capacity bounds with the [agent host registry](agent-placement.md#agent-host-registry). It polls only its own host task queue. When renewing its registration lease, the worker reports metadata, capacity, the most recent capacity change reason, and the execution IDs and generations holding slots. Metadata describes platform, harnesses, sandbox runtime, labels, and resolvable credential references.

The [reservation API](agent-placement.md#reservation-api) selects an eligible host with a free slot. Reserved and claimed slots both count against capacity. Lowering capacity preserves running executions and existing reservations, blocking new reservations until capacity permits them. [Image locality](agent-placement.md#image-locality) is advisory and never overrides requirements, host binding, or slot accounting. [Preheating](agent-workspace-image.md#placement-and-preheating) uses host task queues and respects disk pressure without reserving future slots.

The mediator relays sessions over the host's [authenticated outbound connection](agent-session.md#acp-gate) to the API server. The connection transport remains [open](agent-open-questions.md#agent-session). Preview traffic returns over the mediator's outbound connection to the [preview broker](agent-ports.md#http-previews). Both paths work without an inbound listener on the agent host.

## Runtime privileges and platforms

The [rootless runtime](agent-sandbox.md#enforcement) runs under the host user, without root privileges: a systemd user unit on Linux and a launchd agent on macOS. One-time setup may require an administrator. Whether that setup counts as rootless and how the user service continues running after logout remain [open](agent-open-questions.md#agent-sandbox).

The sandbox's [platform backends](agent-sandbox.md#enforcement) are:

| Host platform | VM backend |
| --- | --- |
| Linux | A Kernel-based Virtual Machine (KVM) microVM driven by Cloud Hypervisor. |
| macOS | Virtualization.framework. Using Apple's Containerization framework or driving Virtualization.framework directly, including the guest init choice, remains [open](agent-open-questions.md#agent-sandbox). |
| Windows | The Linux microVM inside Windows Subsystem for Linux by default. An agent that needs a Windows guest uses a Windows-native utility VM. |

The weaker host sandbox mode serves hosts without virtualization only where deployment policy permits it. Whether a deployment requires sandbox VMs everywhere or permits host sandboxes, and where, remains an on-hold [operator decision](agent-open-questions.md#agent-sandbox). A provider rejects a launch whose sandbox settings it can't enforce.

## Trust boundary

Trust in the host covers VM control, mediator policy enforcement, and [credential custody](agent-sandbox.md#credential-brokering). Host configuration defines the mediator's certificate authority custody. Brokered secrets and signing material remain outside the sandbox. The guest receives neither long-lived secrets nor brokered credential values nor the [execution token](agent-sandbox.md#execution-token). For the agent, the mediator authenticates control requests. A guest can't choose another agent's identity.

Profiles select brokered credentials, and network rules bound each injection. Bounds on repository profile self-grants and credential domains remain an on-hold [operator decision](agent-open-questions.md#agent-profile). Protection of provider-installed hooks and the scope of repository project configuration also remain [undecided](agent-open-questions.md#agent-hooks), as sandbox confinement doesn't settle either question. The session-end wait and the lifetime of the execution token remain [undecided together](agent-open-questions.md#agent-hooks).

## Restart and recovery

A [worker restart](agent-execution.md#reconciliation-after-restart-or-retry) leaves the supervisor, virtual machine monitors, mediators, and running agents alive. Execution records persist on the host. The restarted worker loads them and observes each agent before answering status. An executor checks status before launch and adopts an existing execution. It launches only on `not-found` and retries `unavailable` on the same host. Interrupted preparation fails after verified termination. Interrupted hooks follow the hook failure rules. Output validation and result construction can restart.

A missed watch heartbeat or lost host contact never proves termination. The executor retains the [host binding](agent-execution.md#placement-on-agent-hosts).

An expired registration lease stops new placement until the worker registers again. Executions and reservations are unchanged. Unclaimed reservations still expire under their own bounds. Terminal executions release their slots, and lease renewals reconcile slots no longer reported. See [Agent host registry](agent-placement.md#agent-host-registry) and [Slot reservations](agent-placement.md#slot-reservations).

Recovery after a whole-machine restart remains [open](agent-open-questions.md#agent-host), separately from recovery after a worker restart. Handling a host that never returns also remains [open](agent-open-questions.md#agent-execution). Neither lost contact nor an expired lease authorizes launching the same generation on another host.

[Session suspension](agent-session.md#session-lifecycle) saves a checkpoint at an idle turn boundary. It stops the VM and frees its slot. [Resume](agent-session.md#suspend-and-resume) restores guest paths and reacquires credentials outside the guest. It rechecks channels and port publications. Which process state survives and how resume or migration maps to execution generations remain [open](agent-open-questions.md#agents).
