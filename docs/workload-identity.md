---
type: Design
title: Workload identity
description: SPIFFE identities, authority by workload type, and the separate identity of each overlay node.
tags: [afakt, workload, identity, security]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T23:55:04Z" }
---

# Workload identity

afakt uses Secure Production Identity Framework for Everyone (SPIFFE) for workload identity. Each control plane has one trust domain, without federation between control planes. A SPIFFE Verifiable Identity Document (SVID) proves an identity. The receiver checks authorization separately. See [Workload identity keys](pki-workload.md) and [Service authorization](agent-open-questions.md#%60OQ-SBX-43%60).

SPIFFE's workload includes services, gateways, and host workers. A [Workload resource](workload-resource.md) represents an execution context for an agent, app, or worker running user code. The control plane's own deployments, including its platform Temporal worker, have service identities without becoming Workload resources. See [Where workers run](https://github.com/tbhb-dev/afakt-drafts/blob/18226efca53614bf2e8ac7151f606ff514997c95/workflows/wf-design/DESIGN.md#D.10%20Where%20workers%20run).

These boundaries apply to `linux-host`, `linux-vm`, and `docker-engine` on `linux/amd64` and `linux/arm64`, `macos-host` and `macos-vm` on `darwin/arm64`, and `windows-host` and `windows-vm` on `windows/amd64`. See the [Tier 1 matrix](https://github.com/tbhb-dev/afakt-drafts/blob/18226efca53614bf2e8ac7151f606ff514997c95/workloads/workload-platform-support.md). Host-provider custody depends on the held [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## SVIDs and workload types

Connections from a workload host to the control plane use mutual Transport Layer Security (TLS) with X.509-SVIDs. Host-side consumers get identity material through a local SPIFFE Workload API endpoint, which guests can't access. SVIDs and private keys stay outside guests. The SPIFFE Workload API supplies identity material, while the [workload management API](workload-api.md) manages resources.

Under the proposal, each workload has its own SVID and gains an agent identity when an agent launch binds. Its scope and the binding between identities remain [`OQ-WLD-8`](agent-open-questions.md#%60OQ-WLD-8%60), following [workload design question 6](https://github.com/tbhb-dev/afakt-drafts/blob/18226efca53614bf2e8ac7151f606ff514997c95/workloads/workload-design.md#Open%20questions). Creation and claims remain under [`OQ-WLD-1`](agent-open-questions.md#%60OQ-WLD-1%60).

| Workload type | Authority and open scope |
| --- | --- |
| `agent` | The proposal limits the workload's own SVID to registration, health, and claims. For a bound launch, the mediator stamps launch identity and the workload proxy authenticates with the launch's X.509-SVID. Whether bind supplies another SVID, and which component receives it, remain open. |
| `app` | An app never binds an agent. Its workload role bounds its access. Attaching `afakt/api` gives the app an identity that can call the afakt API. Its permissions come from the attachment and role. Whether that identity is a separate SVID, and whether an app without the add-on has any afakt API access, remain open. |
| `worker` | A worker running user code never binds an agent. The proposal gives it egress under its workload role. Which operations its SVID permits, including service calls and credential exchanges, remain open. |

Registration-only authority can't supply the egress that apps and workers need without a bind. [App requirements](https://github.com/tbhb-dev/afakt-drafts/blob/18226efca53614bf2e8ac7151f606ff514997c95/workloads/workloads-for-apps.md#What%20doesn%27t%20fit%20as%20it%20stands) and workload design question 6 leave that conflict under [`OQ-WLD-8`](agent-open-questions.md#%60OQ-WLD-8%60). Workloads pin their Identity and Access Management (IAM) role and policy revisions, and each agent launch has an IAM role of its own. Composition remains [`OQ-CON-15`](agent-open-questions.md#%60OQ-CON-15%60). Credential grants remain held under [`OQ-PRF-5`](agent-open-questions.md#%60OQ-PRF-5%60).

## Overlay node identity

Each workload proxy has its own overlay node separate from the workload host and every other workload. Its key stays outside the workload sandbox. Node identity establishes overlay membership without granting permission to perform an app operation. Peers authenticate with SPIFFE mutual TLS, and mediators check connection policy. See [Overlay connections](workload-overlay.md#Launch-to-launch%20traffic).

Workload hosts and control plane service instances also have separate nodes. Through its overlay interface, the control plane checks a presented SVID before permitting node enrollment. Enrollment, service tags, and access to a service are distinct checks. See [Overlay transport](workload-overlay.md#Overlay%20transport) and [Overlay credentials](pki-transport.md#Overlay%20credentials).

Agent-hub routes in the identity proposal conflict with the per-workload proxy topology in [Workload firewall and ingress](https://github.com/tbhb-dev/afakt-drafts/blob/18226efca53614bf2e8ac7151f606ff514997c95/workloads/workload-design.md#Workload%20firewall%20and%20ingress). Permitted peers and the ingress gateway connect directly to the workload proxy. The workload SVID's relationship to the launch and overlay node remains [`OQ-WLD-8`](agent-open-questions.md#%60OQ-WLD-8%60).

## Issuance and retirement

afakt uses SPIFFE. The issuer and its ownership remain [`OQ-SBX-33`](agent-open-questions.md#%60OQ-SBX-33%60). Attestation and registration remain [`OQ-SBX-30`](agent-open-questions.md#%60OQ-SBX-30%60). The path scheme for workloads, launches, hosts, and services remains [`OQ-MOD-9`](agent-open-questions.md#%60OQ-MOD-9%60). An illustrative path doesn't establish a permission or prove that a caller belongs to a workload.

[S-086](https://github.com/tbhb-dev/afakt-spikes/blob/14d4d7aa84645e6e1a94634ae3ccbc79ed6f53c4/s-086-what-is-the-spiffe-workload-at-each-tier-host-component/REPORT.md) examines identity boundaries and attestation. Its SPIFFE Runtime Environment (SPIRE) fixture created registration entries but didn't get a workload SVID. [S-093](https://github.com/tbhb-dev/afakt-spikes/blob/14d4d7aa84645e6e1a94634ae3ccbc79ed6f53c4/s-093-control-cell-without-spire-an-afakt-run-ca-issuing-x509-and/REPORT.md) issued and rotated SVIDs in a synthetic issuer fixture. Issuer selection and issuance across Tier 1 cells remain open after these experiments.

S-093 also found that existing connections can survive credential expiry or trust retirement. Credential expiry alone doesn't end their authority. Active retirement remains [`OQ-SBX-40`](agent-open-questions.md#%60OQ-SBX-40%60), trust distribution remains [`OQ-SBX-41`](agent-open-questions.md#%60OQ-SBX-41%60), and delivery remains [`OQ-SBX-47`](agent-open-questions.md#%60OQ-SBX-47%60). Session-end wait and SVID lifetime remain held together under [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).
