---
type: Design
title: Agent ports
description: Loopback listeners, forwarded port records, connections between agents, and authenticated client tunnels.
tags: [afakt, agents, agent-launch, sessions, ports, previews]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent ports

[Public key infrastructure](pki.md).

An agent can run development servers inside its [sandbox](agent-sandbox.md). The guest relay discovers their listening ports. An explicit exposure authorized by the policy engine creates a Hypertext Transfer Protocol (HTTP) preview or permits raw Transmission Control Protocol (TCP) access through an authenticated client tunnel without requiring inbound ports on agent hosts.

## Loopback listeners

The combined agent sandbox configuration's `allowLocalBinding` permits processes to listen on loopback ports in the sandbox's own network view. It never permits binding the agent host's external interfaces. With the setting false, processes can't open listeners. The provider rejects a launch when it can't enforce these restrictions. Listening doesn't widen the sandbox's outbound domain rules. Host-mode loopback isolation and listener attribution remain [`OQ-SBX-16`](agent-open-questions.md#%60OQ-SBX-16%60) and [`OQ-SBX-54`](agent-open-questions.md#%60OQ-SBX-54%60).

## Port discovery and records

The guest relay discovers listening TCP sockets belonging to the agent's process tree, including development servers started by hooks. Discovery supplies metadata, not authority to publish a port. It attributes each listener to its agent launch. Automatic discovery assigns TCP until a declaration or an HTTP protocol check identifies HTTP.

From inside the VM, an agent runs `afakt port expose <port>` or the matching Model Context Protocol (MCP) tool. The request can include an optional name and protocol. It goes through the guest relay to the agent mediator. After stamping the agent's identity, the mediator evaluates whether policy permits that agent to expose the port. A permit creates a publication. The relay checks listener ownership and guest loopback scope before connecting. Neither declaration nor discovery authorizes an unrelated process or an address outside that scope.

A publication is the forwarded port record on the agent's agent launch or session. Discovery and exposure of the same listener update one record.

| Field | Meaning |
| --- | --- |
| Agent | The durable agent that exposed the port. |
| Owner | Agent launch ID, with the session ID when it belongs to a session. |
| Agent host | The host running the agent sandbox VM and mediator. |
| Preview label | The default or vanity name used to route an HTTP preview. |
| Number | The listening port number in the guest's network view. |
| Name | The declared name, or a generated display name for a discovered port. |
| Protocol | `http` for HTTP, including WebSockets and server-sent events, or `tcp` for raw TCP. |
| State | `open` while the attributed listener exists, or `closed` after it goes away or the agent launch ends. |
| First seen | When the relay first observed the listener, as a Coordinated Universal Time (UTC) timestamp. |
| Preview uniform resource locator (URL) | For an open HTTP port, the URL under the configured preview domain using its preview name. Absent for TCP. |

The mediator closes forwarding and marks the record `closed` when the listener goes away outside suspension or the agent launch ends. A closed record never routes to a later process that reuses its port. When observed and exposed again, a listener receives a fresh record and first-seen time. Resume requires validation of listener ownership and publication agent launch ID before routing. Closed publications and their share links never transfer to a new listener. Preview behavior during suspension remains [open](agent-open-questions.md#Agent%20ports%20and%20previews).

## Connections between agents

An agent connects to another agent's exposed port by its [internal name](agent-mediator.md#Agent%20DNS). The policy engine checks both ends: source agent A's network policy must allow destination agent B's name and port, and B must have exposed that port to A. A listener alone doesn't grant access.

Across agent hosts, the [agent hub](agent-hub.md#Launch-to-launch%20traffic) joins streams over the source and destination launches' connections. Each agent proxy opens its connection with its launch's identity over mutual Transport Layer Security (TLS). Each mediator checks its side's authority before the destination guest relay reaches the exposed loopback listener.

On the same agent host, launch traffic uses the overlay like traffic between hosts. The worker stays outside the data path. A direct overlay path remains [undecided](agent-open-questions.md#%60OQ-PORT-8%60). It doesn't introduce a same-host shortcut.

[Agent previews](agent-previews.md) use these publications under [viewer access and share-link rules](agent-previews.md#Access%20and%20share%20links). [Agent services](agent-services.md) publish names and port mappings across agent launches.

## TCP client tunnels

For a person, raw TCP ports, such as databases, are reachable only through an authenticated afakt client tunnel. Agents use the separately authorized [connections between agents](#Connections%20between%20agents). The client signs in through the ingress gateway. Then it requests one forwarded port through the gateway and agent hub. It receives access only after the control plane checks its access to the agent launch or session. Traffic crosses the launch's agent-proxy connection to the agent hub. For each stream, the mediator checks the current publication and agent launch ID before the guest relay connects to its loopback listener. Raw TCP ports never receive public listeners, preview URLs, or public share links.

[Public tunnel providers](agent-previews.md#Public%20tunnel%20providers) publish previews under the same access checks. [Session clients](agent-previews.md#Sessions%20and%20clients) display publications, and [preview telemetry](agent-previews.md#Telemetry) records requests and tunnel state.
