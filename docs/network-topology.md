---
type: Architecture
title: Network topology
description: Cluster, workload host, and agent launch networks with trust boundaries and authenticated traffic flows.
tags: [afakt, architecture, agents, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Network topology

[Public key infrastructure](pki.md).

The cluster connects control plane services and workload hosts through a private overlay. Each workload has its own workload proxy mediator and workload proxy. The guest has no overlay membership. Its proxy transmits traffic in both directions and requests a mediator decision for each connection. Workload proxies connect directly over the overlay with SPIFFE mutual Transport Layer Security (TLS). Client ACP routing, session ownership, and published-port access remain open under [`OQ-SES-17`](agent-open-questions.md#%60OQ-SES-17%60), [`OQ-SES-18`](agent-open-questions.md#%60OQ-SES-18%60), and [`OQ-PORT-14`](agent-open-questions.md#%60OQ-PORT-14%60).

## Context

![Network context](diagrams/structurizr-network-context.svg)

The operator uses browsers, phones, tablets, or native apps to reach the cluster. Public app requests enter through the external ingress gateway. A user can attach their own tailnets to that gateway for private app access. The user tailnet and cluster overlay remain separate trust zones: the gateway proxies app requests and never routes packets between them. Network access depends on a device's identity on the tailnet. Signing in to afakt supplies the user's app identity. Internet services remain outside both zones and receive only permitted destination requests.

User access follows [session authentication](agent-acp.md#The%20protocol%20endpoint) and [preview access](workload-previews.md#Access%20and%20share%20links). Authentication across service hops depends on [user-context binding](agent-open-questions.md#%60OQ-SES-10%60). Tailnet attachment and overlay behavior depend on [overlay qualification](agent-open-questions.md#%60OQ-HST-10%60).

## Cluster

![Cluster topology](diagrams/structurizr-network-cluster.svg)

The cluster overlay uses Tailscale's userspace client library, `tsnet`, against an isolated Headscale. Each workload proxy, workload host, and control plane service instance has its own node. Instances of one service share a tag but have separate node identities. Within the control plane are an Application Programming Interface (API) server, a preview server, a Model Context Protocol (MCP) gateway, and coordination and relay nodes. From its own origin, the API service serves the web dashboard. Headscale coordination sits behind afakt's own interface, which registers nodes and issues single-use tagged pre-authentication keys against a presented workload identity. That interface also sets tags and access rules and deletes nodes. Coordination [instance counts](agent-open-questions.md#%60OQ-HUB-1%60) remain undecided. Each cluster node stores a static map of stable nodes for coordination outages. That map supplies addresses without granting permission to bypass authentication or policy.

The external ingress gateway uses Traefik with afakt authentication middleware for public and user-facing app traffic, under [External ingress gateway](workload-overlay.md#External%20ingress%20gateway). Internal services connect to one another over the overlay without a separate internal gateway. User-tailnet attachment is available to the gateway and optionally a workload host's host-level node. Launch nodes never join user tailnets, and tailnet attachment doesn't select a mechanism for one node to join more than one tailnet. Overlay access control lists (ACLs) bound reachability. Each workload proxy accepts only the ingress gateway and peers named by its ingress rules. Mediators enforce grants for individual launch-to-launch connections.

![Control and transport service boundaries](diagrams/structurizr-network-service-boundaries.svg)

Control plane services authenticate workload connections using mutual Transport Layer Security (TLS) and service identities. A workload proxy stores its launch's X.509 SPIFFE Verifiable Identity Document (SVID), outside the guest. Host access uses the host identity. The ingress gateway and permitted workload peers dial the workload proxy directly. The preview server also dials that proxy over the overlay. Its published-port access depends on [`OQ-PORT-14`](agent-open-questions.md#%60OQ-PORT-14%60). The mediator's own control plane connections use the proxy under a fixed rule set built into the workload. Pushes arrive on a stream opened by the mediator, so the control plane never opens a connection to a workload.

[Workload identity](workload-host.md#Workload%20identity), [service authorization](agent-open-questions.md#%60OQ-SBX-43%60), and [tool access](agent-mcp-gateway.md#Server%20and%20tool%20access) define the authority at these boundaries. [Overlay qualification](agent-open-questions.md#%60OQ-HST-10%60) leaves operational properties undecided. [Client connection ownership](agent-open-questions.md#%60OQ-SES-17%60) leaves remote-control and session routing responsibilities undecided. The [policy engine](agent-open-questions.md#%60OQ-SBX-61%60) is undecided.

## Workload host

![Workload host topology](diagrams/structurizr-network-host.svg)

The workload host worker manages launch lifecycle as a dedicated non-root user. Host-mode launches use a separate dedicated agent user. A privileged helper crosses the setup and launch authority boundary. Each workload host supplies Domain Name System (DNS) resolution of `afakt.internal` and connectivity to workload ports, previews, services, and cluster services. Resolution doesn't grant access: requests require host-level authorization. The host-level network service names this connectivity function. Its mechanism and resolver ownership remain [undecided](agent-open-questions.md#%60OQ-HST-9%60).

Each agent launch has an independent guest, mediator, workload proxy, and launch identity. Launches have no implicit access to one another. As the shared lifecycle controller, the worker stays out of every launch's data path. Operator and host-management requests pass through the host-level network service, which never handles guest egress. User selection for the mediator and workload proxy remains [undecided](agent-open-questions.md#%60OQ-HST-8%60). Traffic between launches on the same host uses the overlay, with no local shortcut.

[Host responsibilities](workload-host.md#What%20runs%20on%20a%20workload%20host) and [runtime privileges](workload-host.md#Runtime%20privileges%20and%20platforms) supply lifecycle context. Helper caller binding depends on [helper authentication](agent-open-questions.md#%60OQ-SBX-27%60). Local caller checks, name routing, and host-client certificate trust depend on [host access](agent-open-questions.md#%60OQ-HST-9%60). The broader [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60) remains on hold.

## Agent launch

![Agent launch topology](diagrams/structurizr-network-launch.svg)

The agent launch contains three separate trust zones. These are untrusted guest code, the workload proxy mediator that enforces policy, and the workload proxy that holds network access. Over a private link, the guest sends traffic to its proxy. It receives neither the launch SVID nor its private key. The mediator stamps launch identity and decides connections without transmitting traffic. It has no independent network access. Using the launch SVID for authentication, the proxy opens permitted service and internet connections and accepts permitted overlay peers. Each receiving service authenticates as its own workload.

The [guest layout](workload-sandbox.md#Guest%20layout%20and%20lifecycle%20control), [control requests](workload-proxy-mediator.md#Control%20requests), and [launch authentication](workload-proxy.md#Agent%20launch%20authentication) define the guest and service authority. Local carrier authentication depends on [guest-link protection](agent-open-questions.md#%60OQ-SBX-13%60). Workload provider capture depends on [mandatory mediation](agent-open-questions.md#%60OQ-SBX-54%60). Enforcement of the mediator's proxy-only connections on each operating system is [undecided](agent-open-questions.md#%60OQ-SBX-62%60). That rule doesn't select a service user or settle the isolation baseline for workload hosts.

## Guest to the internet

![Guest internet flow](diagrams/structurizr-network-flow-internet.svg)

The [workload proxy](workload-proxy.md#Internet%20connections) opens internet connections permitted by the workload proxy mediator.

[Network policy](workload-proxy-mediator.md#Network%20policy) and [credential custody](workload-sandbox.md#Credential%20types%20and%20custody) bound the request. The [policy engine](agent-open-questions.md#%60OQ-SBX-61%60), [destination classification](agent-open-questions.md#%60OQ-SBX-24%60), [upstream name resolution](agent-open-questions.md#%60OQ-SBX-21%60), and [mediator network confinement](agent-open-questions.md#%60OQ-SBX-62%60) remain undecided.

## Guest to a control plane service

![Guest control plane flow](diagrams/structurizr-network-flow-control.svg)

The guest sends a control request through its workload proxy. The mediator evaluates the request and stamps the launch identity without transmitting traffic. Over the overlay, the workload proxy opens a workload-authenticated connection to the API service. From its peer certificate, the service reads the launch identity and checks capabilities and scopes against the immutable launch spec. The guest can't substitute another launch identity. Calls to the MCP gateway cross the same guest, mediator, workload-proxy, and service boundaries with the gateway's tool checks.

[Control requests](workload-proxy-mediator.md#Control%20requests), [capabilities](agent-profile.md#Capabilities), and [tool access](agent-mcp-gateway.md#Server%20and%20tool%20access) supply the request contract. [Local sender binding](agent-open-questions.md#%60OQ-SES-6%60), [service admission](agent-open-questions.md#%60OQ-SBX-43%60), and the [policy engine](agent-open-questions.md#%60OQ-SBX-61%60) remain undecided. [Session-end wait and SVID lifetime](agent-open-questions.md#%60OQ-HOOK-18%60) remain on hold.

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

## Operator browser to a preview

![Host browser preview flow](diagrams/structurizr-network-flow-preview.svg)

The preview request follows the [host browser access rules](workload-previews.md#Host%20browser%20access).

## Phone through a user tailnet to the dashboard

![Phone dashboard flow](diagrams/structurizr-network-flow-phone.svg)

The phone connects through its user tailnet to the external ingress gateway at a MagicDNS name such as `https://afakt.<tailnet>.ts.net`. Hypertext Transfer Protocol Secure (HTTPS) protects the app connection. Tailnet device identity permits network reachability, while afakt sign-in identifies the user. The gateway forwards authenticated app requests to the API service's web dashboard over the cluster overlay. Its tailnet attachment never becomes a packet route into that overlay. Tablets and native apps use the same user-facing boundary, and launch nodes remain outside the user's tailnet.

[Dashboard access](components.md#Web%20dashboard) and [user authentication](agent-acp.md#The%20protocol%20endpoint) define app authority. [User-context binding](agent-open-questions.md#%60OQ-SES-10%60) and [overlay qualification](agent-open-questions.md#%60OQ-HST-10%60) leave authentication propagation, certificate behavior, and attachment details undecided.

## Launch to launch under a grant

![Launch grant flow](diagrams/structurizr-network-flow-grant.svg)

From the guest, traffic goes to the source proxy. Before that proxy dials the destination proxy directly over the overlay, its mediator checks the firewall's egress rules for the destination name and port. Both sides authenticate with SPIFFE mutual TLS. At the destination, the mediator checks ingress rules, the current publication, and the source grant. The [published-port path](agent-open-questions.md#%60OQ-PORT-14%60) from proxy to listener remains open. Overlay ACLs supply only coarse reachability. Neither a listener nor overlay access grants permission to use a port.

[Connections between agents](workload-ports.md#Connections%20between%20agents) defines the two-sided authorization. The [policy engine](agent-open-questions.md#%60OQ-SBX-61%60), [local carrier binding](agent-open-questions.md#%60OQ-SBX-13%60), and [stream retirement](agent-open-questions.md#%60OQ-SBX-40%60) remain undecided. Same-host launch traffic uses the overlay too. Neither the worker nor the host-level network service transmits launch traffic.
