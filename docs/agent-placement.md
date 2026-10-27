---
type: Design
title: Agent placement
description: How the agent host registry and the reservation API bind each agent launch to one agent host with a free agent launch slot before launch.
tags: [afakt, agents, agent-launch, placement]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent placement

Placement binds an agent launch to one agent host. The Application Programming Interface (API) server keeps the agent host registry and serves the reservation API. The agent dispatcher workflow selects a host task queue and reserves an agent launch slot. After finalizing the agent launch spec, it queues the agent launcher workflow there. Migration creates a new agent launch linked to its predecessor. Transfer and source availability remain [open](agent-open-questions.md#%60OQ-AGT-12%60).

[Placement and preheating](agent-workspace-lifecycle.md#Placement%20and%20preheating).

## Agent host registry

The API server keeps a registry of the agent hosts that can run agent launches.

- At startup, the worker on each agent host registers with the API server when it starts. The registration gives the agent host with its host task queue, along with the host's lower and upper slot capacity bounds, current slot capacity, and host metadata. The worker polls only its own host task queue.
- Host metadata describes what the agent host offers, such as its operating system and architecture, each installed agent harness with its version, the agent sandbox runtime and labels, and the credential references it can resolve. Placement requirements match against these values.
- A registration holds a lease that the worker renews. Each renewal reports the worker's current host metadata, slot capacity, and reason for the most recent capacity change. It also reports the agent launch IDs that hold its slots. The registry records the lower and upper capacity bounds, current capacity, and change reason.
- When a lease runs out, the registry stops placing agent launches on that agent host until its worker registers again. The host's reservations and agent launches don't change.

## Slot capacity

The agent host's configuration sets lower and upper bounds on the number of agent launch slots, plus resource targets for central processing unit (CPU) use, memory use, and free disk space on the volume that holds agent workspaces. CPU and memory targets are upper bounds on host use, and the disk target is a lower bound on free space.

A host with no resource targets keeps a fixed slot capacity equal to its upper bound.

A capacity controller in the agent host worker measures host resource use at intervals. It keeps the slot capacity between the configured lower and upper bounds. When every configured target permits more work, it raises capacity one slot at a time. When CPU or memory use exceeds its target, or free workspace disk space is below its target, it lowers capacity at once, without waiting for a sequence of one-slot decreases. At the lower bound, capacity stays there even when host resource use violates a target. For a decrease, the change reason identifies the resource target that caused it: `cpu-use`, `memory-use`, or `free-disk-space`. For an increase, the reason `resource-headroom` indicates that every configured target permits it. The worker reports capacity and the most recent change reason in each lease renewal.

A capacity decrease doesn't stop or preempt a running agent launch or cancel an existing reservation. Reserved and claimed slots may together exceed the lowered capacity. Until their total falls below capacity, the reservation API doesn't reserve a new slot on that host. Provision can still claim an existing reservation because claiming it doesn't add a slot.

## Reservation API

The API server's reservation API matches placement requirements against the registry and reserves agent launch slots. Placement matches against free slots only, using the most recently reported slot capacity. A free slot exists only when the total number of reserved and claimed slots is below that capacity.

- To place an agent launch, the agent dispatcher workflow runs a placement activity on its own task queue, so any afakt worker that polls that queue can run it. The activity sends the agent launch ID and the placement requirements to the reservation API.
- In one step, the reservation API picks a registered agent host that meets the requirements and has a free slot, then records a reservation on it. It returns that agent host and its host task queue, which the workflow history records.
- Each agent launch keeps the agent host chosen for it. A replacement has a new ID and predecessor link. Migration may select another host after the predecessor stops. Cross-host workspace transfer remains [open](agent-open-questions.md#%60OQ-AGT-12%60).
- A specification with an `agent-launch` source goes to the agent host that holds that earlier agent workspace. All `agent-launch` sources in one specification come from agent launches on the same agent host.
- When registered agent hosts meet the requirements and none has a free slot, the placement activity waits and asks again. When no registered agent host meets them, placement fails, and the agent dispatcher workflow fails with unsatisfiable placement as its reason, without starting an agent launcher workflow.

## Image locality

Workers report cached layer digests as advisory host metadata for [workspace images](agent-workspace-image.md#Placement%20and%20preheating). Placement can use known layer digests from a request as a preference. The provider resolves tags during workspace construction. Until that resolution, a tag alone can't supply fixed digests for locality, and placement leaves resolution to the provider.

Locality applies only to hosts with free slots and live registration leases that meet all placement requirements. The reservation API preserves each agent launch's host binding. An `agent-launch` source requires the host holding that agent workspace. Every preference follows the slot capacity and reservation rules. An agent launch keeps its chosen host. The general choice among matching hosts remains open.

A stale cache report is an ordinary cache miss on the chosen host. After publication, the control plane can preheat an image on chosen hosts through their task queues without reserving a slot for a future agent launch. Workers evict least recently used layer blobs and unpacked layers under the free-disk resource target, yielding preheat work to that target and leaving the capacity controller's behavior unchanged.

## Slot reservations

- A reservation belongs to one agent launch ID. A repeated request for that ID returns the existing reservation and never reserves a second slot.
- Provision claims the reservation through the reservation API. The order of reservation claiming and recording the pending agent launch remains [`OQ-PLC-11`](agent-open-questions.md#%60OQ-PLC-11%60). A provision request that doesn't find a reservation, such as after the reservation expired, requests a slot on its own agent host. Without a free slot, the agent launch waits in `pending` until one frees.
- If no provision request claims a reservation within a bounded time, the reservation expires and the slot returns to the pool.
- When an agent launch enters a terminal state, the provider frees its slot through the reservation API. The reservation API also frees each slot that a lease renewal no longer reports, so a lost request to free a slot doesn't keep it in use.
- The reservation API counts reserved and claimed slots together. It reserves a new slot only when doing so keeps their total at or below the host's most recently reported capacity, even when concurrent requests arrive. A lower capacity leaves existing reservations and claimed slots intact.
