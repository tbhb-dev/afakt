---
type: Design
title: Workload host worker resource
description: Desired host-worker configuration, registration observations, and the open mapping from host records to worker identity.
tags: [afakt, workload, resources]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:26:00Z" }
---

# Workload host worker resource

`WorkloadHostWorker` represents the [worker daemon](workload-host-worker.md) on a [workload host](workload-host.md). Mapping the machine's identity to its worker instance, and dividing their records, remain [`OQ-RES-6`](agent-open-questions.md#%60OQ-RES-6%60).

The resource applies to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`.

## Envelope

| Key | Meaning | Who sets it |
| --- | --- | --- |
| `apiVersion` | `afakt/v1alpha1`. | Caller |
| `kind` | `WorkloadHostWorker`. | Caller |
| `metadata` | Worker identity. Binding to a machine and worker instance remains open. | Control plane |
| `spec` | Desired worker configuration. | Authorized operator |
| `status` | Registration and execution observations. | Worker, through its outbound connection |

Worker identifiers and version formats remain [`OQ-MOD-7`](agent-open-questions.md#%60OQ-MOD-7%60). The division between machine and worker information remains [`OQ-RES-6`](agent-open-questions.md#%60OQ-RES-6%60).

## Specification keys

| Proposed key | Meaning |
| --- | --- |
| `task_queue` | Queue the worker polls for placed agent launches. |
| `lower_slot_capacity` | Configured lower bound for launch slots. |
| `upper_slot_capacity` | Configured upper bound for launch slots. |
| `resource_targets` | CPU and memory use targets and the free-disk target. |

See [Workload host registry](workload-placement.md#Workload%20host%20registry) and [Slot capacity](workload-placement.md#Slot%20capacity). The host isolation baseline remains held under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60). Desired-state commands, revisions, and acknowledgement keys remain [`OQ-WLD-7`](agent-open-questions.md#%60OQ-WLD-7%60).

## Status keys

| Proposed key | Meaning |
| --- | --- |
| `operating_system`, `architecture` | Advertised workload host platform. |
| `harnesses`, `sandbox_runtime` | Installed execution capabilities. |
| `credential_references` | References the worker can resolve, without secret values. |
| `slot_capacity`, `capacity_change_reason` | Observed launch-slot capacity and its change reason. |
| `agent_launch_ids` | Agent launches holding slots. |
| `lease_expires_at` | Registration lease expiry, with duration under `OQ-PLC-1`. |
| `cached_layer_digests` | Advisory image locality. |

Workload observations and their freshness, including app and user code worker allocations, need the contracts in [`OQ-RES-6`](agent-open-questions.md#%60OQ-RES-6%60) and [`OQ-WLD-4`](agent-open-questions.md#%60OQ-WLD-4%60). Even after a lease expires, execution can continue, so capacity release depends on observations under [Restart and reclaim](workload-host-worker.md#Restart%20and%20reclaim).

## Example

A worker registers a `darwin/arm64` workload host and reports available launch slots. Its resource-to-machine binding remains [`OQ-RES-6`](agent-open-questions.md#%60OQ-RES-6%60).

```yaml
apiVersion: afakt/v1alpha1
kind: WorkloadHostWorker
metadata:
  name: example-worker
spec:
  task_queue: example-host
  lower_slot_capacity: 0
  upper_slot_capacity: 4
status:
  operating_system: darwin
  architecture: arm64
  slot_capacity: 2
  agent_launch_ids: []
```
