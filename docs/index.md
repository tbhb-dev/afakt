---
okf_version: "0.2"
---

<!-- generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" } -->

# Architecture

* [Architecture](architecture.md) - How afakt's components fit together in one binary and how they talk to each other.
* [Network topology](network-topology.md) - Cluster, agent host, and agent launch networks with trust boundaries and authenticated traffic flows.
* [Public key infrastructure](pki.md) - Authorities, trust boundaries, certificate policies, and key custody.
* [Workload identity keys](pki-workload.md) - Workload authorities, identity documents, enrollment keys, and their custody.
* [Transport certificates and keys](pki-transport.md) - Overlay credentials, public and inspection certificates, and storage transport trust.
* [Signing keys and secrets](pki-signing.md) - Authentication secrets, delegated keys, artifact signatures, and encryption at rest.
* [Public key infrastructure operations](pki-operations.md) - Issuance, renewal, recovery, administrative operations, and threats.

# Agents

* [Agents](agent.md) - Durable agent identity, configuration, owned resources, and operations across agent launches.
* [Agent lifecycle](agent-lifecycle.md) - Agent, launch, and session states from creation through suspension, capture, restoration, and deletion.
* [Agent host](agent-host.md) - A machine's role in the agent fleet, its agent launch services, trust boundary, and recovery behavior.
* [Agent launch](agent-launch.md) - One instantiation of an agent in an agent sandbox on an agent host, with lifecycle, retries, placement, and cleanup.
* [Agent launch model](agent-launch-model.md) - The fields and identity of each agent launch, who sets them, which never change, and the provenance record and evidence.
* [Agent launch contract](agent-launch-contract.md) - The agent launch spec, the agent launch result, and the provision, restore, release, status, cancel, and result operations of an agent sandbox provider.
* [Agent placement](agent-placement.md) - The agent host registry and the reservation API that bind each agent launch to an agent host with a free agent launch slot.
* [Agent launch sequence](agent-launch-sequence.md) - Every message between the parts of an agent launch, from the start of the agent dispatcher workflow to the result.
* [Agent launch resource](agent-launch-resource.md) - The AgentLaunch resource, the resource form of one agent launch, with a full example.
* [Agent workspace](agent-workspace.md) - The workspace sources an agent's workspace comes from and the environment variables the agent and its hooks see.
* [Agent workspace lifecycle](agent-workspace-lifecycle.md) - Workspace creation, materialization, launch attachment, capture, retention, and deletion across agent sandbox providers.
* [Agent workspace image](agent-workspace-image.md) - Prebuilt workspace directory trees in Open Container Initiative (OCI) images, their sandboxed builds, registry and object storage, and agent host caches for fast start.
* [Agent hooks](agent-hooks.md) - The hooks that run on hook events before, during, and after the agent's work, how profiles and specifications combine them, and which events and decisions each agent harness supports.
* [Agent observability](agent-observability.md) - The OpenTelemetry trace and spans of an agent launch, the trace context the harness receives, and the metrics afakt publishes.
* [Agent profile](agent-profile.md) - The blueprint for creating an agent, with configuration templates and bounds for finalizing each agent launch spec.
* [Agent overlay and names](agent-overlay.md) - Workload overlay nodes, direct connections, internal names, and ingress access.
* [Agent workload](agent-workload.md) - Guest, mediator, proxy, firewall, and launch boundary.
* [Agent workload guest](agent-workload-guest.md) - Guest routing, containment, and its private proxy link.
* [Agent proxy](agent-workload-proxy.md) - Workload traffic, direct overlay connections, and mutual authentication.
* [Agent session](agent-session.md) - An agent's durable conversation, retained transcripts, and session workflow.
* [Agent ACP](agent-acp.md) - Agent Client Protocol connections, message policy, turns, and writer and reader roles.
* [Agent session attachment](agent-session-attachment.md) - Session attachment, history replay, and signed agent messages.
* [Agent sandbox](agent-sandbox.md) - The sandbox configuration, sandbox variables, credential brokering, and authentication of an agent launch.
* [Agent mediator](agent-workload-mediator.md) - Per-launch guest connections, network policy, and control requests.
* [Agent sandbox lifecycle](agent-sandbox-lifecycle.md) - Provisioning, readiness, suspension, restoration, and teardown across providers.
* [Agent sandbox provider for macOS VM](agent-sandbox-macos-vm.md) - Isolation, mediation, storage, and lifecycle for macos-vm.
* [Agent sandbox provider for Linux VM](agent-sandbox-linux-vm.md) - Isolation, mediation, storage, and lifecycle for linux-vm.
* [Agent sandbox provider for Windows VM](agent-sandbox-windows-vm.md) - Isolation, mediation, storage, and lifecycle for windows-vm.
* [Agent sandbox provider for macOS host](agent-sandbox-macos-host.md) - Isolation, mediation, storage, and lifecycle for macos-host.
* [Agent sandbox provider for Linux host](agent-sandbox-linux-host.md) - Isolation, mediation, storage, and lifecycle for linux-host.
* [Agent sandbox provider for Windows host](agent-sandbox-windows-host.md) - Isolation, mediation, storage, and lifecycle for windows-host.
* [Agent sandbox provider for Docker Engine](agent-sandbox-docker-engine.md) - Isolation, mediation, storage, and lifecycle for docker-engine.
* [Agent sandbox provider for Docker Desktop on macOS](agent-sandbox-docker-desktop-macos.md) - Isolation, mediation, storage, and lifecycle for docker-desktop-macos.
* [Agent sandbox provider for Docker Desktop on Windows](agent-sandbox-docker-desktop-windows.md) - Isolation, mediation, storage, and lifecycle for docker-desktop-windows.
* [Agent sandbox provider for Podman](agent-sandbox-podman.md) - Isolation, mediation, storage, and lifecycle for Podman.
* [Agent sandbox provider for Claude Managed Agents](agent-sandbox-claude-managed-agents.md) - Remote isolation, mediation, storage, and lifecycle for Claude Managed Agents.
* [Agent sandbox provider for OpenAI Agents API](agent-sandbox-openai-agents-api.md) - Remote isolation, mediation, storage, and lifecycle for OpenAI Agents API.
* [Agent sandbox provider for Modal Sandboxes](agent-sandbox-modal.md) - Remote isolation, mediation, storage, and lifecycle for Modal Sandboxes.
* [Agent sandbox provider for Cloudflare Sandbox SDK](agent-sandbox-cloudflare.md) - Remote isolation, mediation, storage, and lifecycle for Cloudflare Sandbox SDK.
* [Agent MCP gateway](agent-mcp-gateway.md) - One Model Context Protocol (MCP) endpoint per agent harness, with profile tool access, upstream credentials, server hosting, approvals, and audit.
* [Agent ports](agent-ports.md) - Loopback listeners, forwarded port records, connections between agents, and authenticated client tunnels.
* [Agent previews](agent-previews.md) - Preview names, endpoints, viewer access, tunnel providers, and lifetime.
* [Agent services](agent-services.md) - Published service names, port mappings, access, and lifetime across launches.
* [Agent threat model](agent-threat-model.md) - Shared assets, service boundaries, and threats across agent launches.
* [Agent threat model for macOS VM](agent-threat-model-macos-vm.md) - Provider boundaries and threats for macOS VM.
* [Agent threat model for Linux VM](agent-threat-model-linux-vm.md) - Provider boundaries and threats for Linux VM.
* [Agent threat model for Windows VM](agent-threat-model-windows-vm.md) - Provider boundaries and threats for Windows VM.
* [Agent threat model for macOS host](agent-threat-model-macos-host.md) - Provider boundaries and threats for macOS host.
* [Agent threat model for Linux host](agent-threat-model-linux-host.md) - Provider boundaries and threats for Linux host.
* [Agent threat model for Windows host](agent-threat-model-windows-host.md) - Provider boundaries and threats for Windows host.
* [Agent threat model for Docker Engine on Linux](agent-threat-model-docker-engine.md) - Provider boundaries and threats for Docker Engine on Linux.
* [Agent threat model for Docker Desktop on macOS](agent-threat-model-docker-desktop-macos.md) - Provider boundaries and threats for Docker Desktop on macOS.
* [Agent threat model for Docker Desktop on Windows](agent-threat-model-docker-desktop-windows.md) - Provider boundaries and threats for Docker Desktop on Windows.
* [Agent threat model for Podman](agent-threat-model-podman.md) - Provider boundaries and threats for Podman.
* [Agent threat model for Claude Managed Agents](agent-threat-model-claude-managed-agents.md) - Remote execution and vendor custody threats for Claude Managed Agents.
* [Agent threat model for OpenAI Agents API](agent-threat-model-openai-agents-api.md) - Remote execution and vendor custody threats for OpenAI Agents API.
* [Agent threat model for Modal Sandboxes](agent-threat-model-modal.md) - Remote execution and vendor custody threats for Modal Sandboxes.
* [Agent threat model for Cloudflare Sandbox SDK](agent-threat-model-cloudflare.md) - Remote execution and vendor custody threats for Cloudflare Sandbox SDK.
* [Agent open questions](agent-open-questions.md) - Undecided design questions for agents and their agent launches, linked to the sections they concern.

# Command-line interface

* [Command-line reference](cli.md) - Commands grouped by noun with agent and deployment resource boundaries.

# Components

* [Components](components.md) - The parts of afakt and what each one does.

# Glossary

* [Glossary](glossary.md) - Short definitions of afakt's terms, each linked to the section that describes it.
