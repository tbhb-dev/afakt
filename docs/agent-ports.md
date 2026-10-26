---
type: Design
title: Agent ports and previews
description: Loopback development servers, forwarded port records, authenticated previews and client tunnels, and public tunnel providers.
tags: [afakt, agents, execution, sessions, ports, previews]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent ports and previews

An agent can run development servers inside its [sandbox](agent-sandbox.md). The guest relay discovers their listening ports. An explicit exposure authorized by Cedar creates an HTTP preview or permits raw TCP access through an authenticated client tunnel without requiring inbound ports on agent hosts.

## Loopback listeners

The combined sandbox configuration's `allowLocalBinding` permits processes to listen on loopback ports in the sandbox's own network view. It never permits binding the agent host's external interfaces. With the setting false, processes can't open listeners. The provider rejects a launch when it can't enforce these restrictions. Listening doesn't widen the sandbox's outbound domain rules or give access to another execution's listeners.

## Port discovery and records

The guest relay discovers listening TCP sockets belonging to the agent's process tree, including development servers started by hooks. Discovery supplies metadata, not authority to publish a port. It attributes each listener to its execution and generation. Automatic discovery assigns TCP until a declaration or an HTTP protocol check identifies HTTP.

From inside the VM, an agent runs `afakt port expose <port>` or the matching MCP tool. The request can include an optional name and protocol. It goes through the guest relay to the mediator. After stamping the agent's identity, the mediator asks Cedar whether that agent may expose that port. A permit creates a publication. The relay checks listener ownership and guest loopback scope before connecting. Neither declaration nor discovery authorizes an unrelated process or an address outside that scope.

A publication is the forwarded port record on the agent's execution or session. Discovery and exposure of the same listener update one record.

| Field | Meaning |
| --- | --- |
| Agent | The durable agent that exposed the port. |
| Owner | Execution ID and generation, with the session ID when it belongs to a session. |
| Agent host | The host running the sandbox VM and mediator. |
| Preview label | A random label used to route an HTTP preview without revealing agent or execution identifiers. |
| Number | The listening port number in the guest's network view. |
| Name | The declared name, or a generated display name for a discovered port. |
| Protocol | `http` for HTTP, including WebSockets and server-sent events, or `tcp` for raw TCP. |
| State | `open` while the attributed listener exists, or `closed` after it goes away or the execution ends. |
| First seen | When the relay first observed the listener, as a Coordinated Universal Time (UTC) timestamp. |
| Preview URL | For an open HTTP port, the URL under the configured preview domain using the random preview label. Absent for TCP. |

The mediator closes forwarding and marks the record `closed` when the listener goes away outside suspension or the execution ends. A closed record never routes to a later process that reuses its port. When observed and exposed again, a listener receives a fresh record and first-seen time. Resume requires validation of listener ownership and publication generation before routing. Closed publications and their share links never transfer to a new listener. Preview behavior during suspension remains [open](agent-open-questions.md#agent-ports-and-previews).

## HTTP previews

Each published HTTP port gets a preview URL served by the preview proxy inside the API server process. A preview domain is the configured domain beneath which each publication gets its own random host name. Routing also distinguishes generations and listener records, so an old URL can't open another generation's server. Applications see a site root that preserves their paths without a path prefix. WebSockets and server-sent events pass through without response buffering that prevents streaming.

The mediator dials out to a preview broker beside the preview proxy and maintains that connection. Preview requests return as streams over it. For each stream, the mediator checks that the publication is still current and that its generation matches, then connects through the guest relay to the guest's loopback listener. The broker routes only recorded publications, never an arbitrary destination supplied by a browser. No preview transport listens for inbound connections on the agent host or in the guest. When the mediator connection is unavailable, requests fail without opening a direct route.

The preview proxy uses a listener separate from the API and dashboard listener within the API server process. afakt isolates preview origins and cookies from the API and dashboard origins and cookies: API and dashboard cookies are host-only and never scoped to a parent of preview hosts. Preview authentication uses cookies restricted to the exact preview host. Credentialed cross-origin access from preview origins fails because the API and dashboard reject it. Preview content can't read their cookies or use their authenticated browser access.

## Access and share links

A preview requires afakt sign-in and access to its execution or session by default. On every HTTP request and before each WebSocket upgrade or event stream, Cedar checks the viewer's access at the preview proxy. Port records and preview links are visible only to callers with access to their owner.

A user with that access can create a share link scoped to one forwarded port with an expiry and revocation. Signed-in recipients can use a valid link for that port without gaining access to the rest of the execution or session. On use, the control plane checks expiry, revocation, and port state. Grant expiry closes active shared streams. Revocation does too. Whether a share link can also admit a recipient without sign-in remains an [open question](agent-open-questions.md#agent-ports-and-previews).

Previews never pass afakt credentials to the app. Before forwarding, the preview proxy consumes and strips Authorization headers, afakt cookies, and share-link secrets from the request. Other app credentials remain subject to that stripping rule. Authentication redirects and share-link redemption end before the app receives a request. The proxy rewrites the Host header or configures the development server's allowed hosts so preview host names work. It rejects app responses that try to set cookies for a parent domain or another preview host.

## TCP client tunnels

Raw TCP ports, such as databases, are reachable only through an authenticated afakt client tunnel. The client signs in. Then it requests one forwarded port. It receives access only after the control plane checks its access to the execution or session. Traffic crosses the mediator's outbound connection to the preview broker. For each stream, the mediator checks the current publication and generation before the guest relay connects to its loopback listener. Raw TCP ports never receive public listeners, preview URLs, or public share links.

## Public tunnel providers

afakt includes Cloudflare Tunnel and Tailscale Funnel providers in its binary. On the control plane, the operator selects a provider and configures its credentials and public naming. afakt runs the connector inside the API server process and maintains outbound tunnel connections. Credentials and the connector's identity state stay on the control plane and never reach agent hosts or the sandbox.

A provider publishes the preview listener, and can also publish the whole afakt web surface through a separate API and dashboard origin. Publishing preserves the listener separation and cookie restrictions. Every preview follows the same afakt access checks through a tunnel. A public host name never makes a preview public to anonymous callers. Tunnel providers never publish raw execution TCP ports.

### Cloudflare

afakt runs the embedded Cloudflare Tunnel connector in the API server process and routes its HTTP traffic to the preview listener. Its outbound connections connect to Cloudflare's network without opening inbound ports. The operator creates a tunnel in a Cloudflare account and configures connector credentials. They route the preview domain's wildcard DNS entry to that tunnel. For its preview host names, the domain needs edge Transport Layer Security (TLS) coverage. See [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/).

Each port's execution-specific host name appears beneath that configured preview domain. An optional separate public host name routes to the API and dashboard listener. The connector preserves preview host routing and streaming, and afakt still authenticates and authorizes each request.

### Tailscale

afakt runs embedded Tailscale nodes through `tsnet` in the API server process, with Funnel listeners feeding the preview listener. For enrollment, the operator supplies credentials and enables DNS and HTTPS for the tailnet. The operator grants Funnel permission to those nodes in the tailnet policy. Funnel uses the tailnet's domain for its public names. See [Tailscale Funnel](https://tailscale.com/docs/features/tailscale-funnel) and [the embedded server API](https://tailscale.com/docs/reference/tsnet-server-api).

To keep one site root per port and execution, afakt gives each preview its own embedded node identity and host name in the configured tailnet preview domain. These nodes run within the API server process. Each node routes only to its assigned preview record. Closing the record withdraws its public route. An optional separate node host name publishes the API and dashboard listener. The provider reports a publication failure when enrollment, naming, or Funnel policy prevents it from serving a required preview host.

## Sessions and clients

The web client and chat bridges show a session's forwarded ports and preview links as they appear, change, or close. Opening a preview link opens a browser tab beside the session. Port access follows the session's read access and doesn't require its writer role. A chat message contains a link, never an afakt credential or a share-link secret unless a user explicitly creates that share link for sharing.

[Agent session](agent-session.md#forwarded-ports-and-previews) defines how this view fits session attachment and suspend and resume.

## Telemetry

[Ports and previews](agent-observability.md#ports-and-previews) defines the metrics for forwarded port counts, preview requests, transferred bytes and latency, share links created and used, and tunnel connection state. Each preview request emits an `afakt.preview.request` span linked to the execution generation's trace. Metric attributes use bounded protocol, outcome, profile, and tunnel provider values. Port numbers, preview names and URLs, session and execution IDs, and share-link secrets never become metric attributes.
