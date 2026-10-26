---
type: Design
title: Agent placement
description: How the agent host registry and the reservation API bind each agent execution to one agent host with a free execution slot before launch.
tags: [afakt, agents, execution, placement]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent placement

Placement binds a durable [agent's](agent.md) execution to one agent host before launch. Each execution is one harness run. How migration fits generation host binding remains [open](agent-open-questions.md#agents). The API server keeps a registry of agent hosts and serves the reservation API, which picks a host that meets the specification's placement requirements and reserves an execution slot on it. The agent runner workflow then starts the agent executor workflow on that host's task queue. [Agent execution](agent-execution.md#placement-on-agent-hosts) gives the place of placement in the lifecycle.

## Agent host registry

The API server keeps a registry of the agent hosts that can run executions.

- Each afakt worker on an agent host registers with the API server when it starts. The registration gives the agent host with its host task queue, along with the host's lower and upper slot capacity bounds, current slot capacity, and host metadata. The worker polls only its own host task queue.
- Host metadata describes what the agent host offers, such as its operating system and architecture, each installed harness with its version, the sandbox runtime and labels, and the credential references it can resolve. Placement requirements match against these values.
- A registration holds a lease that the worker renews. Each renewal reports the worker's current host metadata, slot capacity, and reason for the most recent capacity change. It also reports the execution IDs and generations that hold its slots. The registry records the lower and upper capacity bounds, current capacity, and change reason.
- When a lease runs out, the registry stops placing executions on that agent host until its worker registers again. The host's reservations and executions don't change.

## Slot capacity

The agent host's configuration sets lower and upper bounds on the number of execution slots, plus resource targets for CPU use, memory use, and free disk space on the volume that holds workspaces. CPU and memory targets are upper bounds on host use, and the disk target is a lower bound on free space.

A host with no resource targets keeps a fixed slot capacity equal to its upper bound.

A capacity controller in the afakt worker measures host resource use at intervals. It keeps the slot capacity between the configured lower and upper bounds. When every configured target permits more work, it raises capacity one slot at a time. When CPU or memory use exceeds its target, or free workspace disk space is below its target, it lowers capacity at once, without waiting for a sequence of one-slot decreases. At the lower bound, capacity stays there even when host resource use violates a target. For a decrease, the change reason identifies the resource target that caused it: `cpu-use`, `memory-use`, or `free-disk-space`. For an increase, the reason `resource-headroom` indicates that every configured target permits it. The worker reports capacity and the most recent change reason in each lease renewal.

A capacity decrease doesn't stop or preempt a running execution or cancel an existing reservation. Reserved and claimed slots may together exceed the lowered capacity. Until their total falls below capacity, the reservation API doesn't reserve a new slot on that host. Launch can still claim an existing reservation because claiming it doesn't add a slot.

## Reservation API

The API server's reservation API matches placement requirements against the registry and reserves execution slots. Placement matches against free slots only, using the most recently reported slot capacity. A free slot exists only when the total number of reserved and claimed slots is below that capacity.

- To place an execution, the agent runner workflow runs a placement activity on its own task queue, so any afakt worker that polls that queue can run it. The activity sends the execution ID, the generation, and the placement requirements to the reservation API.
- In one step, the reservation API picks a registered agent host that meets the requirements and has a free slot, then records a reservation on it. It returns that agent host and its host task queue, which the workflow history records.
- Every generation of one execution ID goes to the agent host of its first placement, and the reservation API reserves each higher generation's slot on that host. The previous generation's agent host can then supersede it.
- A specification with an `execution` source goes to the agent host that holds that earlier workspace. All `execution` sources in one specification come from executions on the same agent host.
- When registered agent hosts meet the requirements and none has a free slot, the placement activity waits and asks again. When no registered agent host meets them, placement fails, and the agent runner workflow fails with unsatisfiable placement as its reason, without starting an agent executor workflow.

## Image locality

Workers report cached layer digests as advisory host metadata for [workspace images](agent-workspace-image.md#placement-and-preheating). Placement can use known layer digests from a request as a preference. The provider resolves tags during workspace construction. Until that resolution, a tag alone can't supply fixed digests for locality, and placement leaves resolution to the provider.

Locality applies only to hosts with free slots and live registration leases that meet all placement requirements. For higher generations and `execution` sources, the reservation API preserves the host binding. Every preference follows the slot capacity and reservation rules. An execution keeps its chosen host. The general choice among matching hosts remains open.

A stale cache report is an ordinary cache miss on the chosen host. After publication, the control plane can preheat an image on chosen hosts through their task queues without reserving a future execution slot. Workers evict least recently used layer blobs and unpacked layers under the free-disk resource target, yielding preheat work to that target and leaving the capacity controller's behavior unchanged.

## Slot reservations

- A reservation belongs to one execution ID and generation. A repeated request for them returns the existing reservation and never reserves a second slot.
- Launch claims the reservation through the reservation API before it records the execution. A launch that doesn't find a reservation, such as after the reservation expired, requests a slot on its own agent host. Without a free slot, the execution waits in `pending` until one frees.
- If no launch claims a reservation within a bounded time, the reservation expires and the slot returns to the pool.
- When an execution enters a terminal state, the provider frees its slot through the reservation API. The reservation API also frees each slot that a lease renewal no longer reports, so a lost request to free a slot doesn't keep it in use.
- The reservation API counts reserved and claimed slots together. It reserves a new slot only when doing so keeps their total at or below the host's most recently reported capacity, even when concurrent requests arrive. A lower capacity leaves existing reservations and claimed slots intact.
