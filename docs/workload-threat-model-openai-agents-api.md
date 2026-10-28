---
type: Threat Model
title: "Workload threat model for OpenAI Agents API"
description: Remote execution, vendor custody, and provider integration threats for OpenAI Agents API.
tags: [afakt, agents, security, threat-model]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload threat model for OpenAI Agents API

[Public key infrastructure](pki.md).

## Scope

The OpenAI Agents Application Programming Interface (API) provider follows its [provider design](workload-provider-openai-agents-api.md).

Vendor lifecycle authority, sandbox isolation, workspace and session storage, network paths, and retained state fall within this model, as do commands requested through `afakt agent exec`. The [common workload threat model](workload-threat-model.md) covers shared control plane authorization, credential exchange, session access, registries, and audit storage because their controls don't depend on this vendor.

STRIDE covers spoofing, tampering, repudiation, information disclosure, denial of service, and elevation of privilege. A virtual machine (VM) can supply a private kernel. Domain Name System (DNS) queries and Agent Client Protocol (ACP) sessions require the launch's workload proxy mediator. Model Context Protocol (MCP) tools require afakt gateway authorization. The API exposes vendor management operations. Hypertext Transfer Protocol Secure (HTTPS) encrypts web traffic, while Hypertext Transfer Protocol (HTTP) names the underlying request protocol.

[Platforms and privileges](workload-provider-openai-agents-api.md#Platforms%20and%20privileges) and [Isolation boundary](workload-provider-openai-agents-api.md#Isolation%20boundary) define the known runtime boundary and missing privilege contract. [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup) distinguishes harness placement from tool execution. [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution) leaves mandatory remote mediation unresolved. [Cancellation and teardown](workload-provider-openai-agents-api.md#Cancellation%20and%20teardown) requires observed termination. Vendor idle state alone is insufficient. Adversaries include malicious sandbox code and hostile repository files. Another agent launch or a limited management caller can also attack. A compromised vendor service or administrator also threatens the launch. Vendor custody forms a separate trust boundary without an accepted residual risk for disclosure. Containment after an attacker compromises the runtime needs independent controls. The host isolation baseline remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Trust boundaries

| Boundary ID | Boundary | Two sides | Enforcement in the design |
| --- | --- | --- | --- |
| `TB-OAA-1` | Vendor custody | afakt authority / vendor management, runtime, storage, and administrators | [Host observations](workload-provider-openai-agents-api.md#Host%20observations). Independent vendor evidence and custody controls remain [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60). |
| `TB-OAA-2` | Sandbox runtime | untrusted execution / vendor kernel, runtime, and management | [Isolation boundary](workload-provider-openai-agents-api.md#Isolation%20boundary). Containment after runtime compromise remains [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60). |
| `TB-OAA-3` | Remote mediation | Sandbox and vendor network paths / launch policy and authenticated afakt services | [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution). Carrier and bypass prevention remain [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). |
| `TB-OAA-4` | Workspace processing | Launch-controlled files / privileged transfer, materialization, and capture | [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). Storage mapping and hostile metadata remain [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60), [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). |
| `TB-OAA-5` | Shared named cache | One launch / state consumed by another launch | [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). Cache access and provenance remain [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). |
| `TB-OAA-6` | Retained state | Current launch / captured content and successor authority | [Workload snapshots and agent checkpoints](workload-provider-openai-agents-api.md#Workload%20snapshots%20and%20agent%20checkpoints). Pairing and authorization remain [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60), [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). |
| `TB-OAA-7` | Workspace command | Command requester / target workspace, process, and launch authority | [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). Command integration remains [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). |

## Data flow diagram

### Element inventory

| Element ID | Element | Location or custody |
| --- | --- | --- |
| `E-OAA-1` | Workspace command caller | afakt caller |
| `E-OAA-2` | Another agent launch | Separate launch authority |
| `E-OAA-3` | Vendor administrator | Privileged vendor authority |
| `E-OAA-4` | External destination | Outside afakt and the launch |
| `P-OAA-1` | Workload host worker and workspace processor | afakt management. Remote storage mapping unresolved |
| `P-OAA-2` | Vendor management and managed harness | Vendor custody outside untrusted execution |
| `P-OAA-3` | Sandbox execution | Vendor sandbox. Managed tools, not the managed loop |
| `P-OAA-4` | Workload proxy mediator | Required per launch. Remote placement unresolved |
| `P-OAA-5` | Workload proxy | Required outside sandbox. Remote placement unresolved |
| `P-OAA-6` | Overlay peers | Workload peers and control plane services. Shared models apply |
| `P-OAA-7` | Successor launch | New launch authority |
| `P-OAA-8` | Workspace command | afakt-sandbox placement unresolved |
| `DS-OAA-1` | Workspace and scratch | Launch data. Remote backing and separation unresolved |
| `DS-OAA-2` | Shared named tool cache | Cross-launch sharing. Provider mapping unresolved |
| `DS-OAA-3` | Verified image cache | Outside launch-writable storage. Remote mapping unresolved |
| `DS-OAA-4` | Retained workspace and session state | Vendor copies or afakt captures. Paired capture unresolved |
| `DS-OAA-5` | Vendor-held credentials and observations | Vendor custody. Access and retention unresolved |

### Level one provider flows

Arrows show requests and responses, with dashed arrows marking afakt exchanges that need a provider carrier or placement decision. Vendor network paths remain visible because they can bypass that intended path. Shared caches and retained state have separate boundaries because neither belongs implicitly to one launch. Physical placement for the mediator and workload proxy remains undecided.

```mermaid
flowchart LR
    subgraph vendor["TB-OAA-1 / vendor custody"]
        P-OAA-2["P-OAA-2: Vendor management and managed harness"]
        E-OAA-3["E-OAA-3: Vendor administrator"]
        DS-OAA-5["DS-OAA-5: Vendor credentials and observations"]
        subgraph execution["TB-OAA-2 / sandbox execution"]
            P-OAA-3["P-OAA-3: Untrusted sandbox execution"]
        end
    end
    subgraph afakt["TB-OAA-1 / afakt authority"]
        P-OAA-1["P-OAA-1: Worker and workspace processor"]
    end
    subgraph mediation["TB-OAA-3 / remote placement undecided"]
        P-OAA-4["P-OAA-4: Workload proxy mediator"]
        P-OAA-5["P-OAA-5: Workload proxy"]
    end
    subgraph services["TB-OAA-3 / overlay peers and afakt services"]
        P-OAA-6["P-OAA-6: Overlay peers"]
    end
    subgraph files["TB-OAA-4 / launch-controlled storage"]
        DS-OAA-1["DS-OAA-1: Workspace and scratch"]
    end
    subgraph images["TB-OAA-4 / verified content"]
        DS-OAA-3["DS-OAA-3: Verified image cache"]
    end
    subgraph shared["TB-OAA-5 / shared across launches"]
        DS-OAA-2["DS-OAA-2: Named tool cache"]
    end
    subgraph other["TB-OAA-5 / other launch"]
        E-OAA-2["E-OAA-2: Another launch"]
    end
    subgraph capture["TB-OAA-6 / retained state"]
        DS-OAA-4["DS-OAA-4: Workspace and session captures"]
    end
    subgraph successor["TB-OAA-6 / new authority"]
        P-OAA-7["P-OAA-7: Successor launch"]
    end
    subgraph caller["TB-OAA-7 / command caller"]
        E-OAA-1["E-OAA-1: Command caller"]
    end
    subgraph command["TB-OAA-7 / command placement unresolved"]
        P-OAA-8["P-OAA-8: Workspace command"]
    end
    E-OAA-4["E-OAA-4: External destination"]
    P-OAA-1 -->|"F-OAA-1"| P-OAA-2
    P-OAA-2 -->|"F-OAA-2"| P-OAA-3
    P-OAA-3 -->|"F-OAA-3"| P-OAA-2
    P-OAA-3 -.->|"F-OAA-4"| P-OAA-4
    P-OAA-4 -.->|"F-OAA-5"| P-OAA-5
    P-OAA-5 -.->|"F-OAA-6"| P-OAA-6
    P-OAA-5 -.->|"F-OAA-6"| E-OAA-4
    P-OAA-3 -->|"F-OAA-7"| E-OAA-4
    P-OAA-3 -->|"F-OAA-8"| DS-OAA-1
    P-OAA-1 -.->|"F-OAA-9"| DS-OAA-1
    DS-OAA-3 -.->|"F-OAA-10"| P-OAA-1
    P-OAA-3 -.->|"F-OAA-11"| DS-OAA-2
    E-OAA-2 -.->|"F-OAA-12"| DS-OAA-2
    P-OAA-2 -->|"F-OAA-13"| DS-OAA-4
    DS-OAA-4 -.->|"F-OAA-14"| P-OAA-7
    E-OAA-1 -.->|"F-OAA-15"| P-OAA-8
    P-OAA-8 -.->|"F-OAA-16"| DS-OAA-1
    P-OAA-8 -.->|"F-OAA-17"| P-OAA-4
    E-OAA-3 -->|"F-OAA-18"| P-OAA-2
    P-OAA-2 -->|"F-OAA-19"| DS-OAA-5
    P-OAA-2 -->|"F-OAA-20"| E-OAA-4
    P-OAA-3 -->|"F-OAA-21"| DS-OAA-5
```

- **`F-OAA-1`.** Creation, configuration, status, and stop operations, including managed session control where present. Crosses `TB-OAA-1`.
- **`F-OAA-2`.** Execution requests, process control, and observed outcomes across the sandbox boundary. Crosses `TB-OAA-2`.
- **`F-OAA-3`.** untrusted runtime input, system calls, and execution output presented to vendor services. Crosses `TB-OAA-2`.
- **`F-OAA-4`.** Required launch-bound network, name queries, control, ACP, MCP, and preview traffic. Crosses `TB-OAA-1`, `TB-OAA-3`.
- **`F-OAA-5`.** Permitted requests passed from the no-network mediator to its workload proxy. Crosses `TB-OAA-3`.
- **`F-OAA-6`.** Direct authenticated overlay connections and permitted internet traffic through the workload proxy. Its placement and carrier remain under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). Session routing depends on [`OQ-SES-17`](agent-open-questions.md#%60OQ-SES-17%60). Published-port routing depends on [`OQ-PORT-14`](agent-open-questions.md#%60OQ-PORT-14%60). Common service controls apply. Crosses `TB-OAA-3`.
- **`F-OAA-7`.** Vendor-permitted external traffic, including potential direct paths outside afakt mediation. Crosses `TB-OAA-1`, `TB-OAA-3`.
- **`F-OAA-8`.** Workspace and scratch reads and writes. Crosses `TB-OAA-4`.
- **`F-OAA-9`.** Remote workspace transfer, capture, materialization, and cleanup of hostile content. Crosses `TB-OAA-1`, `TB-OAA-4`.
- **`F-OAA-10`.** Verified workspace image content for independent writable materialization. Crosses `TB-OAA-4`.
- **`F-OAA-11`.** Named cache access by this launch. Provider backing unresolved. Crosses `TB-OAA-5`.
- **`F-OAA-12`.** Named cache access by another launch, exposing cross-launch sharing. Crosses `TB-OAA-5`.
- **`F-OAA-13`.** Vendor-retained execution and session content. afakt capture scope remains unresolved. Crosses `TB-OAA-1`, `TB-OAA-6`.
- **`F-OAA-14`.** Retained content admitted to a successor with new launch authority. Crosses `TB-OAA-6`.
- **`F-OAA-15`.** Authorized workspace command request. Crosses `TB-OAA-7`.
- **`F-OAA-16`.** Command reads and writes in the selected workspace. Crosses `TB-OAA-4`, `TB-OAA-7`.
- **`F-OAA-17`.** Command traffic requiring the current launch’s mediator. Crosses `TB-OAA-3`, `TB-OAA-7`.
- **`F-OAA-18`.** Privileged vendor control. Exact internal access limits aren't stated. Crosses `TB-OAA-1`.
- **`F-OAA-19`.** Vendor service handling of supplied credentials, runtime observations, and logs. Crosses `TB-OAA-1`.
- **`F-OAA-20`.** Managed tool calls outside the sandbox. Crosses `TB-OAA-1`, `TB-OAA-3`.
- **`F-OAA-21`.** Execution output and file content exposed to vendor processing. Internal retention remains undecided. Crosses `TB-OAA-1`, `TB-OAA-2`.

## STRIDE analysis

Each table covers the six categories for its listed flows and endpoints. A mitigation of ordinary cross-sandbox access doesn't mitigate an exploit of the runtime itself. No vendor mechanism substitutes for an unresolved afakt integration.

### Remote lifecycle and harness control

The OpenAI-managed harness is separate from the environment, and shared environment lifetime can outlast every attached session. See [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-OAA-1` | `F-OAA-1`, `F-OAA-2`, `P-OAA-1`, `P-OAA-2`, `P-OAA-3` | spoofing | A management caller selects another launch through a reused vendor object identifier. | Vendor object binding to the current launch remains undecided under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60)) |
| `TM-OAA-2` | `F-OAA-1`, `F-OAA-2`, `P-OAA-1`, `P-OAA-2`, `P-OAA-3` | tampering | A template, tool configuration, or shared environment change breaks the launch’s pinned execution contract. | Enforcement of the pinned configuration across vendor changes remains undecided under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60)) |
| `TM-OAA-3` | `F-OAA-1`, `F-OAA-2`, `P-OAA-1`, `P-OAA-2`, `P-OAA-3` | repudiation | A vendor service supplies a false startup, cancellation, or completion history. | Independent verification of vendor lifecycle observations has no stated mechanism under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-4` | `F-OAA-1`, `F-OAA-2`, `P-OAA-1`, `P-OAA-2`, `P-OAA-3` | information disclosure | A broad provider credential exposes another launch’s prompts, files, or session records. | The scope of management credentials across launch data remains unspecified under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-5` | `F-OAA-1`, `F-OAA-2`, `P-OAA-1`, `P-OAA-2`, `P-OAA-3` | denial of service | Account limits or unavailable remote capacity prevent startup or reconciliation. | Account capacity and provider availability need placement rules under [Platform dependencies](workload-provider-openai-agents-api.md#Platform%20dependencies). | depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)) |
| `TM-OAA-6` | `F-OAA-1`, `F-OAA-2`, `P-OAA-1`, `P-OAA-2`, `P-OAA-3` | elevation of privilege | Session deletion leaves a shared environment running while afakt starts a replacement launch. | Remote stop evidence and safe replacement ordering remain undecided under [Cancellation and teardown](workload-provider-openai-agents-api.md#Cancellation%20and%20teardown). | depends on an open question ([`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60)) |

### Sandbox runtime and vendor host

The provider supports sessions that share one environment despite keeping their conversation histories separate. The launch mapping is unresolved. See [Isolation boundary](workload-provider-openai-agents-api.md#Isolation%20boundary).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-OAA-7` | `F-OAA-3`, `P-OAA-2`, `P-OAA-3` | spoofing | untrusted execution impersonates a runtime management caller. | Authentication of runtime management callers has no stated afakt mapping under [Isolation boundary](workload-provider-openai-agents-api.md#Isolation%20boundary). | gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)) |
| `TM-OAA-8` | `F-OAA-3`, `P-OAA-2`, `P-OAA-3` | tampering | A session changes files belonging to another launch attached to the same pre-warmed environment. | [Isolation boundary](workload-provider-openai-agents-api.md#Isolation%20boundary) leaves the provider mechanism or authority unresolved under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). | depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-OAA-9` | `F-OAA-3`, `P-OAA-2`, `P-OAA-3` | repudiation | A runtime compromise suppresses evidence of activity on the vendor host. | The vendor boundary supplies no independent evidence after runtime compromise under [Isolation boundary](workload-provider-openai-agents-api.md#Isolation%20boundary). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-10` | `F-OAA-3`, `P-OAA-2`, `P-OAA-3` | information disclosure | A runtime exploit exposes another sandbox’s memory or private workspace. | Containment of a runtime exploit can't rely on the compromised runtime under [Isolation boundary](workload-provider-openai-agents-api.md#Isolation%20boundary). | gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)) |
| `TM-OAA-11` | `F-OAA-3`, `P-OAA-2`, `P-OAA-3` | denial of service | A launch consumes shared runtime resources needed by other launches. | Resource limits need an account and runtime allocation policy under [Platform dependencies](workload-provider-openai-agents-api.md#Platform%20dependencies). | depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)) |
| `TM-OAA-12` | `F-OAA-3`, `P-OAA-2`, `P-OAA-3` | elevation of privilege | Hostile system calls or virtual-device input escape the runtime into the vendor host. | Containment after an escape into the vendor host remains unspecified under [Isolation boundary](workload-provider-openai-agents-api.md#Isolation%20boundary). | gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)) |

### Workload proxy mediator and vendor network paths

Hosted network modes and vault proxies don't establish mandatory mediation of managed tools, DNS, and control traffic. See [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-OAA-13` | `F-OAA-4`, `F-OAA-5`, `F-OAA-6`, `F-OAA-7`, `F-OAA-20`, `P-OAA-2`, `P-OAA-3`, `P-OAA-4`, `P-OAA-5`, `P-OAA-6`, `E-OAA-4` | spoofing | A vendor-side endpoint attributes traffic to another afakt launch. | Provider identity assertions need a binding to the current afakt launch under [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution). | depends on an open question ([`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60)) |
| `TM-OAA-14` | `F-OAA-4`, `F-OAA-5`, `F-OAA-6`, `F-OAA-7`, `F-OAA-20`, `P-OAA-2`, `P-OAA-3`, `P-OAA-4`, `P-OAA-5`, `P-OAA-6`, `E-OAA-4` | tampering | Enabled access required by hosted stdio MCP permits traffic outside the selected mediator route. | The required exclusive mediator route has no selected provider mechanism under [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution). | gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)) |
| `TM-OAA-15` | `F-OAA-4`, `F-OAA-5`, `F-OAA-6`, `F-OAA-7`, `F-OAA-20`, `P-OAA-2`, `P-OAA-3`, `P-OAA-4`, `P-OAA-5`, `P-OAA-6`, `E-OAA-4` | repudiation | A vendor omits traffic on its network path from afakt’s decision history. | Independent accounting for vendor-side network actions remains unspecified under [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-16` | `F-OAA-4`, `F-OAA-5`, `F-OAA-6`, `F-OAA-7`, `F-OAA-20`, `P-OAA-2`, `P-OAA-3`, `P-OAA-4`, `P-OAA-5`, `P-OAA-6`, `E-OAA-4` | information disclosure | A hosted sandbox sends data to an allowed host without afakt’s request inspection or DNS policy. | Name queries and vendor egress lack a selected afakt inspection path under [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution). | gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)) |
| `TM-OAA-17` | `F-OAA-4`, `F-OAA-5`, `F-OAA-6`, `F-OAA-7`, `F-OAA-20`, `P-OAA-2`, `P-OAA-3`, `P-OAA-4`, `P-OAA-5`, `P-OAA-6`, `E-OAA-4` | denial of service | Blocked or stalled remote transport prevents control, session, or preview progress. | Control transport and its progress bounds need a provider mapping under [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution). | depends on an open question ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)) |
| `TM-OAA-18` | `F-OAA-4`, `F-OAA-5`, `F-OAA-6`, `F-OAA-7`, `F-OAA-20`, `P-OAA-2`, `P-OAA-3`, `P-OAA-4`, `P-OAA-5`, `P-OAA-6`, `E-OAA-4` | elevation of privilege | A remote MCP call or existing vault credential remains usable without afakt’s current launch checks. | Vendor tool and proxy authority lack binding to afakt admission under [Mediator connection and name resolution](workload-provider-openai-agents-api.md#Mediator%20connection%20and%20name%20resolution). | gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)) |

### Workspace transfer and processing

Session-created and standalone shared environments have different file ownership and lifetime contracts. Artifacts persist separately. See [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-OAA-19` | `F-OAA-8`, `F-OAA-9`, `F-OAA-10`, `P-OAA-1`, `P-OAA-3`, `DS-OAA-1`, `DS-OAA-3` | spoofing | A file or artifact handle selects another launch’s workspace. | File handles need ownership checks for the intended launch under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-OAA-20` | `F-OAA-8`, `F-OAA-9`, `F-OAA-10`, `P-OAA-1`, `P-OAA-3`, `DS-OAA-1`, `DS-OAA-3` | tampering | Hostile repository metadata changes how remote capture or extraction processes files. | Confinement of hostile metadata processing remains undecided under [Workspace image materialization](workload-provider-openai-agents-api.md#Workspace%20image%20materialization). | depends on an open question ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |
| `TM-OAA-21` | `F-OAA-8`, `F-OAA-9`, `F-OAA-10`, `P-OAA-1`, `P-OAA-3`, `DS-OAA-1`, `DS-OAA-3` | repudiation | A vendor changes file content without independently attributable evidence. | Attribution of vendor file changes lacks independent evidence under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-22` | `F-OAA-8`, `F-OAA-9`, `F-OAA-10`, `P-OAA-1`, `P-OAA-3`, `DS-OAA-1`, `DS-OAA-3` | information disclosure | A shared environment or retained artifact exposes files after removal of the originating session. | Storage reuse and deletion need a launch-specific ownership mapping under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-OAA-23` | `F-OAA-8`, `F-OAA-9`, `F-OAA-10`, `P-OAA-1`, `P-OAA-3`, `DS-OAA-1`, `DS-OAA-3` | denial of service | Workspace transfer or scratch growth exhausts storage before cleanup completes. | Workspace and scratch limits need a provider allocation and cleanup contract under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)) |
| `TM-OAA-24` | `F-OAA-8`, `F-OAA-9`, `F-OAA-10`, `P-OAA-1`, `P-OAA-3`, `DS-OAA-1`, `DS-OAA-3` | elevation of privilege | Hostile workspace metadata executes with vendor capture or materialization authority outside the sandbox. | The transfer and capture path lacks selected process confinement under [Workspace image materialization](workload-provider-openai-agents-api.md#Workspace%20image%20materialization). | depends on an open question ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)) |

### Shared named tool caches

No provider mapping selects isolated named caches or authorization for cache writers. See [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-OAA-25` | `F-OAA-11`, `F-OAA-12`, `P-OAA-3`, `E-OAA-2`, `DS-OAA-2` | spoofing | A cache writer presents another launch’s cache namespace as its own. | Cache namespace authorization remains undecided under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-OAA-26` | `F-OAA-11`, `F-OAA-12`, `P-OAA-3`, `E-OAA-2`, `DS-OAA-2` | tampering | A launch poisons shared dependencies that another launch executes. | Writer trust and dependency integrity need a cache policy under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-OAA-27` | `F-OAA-11`, `F-OAA-12`, `P-OAA-3`, `E-OAA-2`, `DS-OAA-2` | repudiation | A shared cache changes without a record identifying its writer. | Attribution of shared cache writes remains undecided under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-OAA-28` | `F-OAA-11`, `F-OAA-12`, `P-OAA-3`, `E-OAA-2`, `DS-OAA-2` | information disclosure | A shared cache exposes private dependency or build content to another launch. | Shared cache readership needs a confidentiality policy under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-OAA-29` | `F-OAA-11`, `F-OAA-12`, `P-OAA-3`, `E-OAA-2`, `DS-OAA-2` | denial of service | A launch fills a shared cache and prevents other launches from making progress. | Cache capacity and invalidation need allocation rules under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |
| `TM-OAA-30` | `F-OAA-11`, `F-OAA-12`, `P-OAA-3`, `E-OAA-2`, `DS-OAA-2` | elevation of privilege | A writable cache entry becomes trusted executable input in a later launch. | Cache provenance doesn't yet define trust in executable input under [Workspace, scratch, and caches](workload-provider-openai-agents-api.md#Workspace,%20scratch,%20and%20caches). | depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)) |

### Retained state and successor launch

Artifacts and pre-warmed environments don't supply a documented atomic afakt snapshot and checkpoint pair. See [Workload snapshots and agent checkpoints](workload-provider-openai-agents-api.md#Workload%20snapshots%20and%20agent%20checkpoints).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-OAA-31` | `F-OAA-13`, `F-OAA-14`, `P-OAA-2`, `P-OAA-7`, `DS-OAA-4` | spoofing | A retained-state handle restores another agent’s workspace or session. | Restore authorization needs ownership checks on retained state under [Workload snapshots and agent checkpoints](workload-provider-openai-agents-api.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)) |
| `TM-OAA-32` | `F-OAA-13`, `F-OAA-14`, `P-OAA-2`, `P-OAA-7`, `DS-OAA-4` | tampering | afakt treats an artifact from one completed turn as a checkpoint for a different managed conversation state. | The complete workspace and native-session capture pair remains undecided under [Workload snapshots and agent checkpoints](workload-provider-openai-agents-api.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60)) |
| `TM-OAA-33` | `F-OAA-13`, `F-OAA-14`, `P-OAA-2`, `P-OAA-7`, `DS-OAA-4` | repudiation | afakt lacks attribution of a vendor state change to its capture or restore requester. | Capture and restore records need attribution to their requester under [Workload snapshots and agent checkpoints](workload-provider-openai-agents-api.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)) |
| `TM-OAA-34` | `F-OAA-13`, `F-OAA-14`, `P-OAA-2`, `P-OAA-7`, `DS-OAA-4` | information disclosure | A capture preserves credentials or private session data beyond its authorized scope. | Capture scope must separate retained content from active authority under [Workload snapshots and agent checkpoints](workload-provider-openai-agents-api.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)) |
| `TM-OAA-35` | `F-OAA-13`, `F-OAA-14`, `P-OAA-2`, `P-OAA-7`, `DS-OAA-4` | denial of service | A hosted environment expires before afakt captures the files and native session state needed for restoration. | Durable ownership needs retention roots beyond vendor expiry under [Workload snapshots and agent checkpoints](workload-provider-openai-agents-api.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-11`](agent-open-questions.md#%60OQ-AGT-11%60)) |
| `TM-OAA-36` | `F-OAA-13`, `F-OAA-14`, `P-OAA-2`, `P-OAA-7`, `DS-OAA-4` | elevation of privilege | A successor restores stale authority or hostile state as trusted current execution. | Restoration needs destination authorization and exclusion of stale credentials under [Workload snapshots and agent checkpoints](workload-provider-openai-agents-api.md#Workload%20snapshots%20and%20agent%20checkpoints). | depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)) |

### Workspace command execution

Environment execution doesn't establish standalone afakt-sandbox placement, authority, or descendant supervision. See [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-OAA-37` | `F-OAA-15`, `F-OAA-16`, `F-OAA-17`, `E-OAA-1`, `P-OAA-8`, `DS-OAA-1`, `P-OAA-4` | spoofing | A command request attaches to the wrong launch or recycled sandbox identifier. | Workspace command placement and target identity remain undecided under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-OAA-38` | `F-OAA-15`, `F-OAA-16`, `F-OAA-17`, `E-OAA-1`, `P-OAA-8`, `DS-OAA-1`, `P-OAA-4` | tampering | A workspace command races a checkpoint or changes the harness’s active files. | Command ordering relative to capture remains undecided under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-OAA-39` | `F-OAA-15`, `F-OAA-16`, `F-OAA-17`, `E-OAA-1`, `P-OAA-8`, `DS-OAA-1`, `P-OAA-4` | repudiation | A command result lacks binding to the requester, launch, and executed process. | Command evidence needs requester and process attribution under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-OAA-40` | `F-OAA-15`, `F-OAA-16`, `F-OAA-17`, `E-OAA-1`, `P-OAA-8`, `DS-OAA-1`, `P-OAA-4` | information disclosure | A command inherits provider credentials or another session’s environment. | Inherited command state and credential exclusions remain undecided under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-OAA-41` | `F-OAA-15`, `F-OAA-16`, `F-OAA-17`, `E-OAA-1`, `P-OAA-8`, `DS-OAA-1`, `P-OAA-4` | denial of service | A detached command continues after cancellation and retains resources or workspace access. | Command descendants need lifecycle binding through cancellation under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |
| `TM-OAA-42` | `F-OAA-15`, `F-OAA-16`, `F-OAA-17`, `E-OAA-1`, `P-OAA-8`, `DS-OAA-1`, `P-OAA-4` | elevation of privilege | A workspace command bypasses afakt-sandbox or uses vendor execution authority beyond its launch. | The command confinement and authority boundary remains undecided under [Process placement and startup](workload-provider-openai-agents-api.md#Process%20placement%20and%20startup). | depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)) |

### Vendor custody and administration

OpenAI holds hosted workspace and managed history and can hold vault secrets. Physical cleanup may continue after deletion. See [Host observations](workload-provider-openai-agents-api.md#Host%20observations).

| Threat ID | Element | STRIDE category | Threat | Mitigation in the design | Status |
| --- | --- | --- | --- | --- | --- |
| `TM-OAA-43` | `F-OAA-18`, `F-OAA-19`, `F-OAA-21`, `E-OAA-3`, `P-OAA-2`, `P-OAA-3`, `DS-OAA-4`, `DS-OAA-5` | spoofing | A vendor administrator supplies false identity or observation data to afakt. | afakt lacks an independent check on vendor identity observations under [Host observations](workload-provider-openai-agents-api.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-44` | `F-OAA-18`, `F-OAA-19`, `F-OAA-21`, `E-OAA-3`, `P-OAA-2`, `P-OAA-3`, `DS-OAA-4`, `DS-OAA-5` | tampering | A vendor administrator changes retained workspace, session, or credential state. | Vendor custody supplies no independent integrity protection for retained content under [Host observations](workload-provider-openai-agents-api.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-45` | `F-OAA-18`, `F-OAA-19`, `F-OAA-21`, `E-OAA-3`, `P-OAA-2`, `P-OAA-3`, `DS-OAA-4`, `DS-OAA-5` | repudiation | A vendor administrator removes the evidence of privileged content access. | Evidence of vendor administrator access lacks independent protection under [Host observations](workload-provider-openai-agents-api.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-46` | `F-OAA-18`, `F-OAA-19`, `F-OAA-21`, `E-OAA-3`, `P-OAA-2`, `P-OAA-3`, `DS-OAA-4`, `DS-OAA-5` | information disclosure | An OpenAI service or privileged vendor operator discloses hosted workspace, retained conversation, artifacts, or vault credentials. | No stated control prevents disclosure by the vendor holding the content under [Host observations](workload-provider-openai-agents-api.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |
| `TM-OAA-47` | `F-OAA-18`, `F-OAA-19`, `F-OAA-21`, `E-OAA-3`, `P-OAA-2`, `P-OAA-3`, `DS-OAA-4`, `DS-OAA-5` | denial of service | Vendor deletion, retention, or account suspension makes required state unavailable. | Provider retention and account eligibility need availability rules under [Host observations](workload-provider-openai-agents-api.md#Host%20observations). | depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)) |
| `TM-OAA-48` | `F-OAA-18`, `F-OAA-19`, `F-OAA-21`, `E-OAA-3`, `P-OAA-2`, `P-OAA-3`, `DS-OAA-4`, `DS-OAA-5` | elevation of privilege | A compromised vendor service uses supplied management or upstream credentials outside the intended launch. | The vendor trust boundary doesn't limit a compromised custodian’s authority under [Host observations](workload-provider-openai-agents-api.md#Host%20observations). | gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)) |

## Attack trees

### Escape to the vendor host

```mermaid
flowchart TD
    root["Escape to vendor host"] --> alternatives{"OR"}
    alternatives --> a0["TM-OAA-12: Exploit runtime boundary"]
    alternatives --> a1["TM-OAA-7: Reach runtime management"]
    alternatives --> processing{"AND"}
    processing --> b0["TM-OAA-20: Submit hostile metadata"]
    processing --> b1["TM-OAA-24: Reach privileged vendor file processing"]
```

Runtime input already available to malicious sandbox code is the path requiring the least extra authority. The cited controls supply no containment after a runtime exploit. Reaching a privileged vendor file processor also requires that the unresolved transfer or capture path accept the hostile content.

### Read or change another agent launch’s workspace or session

```mermaid
flowchart TD
    root["Read or change another launch"] --> alternatives{"OR"}
    alternatives --> a0["TM-OAA-1: Select another vendor object"]
    alternatives --> a1["TM-OAA-22: Read reused workspace content"]
    alternatives --> a2["TM-OAA-26: Poison shared cache"]
    alternatives --> a3["TM-OAA-31: Select another retained capture"]
```

A shared environment, mount, cache, or broad object handle needs fewer capabilities than a runtime escape. Launch-specific storage and authorization mappings remain unresolved. Separate sandbox filesystems only block direct file access when there is no shared storage or management authority.

### Send traffic past egress policy

```mermaid
flowchart TD
    root["Bypass egress policy"] --> alternatives{"OR"}
    alternatives --> a0["TM-OAA-14: Use alternate vendor network path"]
    alternatives --> a1["TM-OAA-16: Disclose through DNS or permitted host"]
    alternatives --> a2["TM-OAA-18: Use vendor tool or proxy authority"]
```

A malicious sandbox process can use a documented vendor network or tool path directly. Mandatory afakt mediation remains unresolved because destination allowlists leave the required carrier and request policy unspecified.

### Disclose workspace or credentials through the vendor

```mermaid
flowchart TD
    root["Vendor-side disclosure"] --> alternatives{"OR"}
    alternatives --> a["TM-OAA-46: Read workspace or secret custody"]
    alternatives --> conceal{"AND"}
    conceal --> b0["TM-OAA-44: Alter retained state"]
    conceal --> b1["TM-OAA-45: Remove access evidence"]
```

A privileged vendor service already has custody of submitted data. The cited controls supply no independent barrier to its disclosure. The concealment branch requires both changing retained state and removing access evidence.

## Residual risks and gaps

Provider-specific gaps don't settle the on-hold choices for hook protection, repository project configuration, profile self-grants and credential domains, host isolation, or session-end wait and identity lifetime. See [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60), [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60), [`OQ-PRF-5`](agent-open-questions.md#%60OQ-PRF-5%60), [`OQ-SBX-2`](agent-open-questions.md#%60OQ-SBX-2%60), [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60), and [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

### Remote lifecycle and harness control

- `TM-OAA-1`. A management caller selects another launch through a reused vendor object identifier. depends on an open question ([`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60)).
- `TM-OAA-2`. A template, tool configuration, or shared environment change breaks the launch’s pinned execution contract. depends on an open question ([`OQ-SBX-63`](agent-open-questions.md#%60OQ-SBX-63%60)).
- `TM-OAA-3`. A vendor service supplies a false startup, cancellation, or completion history. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-4`. A broad provider credential exposes another launch’s prompts, files, or session records. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-5`. Account limits or unavailable remote capacity prevent startup or reconciliation. depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)).
- `TM-OAA-6`. Session deletion leaves a shared environment running while afakt starts a replacement launch. depends on an open question ([`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60)).

### Sandbox runtime and vendor host

- `TM-OAA-7`. untrusted execution impersonates a runtime management caller. gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)).
- `TM-OAA-8`. A session changes files belonging to another launch attached to the same pre-warmed environment. depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-OAA-9`. A runtime compromise suppresses evidence of activity on the vendor host. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-10`. A runtime exploit exposes another sandbox’s memory or private workspace. gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)).
- `TM-OAA-11`. A launch consumes shared runtime resources needed by other launches. depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)).
- `TM-OAA-12`. Hostile system calls or virtual-device input escape the runtime into the vendor host. gap ([`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60)).

### Workload proxy mediator and vendor network paths

- `TM-OAA-13`. A vendor-side endpoint attributes traffic to another afakt launch. depends on an open question ([`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60)).
- `TM-OAA-14`. Enabled access required by hosted stdio MCP permits traffic outside the selected mediator route. gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)).
- `TM-OAA-15`. A vendor omits traffic on its network path from afakt’s decision history. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-16`. A hosted sandbox sends data to an allowed host without afakt’s request inspection or DNS policy. gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)).
- `TM-OAA-17`. Blocked or stalled remote transport prevents control, session, or preview progress. depends on an open question ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)).
- `TM-OAA-18`. A remote MCP call or existing vault credential remains usable without afakt’s current launch checks. gap ([`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60)).

### Workspace transfer and processing

- `TM-OAA-19`. A file or artifact handle selects another launch’s workspace. depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-OAA-20`. Hostile repository metadata changes how remote capture or extraction processes files. depends on an open question ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).
- `TM-OAA-21`. A vendor changes file content without independently attributable evidence. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-22`. A shared environment or retained artifact exposes files after removal of the originating session. depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-OAA-23`. Workspace transfer or scratch growth exhausts storage before cleanup completes. depends on an open question ([`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60)).
- `TM-OAA-24`. Hostile workspace metadata executes with vendor capture or materialization authority outside the sandbox. depends on an open question ([`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60)).

### Shared named tool caches

- `TM-OAA-25`. A cache writer presents another launch’s cache namespace as its own. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-OAA-26`. A launch poisons shared dependencies that another launch executes. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-OAA-27`. A shared cache changes without a record identifying its writer. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-OAA-28`. A shared cache exposes private dependency or build content to another launch. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-OAA-29`. A launch fills a shared cache and prevents other launches from making progress. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).
- `TM-OAA-30`. A writable cache entry becomes trusted executable input in a later launch. depends on an open question ([`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60)).

### Retained state and successor launch

- `TM-OAA-31`. A retained-state handle restores another agent’s workspace or session. depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)).
- `TM-OAA-32`. afakt treats an artifact from one completed turn as a checkpoint for a different managed conversation state. depends on an open question ([`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60)).
- `TM-OAA-33`. afakt lacks attribution of a vendor state change to its capture or restore requester. depends on an open question ([`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60)).
- `TM-OAA-34`. A capture preserves credentials or private session data beyond its authorized scope. depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)).
- `TM-OAA-35`. A hosted environment expires before afakt captures the files and native session state needed for restoration. depends on an open question ([`OQ-AGT-11`](agent-open-questions.md#%60OQ-AGT-11%60)).
- `TM-OAA-36`. A successor restores stale authority or hostile state as trusted current execution. depends on an open question ([`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60)).

### Workspace command execution

- `TM-OAA-37`. A command request attaches to the wrong launch or recycled sandbox identifier. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-OAA-38`. A workspace command races a checkpoint or changes the harness’s active files. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-OAA-39`. A command result lacks binding to the requester, launch, and executed process. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-OAA-40`. A command inherits provider credentials or another session’s environment. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-OAA-41`. A detached command continues after cancellation and retains resources or workspace access. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).
- `TM-OAA-42`. A workspace command bypasses afakt-sandbox or uses vendor execution authority beyond its launch. depends on an open question ([`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60)).

### Vendor custody and administration

- `TM-OAA-43`. A vendor administrator supplies false identity or observation data to afakt. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-44`. A vendor administrator changes retained workspace, session, or credential state. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-45`. A vendor administrator removes the evidence of privileged content access. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-46`. An OpenAI service or privileged vendor operator discloses hosted workspace, retained conversation, artifacts, or vault credentials. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
- `TM-OAA-47`. Vendor deletion, retention, or account suspension makes required state unavailable. depends on an open question ([`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60)).
- `TM-OAA-48`. A compromised vendor service uses supplied management or upstream credentials outside the intended launch. gap ([`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60)).
