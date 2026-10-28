---
type: Design
title: Workload overlay and names
description: Workload overlay nodes, direct connections, stable agent names, and ingress access.
tags: [afakt, agents, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload overlay and names

[Public key infrastructure](pki.md).

Every workload's [workload proxy](workload-proxy.md) is its own overlay node. The ingress gateway and other workloads dial it directly. Both sides authenticate with SPIFFE mutual Transport Layer Security (TLS), and mediators at both ends check each connection. Neither name resolution nor overlay reachability grants access. See the [cluster topology](network-topology.md#Cluster).

## Names

The internal Domain Name System (DNS) zone is `afakt.internal`. Stable agent names are DNS records that the control plane updates to point to the new workload when an agent moves. Launch names identify one launch on a workload host. The [namespace](agent-open-questions.md#%60OQ-HUB-2%60), [host-label](agent-open-questions.md#%60OQ-MOD-7%60), and [inactive-name](agent-open-questions.md#%60OQ-HUB-3%60) contracts remain undecided.

| Name | Target |
| --- | --- |
| `<agent_id>.<namespace>.agents.afakt.internal` | The agent's active workload proxy |
| `<launch_id>.<host>.hosts.afakt.internal` | One launch on a workload host |
| `<host>.hosts.afakt.internal` | The workload host |
| `<agent_id>-<protocol>-<port>.<namespace>.ports.afakt.internal` | An agent's published port |

[Service names](workload-services.md#Service%20names) and [internal preview names](workload-previews.md#Internal%20preview%20names) use the same internal zone.

## Launch-to-launch traffic

The source workload proxy dials the destination workload proxy directly over the overlay. Both sides authenticate with SPIFFE mutual TLS. At the source, the mediator checks the firewall's egress rules for the destination name and port. At the destination, it checks ingress rules and the current publication's grant to the source. A destination proxy accepts only the ingress gateway and the peers its ingress rules name, and only for the ports those rules publish. The gateway never bypasses ingress rules. Overlay ACLs supply coarse reachability and never grant port access.

Traffic between launches on the same workload host uses the overlay like traffic between hosts. Launch data paths exclude the worker. The [policy engine](agent-open-questions.md#%60OQ-SBX-61%60) remains undecided. Beyond these connection requirements, the [published-port path](agent-open-questions.md#%60OQ-PORT-14%60) remains open.

Cutting off an agent's traffic requires action at more than one point. Overlay node or Access Control List (ACL) revocation together with a policy push cuts off access. Rule changes end existing connections the new rules deny. [Persistent guest egress and User Datagram Protocol (UDP) revocation](agent-open-questions.md#%60OQ-SBX-69%60) remain open.

## Workload host access

Each workload host has its workload host worker, host-level DNS resolution of `afakt.internal` names, and connectivity to workload ports, previews, services, and cluster services. Resolution doesn't grant access. Requests use host-level authorization, with any further authentication supplied by the accessed app. Separate from each launch's workload proxy, the host access role never handles guest egress. See the [host topology](network-topology.md#Workload%20host) for that authority boundary.

The mechanism for host connectivity, local caller authentication, resolver ownership, and certificate trust remains [undecided](agent-open-questions.md#%60OQ-HST-9%60). A host-level network service names that connectivity role without selecting a process or proxy. Host-mode agents and the worker use separate dedicated users. This separation doesn't settle the broader host isolation baseline.

## Overlay transport

The cluster overlay uses Tailscale's userspace client library, `tsnet`, against an isolated Headscale. Each workload proxy, workload host, and control plane service has its own overlay node. Alongside coordination and relay nodes, the API server, preview server, and Model Context Protocol (MCP) gateway are separate service roles. An instance of a role is a separate node with that role's tag. These nodes appear in the [cluster picture](network-topology.md#Cluster). Coordination [instance counts](agent-open-questions.md#%60OQ-HUB-1%60) remain undecided.

Internal service traffic uses the overlay. User-facing and public app traffic passes through the external ingress gateway. For coordination outages, every node stores a static map of stable nodes. That map supplies addresses without bypassing service authentication or policy. Availability, node churn, policy propagation, resource costs, certificate behavior, and tailnet membership behavior remain [undecided](agent-open-questions.md#%60OQ-HST-10%60).

## Coordination interface

The control plane's own interface fronts the coordination server. Against a presented Secure Production Identity Framework for Everyone (SPIFFE) Verifiable Identity Document (SVID), that interface registers and deletes nodes. The interface also uses a presented SVID when issuing single-use tagged pre-authentication keys and setting tags and ACLs. No launch capabilities or port access come from the coordination server. Launch identity, role membership, and service authorization remain distinct checks. [Service authorization](agent-open-questions.md#%60OQ-SBX-43%60) and [host revocation](agent-open-questions.md#%60OQ-HST-6%60) govern admission and revocation questions.

## External ingress gateway

The external ingress gateway uses Traefik with authentication middleware supplied by afakt. It authenticates user-facing requests and routes them to the API server and preview server, and also to workload proxies. Its separate preview listener follows [preview access controls](workload-previews.md#Preview%20routing%20and%20names). How the authentication middleware integrates with Traefik remains [`OQ-SES-13`](agent-open-questions.md#%60OQ-SES-13%60). Process placement and supervision remain [`OQ-SES-7`](agent-open-questions.md#%60OQ-SES-7%60).

## User tailnet attachments

The external ingress gateway can attach to user tailnets, and a workload host's host-level node can optionally attach. Per-launch nodes never join user tailnets. As a proxy for app requests, the gateway never routes packets between the cluster overlay and a user tailnet. afakt authenticates the user even when tailnet access permits network reachability. Whether tailnet identity is an [extra authorization signal](agent-open-questions.md#%60OQ-HUB-17%60) is undecided.

A user creates the service in the Tailscale administration console and tags the gateway node during attachment setup. Phones, tablets, native apps, and browsers connect to afakt through a MagicDNS uniform resource locator (URL) such as `https://afakt.<tailnet>.ts.net`. Both network boundaries appear in the [phone-to-dashboard flow](network-topology.md#Phone%20through%20a%20user%20tailnet%20to%20the%20dashboard). Attachment mechanics and certificates remain [undecided](agent-open-questions.md#%60OQ-HST-10%60).
