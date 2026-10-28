---
type: Threat Model
title: "Agent threat model for Docker Engine on Linux"
description: Isolation, local transport, retained state, and platform threats for the docker-engine agent sandbox provider.
tags: [afakt, agents, security, threat-model]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent threat model for Docker Engine on Linux

[Public key infrastructure](pki.md).

## Scope

The docker-engine agent sandbox provider uses the [provider design](agent-sandbox-docker-engine.md). The [common agent threat model](agent-threat-model.md) covers shared service authority, credentials, network policy, sessions, registries, and audit records.

[Platforms and privileges](agent-sandbox-docker-engine.md#Platforms%20and%20privileges) and [Provider isolation](agent-sandbox-docker-engine.md#Isolation%20boundary) define the platform and its kernel boundary. [Provider startup](agent-sandbox-docker-engine.md#Process%20placement%20and%20startup) places afakt-agent, afakt-sandbox, and the harness.

STRIDE covers spoofing, tampering, repudiation, information disclosure, denial of service, and elevation of privilege. A virtual machine (VM) supplies a separate guest kernel where the provider uses one. Agent Client Protocol (ACP) transports session messages. Domain Name System (DNS) queries use the agent mediator. A User Identifier (UID) names an operating-system user. An Application Programming Interface (API) provides shared service authority.

### Agent mediator access

The engine doesn't select a mandatory mediator path. Internal networks, host services, and runtime management expose distinct authority. The carrier and its isolation remain unresolved. See [Provider mediation](agent-sandbox-docker-engine.md#Mediator%20connection%20and%20name%20resolution).

### Workspace, scratch, and caches

[Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches) distinguishes launch state, shared named tool caches, and verified image caches. Provider storage mappings, scratch cleanup, shared cache trust, and hostile host-side repository processing remain unresolved. [Provider materialization](agent-sandbox-docker-engine.md#Workspace%20image%20materialization) separates verified extraction from later workspace capture.

### Agent snapshots and agent checkpoints

[Provider capture](agent-sandbox-docker-engine.md#Agent%20snapshots%20and%20agent%20checkpoints) defines the capture boundary and unresolved restore contract. This provider has no agent snapshot. A native harness checkpoint doesn't capture its shared kernel or engine.

### Host visibility and workspace commands

Host custody and its limits follow [Provider observations](agent-sandbox-docker-engine.md#Host%20observations). Termination authority and unresolved descendant cleanup follow [Provider teardown](agent-sandbox-docker-engine.md#Cancellation%20and%20teardown). Same-user process separation and the host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

Commands started by `afakt agent exec` use standalone afakt-sandbox. [Provider startup](agent-sandbox-docker-engine.md#Process%20placement%20and%20startup) leaves command placement, inherited state, and launch authority under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). An unsandboxed invocation remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

## Trust boundaries

| Boundary ID | Boundary | Two sides | Enforcement in the design |
| --- | --- | --- | --- |
| `TB-DE-1` | Agent harness and provider control | untrusted process tree / provider lifecycle and platform authority | [Enforcement](agent-sandbox.md#Enforcement) defines the isolation tier. [Host visibility and workspace commands](#Host%20visibility%20and%20workspace%20commands) identifies platform exposure and missing confinement. |
| `TB-DE-2` | Local transport and agent mediator | Launch network and control endpoints / mediator identity and policy | [Agent mediator](agent-workload-mediator.md#Launch%20boundary) Local endpoint binding remains [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). |
| `TB-DE-3` | Agent workspace and host processing | Launch-controlled files / worker capture, materialization, and cleanup | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Host-side handling of repository metadata remains [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). |
| `TB-DE-4` | Shared named tool cache | One launch / cache content consumed by another launch | [Sandbox variables](agent-sandbox.md#Sandbox%20variables) Access, provenance, and capacity controls remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). |
| `TB-DE-5` | Captured state and successor launch | Current launch / retained capture and later launch | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume) State authorization and ownership remain [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). |
| `TB-DE-6` | Workspace command and launch authority | Authorized request for a command / command confinement and workspace access | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) Command placement and lifecycle binding remain [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). |

## Data flow diagram

### Element inventory

| Element ID | Element | Location or custody |
| --- | --- | --- |
| `E-DE-1` | Command caller | Authenticated client. Shared service grants apply |
| `E-DE-2` | Another agent launch | Separate launch authority |
| `P-DE-1` | Agent harness and afakt-sandbox | Container |
| `P-DE-2` | Agent host worker and provider lifecycle | Agent host |
| `P-DE-3` | Host workspace processor | Worker capture and cleanup |
| `P-DE-8` | Agent proxy | Outside guest confinement. Launch identity custody |
| `P-DE-4` | Agent mediator | Outside launch confinement. Process separation unresolved |
| `P-DE-5` | Local control endpoint | Carrier and placement unspecified |
| `P-DE-6` | Successor launch | New launch authority |
| `P-DE-7` | afakt agent exec command | Execution placement unspecified |
| `DS-DE-1` | Agent workspace | Per-launch writable content |
| `DS-DE-2` | Scratch storage | Outside workspace. Backing unspecified |
| `DS-DE-3` | Shared named tool cache | Shared across launches. Backing unspecified |
| `DS-DE-4` | Verified image cache | Agent host. Distinct from named tool cache |
| `DS-DE-5` | Checkpoint capture | Retained state. Access controls unresolved |

### Level one provider flows

Arrows identify exchanges and their return paths. Unspecified carriers and placements remain gaps. The shared named tool cache is outside individual launch confinement and has a separate crossing from another launch. The verified image cache has a different trust relationship.

```mermaid
flowchart LR
    P-DE-8["P-DE-8: Agent proxy"]
    subgraph z0["TB-DE-1 / launch confinement"]
        P-DE-1["P-DE-1: Agent harness and afakt-sandbox"]
    end
    subgraph z1["TB-DE-2 / local control"]
        P-DE-5["P-DE-5: Local control endpoint"]
    end
    subgraph z2["TB-DE-2 / agent host policy"]
        P-DE-4["P-DE-4: Agent mediator"]
    end
    subgraph z3["TB-DE-3 / launch-controlled storage"]
        DS-DE-1["DS-DE-1: Agent workspace"]
        DS-DE-2["DS-DE-2: Scratch storage"]
    end
    subgraph z4["TB-DE-3 / agent host processing"]
        P-DE-2["P-DE-2: Agent host worker and provider lifecycle"]
        P-DE-3["P-DE-3: Host workspace processor"]
        DS-DE-4["DS-DE-4: Verified image cache"]
    end
    subgraph z5["TB-DE-4 / shared across launches"]
        DS-DE-3["DS-DE-3: Shared named tool cache"]
    end
    subgraph z6["TB-DE-4 / other launch"]
        E-DE-2["E-DE-2: Another agent launch"]
    end
    subgraph z7["TB-DE-5 / retained capture"]
        DS-DE-5["DS-DE-5: Checkpoint capture"]
    end
    subgraph z8["TB-DE-5 / successor"]
        P-DE-6["P-DE-6: Successor launch"]
    end
    subgraph z9["TB-DE-6 / caller"]
        E-DE-1["E-DE-1: Command caller"]
    end
    subgraph z10["TB-DE-6 / command placement unresolved"]
        P-DE-7["P-DE-7: afakt agent exec command"]
    end
    P-DE-2 -->|"F-DE-1"| P-DE-1
    P-DE-1 -->|"F-DE-2"| P-DE-5
    P-DE-5 -->|"F-DE-3"| P-DE-8
    P-DE-1 -->|"F-DE-4"| P-DE-8
    P-DE-1 -->|"F-DE-5"| DS-DE-1
    P-DE-1 -->|"F-DE-6"| DS-DE-2
    P-DE-3 -->|"F-DE-7"| DS-DE-1
    DS-DE-4 -->|"F-DE-8"| P-DE-3
    P-DE-1 -->|"F-DE-9"| DS-DE-3
    E-DE-2 -->|"F-DE-10"| DS-DE-3
    P-DE-2 -->|"F-DE-11"| DS-DE-5
    DS-DE-5 -->|"F-DE-12"| P-DE-6
    E-DE-1 -->|"F-DE-13"| P-DE-7
    P-DE-7 -->|"F-DE-14"| DS-DE-1
    P-DE-7 -->|"F-DE-15"| P-DE-8
```

- **`F-DE-1`.** Startup, stop, and observed exit of the process tree. Crosses `TB-DE-1`.
- **`F-DE-2`.** Session and local control messages. Endpoint binding remains unresolved. Crosses `TB-DE-2`.
- **`F-DE-3`.** Control messages and preview streams over the provider transport. Carrier authentication remains unresolved. Crosses `TB-DE-2`.
- **`F-DE-4`.** Network packets and name queries through the workload proxy under mediator decisions. Network and control traffic needs a bound mediator path. Its provider-specific mechanism remains unspecified. Crosses `TB-DE-2`.
- **`F-DE-5`.** Workspace reads and writes under launch confinement. Crosses `TB-DE-3`.
- **`F-DE-6`.** Scratch reads and writes outside the workspace. Allocation and cleanup are unspecified. Crosses `TB-DE-3`.
- **`F-DE-7`.** Host-side capture, metadata interpretation, and cleanup of hostile workspace content. Crosses `TB-DE-3`.
- **`F-DE-8`.** Verified immutable image content used for materialization. Cache ownership is provider-specific. Crosses `TB-DE-3`.
- **`F-DE-9`.** Named cache reads and writes by this launch. Mapping to backing storage is unspecified. Crosses `TB-DE-4`.
- **`F-DE-10`.** Named cache reads and writes by another launch across the same sharing boundary. Crosses `TB-DE-4`.
- **`F-DE-11`.** Capture and restore of harness checkpoint and workspace state. The contract for resuming execution remains unresolved and doesn't include a VM snapshot. Crosses `TB-DE-5`.
- **`F-DE-12`.** State admitted into a successor launch under new authority. Crosses `TB-DE-5`.
- **`F-DE-13`.** Authorized workspace command request. Local placement and credential binding remain unresolved. Crosses `TB-DE-6`.
- **`F-DE-14`.** Command reads and writes in the selected agent workspace. Concurrent access remains unresolved. Crosses `TB-DE-6`, `TB-DE-3`.
- **`F-DE-15`.** Command traffic requiring launch egress policy. Transport and identity binding remain unresolved. Crosses `TB-DE-6`, `TB-DE-2`.

## Agent proxy

Each workload has one agent mediator and one agent proxy. The proxy transmits traffic in both directions and requests a mediator decision for every connection. The mediator never transmits traffic. Direct overlay peers authenticate with SPIFFE mutual Transport Layer Security (TLS), and both workload mediators check the connection. Service users and provider integration remain undecided under [Launch boundary](agent-workload-proxy.md#Launch%20boundary).

```mermaid
flowchart LR
    P-DE-4["P-DE-4: Agent mediator"]
    P-DE-8["P-DE-8: Agent proxy"]
    P-DE-4 -->|"F-DE-16"| P-DE-8
```

- **`F-DE-16`.** Connection decision requests, launch identity, and decisions across the mediator/agent-proxy boundary. The local channel needs an authenticated peer and workload context.

## STRIDE analysis

Each table covers the six categories for its listed flows and every participating endpoint. Shared service checks are in the [common STRIDE analysis](agent-threat-model.md#STRIDE%20analysis). They don't establish provider isolation. A mitigation applies only to the stated attack.

### Isolation and lifecycle

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-DE-1` | `F-DE-1`, `P-DE-2`, `P-DE-1` | spoofing | A local process substitutes for the provider control endpoint or launch process. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) [Runtime privileges](agent-sandbox.md#Runtime%20privileges) Per-launch identities and helper authority remain unresolved. [Provider startup](agent-sandbox-docker-engine.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)) |
| `TM-DE-2` | `F-DE-1`, `P-DE-2`, `P-DE-1` | tampering | A launch changes inherited handles, mounts, descriptors, or confinement inputs to reach provider authority. | [Enforcement](agent-sandbox.md#Enforcement). [Platform exposure](#Host%20visibility%20and%20workspace%20commands). Unspecified. [Provider isolation](agent-sandbox-docker-engine.md#Isolation%20boundary). | depends on an open question ([`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60)) |
| `TM-DE-21` | `F-DE-1`, `P-DE-2`, `P-DE-1` | repudiation | A compromised agent host invents process exits or hides local workspace and command effects. | [Provider observations](agent-sandbox-lifecycle.md#Provider%20observations) Independent verification of a compromised host is unspecified. [Provider observations](agent-sandbox-docker-engine.md#Host%20observations). | gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)) |
| `TM-DE-70` | `F-DE-1`, `P-DE-2`, `P-DE-1` | information disclosure | A same-user host process reads mediator memory, local session traffic, or another launch's files. | [Trust boundary](agent-host.md#Trust%20boundary) The design trusts host custody. Same-user process separation remains on hold. [Provider observations](agent-sandbox-docker-engine.md#Host%20observations). | depends on an open question ([`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60)) |
| `TM-DE-23` | `F-DE-1`, `P-DE-2`, `P-DE-1` | denial of service | A hostile process tree ignores cancellation and continues running. | [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) Bounds on cancellation don't specify how this provider kills and observes every descendant. [Provider teardown](agent-sandbox-docker-engine.md#Cancellation%20and%20teardown). | gap ([`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60)) |
| `TM-DE-6` | `F-DE-1`, `P-DE-2`, `P-DE-1` | elevation of privilege | A process in the container gains authority over the shared kernel or Docker management. | [Enforcement](agent-sandbox.md#Enforcement). [Platform exposure](#Host%20visibility%20and%20workspace%20commands). Unspecified. [Provider isolation](agent-sandbox-docker-engine.md#Isolation%20boundary). | gap ([`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60)) |
| `TM-DE-154` | `F-DE-1`, `P-DE-2`, `P-DE-1` | elevation of privilege | An orphaned process or reused launch user retains local authority after worker failure. | [Restart and recovery](agent-host.md#Restart%20and%20recovery) Durable ownership and safe identity reuse remain unresolved. [Provider teardown](agent-sandbox-docker-engine.md#Cancellation%20and%20teardown). | depends on an open question ([`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60)) |

### Local transport and egress

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-DE-25` | `F-DE-2`, `F-DE-3`, `F-DE-4`, `P-DE-1`, `P-DE-5`, `P-DE-4` | spoofing | A process selects another launch's mediator endpoint and receives its authority. | [Agent mediator](agent-workload-mediator.md#Launch%20boundary) Endpoint authentication remains unresolved. [Provider mediation](agent-sandbox-docker-engine.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-DE-26` | `F-DE-2`, `F-DE-3`, `F-DE-4`, `P-DE-1`, `P-DE-5`, `P-DE-4` | tampering | A process changes proxy settings or uses an interface that bypasses the agent mediator. | [Agent mediator](agent-workload-mediator.md#Launch%20boundary) Network and control traffic needs a bound mediator path. Its provider-specific mechanism remains unspecified. [Provider mediation](agent-sandbox-docker-engine.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)) |
| `TM-DE-27` | `F-DE-2`, `F-DE-3`, `F-DE-4`, `P-DE-1`, `P-DE-5`, `P-DE-4` | repudiation | Unbound local traffic acquires another launch's attribution before shared logging. | [Control requests](agent-workload-mediator.md#Control%20requests) Stamping identity requires a trustworthy local sender binding. [Provider mediation](agent-sandbox-docker-engine.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-DE-28` | `F-DE-2`, `F-DE-3`, `F-DE-4`, `P-DE-1`, `P-DE-5`, `P-DE-4` | information disclosure | A control carrier or network view exposes another launch's session or loopback listener. | [Loopback listeners](agent-ports.md#Loopback%20listeners) The design requires separate network views. The provider binding remains unresolved. [Provider mediation](agent-sandbox-docker-engine.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)) |
| `TM-DE-29` | `F-DE-2`, `F-DE-3`, `F-DE-4`, `P-DE-1`, `P-DE-5`, `P-DE-4` | denial of service | A launch floods its local control carrier and prevents lifecycle or session progress. | Local carrier bounds and sharing remain unresolved under [mediator overload control](agent-open-questions.md#%60OQ-SBX-14%60). [Provider mediation](agent-sandbox-docker-engine.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60)) |
| `TM-DE-30` | `F-DE-2`, `F-DE-3`, `F-DE-4`, `P-DE-1`, `P-DE-5`, `P-DE-4` | elevation of privilege | A local process uses an inherited control handle for host operations outside its launch authority. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox). Control handle ownership and launch binding remain unresolved. [Common mediator threats](agent-threat-model.md#Agent%20mediator%20request%20admission) cover protocol parser attacks. [Provider mediation](agent-sandbox-docker-engine.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-DE-67` | `F-DE-2`, `F-DE-3`, `P-DE-1`, `P-DE-5`, `P-DE-4` | spoofing | A process impersonates the local control endpoint or reuses its previous launch identity. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox). A local endpoint needs the same binding guarantees as a guest relay. Its carrier authentication remains unresolved. [Provider mediation](agent-sandbox-docker-engine.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60)) |

### Workspace and scratch storage

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-DE-19` | `F-DE-5`, `F-DE-6`, `F-DE-7`, `F-DE-8`, `P-DE-1`, `DS-DE-1`, `DS-DE-2`, `P-DE-3`, `DS-DE-4` | spoofing | A same-user process substitutes a workspace owner or capture helper. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Ownership and helper rights depend on launch identity. [Provider startup](agent-sandbox-docker-engine.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)) |
| `TM-DE-32` | `F-DE-5`, `F-DE-6`, `F-DE-7`, `F-DE-8`, `P-DE-1`, `DS-DE-1`, `DS-DE-2`, `P-DE-3`, `DS-DE-4` | tampering | Hostile workspace metadata redirects host capture or cleanup outside the selected workspace. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Content capture supplies no safe metadata interpretation contract. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |
| `TM-DE-33` | `F-DE-5`, `F-DE-6`, `F-DE-7`, `F-DE-8`, `P-DE-1`, `DS-DE-1`, `DS-DE-2`, `P-DE-3`, `DS-DE-4` | repudiation | A workspace or scratch writer denies changes made through a shared local identity. | [Provider observations](agent-sandbox-lifecycle.md#Provider%20observations) Host compromise has no independent evidence mechanism. [Provider observations](agent-sandbox-docker-engine.md#Host%20observations). | gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)) |
| `TM-DE-22` | `F-DE-5`, `F-DE-6`, `F-DE-7`, `F-DE-8`, `P-DE-1`, `DS-DE-1`, `DS-DE-2`, `P-DE-3`, `DS-DE-4` | information disclosure | Capture, scratch reuse, or path aliasing exposes another launch's files. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Capture rights and storage placement remain unresolved. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-DE-35` | `F-DE-5`, `F-DE-6`, `F-DE-7`, `F-DE-8`, `P-DE-1`, `DS-DE-1`, `DS-DE-2`, `P-DE-3`, `DS-DE-4` | denial of service | Workspace or scratch growth exhausts host capacity outside an enforced allocation. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables) Scratch is outside the workspace, but storage quotas and cleanup are unspecified. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-DE-36` | `F-DE-5`, `F-DE-6`, `F-DE-7`, `F-DE-8`, `P-DE-1`, `DS-DE-1`, `DS-DE-2`, `P-DE-3`, `DS-DE-4` | elevation of privilege | Host-side Git processing runs repository hooks or configuration or follows alternates or submodules or accepts hostile worktree paths. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Archive extraction checks don't protect later Git or filesystem consumers. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |

### Shared named tool cache

Cache sharing has no stated authorization, provenance, capacity, or invalidation mechanism.

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-DE-127` | `F-DE-9`, `F-DE-10`, `P-DE-1`, `DS-DE-3`, `E-DE-2` | spoofing | One launch places cache entries that another mistakes for trusted tool output. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-DE-128` | `F-DE-9`, `F-DE-10`, `P-DE-1`, `DS-DE-3`, `E-DE-2` | tampering | A launch changes shared cache content consumed by another launch. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-DE-129` | `F-DE-9`, `F-DE-10`, `P-DE-1`, `DS-DE-3`, `E-DE-2` | repudiation | A cache writer denies modifying content because cache provenance has no stated record. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-DE-130` | `F-DE-9`, `F-DE-10`, `P-DE-1`, `DS-DE-3`, `E-DE-2` | information disclosure | A launch reads private data another launch left in a shared tool cache. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-DE-131` | `F-DE-9`, `F-DE-10`, `P-DE-1`, `DS-DE-3`, `E-DE-2` | denial of service | A launch fills or churns a shared cache and denies other launches useful capacity. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-DE-132` | `F-DE-9`, `F-DE-10`, `P-DE-1`, `DS-DE-3`, `E-DE-2` | elevation of privilege | A cached executable or build result runs under a later launch's broader authority. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-docker-engine.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |

### Agent checkpoints

This provider has no agent snapshot. Checkpoint capture, workspace binding, and resume behavior remain unresolved.

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-DE-121` | `F-DE-11`, `F-DE-12`, `P-DE-2`, `DS-DE-5`, `P-DE-6` | spoofing | A capture or restore request substitutes another agent's owner or launch. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-docker-engine.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-DE-122` | `F-DE-11`, `F-DE-12`, `P-DE-2`, `DS-DE-5`, `P-DE-6` | tampering | A mismatched checkpoint or workspace content restores inconsistent or hostile state. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-docker-engine.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-DE-123` | `F-DE-11`, `F-DE-12`, `P-DE-2`, `DS-DE-5`, `P-DE-6` | repudiation | Rewind hides an earlier external effect or presents it as undone. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-docker-engine.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-DE-124` | `F-DE-11`, `F-DE-12`, `P-DE-2`, `DS-DE-5`, `P-DE-6` | information disclosure | Checkpoint transfer exposes private session or workspace data to an unauthorized reader. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-docker-engine.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-DE-125` | `F-DE-11`, `F-DE-12`, `P-DE-2`, `DS-DE-5`, `P-DE-6` | denial of service | Retained captures or uncertain ownership prevent safe collection and consume capacity. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-docker-engine.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-DE-126` | `F-DE-11`, `F-DE-12`, `P-DE-2`, `DS-DE-5`, `P-DE-6` | elevation of privilege | Restoration carries old identity, executable configuration, or authority into a successor launch. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-docker-engine.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |

### Commands in the agent workspace

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-DE-40` | `F-DE-13`, `F-DE-14`, `F-DE-15`, `E-DE-1`, `P-DE-7`, `DS-DE-1`, `P-DE-4` | spoofing | A workspace command uses the authority of a different agent launch. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) Standalone command confinement doesn't specify launch identity binding. [Provider startup](agent-sandbox-docker-engine.md#Process%20placement%20and%20startup). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-DE-41` | `F-DE-13`, `F-DE-14`, `F-DE-15`, `E-DE-1`, `P-DE-7`, `DS-DE-1`, `P-DE-4` | tampering | A command races capture or changes files outside the selected workspace through a stale path. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Command placement and capture ordering remain unresolved. [Provider startup](agent-sandbox-docker-engine.md#Process%20placement%20and%20startup). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-DE-42` | `F-DE-13`, `F-DE-14`, `F-DE-15`, `E-DE-1`, `P-DE-7`, `DS-DE-1`, `P-DE-4` | repudiation | A command runs without a record binding its caller, launch, and confinement settings. | [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) Command attribution and lifecycle records remain unspecified. [Provider observations](agent-sandbox-docker-engine.md#Host%20observations). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-DE-43` | `F-DE-13`, `F-DE-14`, `F-DE-15`, `E-DE-1`, `P-DE-7`, `DS-DE-1`, `P-DE-4` | information disclosure | A command inherits another launch's environment, session data, or readable workspace paths. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) Execution identity, inherited state, and path ownership remain unresolved. [Provider startup](agent-sandbox-docker-engine.md#Process%20placement%20and%20startup). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-DE-44` | `F-DE-13`, `F-DE-14`, `F-DE-15`, `E-DE-1`, `P-DE-7`, `DS-DE-1`, `P-DE-4` | denial of service | A command continues running after cancellation or blocks workspace capture and teardown. | [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) Independent command lifetime and cancellation remain unspecified. [Provider teardown](agent-sandbox-docker-engine.md#Cancellation%20and%20teardown). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-DE-45` | `F-DE-13`, `F-DE-14`, `F-DE-15`, `E-DE-1`, `P-DE-7`, `DS-DE-1`, `P-DE-4` | elevation of privilege | A command starts outside the selected provider confinement or obtains a relaxation without review. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) afakt-sandbox confines standalone commands. Authority for an unsandboxed invocation remains unresolved. [Provider startup](agent-sandbox-docker-engine.md#Process%20placement%20and%20startup). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60)) |

### Per-launch network boundary

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-DE-155` | `F-DE-16`, `P-DE-4`, `P-DE-8` | spoofing | A process substitutes another launch's local agent-proxy endpoint. | [Per-launch services](agent-workload-proxy.md#Launch%20boundary) requires a local channel bound to the launch. Its peer authentication remains undecided. | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-DE-156` | `F-DE-16`, `P-DE-4`, `P-DE-8` | tampering | A malformed connection request changes its caller or target. | [Overlay connections](agent-workload-proxy.md#Overlay%20connections) require direct mutual authentication and mediator checks. Binding caller and publication context to the guest listener remains open. | depends on an open question ([`OQ-PORT-14`](agent-open-questions.md#%60OQ-PORT-14%60)) |
| `TM-DE-157` | `F-DE-16`, `P-DE-4`, `P-DE-8` | repudiation | Reconnect loses the association between an admitted stream and its authorization record. | [Session message log](agent-acp.md#Session%20message%20log) requires retained history. Its owner, replay, and reconnect attribution remain open. | depends on an open question ([`OQ-SES-18`](agent-open-questions.md#%60OQ-SES-18%60)) |
| `TM-DE-158` | `F-DE-16`, `P-DE-4`, `P-DE-8` | information disclosure | Connection replacement sends a stream to a different launch. | Direct [overlay connections](agent-workload-proxy.md#Overlay%20connections) authenticate both peers and require mediator decisions. How that publication and caller map to the intended listener remains unresolved. | depends on an open question ([`OQ-PORT-14`](agent-open-questions.md#%60OQ-PORT-14%60)) |
| `TM-DE-159` | `F-DE-16`, `P-DE-4`, `P-DE-8` | denial of service | One launch consumes proxy buffers or blocks control streams with published-port traffic. | [Overlay connections](agent-workload-proxy.md#Overlay%20connections) requires bounded admission. The local scheduling and buffer limits remain undecided. | depends on an open question ([`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60)) |
| `TM-DE-160` | `F-DE-16`, `P-DE-4`, `P-DE-8` | elevation of privilege | After compromising the agent proxy, an attacker sends traffic without mediator approval. | [Per-launch services](agent-workload-proxy.md#Launch%20boundary) separates policy from network authority. Their authenticated local binding remains undecided. | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |

## Attack trees

`OR` requires any child. `AND` requires every child. Leaves reference local threats. Shared policy failures use the [common attack trees](agent-threat-model.md#Attack%20trees).

### Escape to the agent host

```mermaid
flowchart TD
    root["Escape to the agent host"] --> choice{"OR"}
    choice --> leaf0["TM-DE-6: Exploit the isolation mechanism"]
    choice --> leaf1["TM-DE-36: Make host processing execute workspace metadata"]
    choice --> leaf2["TM-DE-30: Abuse an inherited control handle"]
    choice --> leaf3["TM-DE-45: Run a workspace command outside intended confinement"]
```

- Cheapest path and blocking controls. Trigger host processing of hostile workspace metadata without a guest kernel exploit. [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60) leaves the host-side consumer unprotected. The [platform attack surface](#Host%20visibility%20and%20workspace%20commands) has separate confinement gaps.

### Read or change another agent launch's workspace or session

```mermaid
flowchart TD
    root["Read or change another agent launch's workspace or session"] --> choice{"OR"}
    choice --> leaf0["TM-DE-130: Read a shared cache"]
    choice --> leaf1["TM-DE-128: Change a cache consumed by another launch"]
    choice --> leaf2["TM-DE-22: Cross workspace or scratch ownership"]
    choice --> leaf3["TM-DE-28: Use an incorrectly bound local endpoint"]
    choice --> leaf4["TM-DE-124: Read retained capture"]
```

- Cheapest path and blocking controls. Use a shared named tool cache already exposed to the launch. Cache authorization and provenance remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Workspace ownership and local endpoint binding require separate controls.

### Send traffic past egress policy

```mermaid
flowchart TD
    root["Send traffic past egress policy"] --> choice{"OR"}
    choice --> leaf0["TM-DE-26: Use a direct network route"]
    choice --> leaf1["TM-DE-25: Select another mediator identity"]
    choice --> leaf2["TM-DE-45: Run a workspace command outside confinement"]
    choice --> leaf3["TM-DE-30: Use a privileged control handle"]
```

- Cheapest path and blocking controls. Use a network path whose proxy routing is unspecified. [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60) covers missing provider routing. [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60) covers sender binding. The [common egress tree](agent-threat-model.md#Send%20traffic%20past%20egress%20policy) covers permitted name queries, passthrough, and relay proxy use.

### Persist hostile state across agent launches

```mermaid
flowchart TD
    root["Persist hostile state across agent launches"] --> choice{"OR"}
    choice --> leaf0["TM-DE-122: Restore mismatched or hostile captured state"]
    choice --> leaf1["TM-DE-126: Carry old authority into a successor launch"]
    choice --> leaf2["TM-DE-132: Execute a shared cache entry under later authority"]
```

- Cheapest path and blocking controls. Leave executable content in a shared named tool cache for a later launch. Cache provenance and invalidation remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Captured state needs a separate binding under [`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60), while authority carried into a successor launch remains [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60). The [common persistence tree](agent-threat-model.md#Persist%20hostile%20state%20across%20agent%20launches) covers workspace images and hooks.

## Residual risks and gaps

Host isolation remains on hold. The [common residual risks](agent-threat-model.md#Residual%20risks%20and%20gaps) keep hook protection, repository configuration scope, repository profile self-grant and credential domain bounds, and session-end wait with workload identity lifetime. No pending option supplies a mitigation.

### Isolation and lifecycle

- `TM-DE-1`. A local process substitutes for the provider control endpoint or launch process. depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)).
- `TM-DE-2`. A launch changes inherited handles, mounts, descriptors, or confinement inputs to reach provider authority. depends on an open question ([`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60)).
- `TM-DE-21`. A compromised agent host invents process exits or hides local workspace and command effects. gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)).
- `TM-DE-70`. A same-user host process reads mediator memory, local session traffic, or another launch's files. depends on an open question ([`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60)).
- `TM-DE-23`. A hostile process tree ignores cancellation and continues running. gap ([`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60)).
- `TM-DE-6`. A process in the container gains authority over the shared kernel or Docker management. gap ([`OQ-SBX-56`](agent-open-questions.md#%60OQ-SBX-56%60)).
- `TM-DE-154`. An orphaned process or reused launch user retains local authority after worker failure. depends on an open question ([`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60)).

### Local transport and egress

- `TM-DE-25`. A process selects another launch's mediator endpoint and receives its authority. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-DE-26`. A process changes proxy settings or uses an interface that bypasses the agent mediator. gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)).
- `TM-DE-27`. Unbound local traffic acquires another launch's attribution before shared logging. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-DE-28`. A control carrier or network view exposes another launch's session or loopback listener. gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)).
- `TM-DE-29`. A launch floods its local control carrier and prevents lifecycle or session progress. depends on an open question ([`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60)).
- `TM-DE-30`. A local process uses an inherited control handle for host operations outside its launch authority. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-DE-67`. A process impersonates the local control endpoint or reuses its previous launch identity. depends on an open question ([`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60)).

### Workspace and scratch storage

- `TM-DE-19`. A same-user process substitutes a workspace owner or capture helper. depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)).
- `TM-DE-32`. Hostile workspace metadata redirects host capture or cleanup outside the selected workspace. gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).
- `TM-DE-33`. A workspace or scratch writer denies changes made through a shared local identity. gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)).
- `TM-DE-22`. Capture, scratch reuse, or path aliasing exposes another launch's files. gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-DE-35`. Workspace or scratch growth exhausts host capacity outside an enforced allocation. gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-DE-36`. Host-side Git processing runs repository hooks or configuration or follows alternates or submodules or accepts hostile worktree paths. gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).

### Shared named tool cache

- `TM-DE-127`. One launch places cache entries that another mistakes for trusted tool output. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-DE-128`. A launch changes shared cache content consumed by another launch. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-DE-129`. A cache writer denies modifying content because cache provenance has no stated record. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-DE-130`. A launch reads private data another launch left in a shared tool cache. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-DE-131`. A launch fills or churns a shared cache and denies other launches useful capacity. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-DE-132`. A cached executable or build result runs under a later launch's broader authority. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).

### Agent checkpoints

- `TM-DE-121`. A capture or restore request substitutes another agent's owner or launch. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-DE-122`. A mismatched checkpoint or workspace content restores inconsistent or hostile state. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-DE-123`. Rewind hides an earlier external effect or presents it as undone. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-DE-124`. Checkpoint transfer exposes private session or workspace data to an unauthorized reader. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-DE-125`. Retained captures or uncertain ownership prevent safe collection and consume capacity. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-DE-126`. Restoration carries old identity, executable configuration, or authority into a successor launch. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).

### Commands in the agent workspace

- `TM-DE-40`. A workspace command uses the authority of a different agent launch. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-DE-41`. A command races capture or changes files outside the selected workspace through a stale path. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-DE-42`. A command runs without a record binding its caller, launch, and confinement settings. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-DE-43`. A command inherits another launch's environment, session data, or readable workspace paths. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-DE-44`. A command continues running after cancellation or blocks workspace capture and teardown. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-DE-45`. A command starts outside the selected provider confinement or obtains a relaxation without review. depends on an open question ([`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60)).
