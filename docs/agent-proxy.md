---
type: Design
title: Agent proxy
description: Per-launch overlay identity, authenticated connections, and streams between the agent mediator and services.
tags: [afakt, agents, network, identity]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent proxy

[Public key infrastructure](pki.md).

## Launch boundary

Each agent launch has one agent mediator and one agent proxy on every agent sandbox provider, including container manager hosts and host-mode launches. Policy enforcement belongs to the agent mediator, which has no network of its own. It passes permitted requests over a local channel bound to the launch. For local providers, the agent proxy opens the hub connection and permitted service and internet connections from the agent host. Remote placement and carriers remain [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). Guests never join the overlay. These separate boundaries appear in the [launch topology](network-topology.md#Agent%20launch).

The agent proxy stores that launch's X.509 SPIFFE Verifiable Identity Document (SVID) and private key outside the guest. SPIFFE means Secure Production Identity Framework for Everyone. On control requests, the agent mediator stamps the launch identity. Using the launch's SVID, the agent proxy authenticates those requests. A receiving service checks the peer identity and the immutable launch spec. Terminal launch identities can't start new authorized work. [Workload identity](agent-host.md#Workload%20identity) governs issuance and registration.

The service users remain [undecided](agent-open-questions.md#%60OQ-HST-8%60). The mechanisms for container manager hosts and host-mode launches are [undecided](agent-open-questions.md#%60OQ-HUB-15%60), as is whether the agent proxy is a [separate program or an adapted proxy](agent-open-questions.md#%60OQ-HUB-16%60). The [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60) remains on hold.

## Hub connection

The agent proxy opens one multiplexed connection to the agent hub over the cluster overlay, authenticated with its launch's own SVID through mutual Transport Layer Security (TLS). As a service workload, the hub authenticates itself. Remote-control streams, Agent Client Protocol (ACP) streams, published-port and service streams, and preview traffic use this connection. Returning streams cross the target agent proxy and mediator before reaching a guest listener or control endpoint. Each stream retains the caller, target launch, and authority for its operation.

Ownership of [hub stream framing](agent-open-questions.md#%60OQ-HUB-14%60) between the mediator and agent proxy is undecided. User Datagram Protocol (UDP) traffic uses the overlay natively. [Datagram framing on the hub connection](agent-open-questions.md#%60OQ-HUB-9%60) is a separate question. [Stream transport](agent-open-questions.md#%60OQ-SES-4%60), [admission under overload](agent-open-questions.md#%60OQ-SBX-14%60), and [stream retirement](agent-open-questions.md#%60OQ-SBX-40%60) have their own bounds and recovery questions.

## Overlay node

The agent proxy has its own overlay node using Tailscale's userspace client library, `tsnet`, against an isolated Headscale. Launch nodes never join user tailnets. The [overlay transport](agent-hub.md#Overlay%20transport) supplies service reachability, while the launch identity and receiving service policy govern authorization.

## Agent launch authentication

Using the agent launch's X.509 SPIFFE Verifiable Identity Document (SVID), the agent proxy authenticates to the afakt Application Programming Interface (API) server over mutual TLS on the overlay. At the API server, TLS ends and the peer certificate identifies the launch. See [Hub connection](agent-proxy.md#Hub%20connection).

From the agent launch spec it already holds, the API server reads the following:

| Spec field | Value |
| --- | --- |
| Agent profile | Source, name, version, and digest of the pinned profile. |
| Capabilities | Each of the agent launch's capabilities, with its scopes after variable replacement. See [Capabilities](agent-profile.md#Capabilities). |

On each request, the API server checks the capabilities and scopes. After the agent launch enters a terminal state, the API server refuses requests from that agent launch's identity.

Both the SVID and its private key remain outside the guest. Inside the sandbox, the afakt command-line interface (CLI) uses the guest relay. Session-end wait, SVID lifetime, and renewal bounds remain [on hold](agent-open-questions.md#%60OQ-HOOK-18%60).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Operations on external services, such as pull request comments and reviews, go through the afakt API under the agent launch's capabilities. The gateway's [built-in afakt Model Context Protocol (MCP) server](agent-mcp-gateway.md#Built-in%20afakt%20MCP%20server) applies the same authority to its API calls. Agent-owned credential references permit brokered use under [Credential types and custody](agent-sandbox.md#Credential%20types%20and%20custody). An attachment change changes the agent spec and takes effect through a new agent launch at a turn boundary. Credential reference bounds remain [on hold](agent-open-questions.md#%60OQ-SBX-2%60).

## Internet connections

The guest's internet request crosses its local launch boundary to the agent mediator. After binding the request to that launch, the mediator permits or denies its destination. Only permitted traffic passes to the agent proxy, which opens the external connection under [Launch boundary](#Launch%20boundary). At the internet service, the request uses its destination protocol and any authorized upstream credential, but that service doesn't receive a launch SVID merely because the request came from a launch.

Encryption and connection policy still apply to passthrough traffic. Destinations that require credentials or Hypertext Transfer Protocol (HTTP) policy require inspection outside the guest. The broker excludes direct secret delivery to the guest. Response reflection remains [`OQ-SBX-26`](agent-open-questions.md#%60OQ-SBX-26%60). Placement of inspection and broker processes remains undecided under [egress processing](agent-open-questions.md#%60OQ-SBX-20%60) and [credential broker placement](agent-open-questions.md#%60OQ-SBX-23%60). Across either path, the agent proxy retains responsibility for opening off-host connections.

## Open questions

The agent proxy's [program boundary](agent-open-questions.md#%60OQ-HUB-16%60), [hub stream ownership](agent-open-questions.md#%60OQ-HUB-14%60), and [provider mechanisms](agent-open-questions.md#%60OQ-HUB-15%60) remain undecided. [Local carrier binding](agent-open-questions.md#%60OQ-SBX-13%60) must bind the guest to its launch when requests cross the agent mediator and agent proxy.
