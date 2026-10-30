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

A dedicated non-root user runs the worker, and container managers run rootless. See [`OQ-SBX-8`](agent-open-questions.md#%60OQ-SBX-8%60). Privileged setup remains [`OQ-SBX-10`](agent-open-questions.md#%60OQ-SBX-10%60), while workload separation remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

This contract applies to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`.

## Registration and control

At startup, the worker registers its workload host and task queue with the control plane. It reports platform metadata and capacity bounds, then renews its registration lease with capacity observations and the agent launch IDs holding slots. Polling uses its own host task queue. See [Workload host registry](workload-placement.md#Workload%20host%20registry).

Control connections originate at the worker for registration, commands, and observations. The control plane never opens a connection to a host. Worker control traffic uses a separate connection from workload data paths, and the worker doesn't forward workload network traffic, including traffic between workloads on the same machine. It forwards supervisor output over the control path described below.

Desired state expresses what afakt requests. Observed state records what the workload provider can establish about execution. Requests don't prove completion. A protocol for registration, desired and observed state, retries, and recovery after either endpoint restarts remains [`OQ-WLD-7`](agent-open-questions.md#%60OQ-WLD-7%60). Lifecycle ownership remains [`OQ-SBX-71`](agent-open-questions.md#%60OQ-SBX-71%60).

## Workload process output

Through the workload's control link, its supervisor sends captured `stdout` and `stderr`, or interactive terminal output, to the host worker. Each chunk includes a stream label, sequence number, and timestamp. A bounded buffer in the workload sandbox stores the output before the worker forwards it to the control-plane log store.

Clients read output from the control plane with a cursor, using `afakt workload logs --follow`. Output remains available after workload destruction. This path preserves the rule that the control plane never opens a connection to a workload host.

Log storage, retention, and control-link transport remain [`OQ-WLD-11`](agent-open-questions.md#%60OQ-WLD-11%60). Under the proposed transport, VM providers use the guest link, while container and host providers use a local socket. Optional `stdin` and signals also need a transport choice under that question.

## Capacity

The local capacity controller observes CPU use, memory use, and free workspace disk space. It adjusts launch-slot capacity within configured bounds, and the worker reports that capacity at lease renewal. Lowering capacity preserves existing reservations and running agent launches. Lease expiry stops new placement but doesn't prove that workloads stopped. See [Slot capacity](workload-placement.md#Slot%20capacity) and [Slot reservations](workload-placement.md#Slot%20reservations).

A stop request alone can't release capacity. How accounting for workloads and retained resources relates to launch-owned slots, including how to avoid counting the same allocation twice, remains [`OQ-WLD-4`](agent-open-questions.md#%60OQ-WLD-4%60).

## Restart and reclaim

After restarting, the worker uses persisted records to identify and reconcile execution before reporting its state. Loss of the worker connection doesn't permit a duplicate launch on another workload host. Whole-machine recovery remains [`OQ-HST-1`](agent-open-questions.md#%60OQ-HST-1%60), separately from worker replacement. See [Failure and reclaim](workload-lifecycle.md#Failure%20and%20reclaim).

Workloads can survive worker loss. Reclaim requires persistent execution identities and authority over the surviving processes and VMs. A different workload account doesn't by itself prevent access to worker control endpoints. Same-account host execution can reach worker state and endpoints. Control socket paths and account separation alone don't establish caller authentication.

Authentication of control endpoints, persistent execution identity, reclaim across workload accounts, and recovery after supervisor loss remain [`OQ-WLD-7`](agent-open-questions.md#%60OQ-WLD-7%60) for each Tier 1 cell. Custom Seatbelt profiles aren't a supported isolation mechanism. The replacement remains under the held [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).
