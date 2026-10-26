---
okf_version: "0.2"
---

# Architecture

* [Architecture](architecture.md) - How afakt's components fit together in one binary and how they talk to each other.

# Agents

* [Agents](agent.md) - Durable agents, their lifecycle, owned resources, and operations across harness runs.
* [Agent host](agent-host.md) - A machine's role in the agent fleet, its execution services, trust boundary, and recovery behavior.
* [Agent execution](agent-execution.md) - One harness run belonging to a durable agent, with provider states, generations, retries, placement, and cleanup.
* [Agent execution model](agent-execution-model.md) - The fields of a logical execution and each execution generation, who sets them, which never change, and the provenance record and evidence.
* [Agent execution contract](agent-execution-contract.md) - The execution specification, the execution result, and the launch, status, cancel, and result operations of an execution provider.
* [Agent placement](agent-placement.md) - The agent host registry and the reservation API that bind each execution to an agent host with a free execution slot.
* [Agent execution sequence](agent-execution-sequence.md) - Every message between the parts of an execution, from the start of the agent runner workflow to the result.
* [Agent execution resource](agent-execution-resource.md) - The AgentExecution resource, a Kubernetes-style form of one execution generation, with a full example.
* [Agent workspace](agent-workspace.md) - The workspace sources an agent's workspace comes from and the environment variables the agent and its hooks see.
* [Agent workspace image](agent-workspace-image.md) - Prebuilt workspace directory trees in Open Container Initiative (OCI) images, their sandboxed builds, registry and object storage, and agent host caches for fast start.
* [Agent hooks](agent-hooks.md) - The hooks that run on hook events before, during, and after the agent's work, how profiles and specifications combine them, and which events and decisions each harness supports.
* [Agent observability](agent-observability.md) - The OpenTelemetry trace and spans of an agent execution, the trace context the harness receives, and the metrics afakt publishes.
* [Agent profile](agent-profile.md) - The versioned definition of an agent's behavior, which bounds what an execution specification may request and which a launch validates against.
* [Agent provider requirements](agent-provider-requirements.md) - The numbered requirements that every execution provider meets, written with the BCP 14 key words.
* [Agent session](agent-session.md) - An agent's conversation and retained transcript, with ACP workflows, control states, and suspend and resume.
* [Agent session attachment](agent-session-attachment.md) - Session attachment, writer and reader roles, history replay, and signed agent messages.
* [Agent sandbox](agent-sandbox.md) - The sandbox configuration, sandbox variables, and credential brokering that confine an agent execution, and the execution token that grants its capabilities.
* [Agent MCP gateway](agent-mcp-gateway.md) - One Model Context Protocol (MCP) endpoint per agent harness, with profile tool access, upstream credentials, server hosting, approvals, and audit.
* [Agent ports and previews](agent-ports.md) - Loopback development servers, forwarded port records, authenticated previews and client tunnels, and built-in public tunnel providers.
* [Agent open questions](agent-open-questions.md) - Undecided design questions for agents and their executions, linked to the sections they concern.

# Command-line interface

* [Command-line reference](cli.md) - Commands grouped by noun with agent and deployment resource boundaries.

# Components

* [Components](components.md) - The parts of afakt and what each one does.

# Glossary

* [Glossary](glossary.md) - Short definitions of afakt's terms, each linked to the section that describes it.
