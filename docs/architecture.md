---
type: Architecture Overview
title: Architecture overview
description: Rust workspace and command structure for afakt-poc.
tags: [afakt, architecture]
status: draft
generated: { by: "codex:gpt-6.1-sol" }
---

# afakt architecture

One unpublished binary crate, `crates/afakt`, forms the workspace. It uses Rust 1.97.1 and edition 2024. A clap parser selects a command, and a dispatch function sends the response through a helper that takes `std::io::Write`. At runtime, `afakt` supplies locked standard output and returns write errors to the caller.

`status` prints a placeholder response. Unit tests cover command output, newlines, and write failures. Process tests run the binary for version, help, and status. Run `mise run check` to test the code, enforce coverage, check dependencies, and check docs. Workspace lints deny print macros and unsafe code.
