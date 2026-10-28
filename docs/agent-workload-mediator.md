---
type: Design
title: Agent mediator
description: Workload connection decisions, firewall policy, DNS, and control requests outside the traffic path.
tags: [afakt, agents, sandbox, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent mediator

[Public key infrastructure](pki.md).

## Launch boundary

Each [workload](agent-workload.md) has one agent mediator and one agent proxy on every provider. The proxy transmits guest traffic in both directions and requests a mediator decision for every inbound and outbound connection. The mediator never transmits traffic and has no independent network access. Container manager and host-mode mechanisms remain [undecided](agent-open-questions.md#%60OQ-HUB-15%60). Remote provider integration remains [undecided](agent-open-questions.md#%60OQ-SBX-36%60).

Enforcement doesn't depend on tools honoring proxy environment variables. A confined process on the agent host is the leading option for the mediator. Whether it needs its own VM remains [open](agent-open-questions.md#%60OQ-SBX-12%60). afakt permits host modes as the weaker tier. The mediator and proxy run separately from the supervisor and virtual machine monitors. They communicate over a local channel bound to the workload. Service users remain [undecided](agent-open-questions.md#%60OQ-HST-8%60), and the [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60) remains on hold.

## Control plane connections

The mediator's own connections to secrets, policy bundles, and the API server pass through its proxy under a fixed rule set built into the workload. Neither the launch's policy nor the guest can change those rules. Control plane pushes arrive on a long-lived stream the mediator opens. The control plane never opens a connection to a workload.

## Agent egress gateway

The agent egress gateway runs on the agent host by default. An agent host worker setting can select a standalone site gateway instead. Over mutual Transport Layer Security (TLS), the agent proxy connects to that gateway with its agent launch's identity. The site gateway has its own workload identity and the same trust boundary for policy enforcement as an agent host. See [Agent egress gateway](components.md#Agent%20egress%20gateway).

Inspected TLS connections end at the agent egress gateway after mediator admission. That gateway decrypts traffic for HTTP-level policy. The credential broker injects credentials into permitted requests outside the guest. Inside the guest, trusted certificates include the gateway's per-agent certificate authority. A destination requiring HTTP policy or credentials can't use passthrough. [Credential brokering](agent-sandbox.md#Credential%20brokering) defines the injection rules. Whether proxy TLS termination takes over the gateway's decryption role remains [open](agent-open-questions.md#%60OQ-SBX-72%60).

## Agent DNS

The agent mediator is the guest's only Domain Name System (DNS) resolver and controls resolution. The agent proxy forwards guest queries to it. It records name-to-address bindings and never falls back to the agent host's resolver. Permitted upstream requests pass through its agent proxy. Names under `afakt.internal` follow the [overlay naming scheme](agent-overlay.md#Names), including stable agent names, launch names, published ports, services, and previews.

Upstream answers remain [open](agent-open-questions.md#%60OQ-SBX-21%60), as do [synthetic guest addresses](agent-open-questions.md#%60OQ-SBX-22%60).

## Network policy

Each workload's firewall has ingress and egress rules. Egress defaults to a cluster-managed deny-all policy. The agent mediator evaluates each connection's destination address, port, and resolved name. Combined domain and resolved-address rules remain bounds on these decisions. TCP, User Datagram Protocol (UDP), and QUIC pass through the proxy under mediator decisions. How a domain rule applies to QUIC remains [open](agent-open-questions.md#%60OQ-SBX-68%60).

A mediator decision can tell the proxy to end Transport Layer Security (TLS) on an outbound connection when the server name matches a domain with a Hypertext Transfer Protocol (HTTP) rule or credential injection. For passthrough destinations, the proxy opens permitted connections without decryption. These connections remain subject to the firewall.

Network, HTTP, Agent Client Protocol (ACP), and control decisions from the mediator and agent egress gateway enter one decision log for the agent, which excludes secrets and request and response content.

## Rule changes

A rule change ends existing connections the new rules deny. Revoking persistent egress that the guest starts and UDP flows remains [open](agent-open-questions.md#%60OQ-SBX-69%60).

## Control requests

On command-line interface (CLI), Model Context Protocol (MCP), and port exposure control requests arriving from the guest relay, the agent mediator stamps the agent's identity and agent launch ID. The agent proxy authenticates over mutual TLS with the launch's X.509 SPIFFE Verifiable Identity Document (SVID). SPIFFE means Secure Production Identity Framework for Everyone. At the Application Programming Interface (API) server, the request arrives over the overlay. The peer certificate identifies the launch whose capabilities and scopes the server checks against the immutable spec. Neither the SVID nor its private key enters the guest, which can't choose another launch identity. The mediator checks every session message against its [default-deny ACP policy](agent-acp.md#ACP%20message%20policy). See [Per-launch services](agent-workload-proxy.md#Launch%20boundary).

With `allowLocalBinding`, a process can listen only on loopback in the agent sandbox's own network view, never on the agent host's external interfaces. Local app traffic stays in that view. Traffic leaving it crosses the agent proxy under mediator decisions. Over the overlay, the preview server connects directly to the workload proxy. How the proxy connects to a guest listener for a [published port](agent-open-questions.md#%60OQ-PORT-14%60) remains open. Connections between launches require the source network policy to permit the destination name and port and the destination publication to grant source access. Same-host traffic uses the overlay too. The worker never joins launch streams. See [Agent ports and previews](agent-ports.md).

## MCP access

Inside the VM, the agent harness connects to the [MCP gateway](agent-mcp-gateway.md#Endpoint%20and%20harness%20integration) through the guest relay and workload proxy under mediator decisions. The mediator stamps agent and agent launch identity. The agent proxy authenticates with the launch's SVID over the overlay. The MCP gateway enforces the allowlists and keeps upstream credentials outside the agent sandbox.

## Open questions

The agent mediator's [process or virtual machine boundary](agent-open-questions.md#%60OQ-SBX-12%60), [guest-link protection](agent-open-questions.md#%60OQ-SBX-13%60), [mandatory mediation across providers](agent-open-questions.md#%60OQ-SBX-54%60), [policy engine](agent-open-questions.md#%60OQ-SBX-61%60), and [network confinement](agent-open-questions.md#%60OQ-SBX-62%60) remain undecided. The [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60) remains on hold.
