---
type: Threat Model
title: "Agent threat model for Cloudflare Sandbox SDK"
description: Remote execution, vendor custody, and provider integration threats for Cloudflare Sandbox SDK.
tags: [afakt, agents, security, threat-model]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent threat model for Cloudflare Sandbox SDK

[Public key infrastructure](pki.md).

## Scope

The Cloudflare Sandbox Software Development Kit (SDK) provider follows its [provider design](agent-sandbox-cloudflare.md).

Vendor lifecycle authority, sandbox isolation, workspace and session storage, network paths, and retained state fall within this model, as do commands requested through `afakt agent exec`. The [common agent threat model](agent-threat-model.md) covers shared control plane authorization, credential exchange, agent hub sessions, registries, and audit storage because their controls don't depend on this vendor.

STRIDE covers spoofing, tampering, repudiation, information disclosure, denial of service, and elevation of privilege. A virtual machine (VM) can supply a private kernel. Domain Name System (DNS) queries and Agent Client Protocol (ACP) sessions require the launch's agent mediator. Model Context Protocol (MCP) tools require afakt gateway authorization. An Application Programming Interface (API) exposes vendor management operations. Hypertext Transfer Protocol Secure (HTTPS) encrypts web traffic, while Hypertext Transfer Protocol (HTTP) names the underlying request protocol.

[Platforms and privileges](agent-sandbox-cloudflare.md#Platforms%20and%20privileges) and [Isolation boundary](agent-sandbox-cloudflare.md#Isolation%20boundary) define the known runtime boundary and missing privilege contract. [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup) distinguishes harness placement from tool execution. [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution) leaves mandatory remote mediation unresolved. [Cancellation and teardown](agent-sandbox-cloudflare.md#Cancellation%20and%20teardown) requires observed termination. Vendor idle state alone is insufficient. Adversaries include malicious sandbox code and hostile repository files. Another agent launch or a limited management caller can also attack. A compromised vendor service or administrator also threatens the launch. Vendor custody forms a separate trust boundary without an accepted residual risk for disclosure. Containment after an attacker compromises the runtime needs independent controls. The host isolation baseline remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Trust boundaries

| Boundary ID | Boundary | Two sides | Enforcement in the design |
| --- | --- | --- | --- |
| `TB-CF-1` | Vendor custody | afakt authority / vendor management, runtime, storage, and administrators | [Host observations](agent-sandbox-cloudflare.md#Host%20observations). Independent vendor evidence and custody controls remain [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60). |
| `TB-CF-2` | Sandbox runtime | untrusted execution / vendor kernel, runtime, and management | [Isolation boundary](agent-sandbox-cloudflare.md#Isolation%20boundary). Containment after runtime compromise remains [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60). |
| `TB-CF-3` | Remote mediation | Sandbox and vendor network paths / launch policy and authenticated afakt services | [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution). Carrier and bypass prevention remain [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). |
| `TB-CF-4` | Workspace processing | Launch-controlled files / privileged transfer, materialization, and capture | [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). Storage mapping and hostile metadata remain [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60), [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). |
| `TB-CF-5` | Shared named cache | One launch / state consumed by another launch | [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). Cache access and provenance remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). |
| `TB-CF-6` | Retained state | Current launch / captured content and successor authority | [Agent snapshots and agent checkpoints](agent-sandbox-cloudflare.md#Agent%20snapshots%20and%20agent%20checkpoints). Pairing and authorization remain [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60), [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). |
| `TB-CF-7` | Workspace command | Command requester / target workspace, process, and launch authority | [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). Command integration remains [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). |

## Data flow diagram

### Element inventory

| Element ID | Element | Location or custody |
| --- | --- | --- |
| `E-CF-1` | Workspace command caller | afakt caller |
| `E-CF-2` | Another agent launch | Separate launch authority |
| `E-CF-3` | Vendor administrator | Privileged vendor authority |
| `E-CF-4` | External destination | Outside afakt and the launch |
| `P-CF-1` | Agent host worker and workspace processor | afakt management. Remote storage mapping unresolved |
| `P-CF-2` | Vendor management and runtime services | Vendor custody outside untrusted execution |
| `P-CF-3` | Sandbox execution | Vendor sandbox. afakt-agent, afakt-sandbox, and harness. Compatibility unresolved |
| `P-CF-4` | Agent mediator | Required per launch. Remote placement unresolved |
| `P-CF-5` | Agent proxy | Required outside sandbox. Remote placement unresolved |
| `P-CF-6` | Agent hub and global egress gateway | afakt control plane. Shared models apply |
| `P-CF-7` | Successor launch | New launch authority |
| `P-CF-8` | Workspace command | afakt-sandbox placement unresolved |
| `DS-CF-1` | Workspace and scratch | Launch data. Remote backing and separation unresolved |
| `DS-CF-2` | Shared named tool cache | Cross-launch sharing. Provider mapping unresolved |
| `DS-CF-3` | Verified image cache | Outside launch-writable storage. Remote mapping unresolved |
| `DS-CF-4` | Retained workspace and session state | Vendor copies or afakt captures. Paired capture unresolved |
| `DS-CF-5` | Vendor-held credentials and observations | Vendor custody. Access and retention unresolved |

### Level one provider flows

Arrows show requests and responses, with dashed arrows marking afakt exchanges that need a provider carrier or placement decision. Vendor network paths remain visible because they can bypass that intended path. Shared caches and retained state have separate boundaries because neither belongs implicitly to one launch. Physical placement for the mediator and agent proxy remains undecided.

```mermaid
flowchart LR
    subgraph vendor["TB-CF-1 / vendor custody"]
        P-CF-2["P-CF-2: Vendor management and runtime"]
        E-CF-3["E-CF-3: Vendor administrator"]
        DS-CF-5["DS-CF-5: Vendor credentials and observations"]
        subgraph execution["TB-CF-2 / sandbox execution"]
            P-CF-3["P-CF-3: Untrusted sandbox execution"]
        end
    end
    subgraph afakt["TB-CF-1 / afakt authority"]
        P-CF-1["P-CF-1: Worker and workspace processor"]
    end
    subgraph mediation["TB-CF-3 / remote placement undecided"]
        P-CF-4["P-CF-4: Agent mediator"]
        P-CF-5["P-CF-5: Agent proxy"]
    end
    subgraph services["TB-CF-3 / afakt services"]
        P-CF-6["P-CF-6: Hub and global egress gateway"]
    end
    subgraph files["TB-CF-4 / launch-controlled storage"]
        DS-CF-1["DS-CF-1: Workspace and scratch"]
    end
    subgraph images["TB-CF-4 / verified content"]
        DS-CF-3["DS-CF-3: Verified image cache"]
    end
    subgraph shared["TB-CF-5 / shared across launches"]
        DS-CF-2["DS-CF-2: Named tool cache"]
    end
    subgraph other["TB-CF-5 / other launch"]
        E-CF-2["E-CF-2: Another launch"]
    end
    subgraph capture["TB-CF-6 / retained state"]
        DS-CF-4["DS-CF-4: Workspace and session captures"]
    end
    subgraph successor["TB-CF-6 / new authority"]
        P-CF-7["P-CF-7: Successor launch"]
    end
    subgraph caller["TB-CF-7 / command caller"]
        E-CF-1["E-CF-1: Command caller"]
    end
    subgraph command["TB-CF-7 / command placement unresolved"]
        P-CF-8["P-CF-8: Workspace command"]
    end
    E-CF-4["E-CF-4: External destination"]
    P-CF-1 -->|"F-CF-1"| P-CF-2
    P-CF-2 -->|"F-CF-2"| P-CF-3
    P-CF-3 -->|"F-CF-3"| P-CF-2
    P-CF-3 -.->|"F-CF-4"| P-CF-4
    P-CF-4 -.->|"F-CF-5"| P-CF-5
    P-CF-5 -.->|"F-CF-6"| P-CF-6
    P-CF-3 -->|"F-CF-7"| E-CF-4
    P-CF-3 -->|"F-CF-8"| DS-CF-1
    P-CF-1 -.->|"F-CF-9"| DS-CF-1
    DS-CF-3 -.->|"F-CF-10"| P-CF-1
    P-CF-3 -.->|"F-CF-11"| DS-CF-2
    E-CF-2 -.->|"F-CF-12"| DS-CF-2
    P-CF-2 -->|"F-CF-13"| DS-CF-4
    DS-CF-4 -.->|"F-CF-14"| P-CF-7
    E-CF-1 -.->|"F-CF-15"| P-CF-8
    P-CF-8 -.->|"F-CF-16"| DS-CF-1
    P-CF-8 -.->|"F-CF-17"| P-CF-4
    E-CF-3 -->|"F-CF-18"| P-CF-2
    P-CF-2 -->|"F-CF-19"| DS-CF-5
    P-CF-2 -->|"F-CF-20"| E-CF-4
    P-CF-3 -->|"F-CF-21"| DS-CF-5
```

- **`F-CF-1`.** Creation, configuration, status, and stop operations, including managed session control where present. Crosses `TB-CF-1`.
- **`F-CF-2`.** Execution requests, process control, and observed outcomes across the sandbox boundary. Crosses `TB-CF-2`.
- **`F-CF-3`.** untrusted runtime input, system calls, and execution output presented to vendor services. Crosses `TB-CF-2`.
- **`F-CF-4`.** Required launch-bound network, name queries, control, ACP, MCP, and preview traffic. Crosses `TB-CF-1`, `TB-CF-3`.
- **`F-CF-5`.** Permitted requests passed from the no-network mediator to its agent proxy. Crosses `TB-CF-3`.
- **`F-CF-6`.** Authenticated hub streams and inspected traffic to the global egress gateway. Common service controls apply. Crosses `TB-CF-3`.
- **`F-CF-7`.** Vendor-permitted external traffic, including potential direct paths outside afakt mediation. Crosses `TB-CF-1`, `TB-CF-3`.
- **`F-CF-8`.** Workspace and scratch reads and writes. Crosses `TB-CF-4`.
- **`F-CF-9`.** Remote workspace transfer, capture, materialization, and cleanup of hostile content. Crosses `TB-CF-1`, `TB-CF-4`.
- **`F-CF-10`.** Verified workspace image content for independent writable materialization. Crosses `TB-CF-4`.
- **`F-CF-11`.** Named cache access by this launch. Provider backing unresolved. Crosses `TB-CF-5`.
- **`F-CF-12`.** Named cache access by another launch, exposing cross-launch sharing. Crosses `TB-CF-5`.
- **`F-CF-13`.** Vendor-retained execution and session content. afakt capture scope remains unresolved. Crosses `TB-CF-1`, `TB-CF-6`.
- **`F-CF-14`.** Retained content admitted to a successor with new launch authority. Crosses `TB-CF-6`.
- **`F-CF-15`.** Authorized workspace command request. Crosses `TB-CF-7`.
- **`F-CF-16`.** Command reads and writes in the selected workspace. Crosses `TB-CF-4`, `TB-CF-7`.
- **`F-CF-17`.** Command traffic requiring the current launch’s mediator. Crosses `TB-CF-3`, `TB-CF-7`.
- **`F-CF-18`.** Privileged vendor control. Exact internal access limits aren't stated. Crosses `TB-CF-1`.
- **`F-CF-19`.** Vendor service handling of supplied credentials, runtime observations, and logs. Crosses `TB-CF-1`.
- **`F-CF-20`.** Worker outbound forwarding outside untrusted execution. Optional vendor features, not selected afakt placement. Crosses `TB-CF-1`, `TB-CF-3`.
- **`F-CF-21`.** Execution output and file content exposed to vendor processing. Internal retention remains undecided. Crosses `TB-CF-1`, `TB-CF-2`.

## STRIDE analysis

Each table covers the six categories for its listed flows and endpoints. A mitigation of ordinary cross-sandbox access doesn't mitigate an exploit of the runtime itself. No vendor mechanism substitutes for an unresolved afakt integration.

### Remote lifecycle and harness control

Sandbox identifiers route operations without authenticating callers, and the Durable Object identifier persists after container destruction. See [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-CF-1` | `F-CF-1`, `F-CF-2`, `P-CF-1`, `P-CF-2`, `P-CF-3` | spoofing | A caller guesses or reuses a sandbox identifier that the app treats as authentication. | Vendor object binding to the current launch remains undecided under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60)) |
| `TM-CF-2` | `F-CF-1`, `F-CF-2`, `P-CF-1`, `P-CF-2`, `P-CF-3` | tampering | A live Worker handler or host-rule change broadens the authority of an existing launch. | Enforcement of the pinned configuration across vendor changes remains undecided under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)) |
| `TM-CF-3` | `F-CF-1`, `F-CF-2`, `P-CF-1`, `P-CF-2`, `P-CF-3` | repudiation | A vendor service supplies a false startup, cancellation, or completion history. | Independent verification of vendor lifecycle observations has no stated mechanism under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-4` | `F-CF-1`, `F-CF-2`, `P-CF-1`, `P-CF-2`, `P-CF-3` | information disclosure | A broad provider credential exposes another launch’s prompts, files, or session records. | The scope of management credentials across launch data remains unspecified under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-5` | `F-CF-1`, `F-CF-2`, `P-CF-1`, `P-CF-2`, `P-CF-3` | denial of service | Account limits or unavailable remote capacity prevent startup or reconciliation. | Account capacity and provider availability need placement rules under [Platform dependencies](agent-sandbox-cloudflare.md#Platform%20dependencies). | depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)) |
| `TM-CF-6` | `F-CF-1`, `F-CF-2`, `P-CF-1`, `P-CF-2`, `P-CF-3` | elevation of privilege | A late operation recreates a destroyed container under its old sandbox identifier after a successor starts. | Remote stop evidence and safe replacement ordering remain undecided under [Cancellation and teardown](agent-sandbox-cloudflare.md#Cancellation%20and%20teardown). | depends on an open question ([`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60)) |

### Sandbox runtime and vendor host

The vendor states separate sandbox filesystems and VM boundaries. This blocks direct access without an exploit. Shell sessions in the same sandbox remain shared. See [Isolation boundary](agent-sandbox-cloudflare.md#Isolation%20boundary).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-CF-7` | `F-CF-3`, `P-CF-2`, `P-CF-3` | spoofing | A shell session receives independent authority despite sharing the sandbox’s processes and files. | Authentication of runtime management callers has no stated afakt mapping under [Isolation boundary](agent-sandbox-cloudflare.md#Isolation%20boundary). | gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)) |
| `TM-CF-8` | `F-CF-3`, `P-CF-2`, `P-CF-3` | tampering | A process directly changes files in a separate sandbox without shared storage or management authority. | [Isolation boundary](agent-sandbox-cloudflare.md#Isolation%20boundary) states separate sandbox filesystems under the vendor runtime boundary. This threat excludes shared mounts, management authority, and runtime exploits. | mitigated by the design |
| `TM-CF-9` | `F-CF-3`, `P-CF-2`, `P-CF-3` | repudiation | A runtime compromise suppresses evidence of activity on the vendor host. | The vendor boundary supplies no independent evidence after runtime compromise under [Isolation boundary](agent-sandbox-cloudflare.md#Isolation%20boundary). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-10` | `F-CF-3`, `P-CF-2`, `P-CF-3` | information disclosure | A runtime exploit exposes another sandbox’s memory or private workspace. | Containment of a runtime exploit can't rely on the compromised runtime under [Isolation boundary](agent-sandbox-cloudflare.md#Isolation%20boundary). | gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)) |
| `TM-CF-11` | `F-CF-3`, `P-CF-2`, `P-CF-3` | denial of service | A launch consumes shared runtime resources needed by other launches. | Resource limits need an account and runtime allocation policy under [Platform dependencies](agent-sandbox-cloudflare.md#Platform%20dependencies). | depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)) |
| `TM-CF-12` | `F-CF-3`, `P-CF-2`, `P-CF-3` | elevation of privilege | Hostile guest input exploits the VM boundary and reaches the vendor’s container host. | Containment after an escape into the vendor host remains unspecified under [Isolation boundary](agent-sandbox-cloudflare.md#Isolation%20boundary). | gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)) |

### Agent mediator and vendor network paths

HTTPS interception is optional, DNS uses vendor resolvers, and Worker handlers aren't selected afakt mediator components. See [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-CF-13` | `F-CF-4`, `F-CF-5`, `F-CF-6`, `F-CF-7`, `F-CF-20`, `P-CF-2`, `P-CF-3`, `P-CF-4`, `P-CF-5`, `P-CF-6`, `E-CF-4` | spoofing | A vendor-side endpoint attributes traffic to another afakt launch. | Provider identity assertions need a binding to the current afakt launch under [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution). | depends on an open question ([`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60)) |
| `TM-CF-14` | `F-CF-4`, `F-CF-5`, `F-CF-6`, `F-CF-7`, `F-CF-20`, `P-CF-2`, `P-CF-3`, `P-CF-4`, `P-CF-5`, `P-CF-6`, `E-CF-4` | tampering | HTTPS bypasses host rules when the operator leaves interception off and internet access on. | The required exclusive mediator route has no selected provider mechanism under [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution). | gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)) |
| `TM-CF-15` | `F-CF-4`, `F-CF-5`, `F-CF-6`, `F-CF-7`, `F-CF-20`, `P-CF-2`, `P-CF-3`, `P-CF-4`, `P-CF-5`, `P-CF-6`, `E-CF-4` | repudiation | A vendor omits traffic on its network path from afakt’s decision history. | Independent accounting for vendor-side network actions remains unspecified under [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-16` | `F-CF-4`, `F-CF-5`, `F-CF-6`, `F-CF-7`, `F-CF-20`, `P-CF-2`, `P-CF-3`, `P-CF-4`, `P-CF-5`, `P-CF-6`, `E-CF-4` | information disclosure | A vendor-resolved DNS query or a leaked preview URL sends private content outside afakt’s mediated path. | Name queries and vendor egress lack a selected afakt inspection path under [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution). | gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)) |
| `TM-CF-17` | `F-CF-4`, `F-CF-5`, `F-CF-6`, `F-CF-7`, `F-CF-20`, `P-CF-2`, `P-CF-3`, `P-CF-4`, `P-CF-5`, `P-CF-6`, `E-CF-4` | denial of service | Blocked or stalled remote transport prevents control, session, or preview progress. | Control transport and its progress bounds need a provider mapping under [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution). | depends on an open question ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)) |
| `TM-CF-18` | `F-CF-4`, `F-CF-5`, `F-CF-6`, `F-CF-7`, `F-CF-20`, `P-CF-2`, `P-CF-3`, `P-CF-4`, `P-CF-5`, `P-CF-6`, `E-CF-4` | elevation of privilege | A Worker handler uses broad bindings or credentials without binding the operation to afakt’s current launch. | Vendor tool and proxy authority lack binding to afakt admission under [Mediator connection and name resolution](agent-sandbox-cloudflare.md#Mediator%20connection%20and%20name%20resolution). | gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)) |

### Workspace transfer and processing

Container state, bucket mounts, and directory backups have distinct lifetimes. Their afakt ownership mapping remains unresolved. See [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-CF-19` | `F-CF-8`, `F-CF-9`, `F-CF-10`, `P-CF-1`, `P-CF-3`, `DS-CF-1`, `DS-CF-3` | spoofing | A file or artifact handle selects another launch’s workspace. | File handles need ownership checks for the intended launch under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-CF-20` | `F-CF-8`, `F-CF-9`, `F-CF-10`, `P-CF-1`, `P-CF-3`, `DS-CF-1`, `DS-CF-3` | tampering | Hostile repository metadata changes how remote capture or extraction processes files. | Confinement of hostile metadata processing remains undecided under [Workspace image materialization](agent-sandbox-cloudflare.md#Workspace%20image%20materialization). | depends on an open question ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |
| `TM-CF-21` | `F-CF-8`, `F-CF-9`, `F-CF-10`, `P-CF-1`, `P-CF-3`, `DS-CF-1`, `DS-CF-3` | repudiation | A vendor changes file content without independently attributable evidence. | Attribution of vendor file changes lacks independent evidence under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-22` | `F-CF-8`, `F-CF-9`, `F-CF-10`, `P-CF-1`, `P-CF-3`, `DS-CF-1`, `DS-CF-3` | information disclosure | A shared sandbox session or bucket mount exposes another launch’s workspace files. | Storage reuse and deletion need a launch-specific ownership mapping under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-CF-23` | `F-CF-8`, `F-CF-9`, `F-CF-10`, `P-CF-1`, `P-CF-3`, `DS-CF-1`, `DS-CF-3` | denial of service | Workspace transfer or scratch growth exhausts storage before cleanup completes. | Workspace and scratch limits need a provider allocation and cleanup contract under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-CF-24` | `F-CF-8`, `F-CF-9`, `F-CF-10`, `P-CF-1`, `P-CF-3`, `DS-CF-1`, `DS-CF-3` | elevation of privilege | Hostile workspace metadata executes with vendor capture or materialization authority outside the sandbox. | The transfer and capture path lacks selected process confinement under [Workspace image materialization](agent-sandbox-cloudflare.md#Workspace%20image%20materialization). | depends on an open question ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |

### Shared named tool caches

No provider mapping selects independent named caches or controls which launches can write their contents. See [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-CF-25` | `F-CF-11`, `F-CF-12`, `P-CF-3`, `E-CF-2`, `DS-CF-2` | spoofing | A cache writer presents another launch’s cache namespace as its own. | Cache namespace authorization remains undecided under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-CF-26` | `F-CF-11`, `F-CF-12`, `P-CF-3`, `E-CF-2`, `DS-CF-2` | tampering | A launch poisons shared dependencies that another launch executes. | Writer trust and dependency integrity need a cache policy under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-CF-27` | `F-CF-11`, `F-CF-12`, `P-CF-3`, `E-CF-2`, `DS-CF-2` | repudiation | A shared cache changes without a record identifying its writer. | Attribution of shared cache writes remains undecided under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-CF-28` | `F-CF-11`, `F-CF-12`, `P-CF-3`, `E-CF-2`, `DS-CF-2` | information disclosure | A shared cache exposes private dependency or build content to another launch. | Shared cache readership needs a confidentiality policy under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-CF-29` | `F-CF-11`, `F-CF-12`, `P-CF-3`, `E-CF-2`, `DS-CF-2` | denial of service | A launch fills a shared cache and prevents other launches from making progress. | Cache capacity and invalidation need allocation rules under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-CF-30` | `F-CF-11`, `F-CF-12`, `P-CF-3`, `E-CF-2`, `DS-CF-2` | elevation of privilege | A writable cache entry becomes trusted executable input in a later launch. | Cache provenance doesn't yet define trust in executable input under [Workspace, scratch, and caches](agent-sandbox-cloudflare.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |

### Retained state and successor launch

Directory backup restores files through an overlay without restoring machine memory or paired native harness history. See [Agent snapshots and agent checkpoints](agent-sandbox-cloudflare.md#Agent%20snapshots%20and%20agent%20checkpoints).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-CF-31` | `F-CF-13`, `F-CF-14`, `P-CF-2`, `P-CF-7`, `DS-CF-4` | spoofing | A retained-state handle restores another agent’s workspace or session. | Restore authorization needs ownership checks on retained state under [Agent snapshots and agent checkpoints](agent-sandbox-cloudflare.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)) |
| `TM-CF-32` | `F-CF-13`, `F-CF-14`, `P-CF-2`, `P-CF-7`, `DS-CF-4` | tampering | Restoring a backup discards the writable overlay while native harness state describes newer files. | The complete workspace and native-session capture pair remains undecided under [Agent snapshots and agent checkpoints](agent-sandbox-cloudflare.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60)) |
| `TM-CF-33` | `F-CF-13`, `F-CF-14`, `P-CF-2`, `P-CF-7`, `DS-CF-4` | repudiation | afakt lacks attribution of a vendor state change to its capture or restore requester. | Capture and restore records need attribution to their requester under [Agent snapshots and agent checkpoints](agent-sandbox-cloudflare.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)) |
| `TM-CF-34` | `F-CF-13`, `F-CF-14`, `P-CF-2`, `P-CF-7`, `DS-CF-4` | information disclosure | A capture preserves credentials or private session data beyond its authorized scope. | Capture scope must separate retained content from active authority under [Agent snapshots and agent checkpoints](agent-sandbox-cloudflare.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)) |
| `TM-CF-35` | `F-CF-13`, `F-CF-14`, `P-CF-2`, `P-CF-7`, `DS-CF-4` | denial of service | Container sleep removes workspace changes before a new directory backup preserves them. | Durable ownership needs retention roots beyond vendor expiry under [Agent snapshots and agent checkpoints](agent-sandbox-cloudflare.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-11`](agent-open-questions.md#%60OQ-AGT-11%60)) |
| `TM-CF-36` | `F-CF-13`, `F-CF-14`, `P-CF-2`, `P-CF-7`, `DS-CF-4` | elevation of privilege | A successor restores stale authority or hostile state as trusted current execution. | Restoration needs destination authorization and exclusion of stale credentials under [Agent snapshots and agent checkpoints](agent-sandbox-cloudflare.md#Agent%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)) |

### Workspace command execution

Shell execution sessions share processes and files, so they don't establish standalone command confinement or authority. See [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-CF-37` | `F-CF-15`, `F-CF-16`, `F-CF-17`, `E-CF-1`, `P-CF-8`, `DS-CF-1`, `P-CF-4` | spoofing | A command request attaches to the wrong launch or recycled sandbox identifier. | Workspace command placement and target identity remain undecided under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-CF-38` | `F-CF-15`, `F-CF-16`, `F-CF-17`, `E-CF-1`, `P-CF-8`, `DS-CF-1`, `P-CF-4` | tampering | A workspace command races a checkpoint or changes the harness’s active files. | Command ordering relative to capture remains undecided under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-CF-39` | `F-CF-15`, `F-CF-16`, `F-CF-17`, `E-CF-1`, `P-CF-8`, `DS-CF-1`, `P-CF-4` | repudiation | A command result lacks binding to the requester, launch, and executed process. | Command evidence needs requester and process attribution under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-CF-40` | `F-CF-15`, `F-CF-16`, `F-CF-17`, `E-CF-1`, `P-CF-8`, `DS-CF-1`, `P-CF-4` | information disclosure | A command inherits provider credentials or another session’s environment. | Inherited command state and credential exclusions remain undecided under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-CF-41` | `F-CF-15`, `F-CF-16`, `F-CF-17`, `E-CF-1`, `P-CF-8`, `DS-CF-1`, `P-CF-4` | denial of service | A detached command continues after cancellation and retains resources or workspace access. | Command descendants need lifecycle binding through cancellation under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-CF-42` | `F-CF-15`, `F-CF-16`, `F-CF-17`, `E-CF-1`, `P-CF-8`, `DS-CF-1`, `P-CF-4` | elevation of privilege | A workspace command bypasses afakt-sandbox or uses vendor execution authority beyond its launch. | The command confinement and authority boundary remains undecided under [Process placement and startup](agent-sandbox-cloudflare.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |

### Vendor custody and administration

Cloudflare runs the container, Worker, routing, and backup services. The documentation omits independent protection from vendor custody. See [Host observations](agent-sandbox-cloudflare.md#Host%20observations).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-CF-43` | `F-CF-18`, `F-CF-19`, `F-CF-21`, `E-CF-3`, `P-CF-2`, `P-CF-3`, `DS-CF-4`, `DS-CF-5` | spoofing | A vendor administrator supplies false identity or observation data to afakt. | afakt lacks an independent check on vendor identity observations under [Host observations](agent-sandbox-cloudflare.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-44` | `F-CF-18`, `F-CF-19`, `F-CF-21`, `E-CF-3`, `P-CF-2`, `P-CF-3`, `DS-CF-4`, `DS-CF-5` | tampering | A vendor administrator changes retained workspace, session, or credential state. | Vendor custody supplies no independent integrity protection for retained content under [Host observations](agent-sandbox-cloudflare.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-45` | `F-CF-18`, `F-CF-19`, `F-CF-21`, `E-CF-3`, `P-CF-2`, `P-CF-3`, `DS-CF-4`, `DS-CF-5` | repudiation | A vendor administrator removes the evidence of privileged content access. | Evidence of vendor administrator access lacks independent protection under [Host observations](agent-sandbox-cloudflare.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-46` | `F-CF-18`, `F-CF-19`, `F-CF-21`, `E-CF-3`, `P-CF-2`, `P-CF-3`, `DS-CF-4`, `DS-CF-5` | information disclosure | A Cloudflare service or privileged vendor operator discloses container files, R2 backups, or Worker-held secrets. | No stated control prevents disclosure by the vendor holding the content under [Host observations](agent-sandbox-cloudflare.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-CF-47` | `F-CF-18`, `F-CF-19`, `F-CF-21`, `E-CF-3`, `P-CF-2`, `P-CF-3`, `DS-CF-4`, `DS-CF-5` | denial of service | Vendor deletion, retention, or account suspension makes required state unavailable. | Provider retention and account eligibility need availability rules under [Host observations](agent-sandbox-cloudflare.md#Host%20observations). | depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)) |
| `TM-CF-48` | `F-CF-18`, `F-CF-19`, `F-CF-21`, `E-CF-3`, `P-CF-2`, `P-CF-3`, `DS-CF-4`, `DS-CF-5` | elevation of privilege | A compromised vendor service uses supplied management or upstream credentials outside the intended launch. | The vendor trust boundary doesn't limit a compromised custodian’s authority under [Host observations](agent-sandbox-cloudflare.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |

## Attack trees

### Escape to the vendor host

```mermaid
flowchart TD
    root["Escape to vendor host"] --> alternatives{"OR"}
    alternatives --> a0["TM-CF-12: Exploit runtime boundary"]
    alternatives --> a1["TM-CF-7: Reach runtime management"]
    alternatives --> processing{"AND"}
    processing --> b0["TM-CF-20: Submit hostile metadata"]
    processing --> b1["TM-CF-24: Reach privileged vendor file processing"]
```

Runtime input already available to malicious sandbox code is the path requiring the least extra authority. The cited controls supply no containment after a runtime exploit. Reaching a privileged vendor file processor also requires that the unresolved transfer or capture path accept the hostile content.

### Read or change another agent launch’s workspace or session

```mermaid
flowchart TD
    root["Read or change another launch"] --> alternatives{"OR"}
    alternatives --> a0["TM-CF-1: Select another vendor object"]
    alternatives --> a1["TM-CF-22: Read reused workspace content"]
    alternatives --> a2["TM-CF-26: Poison shared cache"]
    alternatives --> a3["TM-CF-31: Select another retained capture"]
```

A shared environment, mount, cache, or broad object handle needs fewer capabilities than a runtime escape. Launch-specific storage and authorization mappings remain unresolved. Separate sandbox filesystems only block direct file access when there is no shared storage or management authority.

### Send traffic past egress policy

```mermaid
flowchart TD
    root["Bypass egress policy"] --> alternatives{"OR"}
    alternatives --> a0["TM-CF-14: Use alternate vendor network path"]
    alternatives --> a1["TM-CF-16: Disclose through DNS or permitted host"]
    alternatives --> a2["TM-CF-18: Use vendor tool or proxy authority"]
```

A malicious sandbox process can use a documented vendor network or tool path directly. Mandatory afakt mediation remains unresolved because destination allowlists leave the required carrier and request policy unspecified.

### Disclose workspace or credentials through the vendor

```mermaid
flowchart TD
    root["Vendor-side disclosure"] --> alternatives{"OR"}
    alternatives --> a["TM-CF-46: Read workspace or secret custody"]
    alternatives --> conceal{"AND"}
    conceal --> b0["TM-CF-44: Alter retained state"]
    conceal --> b1["TM-CF-45: Remove access evidence"]
```

A privileged vendor service already has custody of submitted data. The cited controls supply no independent barrier to its disclosure. The concealment branch requires both changing retained state and removing access evidence.

## Residual risks and gaps

Provider-specific gaps don't settle the on-hold choices for hook protection, repository project configuration, profile self-grants and credential domains, host isolation, or session-end wait and identity lifetime. See [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60), [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60), [`OQ-PRF-5`](agent-open-questions.md#%60OQ-PRF-5%60), [`OQ-SBX-2`](agent-open-questions.md#%60OQ-SBX-2%60), [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60), and [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

### Remote lifecycle and harness control

- `TM-CF-1`. A caller guesses or reuses a sandbox identifier that the app treats as authentication. depends on an open question ([`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60)).
- `TM-CF-2`. A live Worker handler or host-rule change broadens the authority of an existing launch. depends on an open question ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)).
- `TM-CF-3`. A vendor service supplies a false startup, cancellation, or completion history. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-4`. A broad provider credential exposes another launch’s prompts, files, or session records. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-5`. Account limits or unavailable remote capacity prevent startup or reconciliation. depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)).
- `TM-CF-6`. A late operation recreates a destroyed container under its old sandbox identifier after a successor starts. depends on an open question ([`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60)).

### Sandbox runtime and vendor host

- `TM-CF-7`. A shell session receives independent authority despite sharing the sandbox’s processes and files. gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)).
- `TM-CF-9`. A runtime compromise suppresses evidence of activity on the vendor host. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-10`. A runtime exploit exposes another sandbox’s memory or private workspace. gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)).
- `TM-CF-11`. A launch consumes shared runtime resources needed by other launches. depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)).
- `TM-CF-12`. Hostile guest input exploits the VM boundary and reaches the vendor’s container host. gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)).

### Agent mediator and vendor network paths

- `TM-CF-13`. A vendor-side endpoint attributes traffic to another afakt launch. depends on an open question ([`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60)).
- `TM-CF-14`. HTTPS bypasses host rules when the operator leaves interception off and internet access on. gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)).
- `TM-CF-15`. A vendor omits traffic on its network path from afakt’s decision history. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-16`. A vendor-resolved DNS query or a leaked preview URL sends private content outside afakt’s mediated path. gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)).
- `TM-CF-17`. Blocked or stalled remote transport prevents control, session, or preview progress. depends on an open question ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)).
- `TM-CF-18`. A Worker handler uses broad bindings or credentials without binding the operation to afakt’s current launch. gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)).

### Workspace transfer and processing

- `TM-CF-19`. A file or artifact handle selects another launch’s workspace. depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-CF-20`. Hostile repository metadata changes how remote capture or extraction processes files. depends on an open question ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).
- `TM-CF-21`. A vendor changes file content without independently attributable evidence. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-22`. A shared sandbox session or bucket mount exposes another launch’s workspace files. depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-CF-23`. Workspace transfer or scratch growth exhausts storage before cleanup completes. depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-CF-24`. Hostile workspace metadata executes with vendor capture or materialization authority outside the sandbox. depends on an open question ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).

### Shared named tool caches

- `TM-CF-25`. A cache writer presents another launch’s cache namespace as its own. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-CF-26`. A launch poisons shared dependencies that another launch executes. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-CF-27`. A shared cache changes without a record identifying its writer. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-CF-28`. A shared cache exposes private dependency or build content to another launch. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-CF-29`. A launch fills a shared cache and prevents other launches from making progress. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-CF-30`. A writable cache entry becomes trusted executable input in a later launch. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).

### Retained state and successor launch

- `TM-CF-31`. A retained-state handle restores another agent’s workspace or session. depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)).
- `TM-CF-32`. Restoring a backup discards the writable overlay while native harness state describes newer files. depends on an open question ([`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60)).
- `TM-CF-33`. afakt lacks attribution of a vendor state change to its capture or restore requester. depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)).
- `TM-CF-34`. A capture preserves credentials or private session data beyond its authorized scope. depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)).
- `TM-CF-35`. Container sleep removes workspace changes before a new directory backup preserves them. depends on an open question ([`OQ-AGT-11`](agent-open-questions.md#%60OQ-AGT-11%60)).
- `TM-CF-36`. A successor restores stale authority or hostile state as trusted current execution. depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)).

### Workspace command execution

- `TM-CF-37`. A command request attaches to the wrong launch or recycled sandbox identifier. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-CF-38`. A workspace command races a checkpoint or changes the harness’s active files. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-CF-39`. A command result lacks binding to the requester, launch, and executed process. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-CF-40`. A command inherits provider credentials or another session’s environment. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-CF-41`. A detached command continues after cancellation and retains resources or workspace access. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-CF-42`. A workspace command bypasses afakt-sandbox or uses vendor execution authority beyond its launch. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).

### Vendor custody and administration

- `TM-CF-43`. A vendor administrator supplies false identity or observation data to afakt. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-44`. A vendor administrator changes retained workspace, session, or credential state. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-45`. A vendor administrator removes the evidence of privileged content access. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-46`. A Cloudflare service or privileged vendor operator discloses container files, R2 backups, or Worker-held secrets. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-CF-47`. Vendor deletion, retention, or account suspension makes required state unavailable. depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)).
- `TM-CF-48`. A compromised vendor service uses supplied management or upstream credentials outside the intended launch. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
