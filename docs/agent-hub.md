---
type: Design
title: Agent hub
description: Per-launch connections for remote control, agent sessions, published ports, services, and previews over the cluster overlay.
tags: [afakt, agents, network]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent hub

[Public key infrastructure](pki.md).

The agent hub is the control plane service that joins a launch's connection to authorized callers for remote control, agent sessions, published ports, services, and previews. Each agent launch has its own connection and identity. The hub checks the caller and target. Enforcement of the launch's policy belongs to the agent mediator. Neither name resolution nor overlay reachability grants access. See the [cluster topology](network-topology.md#Cluster) for the services and their trust boundaries.

[Per-launch services](agent-proxy.md#Launch%20boundary). [Hub connection](agent-proxy.md#Hub%20connection).

## Names

The internal Domain Name System (DNS) zone is `afakt.internal`. Stable agent names resolve toward the active launch through the agent hub. Launch names identify one launch on an agent host. The [namespace](agent-open-questions.md#%60OQ-HUB-2%60), [host-label](agent-open-questions.md#%60OQ-MOD-7%60), and [inactive-name](agent-open-questions.md#%60OQ-HUB-3%60) contracts remain undecided.

| Name | Target |
| --- | --- |
| `<agent_id>.<namespace>.agents.afakt.internal` | The agent's active launch through the agent hub |
| `<launch_id>.<host>.hosts.afakt.internal` | One launch on an agent host |
| `<host>.hosts.afakt.internal` | The agent host |
| `<agent_id>-<protocol>-<port>.<namespace>.ports.afakt.internal` | An agent's published port through the agent hub |
| `agents.default.svc.afakt.internal` | The agent hub |

[Service names](agent-services.md#Service%20names) and [internal preview names](agent-previews.md#Internal%20preview%20names) use the same internal zone.

## Remote control

The Application Programming Interface (API) server exposes remote control at `/api/v1/agents/{agent_id}/remote`. At the agent hub, remote control uses `/api/v1/remote` with a WebSocket upgrade. Through the hub and that launch's connection, the API server reaches the target launch. [Agent Client Protocol (ACP) endpoints and session routing](agent-acp.md#ACP%20through%20the%20agent%20hub) share this connection.

[Published ports and services](agent-services.md#Published%20ports%20and%20services).

## Launch-to-launch traffic

The source agent mediator checks the source launch's network policy for the destination name and port, while the agent mediator at the destination checks that the current publication grants access to the source. When joining their launch connections, the agent hub checks the authenticated caller and target. Overlay access control lists (ACLs) provide coarse reachability. They never grant port access. Both launch boundaries appear in the [grant flow](network-topology.md#Launch%20to%20launch%20under%20a%20grant).

Traffic between launches on the same agent host uses the overlay like traffic between hosts. The worker never joins launch data paths. A [direct overlay path](agent-open-questions.md#%60OQ-PORT-8%60) remains undecided. It doesn't create a local shortcut. Selection of the [policy engine](agent-open-questions.md#%60OQ-SBX-61%60) that evaluates connection and message rules remains undecided.

[Preview routing and names](agent-previews.md#Preview%20routing%20and%20names).

## Agent host access

Each agent host has its agent host worker, host-level DNS resolution of `afakt.internal` names, and connectivity to agent ports, previews, services, and cluster services. Resolution doesn't grant access. Requests use host-level authorization, with any further authentication supplied by the accessed app. Separate from each launch's agent proxy, the host access role never handles guest egress. See the [host topology](network-topology.md#Agent%20host) for that authority boundary.

The mechanism for host connectivity, local caller authentication, resolver ownership, and certificate trust remains [undecided](agent-open-questions.md#%60OQ-HST-9%60). A host-level network service names that connectivity role without selecting a process or proxy. Host-mode agents and the worker use separate dedicated users. This separation doesn't settle the broader host isolation baseline.

## Overlay transport

The cluster overlay uses Tailscale's userspace client library, `tsnet`, against an isolated Headscale. Each agent proxy, agent host, and control plane service has its own overlay node. Alongside coordination and relay nodes, the agent hub, API server, preview server, and Model Context Protocol (MCP) gateway are separate service roles. An instance of a role is a separate node with that role's tag. These nodes appear in the [cluster picture](network-topology.md#Cluster). Hub and coordination [instance counts](agent-open-questions.md#%60OQ-HUB-1%60) remain undecided.

Internal service traffic uses the overlay. User-facing and public app traffic passes through the external ingress gateway. For coordination outages, every node stores a static map of stable nodes. That map supplies addresses without bypassing service authentication or policy. Availability, node churn, policy propagation, resource costs, certificate behavior, and tailnet membership behavior remain [undecided](agent-open-questions.md#%60OQ-HST-10%60).

## Coordination interface

The control plane's own interface fronts the coordination server. Against a presented Secure Production Identity Framework for Everyone (SPIFFE) Verifiable Identity Document (SVID), that interface registers and deletes nodes. The interface also uses a presented SVID when issuing single-use tagged pre-authentication keys and setting tags and ACLs. No launch capabilities or port access come from the coordination server. Launch identity, role membership, and service authorization remain distinct checks. [Service authorization](agent-open-questions.md#%60OQ-SBX-43%60) and [host revocation](agent-open-questions.md#%60OQ-HST-6%60) govern admission and revocation questions.

## External ingress gateway

The external ingress gateway uses Traefik with authentication middleware supplied by afakt. It authenticates user-facing requests and routes them to the API server, agent hub, and preview server. Its separate preview listener follows [preview access controls](agent-previews.md#Preview%20routing%20and%20names). How the authentication middleware integrates with Traefik remains [`OQ-SES-13`](agent-open-questions.md#%60OQ-SES-13%60). Process placement and supervision remain [`OQ-SES-7`](agent-open-questions.md#%60OQ-SES-7%60).

## User tailnet attachments

The external ingress gateway can attach to user tailnets, and an agent host's host-level node can optionally attach. Per-launch nodes never join user tailnets. As a proxy for app requests, the gateway never routes packets between the cluster overlay and a user tailnet. afakt authenticates the user even when tailnet access permits network reachability. Whether tailnet identity is an [extra authorization signal](agent-open-questions.md#%60OQ-HUB-17%60) is undecided.

A user creates the service in the Tailscale administration console and tags the gateway node during attachment setup. Phones, tablets, native apps, and browsers connect to afakt through a MagicDNS uniform resource locator (URL) such as `https://afakt.<tailnet>.ts.net`. Both network boundaries appear in the [phone-to-dashboard flow](network-topology.md#Phone%20through%20a%20user%20tailnet%20to%20the%20dashboard). Attachment mechanics and certificates remain [undecided](agent-open-questions.md#%60OQ-HST-10%60).
