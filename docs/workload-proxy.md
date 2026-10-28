---
type: Design
title: Workload proxy
description: Workload traffic, direct overlay connections, mutual authentication, and mediator decisions.
tags: [afakt, agents, network, identity]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload proxy

[Public key infrastructure](pki.md).

## Launch boundary

Each [workload](workload.md) has one workload proxy mediator and one workload proxy on every workload provider. This applies to container manager hosts and host modes. Traffic follows guest ⇄ workload proxy ⇄ network in both directions. For each inbound and outbound connection, the proxy requests the mediator's decision. The mediator never transmits traffic. TCP, User Datagram Protocol (UDP), and QUIC use the proxy, with the [domain rule for QUIC](agent-open-questions.md#%60OQ-SBX-68%60) still open.

The workload proxy stores the launch's X.509 SPIFFE Verifiable Identity Document (SVID) and private key outside the guest. SPIFFE means Secure Production Identity Framework for Everyone. On control requests, the workload proxy mediator stamps the launch identity. Using the launch's SVID, the workload proxy authenticates those requests. A receiving service checks the peer identity and the immutable launch spec. Terminal launch identities can't start new authorized work. [Workload identity](workload-host.md#Workload%20identity) governs issuance and registration.

Service users remain [undecided](agent-open-questions.md#%60OQ-HST-8%60), as do [provider mechanisms](agent-open-questions.md#%60OQ-HUB-15%60), [remote placement](agent-open-questions.md#%60OQ-SBX-36%60), and the proxy's [program boundary](agent-open-questions.md#%60OQ-HUB-16%60). The [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60) remains on hold.

## Overlay connections

Over the overlay, the ingress gateway and other workloads dial the workload's proxy directly. Both ends authenticate with SPIFFE mutual Transport Layer Security (TLS). Connections between workloads require decisions from both mediators, and the destination proxy accepts only the ingress gateway and the peers its ingress rules name, and only for the ports those rules publish. The gateway never bypasses ingress rules. How the preview server binds a request to a guest listener remains part of the open [published-port path](agent-open-questions.md#%60OQ-PORT-14%60).

Direct reachability doesn't assign responsibility for [client ACP and remote-control connections](agent-open-questions.md#%60OQ-SES-17%60) or [session access, message logs, and replay](agent-open-questions.md#%60OQ-SES-18%60). Their owners remain open.

Cutting off an agent's traffic requires action at more than one point. Overlay node or Access Control List (ACL) revocation, together with a policy push, cuts off access. [Rule changes](workload-proxy-mediator.md#Rule%20changes) end existing connections the new rules deny.

## Overlay node

Each workload's proxy has its own overlay node using Tailscale's userspace client library, `tsnet`, against an isolated Headscale. Launch nodes never join user tailnets. The [overlay transport](workload-overlay.md#Overlay%20transport) supplies service reachability, while the launch identity and receiving service policy govern authorization.

## Agent launch authentication

Using the agent launch's X.509 SPIFFE Verifiable Identity Document (SVID), the workload proxy authenticates to the afakt Application Programming Interface (API) server over mutual TLS on the overlay. At the API server, TLS ends and the peer certificate identifies the launch. See [Overlay connections](#Overlay%20connections).

From the agent launch spec it already holds, the API server reads the following:

| Spec field | Value |
| --- | --- |
| Agent profile | Source, name, version, and digest of the pinned profile. |
| Capabilities | Each of the agent launch's capabilities, with its scopes after variable replacement. See [Capabilities](agent-profile.md#Capabilities). |

On each request, the API server checks the capabilities and scopes. After the agent launch enters a terminal state, the API server refuses requests from that agent launch's identity.

Both the SVID and its private key remain outside the guest. Inside the sandbox, the afakt command-line interface (CLI) uses the guest relay. Session-end wait, SVID lifetime, and renewal bounds remain [on hold](agent-open-questions.md#%60OQ-HOOK-18%60).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Operations on external services, such as pull request comments and reviews, go through the afakt API under the agent launch's capabilities. The gateway's [built-in afakt Model Context Protocol (MCP) server](agent-mcp-gateway.md#Built-in%20afakt%20MCP%20server) applies the same authority to its API calls. Agent-owned credential references permit brokered use under [Credential types and custody](workload-sandbox.md#Credential%20types%20and%20custody). An attachment change changes the agent spec and takes effect through a new agent launch at a turn boundary. Credential reference bounds remain [on hold](agent-open-questions.md#%60OQ-SBX-2%60).

## Internet connections

For internet traffic from the guest, the proxy requests the mediator's connection decision before opening a permitted external connection. Internet services receive their destination protocol and any authorized upstream credential, not the launch SVID merely because the request came from a launch.

When an outbound server name matches a domain with a Hypertext Transfer Protocol (HTTP) rule or credential injection, the proxy ends TLS when the mediator instructs it to do so. The guest's trust store includes a per-workload certificate authority. Its signing material stays outside the guest. A destination requiring HTTP policy or credentials can't use passthrough. The credential broker injects credentials at the workload proxy into permitted requests outside the guest. [Credential types](workload-sandbox.md#Credential%20types%20and%20custody) define the material the broker receives from the control plane, including static secrets the control plane can't exchange. Permitted passthrough traffic leaves without decryption under connection policy. Brokered secrets never go directly to the guest. Response reflection remains [`OQ-SBX-26`](agent-open-questions.md#%60OQ-SBX-26%60).

HTTP-level decisions enter the [one decision log for the workload](workload-proxy-mediator.md#Network%20policy), alongside network, Agent Client Protocol (ACP), and control decisions. The log excludes secrets and request and response content.

## Open questions

The workload proxy's [program boundary](agent-open-questions.md#%60OQ-HUB-16%60), [workload provider mechanisms](agent-open-questions.md#%60OQ-HUB-15%60), and [guest-link conditions](agent-open-questions.md#%60OQ-SBX-70%60) remain undecided. [Local carrier binding](agent-open-questions.md#%60OQ-SBX-13%60) must bind the guest to its workload and launch.
