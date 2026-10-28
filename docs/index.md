---
okf_version: "0.2"
---

<!-- generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" } -->

# Architecture

* [Architecture](architecture.md) - How afakt's components fit together in one binary and how they talk to each other.
* [Network topology](network-topology.md) - Cluster, workload host, and agent launch networks with trust boundaries and authenticated traffic flows.
* [Public key infrastructure](pki.md) - Authorities, trust boundaries, certificate policies, and key custody.
* [Workload identity keys](pki-workload.md) - Workload authorities, identity documents, enrollment keys, and their custody.
* [Transport certificates and keys](pki-transport.md) - Overlay credentials, public and inspection certificates, and storage transport trust.
* [Signing keys and secrets](pki-signing.md) - Authentication secrets, delegated keys, artifact signatures, and encryption at rest.
* [Public key infrastructure operations](pki-operations.md) - Issuance, renewal, recovery, administrative operations, and threats.

# Agents

* [Agents](agent.md) - Durable agent identity, configuration, owned resources, and operations across agent launches.
* [Agent lifecycle](agent-lifecycle.md) - Agent, launch, and session states from creation through suspension, capture, restoration, and deletion.
* [Workload host](workload-host.md) - A machine's role in the agent fleet, its agent launch services, trust boundary, and recovery behavior.
* [Agent launch](agent-launch.md) - One instantiation of an agent in an agent sandbox on a workload host, with lifecycle, retries, placement, and cleanup.
* [Agent launch model](agent-launch-model.md) - The fields and identity of each agent launch, who sets them, which never change, and the provenance record and evidence.
* [Agent launch contract](agent-launch-contract.md) - The agent launch spec, the agent launch result, and the provision, restore, release, status, cancel, and result operations of a workload provider.
* [Workload placement](workload-placement.md) - The workload host registry and the reservation API that bind each agent launch to a workload host with a free agent launch slot.
* [Agent launch sequence](agent-launch-sequence.md) - Every message between the parts of an agent launch, from the start of the agent dispatcher workflow to the result.
* [Agent launch resource](agent-launch-resource.md) - The AgentLaunch resource, the resource form of one agent launch, with a full example.
* [Agent workspace](agent-workspace.md) - The workspace sources an agent's workspace comes from and the environment variables the agent and its hooks see.
* [Agent workspace lifecycle](agent-workspace-lifecycle.md) - Workspace creation, materialization, launch attachment, capture, retention, and deletion across workload providers.
* [Agent workspace image](agent-workspace-image.md) - Prebuilt workspace directory trees in Open Container Initiative (OCI) images, their sandboxed builds, registry and object storage, and workload host caches for fast start.
* [Agent hooks](agent-hooks.md) - The hooks that run on hook events before, during, and after the agent's work, how profiles and specifications combine them, and which events and decisions each agent harness supports.
* [Agent observability](agent-observability.md) - The OpenTelemetry trace and spans of an agent launch, the trace context the harness receives, and the metrics afakt publishes.
* [Agent profile](agent-profile.md) - The blueprint for creating an agent, with configuration templates and bounds for finalizing each agent launch spec.
* [Workload overlay and names](workload-overlay.md) - Workload overlay nodes, direct connections, internal names, and ingress access.
* [Workload](workload.md) - Guest, mediator, proxy, firewall, and launch boundary.
* [Workload guest](workload-guest.md) - Guest routing, containment, and its private proxy link.
* [Workload proxy](workload-proxy.md) - Workload traffic, direct overlay connections, and mutual authentication.
* [Agent session](agent-session.md) - An agent's durable conversation, retained transcripts, and session workflow.
* [Agent ACP](agent-acp.md) - Agent Client Protocol connections, message policy, turns, and writer and reader roles.
* [Agent session attachment](agent-session-attachment.md) - Session attachment, history replay, and signed agent messages.
* [Workload sandbox](workload-sandbox.md) - The sandbox configuration, sandbox variables, credential brokering, and authentication of an agent launch.
* [Workload proxy mediator](workload-proxy-mediator.md) - Per-launch guest connections, network policy, and control requests.
* [Workload lifecycle](workload-lifecycle.md) - Provisioning, readiness, suspension, restoration, and teardown across providers.
* [Workload provider for macOS VM](workload-provider-macos-vm.md) - Isolation, mediation, storage, and lifecycle for macos-vm.
* [Workload provider for Linux VM](workload-provider-linux-vm.md) - Isolation, mediation, storage, and lifecycle for linux-vm.
* [Workload provider for Windows VM](workload-provider-windows-vm.md) - Isolation, mediation, storage, and lifecycle for windows-vm.
* [Workload provider for macOS host](workload-provider-macos-host.md) - Isolation, mediation, storage, and lifecycle for macos-host.
* [Workload provider for Linux host](workload-provider-linux-host.md) - Isolation, mediation, storage, and lifecycle for linux-host.
* [Workload provider for Windows host](workload-provider-windows-host.md) - Isolation, mediation, storage, and lifecycle for windows-host.
* [Workload provider for Docker Engine](workload-provider-docker-engine.md) - Isolation, mediation, storage, and lifecycle for docker-engine.
* [Workload provider for Docker Desktop on macOS](workload-provider-docker-desktop-macos.md) - Isolation, mediation, storage, and lifecycle for docker-desktop-macos.
* [Workload provider for Docker Desktop on Windows](workload-provider-docker-desktop-windows.md) - Isolation, mediation, storage, and lifecycle for docker-desktop-windows.
* [Workload provider for Claude Managed Agents](workload-provider-claude-managed-agents.md) - Remote isolation, mediation, storage, and lifecycle for Claude Managed Agents.
* [Workload provider for OpenAI Agents API](workload-provider-openai-agents-api.md) - Remote isolation, mediation, storage, and lifecycle for OpenAI Agents API.
* [Workload provider for Modal Sandboxes](workload-provider-modal.md) - Remote isolation, mediation, storage, and lifecycle for Modal Sandboxes.
* [Workload provider for Cloudflare Sandbox SDK](workload-provider-cloudflare.md) - Remote isolation, mediation, storage, and lifecycle for Cloudflare Sandbox SDK.
* [Agent MCP gateway](agent-mcp-gateway.md) - One Model Context Protocol (MCP) endpoint per agent harness, with profile tool access, upstream credentials, server hosting, approvals, and audit.
* [Workload ports](workload-ports.md) - Loopback listeners, forwarded port records, connections between agents, and authenticated client tunnels.
* [Workload previews](workload-previews.md) - Preview names, endpoints, viewer access, tunnel providers, and lifetime.
* [Workload services](workload-services.md) - Published service names, port mappings, access, and lifetime across launches.
* [Workload threat model](workload-threat-model.md) - Shared assets, service boundaries, and threats across agent launches.
* [Workload threat model for macOS VM](workload-threat-model-macos-vm.md) - Provider boundaries and threats for macOS VM.
* [Workload threat model for Linux VM](workload-threat-model-linux-vm.md) - Provider boundaries and threats for Linux VM.
* [Workload threat model for Windows VM](workload-threat-model-windows-vm.md) - Provider boundaries and threats for Windows VM.
* [Workload threat model for macOS host](workload-threat-model-macos-host.md) - Provider boundaries and threats for macOS host.
* [Workload threat model for Linux host](workload-threat-model-linux-host.md) - Provider boundaries and threats for Linux host.
* [Workload threat model for Windows host](workload-threat-model-windows-host.md) - Provider boundaries and threats for Windows host.
* [Workload threat model for Docker Engine on Linux](workload-threat-model-docker-engine.md) - Provider boundaries and threats for Docker Engine on Linux.
* [Workload threat model for Docker Desktop on macOS](workload-threat-model-docker-desktop-macos.md) - Provider boundaries and threats for Docker Desktop on macOS.
* [Workload threat model for Docker Desktop on Windows](workload-threat-model-docker-desktop-windows.md) - Provider boundaries and threats for Docker Desktop on Windows.
* [Workload threat model for Claude Managed Agents](workload-threat-model-claude-managed-agents.md) - Remote execution and vendor custody threats for Claude Managed Agents.
* [Workload threat model for OpenAI Agents API](workload-threat-model-openai-agents-api.md) - Remote execution and vendor custody threats for OpenAI Agents API.
* [Workload threat model for Modal Sandboxes](workload-threat-model-modal.md) - Remote execution and vendor custody threats for Modal Sandboxes.
* [Workload threat model for Cloudflare Sandbox SDK](workload-threat-model-cloudflare.md) - Remote execution and vendor custody threats for Cloudflare Sandbox SDK.
* [Agent open questions](agent-open-questions.md) - Undecided design questions for agents and their agent launches, linked to the sections they concern.

# Command-line interface

* [Command-line reference](cli.md) - Commands grouped by noun with agent and deployment resource boundaries.

# Components

* [Components](components.md) - The parts of afakt and what each one does.

# Glossary

* [Glossary](glossary.md) - Short definitions of afakt's terms, each linked to the section that describes it.
