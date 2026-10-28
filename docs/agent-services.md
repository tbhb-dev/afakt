---
type: Design
title: Agent services
description: Published service names, port mappings, access, and lifetime across agent launches.
tags: [afakt, agents, services, ports, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent services

[Public key infrastructure](pki.md).

An agent service publishes an endpoint through its [workload proxy](agent-workload-proxy.md). The [published-port path](agent-open-questions.md#%60OQ-PORT-14%60) remains open. Whether services and [previews](agent-previews.md) use one model and lifecycle or separate models remains [undecided](agent-open-questions.md#%60OQ-HUB-18%60).

## Published ports and services

An agent can publish a Transmission Control Protocol (TCP) listener on guest loopback and grant another agent access. That agent can map the publication to its own loopback listener. An authorized caller can map a publication of guest port 3000. A service can map that guest port to service port 443. The port number alone doesn't select [Transport Layer Security (TLS) termination or certificate ownership](agent-open-questions.md#%60OQ-HUB-10%60). The [publication operations](agent-open-questions.md#%60OQ-HUB-7%60) and [host-mode mapping](agent-open-questions.md#%60OQ-HUB-8%60) remain undecided.

A service remains after its current launch ends and can point to a successor. With no running launch, the service returns Hypertext Transfer Protocol (HTTP) status 503. [Ownership transfer and in-flight streams](agent-open-questions.md#%60OQ-HUB-11%60) remain undecided. Publication and listener checks follow [agent ports and previews](agent-ports.md#Port%20discovery%20and%20records).

## Service names

A service name can remain after a launch ends and move to its successor. Internal service Domain Name System (DNS) names use this target.

| Name | Target |
| --- | --- |
| `<service_name>.<namespace>.svc.afakt.internal` | A published service |

Service names use the [shared namespace and resolution rules](agent-overlay.md#Names).

## Access

[Connections between agents](agent-ports.md#Connections%20between%20agents) require source network permission and a destination publication grant. [Agent host access](agent-overlay.md#Agent%20host%20access) uses host-level authorization. Neither name resolution nor overlay reachability grants access. See the [grant flow](network-topology.md#Launch%20to%20launch%20under%20a%20grant) for both launch boundaries.

A service name can remain after a launch ends without reopening its closed publication. Binding it to a successor requires that launch's authorized publication under [Port discovery and records](agent-ports.md#Port%20discovery%20and%20records). The shared resource and transfer rules remain [`OQ-HUB-18`](agent-open-questions.md#%60OQ-HUB-18%60) and [`OQ-HUB-11`](agent-open-questions.md#%60OQ-HUB-11%60).
