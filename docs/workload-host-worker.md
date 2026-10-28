---
type: Design
title: Workload host worker
description: Host registration, execution observations, capacity reporting, and recovery through an outbound control connection.
tags: [afakt, workload, hosts]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:26:00Z" }
---

# Workload host worker

The workload host worker is the daemon on a [workload host](workload-host.md). It connects the machine to the control plane and runs placed agent launcher workflows. This host service is distinct from a worker-type workload that runs user code. The [host-worker resource](workload-host-worker-resource.md) separates desired configuration from observations, with the precise resource boundary still open.

## Runtime authority and targets

A dedicated non-root user runs the worker, and container managers run rootless under the [worker user decision](https://github.com/tbhb-dev/afakt-drafts/blob/744ed448da1a77085e3a83cfb81550466718e5c8/workloads/workload-readiness-checklist.md#Host%20worker), `OQ-SBX-8`. Privileged setup remains [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), while workload separation remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

This contract applies to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`. The [platform matrix](https://github.com/tbhb-dev/afakt-drafts/blob/744ed448da1a77085e3a83cfb81550466718e5c8/workloads/workload-platform-support.md) defines those Tier 1 cells. Start with the targets in this [build order](https://github.com/tbhb-dev/afakt-drafts/blob/744ed448da1a77085e3a83cfb81550466718e5c8/workloads/workload-readiness-checklist.md#Scope%20of%20the%20first%20build): `macos-vm`, `macos-host`, `linux-vm`, then `windows-vm`.

## Registration and control

At startup, the worker registers its workload host and task queue with the control plane. It reports platform metadata and capacity bounds, then renews its registration lease with capacity observations and the agent launch IDs holding slots. Polling uses its own host task queue. See [Workload host registry](workload-placement.md#Workload%20host%20registry).

Control connections originate at the worker for registration, commands, and observations. The control plane never opens a connection to a host. Worker control traffic uses a separate connection from workload data paths, and the worker forwards no workload traffic, including traffic between workloads on the same machine.

Desired state expresses what afakt requests. Observed state records what the workload provider can establish about execution. Requests don't prove completion. A protocol for registration, desired and observed state, retries, and recovery after either endpoint restarts remains [`OQ-WLD-7`](agent-open-questions.md#%60OQ-WLD-7%60), following [workload design question 23](https://github.com/tbhb-dev/afakt-drafts/blob/744ed448da1a77085e3a83cfb81550466718e5c8/workloads/workload-design.md#Open%20questions). Lifecycle ownership remains [`OQ-SBX-71`](agent-open-questions.md#%60OQ-SBX-71%60).

## Capacity

The local capacity controller observes CPU use, memory use, and free workspace disk space. It adjusts launch-slot capacity within configured bounds, and the worker reports that capacity at lease renewal. Lowering capacity preserves existing reservations and running agent launches. Lease expiry stops new placement but doesn't prove that workloads stopped. See [Slot capacity](workload-placement.md#Slot%20capacity) and [Slot reservations](workload-placement.md#Slot%20reservations).

The [capacity proposal](https://github.com/tbhb-dev/afakt-drafts/blob/744ed448da1a77085e3a83cfb81550466718e5c8/workloads/capacity-budget-design.md) extends accounting to workloads and retained resources. Its relationship to launch-owned slots remains [`OQ-WLD-4`](agent-open-questions.md#%60OQ-WLD-4%60). A worker can't report capacity as released solely because it sent a stop request.

## Restart and reclaim

After restarting, the worker uses persisted records to identify and reconcile execution before reporting its state. Loss of the worker connection doesn't permit a duplicate launch on another workload host. Whole-machine recovery remains [`OQ-HST-1`](agent-open-questions.md#%60OQ-HST-1%60), separately from worker replacement. See [Failure and reclaim](workload-lifecycle.md#Failure%20and%20reclaim).

| Evidence | Result and limit |
| --- | --- |
| [S-188 Linux](https://github.com/tbhb-dev/afakt-spikes/blob/23716546c0a7ff73d2f630a53cdd1c8bf55fb785/s-188-host-worker-isolation/linux/REPORT.md) | The worker-owned supplement reclaims surviving host processes using recorded process identities. A host workload can connect to a control socket by its path in the tested confinement. Second-account behavior remains unqualified. |
| [S-188 Windows run 2](https://github.com/tbhb-dev/afakt-spikes/blob/6bbf817446c5a941cfbf7904a32c199ef9b2ecff/s-188-host-worker-isolation/windows/run2/REPORT.md) | Workloads survive worker loss. The restarted worker reclaims same-account processes and recorded VMs, but can't reclaim the second-account host workload. A separate account can still connect to the worker over TCP. |
| [S-188 macOS](https://github.com/tbhb-dev/afakt-spikes/blob/14d4d7aa84645e6e1a94634ae3ccbc79ed6f53c4/s-188-host-worker-isolation/macos/REPORT.md) | Provisional results establish reclaim of surviving VM supervisors and same-account host processes. Same-account host execution reaches worker state and endpoints. Second-account behavior and recovery after supervisor loss remain unqualified. |

These results don't establish a recovery contract for every Tier 1 cell. Authentication of control endpoints, persistent execution identity, and reclaim authority remain [`OQ-WLD-7`](agent-open-questions.md#%60OQ-WLD-7%60). The macOS report excludes unsupported custom Seatbelt profiles, while the host mediation proposal assumes them. That conflict stays under the held [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60), without selecting a replacement isolation mechanism.
