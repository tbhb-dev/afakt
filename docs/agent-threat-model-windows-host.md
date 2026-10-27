---
type: Threat Model
title: "Agent threat model for Windows host"
description: Isolation, local transport, retained state, and platform threats for the windows-host agent sandbox provider.
tags: [afakt, agents, security, threat-model]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent threat model for Windows host

[Public key infrastructure](pki.md).

## Scope

The windows-host agent sandbox provider uses the [provider design](agent-sandbox-windows-host.md). The [common agent threat model](agent-threat-model.md) covers shared service authority, credentials, network policy, sessions, registries, and audit records.

[Platforms and privileges](agent-sandbox-windows-host.md#Platforms%20and%20privileges) and [Provider isolation](agent-sandbox-windows-host.md#Isolation%20boundary) define the platform and its kernel boundary. [Provider startup](agent-sandbox-windows-host.md#Process%20placement%20and%20startup) places afakt-agent, afakt-sandbox, and the harness.

STRIDE covers spoofing, tampering, repudiation, information disclosure, denial of service, and elevation of privilege. A virtual machine (VM) supplies a separate guest kernel where the provider uses one. Agent Client Protocol (ACP) transports session messages. Domain Name System (DNS) queries use the agent mediator. A User Identifier (UID) names an operating-system user. An Application Programming Interface (API) provides shared service authority. Windows Filtering Platform (WFP) redirects flows using an AppContainer Security Identifier (SID).

### Agent mediator access

The callout driver redirects launch traffic to the mediator through the registration service. Persistent block filters deny external connectivity when capture components are absent. Complete redirect handling, resolver exclusion, and existing-flow behavior remain unresolved. See [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution).

### Workspace, scratch, and caches

[Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches) distinguishes launch state, shared named tool caches, and verified image caches. Provider storage mappings, scratch cleanup, shared cache trust, and hostile host-side repository processing remain unresolved. [Provider materialization](agent-sandbox-windows-host.md#Workspace%20image%20materialization) separates verified extraction from later workspace capture.

### Agent snapshots and agent checkpoints

[Provider capture](agent-sandbox-windows-host.md#Agent%20snapshots%20and%20agent%20checkpoints) defines the capture boundary and unresolved restore contract. This provider has no agent snapshot. A native harness checkpoint doesn't capture its shared kernel or engine.

### Host visibility and workspace commands

Host custody and its limits follow [Provider observations](agent-sandbox-windows-host.md#Host%20observations). Termination authority and unresolved descendant cleanup follow [Provider teardown](agent-sandbox-windows-host.md#Cancellation%20and%20teardown). Same-user process separation and the host isolation baseline remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

Commands started by `afakt agent exec` use standalone afakt-sandbox. [Provider startup](agent-sandbox-windows-host.md#Process%20placement%20and%20startup) leaves command placement, inherited state, and launch authority under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). An unsandboxed invocation remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

## Trust boundaries

| Boundary ID | Boundary | Two sides | Enforcement in the design |
| --- | --- | --- | --- |
| `TB-WH-1` | Agent harness and provider control | untrusted process tree / provider lifecycle and platform authority | [Enforcement](agent-sandbox.md#Enforcement) defines the isolation tier. [Host visibility and workspace commands](#Host%20visibility%20and%20workspace%20commands) identifies platform exposure and missing confinement. |
| `TB-WH-2` | Local transport and agent mediator | Launch network and control endpoints / mediator identity and policy | [Agent mediator](agent-mediator.md#Launch%20boundary) Local endpoint binding remains [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). |
| `TB-WH-3` | Agent workspace and host processing | Launch-controlled files / worker capture, materialization, and cleanup | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Host-side handling of repository metadata remains [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). |
| `TB-WH-4` | Shared named tool cache | One launch / cache content consumed by another launch | [Sandbox variables](agent-sandbox.md#Sandbox%20variables) Access, provenance, and capacity controls remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). |
| `TB-WH-5` | Captured state and successor launch | Current launch / retained capture and later launch | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume) State authorization and ownership remain [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). |
| `TB-WH-6` | Workspace command and launch authority | Authorized request for a command / command confinement and workspace access | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) Command placement and lifecycle binding remain [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). |

## Data flow diagram

### Element inventory

| Element ID | Element | Location or custody |
| --- | --- | --- |
| `E-WH-1` | Command caller | Authenticated client. Shared service grants apply |
| `E-WH-2` | Another agent launch | Separate launch authority |
| `P-WH-1` | Agent harness and afakt-sandbox | Host process confinement |
| `P-WH-2` | Agent host worker and provider lifecycle | Agent host |
| `P-WH-3` | Host workspace processor | Worker capture and cleanup |
| `P-WH-9` | Agent proxy | Outside guest confinement. Launch identity custody |
| `P-WH-4` | Agent mediator | Outside launch confinement. Process separation unresolved |
| `P-WH-5` | Local control endpoint | Carrier and placement unspecified |
| `P-WH-6` | Successor launch | New launch authority |
| `P-WH-7` | afakt agent exec command | Execution placement unspecified |
| `DS-WH-1` | Agent workspace | Per-launch writable content |
| `DS-WH-2` | Scratch storage | Outside workspace. Backing unspecified |
| `DS-WH-3` | Shared named tool cache | Shared across launches. Backing unspecified |
| `DS-WH-4` | Verified image cache | Agent host. Distinct from named tool cache |
| `DS-WH-5` | Checkpoint capture | Retained state. Access controls unresolved |
| `P-WH-8` | callout driver and registration service | Outside harness confinement. Registration and recovery unresolved |

### Level one provider flows

Arrows identify exchanges and their return paths. Unspecified carriers and placements remain gaps. The shared named tool cache is outside individual launch confinement and has a separate crossing from another launch. The verified image cache has a different trust relationship.

```mermaid
flowchart LR
    subgraph z0["TB-WH-1 / launch confinement"]
        P-WH-1["P-WH-1: Agent harness and afakt-sandbox"]
    end
    subgraph z1["TB-WH-2 / local control"]
        P-WH-5["P-WH-5: Local control endpoint"]
    end
    subgraph z2["TB-WH-2 / agent host policy"]
        P-WH-4["P-WH-4: Agent mediator"]
    end
    subgraph z3["TB-WH-3 / launch-controlled storage"]
        DS-WH-1["DS-WH-1: Agent workspace"]
        DS-WH-2["DS-WH-2: Scratch storage"]
    end
    subgraph z4["TB-WH-3 / agent host processing"]
        P-WH-2["P-WH-2: Agent host worker and provider lifecycle"]
        P-WH-3["P-WH-3: Host workspace processor"]
        DS-WH-4["DS-WH-4: Verified image cache"]
    end
    subgraph z5["TB-WH-4 / shared across launches"]
        DS-WH-3["DS-WH-3: Shared named tool cache"]
    end
    subgraph z6["TB-WH-4 / other launch"]
        E-WH-2["E-WH-2: Another agent launch"]
    end
    subgraph z7["TB-WH-5 / retained capture"]
        DS-WH-5["DS-WH-5: Checkpoint capture"]
    end
    subgraph z8["TB-WH-5 / successor"]
        P-WH-6["P-WH-6: Successor launch"]
    end
    subgraph z9["TB-WH-6 / caller"]
        E-WH-1["E-WH-1: Command caller"]
    end
    subgraph z10["TB-WH-6 / command placement unresolved"]
        P-WH-7["P-WH-7: afakt agent exec command"]
    end
    P-WH-2 -->|"F-WH-1"| P-WH-1
    P-WH-1 -->|"F-WH-2"| P-WH-5
    P-WH-5 -->|"F-WH-3"| P-WH-4
    P-WH-8["P-WH-8: callout driver and registration service"]
    P-WH-1 -->|"F-WH-4"| P-WH-8
    P-WH-8 -->|"F-WH-16"| P-WH-4
    P-WH-2 -->|"F-WH-17"| P-WH-8
    P-WH-1 -->|"F-WH-5"| DS-WH-1
    P-WH-1 -->|"F-WH-6"| DS-WH-2
    P-WH-3 -->|"F-WH-7"| DS-WH-1
    DS-WH-4 -->|"F-WH-8"| P-WH-3
    P-WH-1 -->|"F-WH-9"| DS-WH-3
    E-WH-2 -->|"F-WH-10"| DS-WH-3
    P-WH-2 -->|"F-WH-11"| DS-WH-5
    DS-WH-5 -->|"F-WH-12"| P-WH-6
    E-WH-1 -->|"F-WH-13"| P-WH-7
    P-WH-7 -->|"F-WH-14"| DS-WH-1
    P-WH-7 -->|"F-WH-15"| P-WH-4
```

- **`F-WH-1`.** Startup, stop, and observed exit of the process tree. Crosses `TB-WH-1`.
- **`F-WH-2`.** Session and local control messages. Endpoint binding remains unresolved. Crosses `TB-WH-2`.
- **`F-WH-3`.** Control messages and preview streams over the provider transport. Carrier authentication remains unresolved. Crosses `TB-WH-2`.
- **`F-WH-4`.** Launch packets and name queries enter the provider capture component. Indirect service paths remain unresolved. Crosses `TB-WH-2`.
- **`F-WH-5`.** Workspace reads and writes under launch confinement. Crosses `TB-WH-3`.
- **`F-WH-6`.** Scratch reads and writes outside the workspace. Allocation and cleanup are unspecified. Crosses `TB-WH-3`.
- **`F-WH-7`.** Host-side capture, metadata interpretation, and cleanup of hostile workspace content. Crosses `TB-WH-3`.
- **`F-WH-8`.** Verified immutable image content used for materialization. Cache ownership is provider-specific. Crosses `TB-WH-3`.
- **`F-WH-9`.** Named cache reads and writes by this launch. Mapping to backing storage is unspecified. Crosses `TB-WH-4`.
- **`F-WH-10`.** Named cache reads and writes by another launch across the same sharing boundary. Crosses `TB-WH-4`.
- **`F-WH-11`.** Capture and restore of harness checkpoint and workspace state. The contract for resuming execution remains unresolved and doesn't include a VM snapshot. Crosses `TB-WH-5`.
- **`F-WH-12`.** State admitted into a successor launch under new authority. Crosses `TB-WH-5`.
- **`F-WH-13`.** Authorized workspace command request. Local placement and credential binding remain unresolved. Crosses `TB-WH-6`.
- **`F-WH-14`.** Command reads and writes in the selected agent workspace. Concurrent access remains unresolved. Crosses `TB-WH-6`, `TB-WH-3`.
- **`F-WH-15`.** Command traffic requiring launch egress policy. Transport and identity binding remain unresolved. Crosses `TB-WH-6`, `TB-WH-2`.
- **`F-WH-16`.** Captured traffic and original destinations pass to the selected mediator. Registration and failure behavior remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60). Crosses `TB-WH-2`.
- **`F-WH-17`.** The worker registers and retires launch capture mappings. Management authentication and privilege confinement remain under [`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60). Crosses `TB-WH-1`, `TB-WH-2`.

## Agent proxy

For each launch, the provider supplies one agent mediator and one agent proxy. Without network access of its own, the mediator passes permitted requests to the agent proxy. Using the launch identity outside guest confinement, that service opens the authenticated hub connection. Local framing, service users, and provider integration remain undecided under [Per-launch services](agent-proxy.md#Launch%20boundary).

```mermaid
flowchart LR
    P-WH-4["P-WH-4: Agent mediator"]
    P-WH-9["P-WH-9: Agent proxy"]
    P-WH-4 -->|"F-WH-18"| P-WH-9
```

- **`F-WH-18`.** Permitted requests, launch identity, and returning streams across the mediator/agent-proxy boundary. The local channel needs an authenticated peer and stream context.

## STRIDE analysis

Each table covers the six categories for its listed flows and every participating endpoint. Shared service checks are in the [common STRIDE analysis](agent-threat-model.md#STRIDE%20analysis). They don't establish provider isolation. A mitigation applies only to the stated attack.

### Isolation and lifecycle

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WH-1` | `F-WH-1`, `P-WH-2`, `P-WH-1` | spoofing | A local process substitutes for the provider control endpoint or launch process. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) [Runtime privileges](agent-sandbox.md#Runtime%20privileges) Per-launch identities and helper authority remain unresolved. [Provider startup](agent-sandbox-windows-host.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)) |
| `TM-WH-2` | `F-WH-1`, `P-WH-2`, `P-WH-1` | tampering | A launch changes inherited handles, mounts, descriptors, or confinement inputs to reach provider authority. | [Enforcement](agent-sandbox.md#Enforcement). [Platform exposure](#Host%20visibility%20and%20workspace%20commands). Unspecified. [Provider isolation](agent-sandbox-windows-host.md#Isolation%20boundary). | depends on an open question ([`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60)) |
| `TM-WH-21` | `F-WH-1`, `P-WH-2`, `P-WH-1` | repudiation | A compromised agent host invents process exits or hides local workspace and command effects. | [Provider observations](agent-sandbox-lifecycle.md#Provider%20observations) Independent verification of a compromised host is unspecified. [Provider observations](agent-sandbox-windows-host.md#Host%20observations). | gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)) |
| `TM-WH-70` | `F-WH-1`, `P-WH-2`, `P-WH-1` | information disclosure | A same-user host process reads mediator memory, local session traffic, or another launch's files. | [Trust boundary](agent-host.md#Trust%20boundary) The design trusts host custody. Same-user process separation remains on hold. [Provider observations](agent-sandbox-windows-host.md#Host%20observations). | depends on an open question ([`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60)) |
| `TM-WH-23` | `F-WH-1`, `P-WH-2`, `P-WH-1` | denial of service | A hostile process tree ignores cancellation and continues running. | [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) Bounds on cancellation don't specify how this provider kills and observes every descendant. [Provider teardown](agent-sandbox-windows-host.md#Cancellation%20and%20teardown). | gap ([`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60)) |
| `TM-WH-24` | `F-WH-1`, `P-WH-2`, `P-WH-1` | elevation of privilege | An agent harness exploits the shared host kernel or escapes process confinement. | [Enforcement](agent-sandbox.md#Enforcement). [Platform exposure](#Host%20visibility%20and%20workspace%20commands). Unspecified. [Provider isolation](agent-sandbox-windows-host.md#Isolation%20boundary). | gap ([`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60)) |
| `TM-WH-154` | `F-WH-1`, `P-WH-2`, `P-WH-1` | elevation of privilege | An orphaned process or reused launch user retains local authority after worker failure. | [Restart and recovery](agent-host.md#Restart%20and%20recovery) Durable ownership and safe identity reuse remain unresolved. [Provider teardown](agent-sandbox-windows-host.md#Cancellation%20and%20teardown). | depends on an open question ([`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60)) |

### Local transport and egress

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WH-25` | `F-WH-2`, `F-WH-3`, `F-WH-4`, `P-WH-1`, `P-WH-5`, `P-WH-4` | spoofing | A process selects another launch's mediator endpoint and receives its authority. | [Agent mediator](agent-mediator.md#Launch%20boundary) Endpoint authentication remains unresolved. [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-WH-26` | `F-WH-2`, `F-WH-3`, `F-WH-4`, `P-WH-1`, `P-WH-5`, `P-WH-4` | tampering | A process changes proxy settings or uses an interface that bypasses the agent mediator. | [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution) selects traffic interception outside the harness. Complete resolver, indirect-service, and capture-failure enforcement remains unresolved. See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)) |
| `TM-WH-27` | `F-WH-2`, `F-WH-3`, `F-WH-4`, `P-WH-1`, `P-WH-5`, `P-WH-4` | repudiation | Unbound local traffic acquires another launch's attribution before shared logging. | [Control requests](agent-mediator.md#Control%20requests) Stamping identity requires a trustworthy local sender binding. [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-WH-28` | `F-WH-2`, `F-WH-3`, `F-WH-4`, `P-WH-1`, `P-WH-5`, `P-WH-4` | information disclosure | A control carrier or network view exposes another launch's session or loopback listener. | [Loopback listeners](agent-ports.md#Loopback%20listeners) The design requires separate network views. The provider binding remains unresolved. [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)) |
| `TM-WH-29` | `F-WH-2`, `F-WH-3`, `F-WH-4`, `P-WH-1`, `P-WH-5`, `P-WH-4` | denial of service | A launch floods its local control carrier and prevents lifecycle or session progress. | Local carrier bounds and sharing remain unresolved under [mediator overload control](agent-open-questions.md#%60OQ-SBX-14%60). [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60)) |
| `TM-WH-30` | `F-WH-2`, `F-WH-3`, `F-WH-4`, `P-WH-1`, `P-WH-5`, `P-WH-4` | elevation of privilege | A local process uses an inherited control handle for host operations outside its launch authority. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox). Control handle ownership and launch binding remain unresolved. [Common mediator threats](agent-threat-model.md#Agent%20mediator%20request%20admission) cover protocol parser attacks. [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-WH-67` | `F-WH-2`, `F-WH-3`, `P-WH-1`, `P-WH-5`, `P-WH-4` | spoofing | A process impersonates the local control endpoint or reuses its previous launch identity. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox). A local endpoint needs the same binding guarantees as a guest relay. Its carrier authentication remains unresolved. [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60)) |
| `TM-WH-82` | `F-WH-4`, `P-WH-1`, `P-WH-4` | information disclosure | A host-mode launch reads another launch's loopback listener through a shared network view. | [Loopback listeners](agent-ports.md#Loopback%20listeners) Host-mode enforcement remains unresolved. [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](agent-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-16`](agent-open-questions.md#%60OQ-SBX-16%60)) |

### Workspace and scratch storage

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WH-19` | `F-WH-5`, `F-WH-6`, `F-WH-7`, `F-WH-8`, `P-WH-1`, `DS-WH-1`, `DS-WH-2`, `P-WH-3`, `DS-WH-4` | spoofing | A same-user process substitutes a workspace owner or capture helper. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Ownership and helper rights depend on launch identity. [Provider startup](agent-sandbox-windows-host.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)) |
| `TM-WH-32` | `F-WH-5`, `F-WH-6`, `F-WH-7`, `F-WH-8`, `P-WH-1`, `DS-WH-1`, `DS-WH-2`, `P-WH-3`, `DS-WH-4` | tampering | Hostile workspace metadata redirects host capture or cleanup outside the selected workspace. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Content capture supplies no safe metadata interpretation contract. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |
| `TM-WH-33` | `F-WH-5`, `F-WH-6`, `F-WH-7`, `F-WH-8`, `P-WH-1`, `DS-WH-1`, `DS-WH-2`, `P-WH-3`, `DS-WH-4` | repudiation | A workspace or scratch writer denies changes made through a shared local identity. | [Provider observations](agent-sandbox-lifecycle.md#Provider%20observations) Host compromise has no independent evidence mechanism. [Provider observations](agent-sandbox-windows-host.md#Host%20observations). | gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)) |
| `TM-WH-22` | `F-WH-5`, `F-WH-6`, `F-WH-7`, `F-WH-8`, `P-WH-1`, `DS-WH-1`, `DS-WH-2`, `P-WH-3`, `DS-WH-4` | information disclosure | Capture, scratch reuse, or path aliasing exposes another launch's files. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Capture rights and storage placement remain unresolved. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60)) |
| `TM-WH-35` | `F-WH-5`, `F-WH-6`, `F-WH-7`, `F-WH-8`, `P-WH-1`, `DS-WH-1`, `DS-WH-2`, `P-WH-3`, `DS-WH-4` | denial of service | Workspace or scratch growth exhausts host capacity outside an enforced allocation. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables) Scratch is outside the workspace, but storage quotas and cleanup are unspecified. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-WH-36` | `F-WH-5`, `F-WH-6`, `F-WH-7`, `F-WH-8`, `P-WH-1`, `DS-WH-1`, `DS-WH-2`, `P-WH-3`, `DS-WH-4` | elevation of privilege | Host-side Git processing runs repository hooks or configuration or follows alternates or submodules or accepts hostile worktree paths. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Archive extraction checks don't protect later Git or filesystem consumers. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |

### Shared named tool cache

Cache sharing has no stated authorization, provenance, capacity, or invalidation mechanism.

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WH-127` | `F-WH-9`, `F-WH-10`, `P-WH-1`, `DS-WH-3`, `E-WH-2` | spoofing | One launch places cache entries that another mistakes for trusted tool output. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WH-128` | `F-WH-9`, `F-WH-10`, `P-WH-1`, `DS-WH-3`, `E-WH-2` | tampering | A launch changes shared cache content consumed by another launch. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WH-129` | `F-WH-9`, `F-WH-10`, `P-WH-1`, `DS-WH-3`, `E-WH-2` | repudiation | A cache writer denies modifying content because cache provenance has no stated record. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WH-130` | `F-WH-9`, `F-WH-10`, `P-WH-1`, `DS-WH-3`, `E-WH-2` | information disclosure | A launch reads private data another launch left in a shared tool cache. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WH-131` | `F-WH-9`, `F-WH-10`, `P-WH-1`, `DS-WH-3`, `E-WH-2` | denial of service | A launch fills or churns a shared cache and denies other launches useful capacity. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WH-132` | `F-WH-9`, `F-WH-10`, `P-WH-1`, `DS-WH-3`, `E-WH-2` | elevation of privilege | A cached executable or build result runs under a later launch's broader authority. | [Sandbox variables](agent-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](agent-sandbox-windows-host.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |

### Agent checkpoints

This provider has no agent snapshot. Checkpoint capture, workspace binding, and resume behavior remain unresolved.

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WH-121` | `F-WH-11`, `F-WH-12`, `P-WH-2`, `DS-WH-5`, `P-WH-6` | spoofing | A capture or restore request substitutes another agent's owner or launch. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-windows-host.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-WH-122` | `F-WH-11`, `F-WH-12`, `P-WH-2`, `DS-WH-5`, `P-WH-6` | tampering | A mismatched checkpoint or workspace content restores inconsistent or hostile state. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-windows-host.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-WH-123` | `F-WH-11`, `F-WH-12`, `P-WH-2`, `DS-WH-5`, `P-WH-6` | repudiation | Rewind hides an earlier external effect or presents it as undone. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-windows-host.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-WH-124` | `F-WH-11`, `F-WH-12`, `P-WH-2`, `DS-WH-5`, `P-WH-6` | information disclosure | Checkpoint transfer exposes private session or workspace data to an unauthorized reader. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-windows-host.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-WH-125` | `F-WH-11`, `F-WH-12`, `P-WH-2`, `DS-WH-5`, `P-WH-6` | denial of service | Retained captures or uncertain ownership prevent safe collection and consume capacity. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-windows-host.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |
| `TM-WH-126` | `F-WH-11`, `F-WH-12`, `P-WH-2`, `DS-WH-5`, `P-WH-6` | elevation of privilege | Restoration carries old identity, executable configuration, or authority into a successor launch. | [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Resume behavior remains unresolved. [Provider capture](agent-sandbox-windows-host.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)) |

### Commands in the agent workspace

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WH-40` | `F-WH-13`, `F-WH-14`, `F-WH-15`, `E-WH-1`, `P-WH-7`, `DS-WH-1`, `P-WH-4` | spoofing | A workspace command uses the authority of a different agent launch. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) Standalone command confinement doesn't specify launch identity binding. [Provider startup](agent-sandbox-windows-host.md#Process%20placement%20and%20startup). See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WH-41` | `F-WH-13`, `F-WH-14`, `F-WH-15`, `E-WH-1`, `P-WH-7`, `DS-WH-1`, `P-WH-4` | tampering | A command races capture or changes files outside the selected workspace through a stale path. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Command placement and capture ordering remain unresolved. [Provider startup](agent-sandbox-windows-host.md#Process%20placement%20and%20startup). See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WH-42` | `F-WH-13`, `F-WH-14`, `F-WH-15`, `E-WH-1`, `P-WH-7`, `DS-WH-1`, `P-WH-4` | repudiation | A command runs without a record binding its caller, launch, and confinement settings. | [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) Command attribution and lifecycle records remain unspecified. [Provider observations](agent-sandbox-windows-host.md#Host%20observations). See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WH-43` | `F-WH-13`, `F-WH-14`, `F-WH-15`, `E-WH-1`, `P-WH-7`, `DS-WH-1`, `P-WH-4` | information disclosure | A command inherits another launch's environment, session data, or readable workspace paths. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) Execution identity, inherited state, and path ownership remain unresolved. [Provider startup](agent-sandbox-windows-host.md#Process%20placement%20and%20startup). See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WH-44` | `F-WH-13`, `F-WH-14`, `F-WH-15`, `E-WH-1`, `P-WH-7`, `DS-WH-1`, `P-WH-4` | denial of service | A command continues running after cancellation or blocks workspace capture and teardown. | [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) Independent command lifetime and cancellation remain unspecified. [Provider teardown](agent-sandbox-windows-host.md#Cancellation%20and%20teardown). See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WH-45` | `F-WH-13`, `F-WH-14`, `F-WH-15`, `E-WH-1`, `P-WH-7`, `DS-WH-1`, `P-WH-4` | elevation of privilege | A command starts outside the selected provider confinement or obtains a relaxation without review. | [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) afakt-sandbox confines standalone commands. Authority for an unsandboxed invocation remains unresolved. [Provider startup](agent-sandbox-windows-host.md#Process%20placement%20and%20startup). See [Per-launch services](agent-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60)) |

### Host traffic capture

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WH-155` | `F-WH-4`, `F-WH-16`, `F-WH-17`, `P-WH-1`, `P-WH-2`, `P-WH-8`, `P-WH-4` | spoofing | A local process registers its endpoint as another launch’s mediator. | [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution) Registration authentication and current-launch binding remain unresolved. See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)) |
| `TM-WH-156` | `F-WH-4`, `F-WH-16`, `F-WH-17`, `P-WH-1`, `P-WH-2`, `P-WH-8`, `P-WH-4` | tampering | A process changes a capture mapping or original destination to redirect an admitted flow. | [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution) Capture mapping and destination-record integrity remain unresolved. See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)) |
| `TM-WH-157` | `F-WH-4`, `F-WH-16`, `F-WH-17`, `P-WH-1`, `P-WH-2`, `P-WH-8`, `P-WH-4` | repudiation | A capture component denies a mapping change or attributes a flow to a retired launch. | [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution) Registration attribution and retention of launch-generation evidence remain unresolved. See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)) |
| `TM-WH-158` | `F-WH-4`, `F-WH-16`, `F-WH-17`, `P-WH-1`, `P-WH-2`, `P-WH-8`, `P-WH-4` | information disclosure | A stale mapping delivers one launch’s captured traffic to another launch. | [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution) Identity reuse and existing-flow retirement remain unresolved. See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)) |
| `TM-WH-159` | `F-WH-4`, `F-WH-16`, `F-WH-17`, `P-WH-1`, `P-WH-2`, `P-WH-8`, `P-WH-4` | denial of service | A failed or flooded capture component prevents useful traffic or leaves teardown waiting. | [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution) Capture recovery, resource limits, and existing-flow shutdown remain unresolved. See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)) |
| `TM-WH-160` | `F-WH-4`, `F-WH-16`, `F-WH-17`, `P-WH-1`, `P-WH-2`, `P-WH-8`, `P-WH-4` | elevation of privilege | Hostile registration or traffic input exploits the capture component and reaches host management authority. | [Provider mediation](agent-sandbox-windows-host.md#Mediator%20connection%20and%20name%20resolution) Containment of capture management and privileged components remains unresolved. See [Per-launch services](agent-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)) |

### Per-launch network boundary

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WH-161` | `F-WH-18`, `P-WH-4`, `P-WH-9` | spoofing | A process substitutes another launch's local agent-proxy endpoint. | [Per-launch services](agent-proxy.md#Launch%20boundary) requires a local channel bound to the launch. Its peer authentication remains undecided. | depends on an open question ([`OQ-HUB-14`](agent-open-questions.md#%60OQ-HUB-14%60)) |
| `TM-WH-162` | `F-WH-18`, `P-WH-4`, `P-WH-9` | tampering | A malformed hub frame changes the caller or target associated with a stream. | [Hub connection](agent-proxy.md#Hub%20connection) requires caller and target context on each stream. Framing validation remains undecided. | depends on an open question ([`OQ-HUB-14`](agent-open-questions.md#%60OQ-HUB-14%60)) |
| `TM-WH-163` | `F-WH-18`, `P-WH-4`, `P-WH-9` | repudiation | Reconnect loses the association between an admitted stream and its authorization record. | [Remote control and sessions](agent-acp.md#ACP%20through%20the%20agent%20hub) keeps message history. Reconnect attribution needs the local sender binding. | depends on an open question ([`OQ-HUB-14`](agent-open-questions.md#%60OQ-HUB-14%60)) |
| `TM-WH-164` | `F-WH-18`, `P-WH-4`, `P-WH-9` | information disclosure | Connection replacement sends a stream to a different launch. | [Hub connection](agent-proxy.md#Hub%20connection) uses a separate authenticated launch connection. Replacement must also preserve local stream ownership. | depends on an open question ([`OQ-HUB-14`](agent-open-questions.md#%60OQ-HUB-14%60)) |
| `TM-WH-165` | `F-WH-18`, `P-WH-4`, `P-WH-9` | denial of service | One launch consumes hub buffers or blocks control streams with published-port traffic. | [Hub connection](agent-proxy.md#Hub%20connection) requires bounded admission. The local scheduling and buffer limits remain undecided. | depends on an open question ([`OQ-HUB-14`](agent-open-questions.md#%60OQ-HUB-14%60)) |
| `TM-WH-166` | `F-WH-18`, `P-WH-4`, `P-WH-9` | elevation of privilege | After compromising the agent proxy, an attacker sends traffic without mediator approval. | [Per-launch services](agent-proxy.md#Launch%20boundary) separates policy from network authority. Their authenticated local binding remains undecided. | depends on an open question ([`OQ-HUB-14`](agent-open-questions.md#%60OQ-HUB-14%60)) |

## Attack trees

`OR` requires any child. `AND` requires every child. Leaves reference local threats. Shared policy failures use the [common attack trees](agent-threat-model.md#Attack%20trees).

### Escape to the agent host

```mermaid
flowchart TD
    root["Escape to the agent host"] --> choice{"OR"}
    choice --> leaf0["TM-WH-24: Exploit the isolation mechanism"]
    choice --> leaf1["TM-WH-36: Make host processing execute workspace metadata"]
    choice --> leaf2["TM-WH-30: Abuse an inherited control handle"]
    choice --> leaf3["TM-WH-45: Run a workspace command outside intended confinement"]
```

- Cheapest path and blocking controls. Trigger host processing of hostile workspace metadata without a guest kernel exploit. [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60) leaves the host-side consumer unprotected. The [platform attack surface](#Host%20visibility%20and%20workspace%20commands) has separate confinement gaps.

### Read or change another agent launch's workspace or session

```mermaid
flowchart TD
    root["Read or change another agent launch's workspace or session"] --> choice{"OR"}
    choice --> leaf0["TM-WH-130: Read a shared cache"]
    choice --> leaf1["TM-WH-128: Change a cache consumed by another launch"]
    choice --> leaf2["TM-WH-22: Cross workspace or scratch ownership"]
    choice --> leaf3["TM-WH-28: Use an incorrectly bound local endpoint"]
    choice --> leaf4["TM-WH-124: Read retained capture"]
```

- Cheapest path and blocking controls. Use a shared named tool cache already exposed to the launch. Cache authorization and provenance remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Workspace ownership and local endpoint binding require separate controls.

### Send traffic past egress policy

```mermaid
flowchart TD
    root["Send traffic past egress policy"] --> choice{"OR"}
    choice --> leaf0["TM-WH-26: Use a direct network route"]
    choice --> leaf1["TM-WH-25: Select another mediator identity"]
    choice --> leaf2["TM-WH-45: Run a workspace command outside confinement"]
    choice --> leaf3["TM-WH-30: Use a privileged control handle"]
```

- Cheapest path and blocking controls. Use a network path whose mediator routing is unspecified. [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60) covers missing provider routing. [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60) covers sender binding. The [common egress tree](agent-threat-model.md#Send%20traffic%20past%20egress%20policy) covers permitted name queries, passthrough, and relay proxy use.

### Persist hostile state across agent launches

```mermaid
flowchart TD
    root["Persist hostile state across agent launches"] --> choice{"OR"}
    choice --> leaf0["TM-WH-122: Restore mismatched or hostile captured state"]
    choice --> leaf1["TM-WH-126: Carry old authority into a successor launch"]
    choice --> leaf2["TM-WH-132: Execute a shared cache entry under later authority"]
```

- Cheapest path and blocking controls. Leave executable content in a shared named tool cache for a later launch. Cache provenance and invalidation remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Captured state needs a separate binding under [`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60), while authority carried into a successor launch remains [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60). The [common persistence tree](agent-threat-model.md#Persist%20hostile%20state%20across%20agent%20launches) covers workspace images and hooks.

## Residual risks and gaps

Host isolation remains on hold. The [common residual risks](agent-threat-model.md#Residual%20risks%20and%20gaps) keep hook protection, repository configuration scope, repository profile self-grant and credential domain bounds, and session-end wait with workload identity lifetime. No pending option supplies a mitigation.

### Isolation and lifecycle

- `TM-WH-1`. A local process substitutes for the provider control endpoint or launch process. depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)).
- `TM-WH-2`. A launch changes inherited handles, mounts, descriptors, or confinement inputs to reach provider authority. depends on an open question ([`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60)).
- `TM-WH-21`. A compromised agent host invents process exits or hides local workspace and command effects. gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)).
- `TM-WH-70`. A same-user host process reads mediator memory, local session traffic, or another launch's files. depends on an open question ([`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60)).
- `TM-WH-23`. A hostile process tree ignores cancellation and continues running. gap ([`OQ-SBX-57`](agent-open-questions.md#%60OQ-SBX-57%60)).
- `TM-WH-24`. An agent harness exploits the shared host kernel or escapes process confinement. gap ([`OQ-SBX-55`](agent-open-questions.md#%60OQ-SBX-55%60)).
- `TM-WH-154`. An orphaned process or reused launch user retains local authority after worker failure. depends on an open question ([`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60)).

### Local transport and egress

- `TM-WH-25`. A process selects another launch's mediator endpoint and receives its authority. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-WH-26`. A process changes proxy settings or uses an interface that bypasses the agent mediator. gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)).
- `TM-WH-27`. Unbound local traffic acquires another launch's attribution before shared logging. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-WH-28`. A control carrier or network view exposes another launch's session or loopback listener. gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)).
- `TM-WH-29`. A launch floods its local control carrier and prevents lifecycle or session progress. depends on an open question ([`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60)).
- `TM-WH-30`. A local process uses an inherited control handle for host operations outside its launch authority. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-WH-67`. A process impersonates the local control endpoint or reuses its previous launch identity. depends on an open question ([`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60)).
- `TM-WH-82`. A host-mode launch reads another launch's loopback listener through a shared network view. depends on an open question ([`OQ-SBX-16`](agent-open-questions.md#%60OQ-SBX-16%60)).

### Workspace and scratch storage

- `TM-WH-19`. A same-user process substitutes a workspace owner or capture helper. depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)).
- `TM-WH-32`. Hostile workspace metadata redirects host capture or cleanup outside the selected workspace. gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).
- `TM-WH-33`. A workspace or scratch writer denies changes made through a shared local identity. gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)).
- `TM-WH-22`. Capture, scratch reuse, or path aliasing exposes another launch's files. depends on an open question ([`OQ-WSP-15`](agent-open-questions.md#%60OQ-WSP-15%60)).
- `TM-WH-35`. Workspace or scratch growth exhausts host capacity outside an enforced allocation. gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-WH-36`. Host-side Git processing runs repository hooks or configuration or follows alternates or submodules or accepts hostile worktree paths. gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).

### Shared named tool cache

- `TM-WH-127`. One launch places cache entries that another mistakes for trusted tool output. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WH-128`. A launch changes shared cache content consumed by another launch. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WH-129`. A cache writer denies modifying content because cache provenance has no stated record. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WH-130`. A launch reads private data another launch left in a shared tool cache. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WH-131`. A launch fills or churns a shared cache and denies other launches useful capacity. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WH-132`. A cached executable or build result runs under a later launch's broader authority. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).

### Agent checkpoints

- `TM-WH-121`. A capture or restore request substitutes another agent's owner or launch. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-WH-122`. A mismatched checkpoint or workspace content restores inconsistent or hostile state. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-WH-123`. Rewind hides an earlier external effect or presents it as undone. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-WH-124`. Checkpoint transfer exposes private session or workspace data to an unauthorized reader. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-WH-125`. Retained captures or uncertain ownership prevent safe collection and consume capacity. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).
- `TM-WH-126`. Restoration carries old identity, executable configuration, or authority into a successor launch. depends on an open question ([`OQ-AGT-28`](agent-open-questions.md#%60OQ-AGT-28%60)).

### Commands in the agent workspace

- `TM-WH-40`. A workspace command uses the authority of a different agent launch. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WH-41`. A command races capture or changes files outside the selected workspace through a stale path. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WH-42`. A command runs without a record binding its caller, launch, and confinement settings. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WH-43`. A command inherits another launch's environment, session data, or readable workspace paths. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WH-44`. A command continues running after cancellation or blocks workspace capture and teardown. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WH-45`. A command starts outside the selected provider confinement or obtains a relaxation without review. depends on an open question ([`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60)).

### Host traffic capture

- `TM-WH-155`. A local process registers its endpoint as another launch’s mediator. gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)).
- `TM-WH-156`. A process changes a capture mapping or original destination to redirect an admitted flow. gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)).
- `TM-WH-157`. A capture component denies a mapping change or attributes a flow to a retired launch. gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)).
- `TM-WH-158`. A stale mapping delivers one launch’s captured traffic to another launch. gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)).
- `TM-WH-159`. A failed or flooded capture component prevents useful traffic or leaves teardown waiting. gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)).
- `TM-WH-160`. Hostile registration or traffic input exploits the capture component and reaches host management authority. gap ([`OQ-SBX-59`](agent-open-questions.md#%60OQ-SBX-59%60)).
