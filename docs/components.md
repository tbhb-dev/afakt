---
type: Component Reference
title: Components
description: The parts of afakt and what each one does.
tags: [afakt, components]
status: draft
generated: { by: "claude-code:claude-fable-5-1", at: "2026-10-24T20:00:00Z" }
---

# Components

## API server

The API server serves afakt's REST API, which clients like the CLI and dashboard consume, as well as its coordination workflows.

## Command-line interface (CLI)

The `afakt` command provides one command-line entrypoint for both running afakt's server-side components and communicating with them.

## GitHub webhook receiver

The GitHub webhook receiver accepts webhook deliveries from GitHub.

## Web dashboard

The web dashboard provides afakt's web UI.

## Workflow orchestrator

The workflow orchestrator ([Temporal](https://temporal.io)) provides durable execution for afakt's workflows.

## Workflows

afakt coordinates work using Temporal workflows.

### Agent runner

The Agent runner workflow runs AI coding agents.
