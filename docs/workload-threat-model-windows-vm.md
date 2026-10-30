---
type: Threat Model
title: "Workload threat model for Windows VM"
description: Isolation, local transport, retained state, and platform threats for the windows-vm workload provider.
tags: [afakt, agents, security, threat-model]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload threat model for Windows VM

[Public key infrastructure](pki.md).

## Scope

The windows-vm workload provider uses the [provider design](workload-provider-windows-vm.md). The [common workload threat model](workload-threat-model.md) covers shared service authority, credentials, network policy, sessions, registries, and audit records.

[Platforms and privileges](workload-provider-windows-vm.md#Platforms%20and%20privileges) and [Provider isolation](workload-provider-windows-vm.md#Isolation%20boundary) define the platform and its kernel boundary. [Provider startup](workload-provider-windows-vm.md#Process%20placement%20and%20startup) places afakt-agent, afakt-sandbox, and the harness.

STRIDE covers spoofing, tampering, repudiation, information disclosure, denial of service, and elevation of privilege. A virtual machine (VM) supplies a separate guest kernel where the provider uses one. Agent Client Protocol (ACP) transports session messages. Domain Name System (DNS) queries use the workload proxy mediator. A User Identifier (UID) names an operating-system user. An Application Programming Interface (API) provides shared service authority. Hyper-V runs the Linux guest VM directly.

### Workload proxy mediator access

The Linux workload guest has device and management boundaries distinct from its Windows workload host. Forced proxy routing and local carrier authentication remain unresolved. See [Workload provider mediation](workload-provider-windows-vm.md#Mediator%20connection%20and%20name%20resolution).

### Workspace, scratch, and caches

[Workload provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches) distinguishes launch state, shared named tool caches, and verified workspace image caches. Provider storage mappings, scratch cleanup, shared cache trust, and hostile host-side repository processing remain unresolved. [Provider materialization](workload-provider-windows-vm.md#Workspace%20image%20materialization) separates verified extraction from later workspace capture.

### Workload snapshots and agent checkpoints

[Workload provider capture](workload-provider-windows-vm.md#Workload%20snapshots%20and%20agent%20checkpoints) defines the capture boundary and unresolved restore contract. Machine state and native harness state need a consistent capture. Both leave external effects unchanged.

### Host visibility and workspace commands

Host custody and its limits follow [Workload provider observations](workload-provider-windows-vm.md#Host%20observations). Termination authority and unresolved descendant cleanup follow [Provider teardown](workload-provider-windows-vm.md#Cancellation%20and%20teardown). Same-user process separation and the isolation baseline for workload hosts remain on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

Commands started by `afakt workload exec` use standalone afakt-sandbox. [Provider startup](workload-provider-windows-vm.md#Process%20placement%20and%20startup) leaves command placement, inherited state, and launch authority under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). An unsandboxed invocation remains under [`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60).

## Trust boundaries

| Boundary ID | Boundary | Two sides | Enforcement in the design |
| --- | --- | --- | --- |
| `TB-WVM-1` | Agent harness and workload provider control | untrusted process tree / provider lifecycle and platform authority | [Enforcement](workload-sandbox.md#Enforcement) defines the isolation tier. [Host visibility and workspace commands](#Host%20visibility%20and%20workspace%20commands) identifies platform exposure and missing confinement. |
| `TB-WVM-2` | Local transport and workload proxy mediator | Launch network and control endpoints / mediator identity and policy | [Workload proxy mediator](workload-proxy-mediator.md#Launch%20boundary) Local endpoint binding remains [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60). |
| `TB-WVM-3` | Agent workspace and host processing | Launch-controlled files / worker capture, materialization, and cleanup | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Host-side handling of repository metadata remains [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). |
| `TB-WVM-4` | Shared named tool cache | One launch / cache content consumed by another launch | [Sandbox variables](workload-sandbox.md#Sandbox%20variables) Access, provenance, and capacity controls remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). |
| `TB-WVM-5` | Captured state and successor launch | Current launch / retained capture and later launch | [Suspend and resume](workload-lifecycle.md#Suspend%20and%20resume) State authorization and ownership remain [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). |
| `TB-WVM-6` | Workspace command and launch authority | Authorized request for a command / command confinement and workspace access | [Host process sandbox](workload-sandbox.md#Host%20process%20sandbox) Command placement and lifecycle binding remain [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). |

## Data flow diagram

### Element inventory

| Element ID | Element | Location or custody |
| --- | --- | --- |
| `E-WVM-1` | Command caller | Authenticated client. Shared service grants apply |
| `E-WVM-2` | Another agent launch | Separate launch authority |
| `P-WVM-1` | Agent harness and afakt-sandbox | Guest |
| `P-WVM-2` | Workload host worker and workload provider lifecycle | Workload host |
| `P-WVM-3` | Host workspace processor | Worker capture and cleanup |
| `P-WVM-9` | Workload proxy | Outside guest confinement. Launch identity custody |
| `P-WVM-4` | Workload proxy mediator | Outside launch confinement. Process separation unresolved |
| `P-WVM-5` | Guest relay | Separate user in the Linux guest |
| `P-WVM-6` | Successor launch | New launch authority |
| `P-WVM-7` | `afakt workload exec` | Execution placement unspecified |
| `DS-WVM-1` | Agent workspace | Per-launch writable content |
| `DS-WVM-2` | Scratch storage | Outside workspace. Backing unspecified |
| `DS-WVM-3` | Shared named tool cache | Shared across launches. Backing unspecified |
| `DS-WVM-4` | Verified workspace image cache | Workload host. Distinct from named tool cache |
| `DS-WVM-5` | Snapshot and checkpoint capture | Retained state. Access controls unresolved |
| `P-WVM-8` | Windows VM management adapter | Linux guest VM control through Hyper-V. Interface, authority, and stop observations unresolved |

### Level one workload provider flows

Inside the Linux workload guest, the agent harness uses storage managed by the Windows workload host. Arrows identify exchanges and their return paths. Unspecified carriers and placements remain gaps. The shared named tool cache is outside individual launch confinement and has a separate crossing from another launch. For the verified workspace image cache, the trust relationship differs.

```mermaid
flowchart LR
    P-WVM-9["P-WVM-9: Workload proxy"]
    subgraph z0["TB-WVM-1 / launch confinement"]
        P-WVM-1["P-WVM-1: Agent harness and afakt-sandbox"]
    end
    subgraph z1["TB-WVM-2 / local control"]
        P-WVM-5["P-WVM-5: Guest relay"]
    end
    subgraph z2["TB-WVM-2 / workload host policy"]
        P-WVM-4["P-WVM-4: Workload proxy mediator"]
    end
    subgraph z3["TB-WVM-3 / launch-controlled storage"]
        DS-WVM-1["DS-WVM-1: Agent workspace"]
        DS-WVM-2["DS-WVM-2: Scratch storage"]
    end
    subgraph z4["TB-WVM-3 / workload host processing"]
        P-WVM-2["P-WVM-2: Workload host worker and provider lifecycle"]
        P-WVM-3["P-WVM-3: Host workspace processor"]
        DS-WVM-4["DS-WVM-4: Verified workspace image cache"]
    end
    subgraph z5["TB-WVM-4 / shared across launches"]
        DS-WVM-3["DS-WVM-3: Shared named tool cache"]
    end
    subgraph z6["TB-WVM-4 / other launch"]
        E-WVM-2["E-WVM-2: Another agent launch"]
    end
    subgraph z7["TB-WVM-5 / retained capture"]
        DS-WVM-5["DS-WVM-5: Snapshot and checkpoint capture"]
    end
    subgraph z8["TB-WVM-5 / successor"]
        P-WVM-6["P-WVM-6: Successor launch"]
    end
    subgraph z9["TB-WVM-6 / caller"]
        E-WVM-1["E-WVM-1: Command caller"]
    end
    subgraph z10["TB-WVM-6 / command placement unresolved"]
        P-WVM-7["P-WVM-7: `afakt workload exec`"]
    end
    P-WVM-8["P-WVM-8: Windows VM management adapter"]
    P-WVM-2 -->|"F-WVM-16"| P-WVM-8
    P-WVM-8 -->|"F-WVM-1"| P-WVM-1
    P-WVM-1 -->|"F-WVM-2"| P-WVM-5
    P-WVM-5 -->|"F-WVM-3"| P-WVM-9
    P-WVM-1 -->|"F-WVM-4"| P-WVM-9
    P-WVM-1 -->|"F-WVM-5"| DS-WVM-1
    P-WVM-1 -->|"F-WVM-6"| DS-WVM-2
    P-WVM-3 -->|"F-WVM-7"| DS-WVM-1
    DS-WVM-4 -->|"F-WVM-8"| P-WVM-3
    P-WVM-1 -->|"F-WVM-9"| DS-WVM-3
    E-WVM-2 -->|"F-WVM-10"| DS-WVM-3
    P-WVM-2 -->|"F-WVM-11"| DS-WVM-5
    DS-WVM-5 -->|"F-WVM-12"| P-WVM-6
    E-WVM-1 -->|"F-WVM-13"| P-WVM-7
    P-WVM-7 -->|"F-WVM-14"| DS-WVM-1
    P-WVM-7 -->|"F-WVM-15"| P-WVM-9
```

- **`F-WVM-1`.** Startup, stop, and observed exit of the process tree. Crosses `TB-WVM-1`.
- **`F-WVM-2`.** Session and local control messages. Endpoint binding remains unresolved. Crosses `TB-WVM-2`.
- **`F-WVM-3`.** Control messages and preview streams over the workload provider transport. The carrier for each guest platform remains unspecified. Crosses `TB-WVM-2`.
- **`F-WVM-4`.** Network packets and name queries through the workload proxy under mediator decisions. Network and control traffic needs a bound mediator path. Its provider-specific mechanism remains unspecified. Crosses `TB-WVM-2`.
- **`F-WVM-5`.** Workspace reads and writes under launch confinement. Crosses `TB-WVM-3`.
- **`F-WVM-6`.** Scratch reads and writes outside the workspace. Allocation and cleanup are unspecified. Crosses `TB-WVM-3`.
- **`F-WVM-7`.** Host-side capture, metadata interpretation, and cleanup of hostile workspace content. Crosses `TB-WVM-3`.
- **`F-WVM-8`.** Verified immutable workspace image content used for materialization. Cache ownership is provider-specific. Crosses `TB-WVM-3`.
- **`F-WVM-9`.** Named cache reads and writes by this launch. Mapping to backing storage is unspecified. Crosses `TB-WVM-4`.
- **`F-WVM-10`.** Named cache reads and writes by another launch across the same sharing boundary. Crosses `TB-WVM-4`.
- **`F-WVM-11`.** Capture and restore of VM state, workspace delta, and harness checkpoint at one idle boundary. Crosses `TB-WVM-5`.
- **`F-WVM-12`.** State admitted into a successor launch under new authority. Crosses `TB-WVM-5`.
- **`F-WVM-13`.** Authorized workspace command request. Local placement and credential binding remain unresolved. Crosses `TB-WVM-6`.
- **`F-WVM-14`.** Command reads and writes in the selected agent workspace. Concurrent access remains unresolved. Crosses `TB-WVM-6`, `TB-WVM-3`.
- **`F-WVM-15`.** Command traffic requiring launch egress policy. Transport and identity binding remain unresolved. Crosses `TB-WVM-6`, `TB-WVM-2`.
- **`F-WVM-16`.** Windows worker management calls use the Hyper-V interface that controls the Linux guest VM. The interface, authentication, and independent stop observations remain under [`OQ-SBX-60`](agent-open-questions.md#%60OQ-SBX-60%60). Crosses `TB-WVM-1`.

## Workload proxy

Each workload has one workload proxy mediator and one workload proxy. The proxy transmits traffic in both directions and requests a mediator decision for every connection. The mediator never transmits traffic. Direct overlay peers authenticate with SPIFFE mutual Transport Layer Security (TLS), and both workload mediators check the connection. Service users and workload provider integration remain undecided under [Launch boundary](workload-proxy.md#Launch%20boundary).

```mermaid
flowchart LR
    P-WVM-4["P-WVM-4: Workload proxy mediator"]
    P-WVM-9["P-WVM-9: Workload proxy"]
    P-WVM-4 -->|"F-WVM-17"| P-WVM-9
```

- **`F-WVM-17`.** Connection decision requests, launch identity, and decisions across the mediator/workload-proxy boundary. The local channel needs an authenticated peer and workload context.

## STRIDE analysis

Each table covers the six categories for its listed flows and every participating endpoint. Shared service checks are in the [common STRIDE analysis](workload-threat-model.md#STRIDE%20analysis). They don't establish workload provider isolation. A mitigation applies only to the stated attack.

### Isolation and lifecycle

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WVM-1` | `F-WVM-1`, `P-WVM-2`, `P-WVM-1`, `F-WVM-16`, `P-WVM-8` | spoofing | A local process substitutes for the workload provider's control endpoint or launch process. | [Host process sandbox](workload-sandbox.md#Host%20process%20sandbox) [Runtime privileges](workload-sandbox.md#Runtime%20privileges) Per-launch identities and helper authority remain unresolved. [Provider startup](workload-provider-windows-vm.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)) |
| `TM-WVM-2` | `F-WVM-1`, `P-WVM-2`, `P-WVM-1`, `F-WVM-16`, `P-WVM-8` | tampering | A launch changes inherited handles, mounts, descriptors, or confinement inputs to reach provider authority. | [Enforcement](workload-sandbox.md#Enforcement). [Platform exposure](#Host%20visibility%20and%20workspace%20commands). Unspecified. [Provider isolation](workload-provider-windows-vm.md#Isolation%20boundary). | depends on an open question ([`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60)) |
| `TM-WVM-21` | `F-WVM-1`, `P-WVM-2`, `P-WVM-1`, `F-WVM-16`, `P-WVM-8` | repudiation | A compromised workload host invents process exits or hides local workspace and command effects. | [Provider observations](workload-lifecycle.md#Workload%20provider%20observations) Independent verification of a compromised host is unspecified. [Provider observations](workload-provider-windows-vm.md#Host%20observations). | gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)) |
| `TM-WVM-70` | `F-WVM-1`, `P-WVM-2`, `P-WVM-1`, `F-WVM-16`, `P-WVM-8` | information disclosure | A same-user host process reads mediator memory, local session traffic, or another launch's files. | [Trust boundary](workload-host.md#Trust%20boundary) The design trusts host custody. Same-user process separation remains on hold. [Provider observations](workload-provider-windows-vm.md#Host%20observations). | depends on an open question ([`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60)) |
| `TM-WVM-23` | `F-WVM-1`, `P-WVM-2`, `P-WVM-1`, `F-WVM-16`, `P-WVM-8` | denial of service | A hostile process tree ignores cancellation and continues running. | [Provider teardown](workload-provider-windows-vm.md#Cancellation%20and%20teardown) requires independent stop observations for the Linux workload guest. The management adapters and termination proof remain unresolved. | depends on an open question ([`OQ-SBX-60`](agent-open-questions.md#%60OQ-SBX-60%60)) |
| `TM-WVM-151` | `F-WVM-1`, `P-WVM-2`, `P-WVM-1`, `F-WVM-16`, `P-WVM-8` | elevation of privilege | A guest exploits a virtual device or the guest hypervisor to gain workload host authority. | [Enforcement](workload-sandbox.md#Enforcement). [Platform exposure](#Host%20visibility%20and%20workspace%20commands). Unspecified. [Provider isolation](workload-provider-windows-vm.md#Isolation%20boundary). | gap ([`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60)) |
| `TM-WVM-154` | `F-WVM-1`, `P-WVM-2`, `P-WVM-1`, `F-WVM-16`, `P-WVM-8` | elevation of privilege | An orphaned process or reused launch user retains local authority after worker failure. | [Restart and recovery](workload-host.md#Restart%20and%20recovery) Durable ownership and safe identity reuse remain unresolved. [Provider teardown](workload-provider-windows-vm.md#Cancellation%20and%20teardown). | depends on an open question ([`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60)) |

### Local transport and egress

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WVM-25` | `F-WVM-2`, `F-WVM-3`, `F-WVM-4`, `P-WVM-1`, `P-WVM-5`, `P-WVM-4` | spoofing | A process selects another launch's mediator endpoint and receives its authority. | [Workload proxy mediator](workload-proxy-mediator.md#Launch%20boundary) Endpoint authentication remains unresolved. [Workload provider mediation](workload-provider-windows-vm.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-WVM-26` | `F-WVM-2`, `F-WVM-3`, `F-WVM-4`, `P-WVM-1`, `P-WVM-5`, `P-WVM-4` | tampering | A process changes proxy settings or uses an interface that bypasses the workload proxy mediator. | [Workload proxy mediator](workload-proxy-mediator.md#Launch%20boundary) Network and control traffic needs a bound mediator path. Its provider-specific mechanism remains unspecified. [Provider mediation](workload-provider-windows-vm.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](workload-proxy.md#Launch%20boundary). | gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)) |
| `TM-WVM-27` | `F-WVM-2`, `F-WVM-3`, `F-WVM-4`, `P-WVM-1`, `P-WVM-5`, `P-WVM-4` | repudiation | Unbound local traffic acquires another launch's attribution before shared logging. | [Control requests](workload-proxy-mediator.md#Control%20requests) Stamping identity requires a trustworthy local sender binding. [Provider mediation](workload-provider-windows-vm.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-WVM-28` | `F-WVM-2`, `F-WVM-3`, `F-WVM-4`, `P-WVM-1`, `P-WVM-5`, `P-WVM-4` | information disclosure | A control carrier or network view exposes another launch's session or loopback listener. | [Loopback listeners](workload-ports.md#Loopback%20listeners) The design requires separate network views. The workload provider binding remains unresolved. [Provider mediation](workload-provider-windows-vm.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60)) |
| `TM-WVM-29` | `F-WVM-2`, `F-WVM-3`, `F-WVM-4`, `P-WVM-1`, `P-WVM-5`, `P-WVM-4` | denial of service | A launch floods its local control carrier and prevents lifecycle or session progress. | Local carrier bounds and sharing remain unresolved under [mediator overload control](agent-open-questions.md#%60OQ-SBX-14%60). [Provider mediation](workload-provider-windows-vm.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60)) |
| `TM-WVM-30` | `F-WVM-2`, `F-WVM-3`, `F-WVM-4`, `P-WVM-1`, `P-WVM-5`, `P-WVM-4` | elevation of privilege | A local process uses an inherited control handle for workload host operations outside its launch authority. | [Guest layout and lifecycle control](workload-sandbox.md#Guest%20layout%20and%20lifecycle%20control) separates relay and harness users. Handle ownership and launch binding remain unresolved. See [common mediator threats](workload-threat-model.md#Admission%20of%20requests%20by%20the%20workload%20proxy%20mediator) for protocol parser attacks. [Provider mediation](workload-provider-windows-vm.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-WVM-67` | `F-WVM-2`, `F-WVM-3`, `P-WVM-1`, `P-WVM-5`, `P-WVM-4` | spoofing | A process impersonates the guest relay or reuses a previous launch's control endpoint. | [Guest layout and lifecycle control](workload-sandbox.md#Guest%20layout%20and%20lifecycle%20control) gives the Linux guest relay a separate user. Endpoint custody and carrier authentication remain unresolved. [Provider mediation](workload-provider-windows-vm.md#Mediator%20connection%20and%20name%20resolution). See [Per-launch services](workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60)) |

### Workspace and scratch storage

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WVM-19` | `F-WVM-5`, `F-WVM-6`, `F-WVM-7`, `F-WVM-8`, `P-WVM-1`, `DS-WVM-1`, `DS-WVM-2`, `P-WVM-3`, `DS-WVM-4` | spoofing | A same-user process substitutes a workspace owner or capture helper. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Ownership and helper rights depend on launch identity. [Workload provider startup](workload-provider-windows-vm.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)) |
| `TM-WVM-32` | `F-WVM-5`, `F-WVM-6`, `F-WVM-7`, `F-WVM-8`, `P-WVM-1`, `DS-WVM-1`, `DS-WVM-2`, `P-WVM-3`, `DS-WVM-4` | tampering | Hostile workspace metadata redirects workload host capture or cleanup outside the selected workspace. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Content capture supplies no safe metadata interpretation contract. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |
| `TM-WVM-33` | `F-WVM-5`, `F-WVM-6`, `F-WVM-7`, `F-WVM-8`, `P-WVM-1`, `DS-WVM-1`, `DS-WVM-2`, `P-WVM-3`, `DS-WVM-4` | repudiation | A workspace or scratch writer denies changes made through a shared local identity. | [Provider observations](workload-lifecycle.md#Workload%20provider%20observations) Host compromise has no independent evidence mechanism. [Provider observations](workload-provider-windows-vm.md#Host%20observations). | gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)) |
| `TM-WVM-22` | `F-WVM-5`, `F-WVM-6`, `F-WVM-7`, `F-WVM-8`, `P-WVM-1`, `DS-WVM-1`, `DS-WVM-2`, `P-WVM-3`, `DS-WVM-4` | information disclosure | Capture, scratch reuse, or path aliasing exposes another launch's files. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Capture rights and storage placement remain unresolved. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-WVM-35` | `F-WVM-5`, `F-WVM-6`, `F-WVM-7`, `F-WVM-8`, `P-WVM-1`, `DS-WVM-1`, `DS-WVM-2`, `P-WVM-3`, `DS-WVM-4` | denial of service | Workspace or scratch growth exhausts host capacity outside an enforced allocation. | [Sandbox variables](workload-sandbox.md#Sandbox%20variables) Scratch is outside the workspace, but storage quotas and cleanup are unspecified. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-WVM-36` | `F-WVM-5`, `F-WVM-6`, `F-WVM-7`, `F-WVM-8`, `P-WVM-1`, `DS-WVM-1`, `DS-WVM-2`, `P-WVM-3`, `DS-WVM-4` | elevation of privilege | Host-side Git processing runs repository hooks or configuration or follows alternates or submodules or accepts hostile worktree paths. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Archive extraction checks don't protect later Git or filesystem consumers. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |

### Shared named tool cache

Cache sharing has no stated authorization, provenance, capacity, or invalidation mechanism.

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WVM-127` | `F-WVM-9`, `F-WVM-10`, `P-WVM-1`, `DS-WVM-3`, `E-WVM-2` | spoofing | One launch places cache entries that another mistakes for trusted tool output. | [Sandbox variables](workload-sandbox.md#Sandbox%20variables). Unspecified. [Workload provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WVM-128` | `F-WVM-9`, `F-WVM-10`, `P-WVM-1`, `DS-WVM-3`, `E-WVM-2` | tampering | A launch changes shared cache content consumed by another launch. | [Sandbox variables](workload-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WVM-129` | `F-WVM-9`, `F-WVM-10`, `P-WVM-1`, `DS-WVM-3`, `E-WVM-2` | repudiation | A cache writer denies modifying content because cache provenance has no stated record. | [Sandbox variables](workload-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WVM-130` | `F-WVM-9`, `F-WVM-10`, `P-WVM-1`, `DS-WVM-3`, `E-WVM-2` | information disclosure | A launch reads private data another launch left in a shared tool cache. | [Sandbox variables](workload-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WVM-131` | `F-WVM-9`, `F-WVM-10`, `P-WVM-1`, `DS-WVM-3`, `E-WVM-2` | denial of service | A launch fills or churns a shared cache and denies other launches useful capacity. | [Sandbox variables](workload-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-WVM-132` | `F-WVM-9`, `F-WVM-10`, `P-WVM-1`, `DS-WVM-3`, `E-WVM-2` | elevation of privilege | A cached executable or build result runs under a later launch's broader authority. | [Sandbox variables](workload-sandbox.md#Sandbox%20variables). Unspecified. [Provider storage](workload-provider-windows-vm.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |

### Workload snapshots and agent checkpoints

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WVM-121` | `F-WVM-11`, `F-WVM-12`, `P-WVM-2`, `DS-WVM-5`, `P-WVM-6` | spoofing | A capture or restore request substitutes another agent's owner or launch. | [Workload snapshots and the snapshot registry](agent.md#Workload%20snapshots%20and%20the%20snapshot%20registry) defines stored VM state. Snapshot and checkpoint authorization remains unspecified. [Workload provider capture](workload-provider-windows-vm.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)) |
| `TM-WVM-122` | `F-WVM-11`, `F-WVM-12`, `P-WVM-2`, `DS-WVM-5`, `P-WVM-6` | tampering | A mismatched snapshot, checkpoint, or workspace delta restores inconsistent or hostile state. | [Suspend and resume](workload-lifecycle.md#Suspend%20and%20resume) freezes capture at an idle boundary. Exact state binding remains open. [Provider capture](workload-provider-windows-vm.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60)) |
| `TM-WVM-123` | `F-WVM-11`, `F-WVM-12`, `P-WVM-2`, `DS-WVM-5`, `P-WVM-6` | repudiation | Rewind hides an earlier external effect or presents it as undone. | [Agent checkpoints](agent.md#Agent%20checkpoints) retains the original branch. External-effect representation remains open. [Provider capture](workload-provider-windows-vm.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-4`](agent-open-questions.md#%60OQ-AGT-4%60)) |
| `TM-WVM-124` | `F-WVM-11`, `F-WVM-12`, `P-WVM-2`, `DS-WVM-5`, `P-WVM-6` | information disclosure | Snapshot transfer exposes private session or workspace data to an unauthorized reader. | [Suspend and resume](workload-lifecycle.md#Suspend%20and%20resume) excludes credentials from capture. Storage and transfer confidentiality remain unspecified. [Provider capture](workload-provider-windows-vm.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)) |
| `TM-WVM-125` | `F-WVM-11`, `F-WVM-12`, `P-WVM-2`, `DS-WVM-5`, `P-WVM-6` | denial of service | Retained captures or uncertain ownership prevent safe collection and consume capacity. | [Owned resources](agent.md#Owned%20resources) retains durable resources. Retention roots and late holds remain open. [Provider capture](workload-provider-windows-vm.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-11`](agent-open-questions.md#%60OQ-AGT-11%60)) |
| `TM-WVM-126` | `F-WVM-11`, `F-WVM-12`, `P-WVM-2`, `DS-WVM-5`, `P-WVM-6` | elevation of privilege | Restoration carries old identity, executable configuration, or authority into a successor launch. | [Suspend and resume](workload-lifecycle.md#Suspend%20and%20resume) reacquires credentials and rechecks channels. Full transfer scope remains open. [Provider capture](workload-provider-windows-vm.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)) |

### Commands in the agent workspace

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WVM-40` | `F-WVM-13`, `F-WVM-14`, `F-WVM-15`, `E-WVM-1`, `P-WVM-7`, `DS-WVM-1`, `P-WVM-4` | spoofing | A workspace command uses the authority of a different agent launch. | [Host process sandbox](workload-sandbox.md#Host%20process%20sandbox) Standalone command confinement doesn't specify launch identity binding. [Workload provider startup](workload-provider-windows-vm.md#Process%20placement%20and%20startup). See [Per-launch services](workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WVM-41` | `F-WVM-13`, `F-WVM-14`, `F-WVM-15`, `E-WVM-1`, `P-WVM-7`, `DS-WVM-1`, `P-WVM-4` | tampering | A command races capture or changes files outside the selected workspace through a stale path. | [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture) Command placement and capture ordering remain unresolved. [Provider startup](workload-provider-windows-vm.md#Process%20placement%20and%20startup). See [Per-launch services](workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WVM-42` | `F-WVM-13`, `F-WVM-14`, `F-WVM-15`, `E-WVM-1`, `P-WVM-7`, `DS-WVM-1`, `P-WVM-4` | repudiation | A command runs without a record binding its caller, launch, and confinement settings. | [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) Command attribution and lifecycle records remain unspecified. [Provider observations](workload-provider-windows-vm.md#Host%20observations). See [Per-launch services](workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WVM-43` | `F-WVM-13`, `F-WVM-14`, `F-WVM-15`, `E-WVM-1`, `P-WVM-7`, `DS-WVM-1`, `P-WVM-4` | information disclosure | A command inherits another launch's environment, session data, or readable workspace paths. | [Host process sandbox](workload-sandbox.md#Host%20process%20sandbox) Execution identity, inherited state, and path ownership remain unresolved. [Provider startup](workload-provider-windows-vm.md#Process%20placement%20and%20startup). See [Per-launch services](workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WVM-44` | `F-WVM-13`, `F-WVM-14`, `F-WVM-15`, `E-WVM-1`, `P-WVM-7`, `DS-WVM-1`, `P-WVM-4` | denial of service | A command continues running after cancellation or blocks workspace capture and teardown. | [Cancellation and limits](workload-lifecycle.md#Cancellation%20and%20limits) Independent command lifetime and cancellation remain unspecified. [Provider teardown](workload-provider-windows-vm.md#Cancellation%20and%20teardown). See [Per-launch services](workload-proxy.md#Launch%20boundary). | gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-WVM-45` | `F-WVM-13`, `F-WVM-14`, `F-WVM-15`, `E-WVM-1`, `P-WVM-7`, `DS-WVM-1`, `P-WVM-4` | elevation of privilege | A command starts outside the selected workload provider confinement or obtains a relaxation without review. | [Host process sandbox](workload-sandbox.md#Host%20process%20sandbox) afakt-sandbox confines standalone commands. Authority for an unsandboxed invocation remains unresolved. [Provider startup](workload-provider-windows-vm.md#Process%20placement%20and%20startup). See [Per-launch services](workload-proxy.md#Launch%20boundary). | depends on an open question ([`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60)) |

### Per-launch network boundary

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-WVM-155` | `F-WVM-17`, `P-WVM-4`, `P-WVM-9` | spoofing | A process substitutes another launch's local workload-proxy endpoint. | [Per-launch services](workload-proxy.md#Launch%20boundary) requires a local channel bound to the launch. Its peer authentication remains undecided. | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |
| `TM-WVM-156` | `F-WVM-17`, `P-WVM-4`, `P-WVM-9` | tampering | A malformed connection request changes its caller or target. | [Overlay connections](workload-proxy.md#Overlay%20connections) require direct mutual authentication and mediator checks. Binding caller and publication context to the guest listener remains open. | depends on an open question ([`OQ-PORT-14`](agent-open-questions.md#%60OQ-PORT-14%60)) |
| `TM-WVM-157` | `F-WVM-17`, `P-WVM-4`, `P-WVM-9` | repudiation | Reconnect loses the association between an admitted stream and its authorization record. | [Session message log](agent-acp.md#Session%20message%20log) requires retained history. Its owner, replay, and reconnect attribution remain open. | depends on an open question ([`OQ-SES-18`](agent-open-questions.md#%60OQ-SES-18%60)) |
| `TM-WVM-158` | `F-WVM-17`, `P-WVM-4`, `P-WVM-9` | information disclosure | Connection replacement sends a stream to a different launch. | Direct [overlay connections](workload-proxy.md#Overlay%20connections) authenticate both peers and require mediator decisions. How that publication and caller map to the intended listener remains unresolved. | depends on an open question ([`OQ-PORT-14`](agent-open-questions.md#%60OQ-PORT-14%60)) |
| `TM-WVM-159` | `F-WVM-17`, `P-WVM-4`, `P-WVM-9` | denial of service | One launch consumes proxy buffers or blocks control streams with published-port traffic. | [Overlay connections](workload-proxy.md#Overlay%20connections) requires bounded admission. The local scheduling and buffer limits remain undecided. | depends on an open question ([`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60)) |
| `TM-WVM-160` | `F-WVM-17`, `P-WVM-4`, `P-WVM-9` | elevation of privilege | After compromising the workload proxy, an attacker sends traffic without mediator approval. | [Per-launch services](workload-proxy.md#Launch%20boundary) separates policy from network authority. Their authenticated local binding remains undecided. | depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)) |

## Attack trees

`OR` requires any child. `AND` requires every child. Leaves reference local threats. Shared policy failures use the [common attack trees](workload-threat-model.md#Attack%20trees).

### Escape to the workload host

```mermaid
flowchart TD
    root["Escape to the workload host"] --> choice{"OR"}
    choice --> leaf0["TM-WVM-151: Exploit the isolation mechanism"]
    choice --> leaf1["TM-WVM-36: Make host processing execute workspace metadata"]
    choice --> leaf2["TM-WVM-30: Abuse an inherited control handle"]
    choice --> leaf3["TM-WVM-45: Run a workspace command outside intended confinement"]
```

- Cheapest path and blocking controls. Trigger host processing of hostile workspace metadata without a guest kernel exploit. [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60) leaves the host-side consumer unprotected. The [platform attack surface](#Host%20visibility%20and%20workspace%20commands) has separate confinement gaps.

### Read or change another agent launch's workspace or session

```mermaid
flowchart TD
    root["Read or change another agent launch's workspace or session"] --> choice{"OR"}
    choice --> leaf0["TM-WVM-130: Read a shared cache"]
    choice --> leaf1["TM-WVM-128: Change a cache consumed by another launch"]
    choice --> leaf2["TM-WVM-22: Cross workspace or scratch ownership"]
    choice --> leaf3["TM-WVM-28: Use an incorrectly bound local endpoint"]
    choice --> leaf4["TM-WVM-124: Read retained capture"]
```

- Cheapest path and blocking controls. Use a shared named tool cache already exposed to the launch. Cache authorization and provenance remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Workspace ownership and local endpoint binding require separate controls.

### Send traffic past egress policy

```mermaid
flowchart TD
    root["Send traffic past egress policy"] --> choice{"OR"}
    choice --> leaf0["TM-WVM-26: Use a direct network route"]
    choice --> leaf1["TM-WVM-25: Select another mediator identity"]
    choice --> leaf2["TM-WVM-45: Run a workspace command outside confinement"]
    choice --> leaf3["TM-WVM-30: Use a privileged control handle"]
```

- Cheapest path and blocking controls. Use a network path whose proxy routing is unspecified. [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60) covers missing workload provider routing. [`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60) covers sender binding. The [common egress tree](workload-threat-model.md#Send%20traffic%20past%20egress%20policy) covers permitted name queries, passthrough, and relay proxy use.

### Persist hostile state across agent launches

```mermaid
flowchart TD
    root["Persist hostile state across agent launches"] --> choice{"OR"}
    choice --> leaf0["TM-WVM-122: Restore mismatched or hostile captured state"]
    choice --> leaf1["TM-WVM-126: Carry old authority into a successor launch"]
    choice --> leaf2["TM-WVM-132: Execute a shared cache entry under later authority"]
```

- Cheapest path and blocking controls. Leave executable content in a shared named tool cache for a later launch. Cache provenance and invalidation remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Captured state needs a separate binding under [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60), while authority carried into a successor launch remains [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60). The [common persistence tree](workload-threat-model.md#Persist%20hostile%20state%20across%20agent%20launches) covers workspace images and hooks.

## Residual risks and gaps

Host isolation remains on hold. The [common residual risks](workload-threat-model.md#Residual%20risks%20and%20gaps) keep hook protection, repository configuration scope, repository profile self-grant and credential domain bounds, and session-end wait with workload identity lifetime. No pending option supplies a mitigation.

### Isolation and lifecycle

- `TM-WVM-23`. A hostile process tree ignores cancellation and continues running. depends on an open question ([`OQ-SBX-60`](agent-open-questions.md#%60OQ-SBX-60%60)).
- `TM-WVM-1`. A local process substitutes for the workload provider's control endpoint or launch process. depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)).
- `TM-WVM-2`. A launch changes inherited handles, mounts, descriptors, or confinement inputs to reach provider authority. depends on an open question ([`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60)).
- `TM-WVM-21`. A compromised workload host invents process exits or hides local workspace and command effects. gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)).
- `TM-WVM-70`. A same-user host process reads mediator memory, local session traffic, or another launch's files. depends on an open question ([`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60)).
- `TM-WVM-151`. A guest exploits a virtual device or the guest hypervisor to gain workload host authority. gap ([`OQ-SBX-53`](agent-open-questions.md#%60OQ-SBX-53%60)).
- `TM-WVM-154`. An orphaned process or reused launch user retains local authority after worker failure. depends on an open question ([`OQ-SBX-29`](agent-open-questions.md#%60OQ-SBX-29%60)).

### Local transport and egress

- `TM-WVM-25`. A process selects another launch's mediator endpoint and receives its authority. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-WVM-26`. A process changes proxy settings or uses an interface that bypasses the workload proxy mediator. gap ([`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60)).
- `TM-WVM-27`. Unbound local traffic acquires another launch's attribution before shared logging. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-WVM-28`. A control carrier or network view exposes another launch's session or loopback listener. depends on an open question ([`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60)).
- `TM-WVM-29`. A launch floods its local control carrier and prevents lifecycle or session progress. depends on an open question ([`OQ-SBX-14`](agent-open-questions.md#%60OQ-SBX-14%60)).
- `TM-WVM-30`. A local process uses an inherited control handle for workload host operations outside its launch authority. depends on an open question ([`OQ-SBX-13`](agent-open-questions.md#%60OQ-SBX-13%60)).
- `TM-WVM-67`. A process impersonates the guest relay or reuses a previous launch's control endpoint. depends on an open question ([`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60)).

### Workspace and scratch storage

- `TM-WVM-19`. A same-user process substitutes a workspace owner or capture helper. depends on an open question ([`OQ-SBX-27`](agent-open-questions.md#%60OQ-SBX-27%60)).
- `TM-WVM-32`. Hostile workspace metadata redirects workload host capture or cleanup outside the selected workspace. gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).
- `TM-WVM-33`. A workspace or scratch writer denies changes made through a shared local identity. gap ([`OQ-SBX-37`](agent-open-questions.md#%60OQ-SBX-37%60)).
- `TM-WVM-22`. Capture, scratch reuse, or path aliasing exposes another launch's files. gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-WVM-35`. Workspace or scratch growth exhausts host capacity outside an enforced allocation. gap ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-WVM-36`. Host-side Git processing runs repository hooks or configuration or follows alternates or submodules or accepts hostile worktree paths. gap ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).

### Shared named tool cache

- `TM-WVM-127`. One launch places cache entries that another mistakes for trusted tool output. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WVM-128`. A launch changes shared cache content consumed by another launch. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WVM-129`. A cache writer denies modifying content because cache provenance has no stated record. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WVM-130`. A launch reads private data another launch left in a shared tool cache. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WVM-131`. A launch fills or churns a shared cache and denies other launches useful capacity. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-WVM-132`. A cached executable or build result runs under a later launch's broader authority. gap ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).

### Workload snapshots and agent checkpoints

- `TM-WVM-121`. A capture or restore request substitutes another agent's owner or launch. depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)).
- `TM-WVM-122`. A mismatched snapshot, checkpoint, or workspace delta restores inconsistent or hostile state. depends on an open question ([`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60)).
- `TM-WVM-123`. Rewind hides an earlier external effect or presents it as undone. depends on an open question ([`OQ-AGT-4`](agent-open-questions.md#%60OQ-AGT-4%60)).
- `TM-WVM-124`. Snapshot transfer exposes private session or workspace data to an unauthorized reader. depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)).
- `TM-WVM-125`. Retained captures or uncertain ownership prevent safe collection and consume capacity. depends on an open question ([`OQ-AGT-11`](agent-open-questions.md#%60OQ-AGT-11%60)).
- `TM-WVM-126`. Restoration carries old identity, executable configuration, or authority into a successor launch. depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)).

### Commands in the agent workspace

- `TM-WVM-40`. A workspace command uses the authority of a different agent launch. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WVM-41`. A command races capture or changes files outside the selected workspace through a stale path. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WVM-42`. A command runs without a record binding its caller, launch, and confinement settings. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WVM-43`. A command inherits another launch's environment, session data, or readable workspace paths. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WVM-44`. A command continues running after cancellation or blocks workspace capture and teardown. gap ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-WVM-45`. A command starts outside the selected workload provider confinement or obtains a relaxation without review. depends on an open question ([`OQ-AGT-26`](agent-open-questions.md#%60OQ-AGT-26%60)).
