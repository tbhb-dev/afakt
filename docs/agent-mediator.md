---
type: Design
title: Agent mediator
description: Per-launch guest connections, network policy, and control requests through the agent proxy.
tags: [afakt, agents, sandbox, network]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent mediator

[Public key infrastructure](pki.md).

## Launch boundary

Each agent launch has one agent mediator and one agent proxy, on every provider. The mediator controls guest traffic and policy without a network of its own. For a virtual machine (VM), its user-space Transmission Control Protocol (TCP)/IP stack serves the guest's only network device. The agent proxy opens permitted off-host connections and the launch's authenticated [hub connection](agent-proxy.md#Hub%20connection). Container manager and host-mode mechanisms remain [undecided](agent-open-questions.md#%60OQ-HUB-15%60). Remote provider integration remains [undecided](agent-open-questions.md#%60OQ-SBX-36%60).

The guest has no other route out.

Enforcement doesn't depend on tools honoring proxy environment variables. A confined process on the agent host is the leading option for the mediator. Whether it needs its own VM remains [open](agent-open-questions.md#%60OQ-SBX-12%60). afakt permits host modes as the weaker tier. Each launch has its own mediator and agent proxy. Both run separately from the supervisor and virtual machine monitors. They communicate over a launch-bound local channel. The mediator never opens an off-host connection. Service users remain [undecided](agent-open-questions.md#%60OQ-HST-8%60), and the [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60) remains on hold.

## Egress gateway routing

The agent egress gateway runs on the agent host by default. An agent host worker setting can select a standalone site gateway instead. The agent proxy connects to that gateway over mutual Transport Layer Security (TLS) with its agent launch's identity. The site gateway has its own workload identity and the same trust boundary for policy enforcement as an agent host. See [Agent egress gateway](components.md#Agent%20egress%20gateway).

## Agent DNS

The agent mediator is the guest's only Domain Name System (DNS) resolver and controls resolution. It records name-to-address bindings and never falls back to the agent host's resolver. Permitted upstream requests pass through its agent proxy. Names under `afakt.internal` follow the [agent hub naming scheme](agent-hub.md#Names), including stable agent names, launch names, published ports, services, and previews.

Upstream answers remain [open](agent-open-questions.md#%60OQ-SBX-21%60), as do [synthetic guest addresses](agent-open-questions.md#%60OQ-SBX-22%60).

## Network policy

By default, the agent mediator denies network access and doesn't decrypt traffic. For each connection, the policy engine evaluates the destination address, port, and resolved name. Unless policy permits it, the mediator denies User Datagram Protocol (UDP) and Internet Control Message Protocol (ICMP) traffic. Combined domain and resolved-address rules remain bounds on these decisions. Network, Hypertext Transfer Protocol (HTTP), Agent Client Protocol (ACP), and control decisions from the mediator and agent egress gateway enter one decision log for the agent, which excludes secrets and request and response content.

Inspected destinations need credentials or HTTP policy. At the agent egress gateway, inspected TLS connections end after mediator admission. That gateway decrypts traffic for HTTP-level policy. The credential broker injects credentials into permitted requests outside the guest. The guest's trusted certificates include the gateway's per-agent certificate authority. A destination requiring HTTP policy or credentials can't use passthrough. [Credential brokering](agent-sandbox.md#Credential%20brokering) defines the injection rules.

For passthrough destinations, the agent mediator allows or denies each connection by TLS server name, address, and port under policy. The agent proxy opens permitted connections off the agent host without decryption. This traffic remains mediated and subject to network policy.

## Control requests

On command-line interface (CLI), Model Context Protocol (MCP), and port exposure control requests arriving from the guest relay, the agent mediator stamps the agent's identity and agent launch ID. The agent proxy authenticates over mutual TLS with the launch's X.509 SPIFFE Verifiable Identity Document (SVID). SPIFFE means Secure Production Identity Framework for Everyone. At the Application Programming Interface (API) server, the request arrives over the overlay. The peer certificate identifies the launch whose capabilities and scopes the server checks against the immutable spec. Neither the SVID nor its private key enters the guest, which can't choose another launch identity. The mediator checks every session message against its [default-deny ACP policy](agent-acp.md#ACP%20message%20policy). See [Per-launch services](agent-proxy.md#Launch%20boundary).

With `allowLocalBinding`, a process can listen only on loopback in the agent sandbox's own network view, never on the agent host's external interfaces. Local app traffic stays in that view. Traffic leaving it crosses the mediator and agent proxy. Authorized preview streams pass through the agent hub and the launch's connection to guest listeners. Connections between launches require the source network policy to permit the destination name and port and the destination publication to grant source access. Same-host traffic uses the overlay too. The worker never joins launch streams. See [Agent ports and previews](agent-ports.md).

## MCP access

Inside the VM, the agent harness connects to the [MCP gateway](agent-mcp-gateway.md#Endpoint%20and%20harness%20integration) through the guest relay's control channel and agent mediator. The mediator stamps agent and agent launch identity. The agent proxy authenticates with the launch's SVID over the overlay. The MCP gateway enforces the allowlists and keeps upstream credentials outside the agent sandbox.

## Open questions

The agent mediator's [process or virtual machine boundary](agent-open-questions.md#%60OQ-SBX-12%60), [guest-link protection](agent-open-questions.md#%60OQ-SBX-13%60), [mandatory mediation across providers](agent-open-questions.md#%60OQ-SBX-54%60), [policy engine](agent-open-questions.md#%60OQ-SBX-61%60), and [network confinement](agent-open-questions.md#%60OQ-SBX-62%60) remain undecided. The [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60) remains on hold.
