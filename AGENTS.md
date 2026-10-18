---
type: Agent Instructions
title: Agent instructions
description: Guidance for coding agents working in this repository.
tags: [afakt, agents]
generated: { by: "codex:gpt-6.1-sol" }
---

# Agent instructions

## What this is

`afakt-poc` is a proof of concept repository. P1 sets up shared development tools and starter docs. The repository uses devtools v0.12.0.

## Crate map

P1 contains no Rust crates. P2 adds the `afakt` crate and binary.

## Commands

Run commands from the repository root through mise.

| Command | Purpose |
| --- | --- |
| `mise run bootstrap` | Sync devtools, install tools, and install hooks |
| `mise run lint` | Run the source and configuration checks |
| `mise run check` | Check docs and run the local skill tests |
| `mise run check-docs` | Check Markdown layout and OKF frontmatter |
| `mise run test-skills` | Run the local skill tests with bats |
| `mise run lint-commit-msg` | Check the draft in `COMMIT_AGENTMSG` |

## Rust toolchain

Rust code, toolchain pins, and Rust checks belong to P2. The P1 `lint` and `check` tasks cover the files that exist here.

## Conventions

Change only the files the task needs and stage paths by name. Run the checks before committing. Keep private values in the ignored `mise.local.toml` file.

Write each Markdown paragraph on one line. The guard-markdown hook checks this rule. Use Open Knowledge Format (OKF) v0.2 for docs: `docs/index.md` declares `okf_version`, each doc declares its `type`, and draft stubs include `status: draft` and `generated.by`. Run `mise run check-docs` after changing docs.

The devtools payload owns `.devtools/` and `.config/mise/conf.d/`. Edit repository policy in `mise.toml`. Keep every devtools pin on the same release and run `mise run devtools:check-pins` after a pin change.

## Identities

Create commits as `tbhb-agent`. Route reviews through `gh-reviewer` as `tbhb-agent-reviewer`. The local agent skills keep these identities separate.
