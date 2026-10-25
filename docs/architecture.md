---
type: Architecture Overview
title: Architecture
description: How afakt's components fit together in one binary and how they talk to each other.
tags: [afakt, architecture]
status: draft
generated: { by: "claude-code:claude-fable-5-1", at: "2026-10-24T22:00:00Z" }
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
