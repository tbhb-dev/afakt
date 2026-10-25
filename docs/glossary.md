---
type: Glossary
title: Glossary
description: Short definitions of afakt's component and workflow names, each linked to its section in the component reference.
tags: [afakt, glossary]
status: draft
generated: { by: "claude-code:claude-fable-5-1", at: "2026-10-24T20:00:00Z" }
---

# Glossary

- **Agent runner**: the workflow that runs AI coding agents. See [Agent runner](components.md#agent-runner).
- **API server**: the component that serves afakt's REST API to clients like the CLI and dashboard, and to its coordination workflows. See [API server](components.md#api-server).
- **Command-line interface (CLI)**: the `afakt` command, one command-line entrypoint for running afakt's server-side components and communicating with them. See [Command-line interface (CLI)](components.md#command-line-interface-cli).
- **GitHub webhook receiver**: the component that accepts webhook deliveries from GitHub. See [GitHub webhook receiver](components.md#github-webhook-receiver).
- **Web dashboard**: the component that provides afakt's web UI. See [Web dashboard](components.md#web-dashboard).
- **Workflow**: a Temporal workflow through which afakt coordinates work, such as Agent runner. See [Workflows](components.md#workflows).
- **Workflow orchestrator**: the component, Temporal, that provides durable execution for afakt's workflows. See [Workflow orchestrator](components.md#workflow-orchestrator).
