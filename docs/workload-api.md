---
type: Design
title: Workload API
description: REST resource operations for workloads, templates, images, and host workers, with errors and pagination.
tags: [afakt, workload, api]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:55:04Z" }
---

# Workload API

The [API server](components.md#API%20server) exposes workload management through Representational State Transfer (REST) resources. Clients authenticate through the ingress gateway. Workload and host callers use [SPIFFE identity](workload-identity.md). Authentication establishes the caller, while authorization determines its permitted operations. User authentication and workload scopes remain open under [`OQ-WLD-9`](agent-open-questions.md#%60OQ-WLD-9%60) and [`OQ-WLD-8`](agent-open-questions.md#%60OQ-WLD-8%60).

The same resource boundaries apply to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`.

## Resources and authority

Resource representations have `apiVersion`, `kind`, `metadata`, `spec`, and `status`. Desired configuration and observed state have different writers. A request doesn't prove that its execution has completed.

| Resource | Operations and contract |
| --- | --- |
| [Workload](workload-resource.md) | Inspect identity, immutable creation fields, lifecycle state, and conditions. The proposed deletion operation requests teardown. Reclaim depends on execution observations. Lifecycle ownership remains `OQ-SBX-71`. |
| [WorkloadTemplate](workload-template-resource.md) | Create and inspect creation settings, list templates, and request instantiation. Version publication, updates, retirement, and selection remain `OQ-WLD-5`. Each workload records the version it uses. |
| [WorkloadImage](workload-image-resource.md) | Select the container image for the workload process through this resource. Registration, listing, lookup, admission, and retirement operations remain `OQ-WLD-6` and `OQ-WLD-9`. |
| [WorkloadHostWorker](workload-host-worker-resource.md) | Register, renew, report observations, and read desired state over worker-initiated connections. A separate worker resource contract remains `OQ-RES-6` and `OQ-WLD-7`. |

Workload creation fixes type, Identity and Access Management (IAM) role, workload image, execution mode, shape, host, namespace, and workload template version. An `empty` workload doesn't select a workload image. Guest images come from afakt, and the workload records their component digests separately. Changing the creation-time fields requires another workload. Wire fields and request projections remain [`OQ-RES-4`](agent-open-questions.md#%60OQ-RES-4%60) and [`OQ-RES-5`](agent-open-questions.md#%60OQ-RES-5%60). A complete resource representation doesn't establish which fields a caller may write. Authorization of role selection remains part of [`OQ-WLD-9`](agent-open-questions.md#%60OQ-WLD-9%60), separately from the workload's own pinned authority.

## Proposed routes

The proposed HTTP methods, paths, payloads, and response mapping remain [`OQ-WLD-9`](agent-open-questions.md#%60OQ-WLD-9%60).

| Method | Proposed path | Operation |
| --- | --- | --- |
| `GET` | `/api/v1/workloads/{workload_id}` | Inspect a workload. |
| `DELETE` | `/api/v1/workloads/{workload_id}` | Request workload teardown. |
| `GET`, `POST` | `/api/v1/workload-templates` | List or create workload templates. |
| `GET` | `/api/v1/workload-templates/{template_id}` | Inspect a workload template. |
| `POST` | `/api/v1/workload-templates/{template_id}/instantiate` | Request instantiation from a workload template. |
| `POST` | `/api/v1/workloads/prepare` | Request preparation. Its relationship to workload creation remains `OQ-WLD-1`. |
| `POST` | `/api/v1/workload-claims` | Submit a claim. Its relationship to preparation and creation remains `OQ-WLD-1`. |
| `GET` | `/api/v1/workload-claims/{claim_id}` | Inspect a claim. |
| `GET`, `POST` | `/api/v1/agent-hosts` | List or register hosts under the proposed combined host record. |
| `GET` | `/api/v1/agent-hosts/{agent_host_id}` | Inspect that host record. |
| `POST` | `/api/v1/agent-hosts/{agent_host_id}/renew` | Renew registration with capacity observations. |

Host registration needs reconciliation with WorkloadHostWorker under [`OQ-RES-6`](agent-open-questions.md#%60OQ-RES-6%60). Workload listing, workload image operations, and a separate worker surface have no route contract. Template version routes, idempotency, concurrent changes, and operation completion also remain [`OQ-WLD-9`](agent-open-questions.md#%60OQ-WLD-9%60).

Preparation and claims remain under [`OQ-WLD-1`](agent-open-questions.md#%60OQ-WLD-1%60). Claims are one-shot. After one agent launch, afakt destroys the workload instead of making it reusable through the API. See [Workload lifecycle](workload-lifecycle.md).

## Commands in a workload

`afakt workload exec` and `afakt workload shell` are workload provider capabilities. afakt refuses a command when its provider declares that capability unsupported.

Exec takes an argument list and doesn't assume a shell. Shell runs the shell declared by the workload image and refuses to run when the image doesn't declare one. The supervisor starts each command inside a fresh `afakt-sandbox` by default. Permission to run without that sandbox remains [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

### Proposed command transport

The proposed request path passes through the control plane for an IAM check. Command streams would cross the overlay through the workload proxy, with a mediator check. An audit record would include the caller, command, and exit status, excluding output. Placement, launch binding, inherited environment, and capture ordering remain [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60).

### Proposed exec guarantees

- **Shared files.** Commands in one workload share its working tree. Each sandbox permits writes there and denies writes elsewhere.
- **Timeout and cancellation.** Each command has a time limit and supports cancellation. Either event sends `SIGTERM`, followed by `SIGKILL` after a grace period, with the cause in the result.
- **Output and result.** Capture has a limit per exec and reports truncation. An exec always returns a result with its exit status, or with cancellation, timeout, or loss after the supervisor or workload dies.
- **Operation IDs.** A caller can supply an operation ID. The supervisor retains it for the workload's lifetime. A repeat attaches to the running command or returns its recorded result without starting it again. Workload destruction loses that result. The caller then decides what happens next.

Mapping each script step of a workflow job to an exec in one shared `empty` workload remains proposed. A workflow job is distinct from the `job` workload type, which runs one process to completion.

## Worker control

Every control connection originates at the [workload host worker](workload-host-worker.md). The control plane never opens a connection to a host. Requests use the worker's control path without an inbound listener on a workload host.

When registering or renewing its lease, the worker reports availability and capacity. Expiry stops new placement without proving that execution stopped. Desired-state commands, observations, revisions, acknowledgements, and recovery remain [`OQ-WLD-7`](agent-open-questions.md#%60OQ-WLD-7%60).

## Errors

HTTP failures use [RFC 9457 problem details](https://www.rfc-editor.org/rfc/rfc9457.html) with media type `application/problem+json`. Its members distinguish the problem type (`type`), a short description (`title`), the HTTP status (`status`), an explanation of this occurrence (`detail`), and an occurrence reference (`instance`). When supplied, `status` matches the HTTP response. Details exclude secrets.

Problem types, extension fields, mappings from domain failures to HTTP status, and retry behavior remain [`OQ-WLD-10`](agent-open-questions.md#%60OQ-WLD-10%60). A `code` extension remains open under that question.

## Pagination by key

List operations use cursors based on keys. A continuation cursor identifies a position in the same authorized collection with the same filters. The response fields remain open: [`OQ-WLD-10`](agent-open-questions.md#%60OQ-WLD-10%60) covers whether `items` holds the results and `next_cursor` supplies continuation, with the continuation field absent when the list ends.

Ordering, cursor encoding, expiry, invalidation, page limits, and behavior under concurrent changes remain [`OQ-WLD-10`](agent-open-questions.md#%60OQ-WLD-10%60). Continuation state in a cursor doesn't grant access to the collection.
