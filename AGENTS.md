---
type: Agent Instructions
title: Agent instructions
description: Guidance for coding agents working in this repository.
tags: [afakt, agents]
generated: { by: "codex:gpt-6.1-sol" }
---

# Agent instructions

## What this is

`afakt-poc` is a proof of concept repository. P2 adds a Rust command-line skeleton to the shared development tools and docs. The repository uses devtools v0.12.0.

## Crate map

`crates/afakt` builds the `afakt` binary. It uses clap derive for argument parsing and a writer helper for command output. Unit tests cover command dispatch and output. Process tests run the built binary.

## Commands

Run commands from the repository root through mise.

| Command | Purpose |
| --- | --- |
| `mise run bootstrap` | Sync devtools, install tools, and install hooks |
| `mise run lint` | Run the source and configuration checks |
| `mise run check` | Run Rust checks, coverage, docs, and local skill tests |
| `mise run test` | Run unit and integration tests with `nextest` |
| `mise run lint-rust` | Run `clippy` with warnings denied |
| `mise run lint-rustdoc` | Build private-item docs with warnings denied |
| `mise run lint-shear` | Check for unused dependencies |
| `mise run deny` | Check dependency advisories, bans, and sources |
| `mise run vuln` | Check dependency advisories |
| `mise run fmt` | Format Rust code |
| `mise run fmt-check` | Check Rust formatting |
| `mise run coverage` | Enforce measured line and region coverage floors |
| `mise run coverage-report` | Write an HTML coverage report |
| `mise run mutants -- --list` | List mutation candidates without running them |
| `mise run check-docs` | Check Markdown layout and OKF frontmatter |
| `mise run test-skills` | Run the local skill tests with bats |
| `mise run lint-commit-msg` | Check the draft in `COMMIT_AGENTMSG` |

## Rust toolchain

The toolchain pin in `mise.toml` selects Rust 1.97.1 with `clippy`, `rustfmt`, and `llvm-tools`. Workspace crates use edition 2024 and inherit strict `clippy`, Rust, and `rustdoc` lints. Strict lints reject standard output and standard error print macros, unsafe code, `unwrap`, and `expect` in both code and tests. Keep `Cargo.lock` under version control. `nextest` runs without retries and fails on flaky tests. `cargo-deny` rejects duplicate dependency versions. Coverage floors match afakt's measured line and region coverage.

## Conventions

Change only the files the task needs and stage paths by name. Run the checks before committing. Keep private values in the ignored `mise.local.toml` file.

Write each Markdown paragraph on one line. The guard-markdown hook checks this rule. Use Open Knowledge Format (OKF) v0.2 for docs: `docs/index.md` declares `okf_version`, each doc declares its `type`, and draft stubs include `status: draft` and `generated.by`. Run `mise run check-docs` after changing docs.

The devtools payload owns `.devtools/` and `.config/mise/conf.d/`. Edit repository policy in `mise.toml`. Keep every devtools pin on the same release and run `mise run devtools:check-pins` after a pin change.

## Identities

Create commits as `tbhb-agent`. Route reviews through `gh-reviewer` as `tbhb-agent-reviewer`. The local agent skills keep these identities separate.
