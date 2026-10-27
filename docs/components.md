---
type: Component Reference
title: Components
description: The parts of afakt and what each one does.
tags: [afakt, components]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-25T17:30:00Z" }
---

# Components

## API server

The API server serves afakt's REST API, which clients like the CLI and dashboard consume, as well as its coordination workflows.

### Health endpoints

The server answers `GET /livez`, `GET /readyz`, and `GET /healthz` at its root, outside `/api/v1/`, without authentication. Each returns `application/json`, and none reports on workflows or agents.

- `/livez` returns 200 with `{"status":"pass"}` whenever the server can answer. That includes startup and draining.
- `/readyz` returns 200 with `{"status":"pass"}` only while every required service runs and, for the worker target, the Temporal namespace is reachable. It returns 503 with `{"status":"fail"}` while the server starts, while it drains after SIGINT or SIGTERM, after a required service fails, and while the namespace is unreachable.
- `/healthz` returns `{"status":S,"checks":{"livez":{"status":L},"readyz":{"status":R}}}`, where `L` and `R` match what `/livez` and `/readyz` report at that moment. `S` and the status code are `"pass"` and 200 when both pass, and `"fail"` and 503 otherwise.

## Command-line interface (CLI)

The `afakt` command provides one command-line entrypoint for both running afakt's server-side components and communicating with them.

## GitHub webhook receiver

The GitHub webhook receiver accepts webhook deliveries from GitHub.

## Web dashboard

The web dashboard provides afakt's web UI.

The dashboard shares the origin and HTTP listener of the API, including when every runtime role runs together. Its routes belong under `/dashboard/`. Readiness requires every selected service to run and fails during startup and shutdown.

## Workflow orchestrator

The workflow orchestrator ([Temporal](https://temporal.io)) provides durable execution for afakt's workflows. A worker connects to an external Temporal service or runs an embedded Temporal server inside its own process. See [Workflows](architecture.md#workflows).

The Temporal Web UI shares the HTTP listener with the other runtime roles and gives developers and testers access to the embedded server's namespaces, workflows, and histories. Embedded Temporal operates independently of afakt's operational store.

## Workflows

afakt coordinates work using Temporal workflows.

### Agent runner

The Agent runner workflow runs AI coding agents.
