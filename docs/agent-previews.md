---
type: Design
title: Agent previews
description: Preview publication, names, endpoints, viewer access, and lifetime through workload proxies.
tags: [afakt, agents, previews, ports, network]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent previews

[Public key infrastructure](pki.md).

An agent preview exposes a published Hypertext Transfer Protocol (HTTP) port through an authenticated uniform resource locator (URL). The relationship between previews and [services](agent-services.md) remains [undecided](agent-open-questions.md#%60OQ-HUB-18%60).

## HTTP previews

Each published HTTP port gets a preview URL served by the [ingress gateway](components.md#Ingress%20gateway). A preview domain is the configured domain beneath which each publication gets a host name under the [preview naming scheme](agent-previews.md#Preview%20routing%20and%20names). Routing also distinguishes agent launches and listener records, so an old URL can't open another agent launch's server. Internal-only preview names resolve under `afakt.internal` for agents through the agent mediator's resolver. Applications see a site root that preserves their paths without a path prefix. WebSockets and server-sent events pass through without response buffering that prevents streaming.

The preview server is a published-port client. After viewer authorization, it dials the target workload's proxy directly over the overlay. Both sides authenticate with SPIFFE mutual Transport Layer Security (TLS). At the target, the mediator checks ingress rules for that connection. The [published-port path](agent-open-questions.md#%60OQ-PORT-14%60), including caller context and binding to the guest listener, remains open. Routing admits only recorded publications, never an arbitrary browser-supplied destination.

The ingress gateway serves previews on a second listener in the same gateway, separate from the ingress gateway's Application Programming Interface (API) and dashboard listener. afakt isolates preview origins and cookies from the API and dashboard origins and cookies: API and dashboard cookies are host-only and never scoped to a parent of preview hosts. Preview authentication uses cookies restricted to the exact preview host. Credentialed cross-origin access from preview origins fails because the API and dashboard reject it. Preview content can't read their cookies or use their authenticated browser access.

## Preview routing and names

The external ingress gateway forwards viewer requests to the preview server. It connects to workload proxies directly over the overlay. Authorization of the publication and current launch remains required. [How the proxy connects that request to the guest listener](agent-open-questions.md#%60OQ-PORT-14%60) is open. The preview server never accepts a browser-supplied arbitrary destination. See the [preview flow](network-topology.md#Operator%20browser%20to%20a%20preview) for host access.

Default public preview names use `<namespace>-<agent_id>-<port_number or port_name>.<preview_domain>`. Vanity names use `<preview_name>.<preview_domain>`. Internal previews use the internal preview zone. [Vanity-name ownership](agent-open-questions.md#%60OQ-HUB-12%60) and [authorization across launch changes](agent-open-questions.md#%60OQ-HUB-13%60) remain undecided. Preview content remains separate from the API and dashboard origin. See [browser isolation](agent-open-questions.md#%60OQ-PORT-9%60) for the remaining boundary questions.

Viewer authorization, afakt credential removal, and host-header rewriting apply before forwarding to guest content. Their division between the external ingress gateway and preview server remains [undecided](agent-open-questions.md#%60OQ-PORT-13%60). [Access and share links](agent-previews.md#Access%20and%20share%20links) govern viewer authority.

## Internal preview names

Internal preview Domain Name System (DNS) names use these targets.

| Name | Target |
| --- | --- |
| `<preview_name>.<namespace>.previews.afakt.internal` | An internal preview through the preview server |
| `previews.default.svc.afakt.internal` | The preview server |

Internal preview names use the [shared namespace and resolution rules](agent-overlay.md#Names).

## Access and share links

A preview requires afakt sign-in and access to its agent launch or session by default. On every HTTP request and before each WebSocket upgrade or event stream, the policy engine checks the viewer's access before forwarding. Port records and preview links are visible only to callers with access to their owner.

A user with that access can create a share link scoped to one forwarded port with an expiry and revocation. Signed-in recipients can use a valid link for that port without gaining access to the rest of the agent launch or session. On use, the control plane checks expiry, revocation, and port state. Grant expiry closes active shared streams. Revocation does too. Whether a share link can also admit a recipient without sign-in remains an [open question](agent-open-questions.md#Agent%20ports%20and%20previews).

Previews never pass afakt credentials to the app. Before forwarding, the preview path consumes and strips Authorization headers, afakt cookies, and share-link secrets from the request. Other app credentials remain subject to that stripping rule. Authentication redirects and share-link redemption end before the app receives a request. The preview path rewrites the Host header so preview host names work. It rejects app responses that try to set cookies for a parent domain or another preview host. The division of these checks between the ingress gateway and preview server remains [undecided](agent-open-questions.md#%60OQ-PORT-13%60).

## Host browser access

Through host-level DNS, the operator's browser resolves an internal preview name and requests access under host-level authorization. Using host authority, the host connectivity role reaches the preview server. Names for published services and ports cross the same boundary. Any further authentication belongs to the accessed app. The local connection and caller-authentication mechanisms remain [undecided](agent-open-questions.md#%60OQ-HST-9%60).

As a service client, the preview server dials the workload proxy over the overlay using SPIFFE mutual Transport Layer Security (TLS). The proxy requests an inbound connection decision from its mediator, accepting only the ingress gateway and peers named by its ingress rules. Publication and launch checks remain required. Their [binding to the guest listener](agent-open-questions.md#%60OQ-PORT-14%60) remains open. No direct route connects the browser to a guest or its proxy.

[Published ports](agent-ports.md#Port%20discovery%20and%20records) and [preview access](agent-previews.md#Access%20and%20share%20links) define target authority. [Host access](agent-open-questions.md#%60OQ-HST-9%60), [preview request processing](agent-open-questions.md#%60OQ-PORT-13%60), [listener binding](agent-open-questions.md#%60OQ-PORT-6%60), and [active-route changes](agent-open-questions.md#%60OQ-PORT-10%60) remain undecided.

See the [host browser preview flow](network-topology.md#Operator%20browser%20to%20a%20preview) for its trust boundaries.

## Sessions and clients

The web client and chat bridges show an agent session's forwarded ports and preview links as they appear, change, or close. Preview links open beside the session in a browser tab. Users with session read access can use the ports. The writer role isn't required. A chat message contains a link, never an afakt credential or a share-link secret unless a user explicitly creates that share link for sharing.

Previews use afakt sign-in and the session's access checks, with expiring, revocable share links for one port. Raw Transmission Control Protocol (TCP) access requires an [authenticated afakt client tunnel](agent-ports.md#TCP%20client%20tunnels).

## Session previews and suspension

Agent session suspension stops the virtual machine (VM). Whether a preview request wakes it or fails remains [open](agent-open-questions.md#Agent%20ports%20and%20previews). Resume rechecks listener ownership and publication agent launch IDs before routing traffic. Closed records and their share links never attach to another listener.

## Public tunnel providers

Public tunnel providers include Cloudflare Tunnel and Tailscale Funnel in afakt's binary. On the control plane, the operator selects a provider and configures its credentials and public naming. afakt runs the connector inside the API server process and maintains outbound tunnel connections. Credentials and the connector's identity state stay on the control plane and never reach agent hosts or the agent sandbox.

A provider publishes the ingress gateway's listener for previews, and can also publish the whole afakt web surface through a separate API and dashboard origin. Publishing preserves the listener separation and cookie restrictions. Every preview follows the same afakt access checks through a tunnel. A public host name never makes a preview public to anonymous callers. Tunnel providers never publish raw agent launch TCP ports.

### Cloudflare

afakt runs the embedded Cloudflare Tunnel connector in the API server process and routes its HTTP traffic to the ingress gateway's listener for previews. Its outbound connections connect to Cloudflare's network without opening inbound ports. The operator creates a tunnel in a Cloudflare account and configures connector credentials. They route the preview domain's wildcard DNS entry to that tunnel. For its preview host names, the domain needs edge Transport Layer Security (TLS) coverage. See [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/).

Each port's agent launch-specific host name appears beneath that configured preview domain. An optional separate public host name routes to the ingress gateway's API and dashboard listener. The connector preserves preview host routing and streaming, and afakt still authenticates and authorizes each request.

### Tailscale

afakt runs embedded Tailscale nodes through `tsnet` in the API server process, with Funnel listeners feeding the ingress gateway's listener for previews. For enrollment, the operator supplies credentials and enables DNS and Hypertext Transfer Protocol Secure (HTTPS) for the tailnet. The operator grants Funnel permission to those nodes in the tailnet policy. Funnel uses the tailnet's domain for its public names. See [Tailscale Funnel](https://tailscale.com/docs/features/tailscale-funnel) and [the embedded server API](https://tailscale.com/docs/reference/tsnet-server-api).

Each preview needs its own site root and a route bound to its publication record. Closing the record withdraws its public route. Whether Funnel supplies these names through one embedded node per preview remains [`OQ-PORT-4`](agent-open-questions.md#%60OQ-PORT-4%60). Publishing the API and dashboard listener preserves its separate origin. The provider reports a publication failure when enrollment, naming, or Funnel policy prevents it from serving a required preview host.

## Telemetry

[Preview telemetry](agent-observability.md#Ports%20and%20previews) defines the metrics for forwarded port counts, preview requests, transferred bytes and latency, share links created and used, and tunnel connection state. Each preview request emits an `afakt.preview.request` span linked to the agent launch's trace. Metric attributes use bounded protocol, outcome, profile, and tunnel provider values. Port numbers, preview names and URLs, session and agent launch IDs, and share-link secrets never become metric attributes.
