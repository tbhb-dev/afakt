---
type: Architecture Overview
title: Architecture
description: How afakt's components fit together in one binary and how they talk to each other.
tags: [afakt, architecture]
status: draft
generated: { by: "claude-code:claude-opus-5-5", at: "2026-10-25T11:44:54Z" }
---

# Architecture

afakt is a modular monolith. One binary contains every component in [Components](components.md), and `afakt server --target=api|webhook|dashboard|worker|all` selects which runtime roles a process takes on.

![Component architecture](diagrams/component-architecture.svg)

## Clients and the API

The API server exposes afakt's REST API. The web dashboard, the command-line interface, and the coordination workflows call it, and native apps follow later. The binary serves the web dashboard from the same origin as the API. The API server also starts workflows on the workflow orchestrator.

## GitHub

The GitHub webhook receiver accepts webhook deliveries from GitHub. It calls the API server and starts workflows on the workflow orchestrator. In the other direction, the API server and the workflows call the GitHub API.

## Workflows

Temporal is the workflow orchestrator. It runs afakt's workflows on workers inside the binary, and a workflow can start other workflows. The Agent runner workflow runs AI coding agents.

A process with the worker target connects to Temporal in the mode that `--temporal-mode` names. In `external` mode it connects to the Temporal frontend at `--temporal-address`. In `embedded` mode, for development and tests, it runs a Temporal server inside the process on its own loopback ports, with its data in memory or in `--temporal-embedded-dir`. Both modes use the namespace that `--temporal-namespace` names, `default` unless set. The worker target establishes the connection. Workflow registration and polling remain separate changes.

The embedded adapter owns startup and cleanup. Cancellation starts the shutdown timeout while startup is still in progress. If a dependency call ignores cancellation, the adapter stops waiting at the deadline, reports an error, and keeps cleanup tracked for when that call returns. Construction failures release the services and public persistence handles in reverse order. The HTTP constructor still closes and rebinds a reserved port, and Temporal's SQLite plugin retains its private connection pool after public handles close. These remaining ownership limits are tracked in [issue #73](https://github.com/tbhb-dev/afakt/issues/73).

Every process answers `/livez` and `/readyz` on its listener. `/livez` reports that the process answers at all. `/readyz` passes once every service has started and, with the worker target, only while the Temporal namespace is reachable. It fails again when the shutdown begins.
