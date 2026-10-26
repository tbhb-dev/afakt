---
type: README
title: afakt
description: A proof of concept Go command-line tool with shared development tools and docs.
tags: [afakt, readme]
generated: { by: "claude-code:claude-sonnet-5-5", at: "2026-10-20T05:55:00Z" }
---

# afakt

`afakt` is a Go command-line tool. Build `bin/afakt` with `mise run build`, and run `bin/afakt --help` to list its commands. See [docs/components.md](docs/components.md) for the parts of afakt.

`afakt server --target=webhook` serves only health endpoints and has no receivers yet.

Run `afakt server --target=all --temporal-mode=embedded` to serve every target with the Temporal Web UI. It runs without external services or a store backend. `all` expands to `api,webhook,dashboard,worker` and takes the same configuration. Run `mise run check-all-targets` to build the binary and check the pages and health transitions in the pinned browser.

`afakt server --target=dashboard` serves the embedded dashboard at `/dashboard` and `/dashboard/`, with `/livez` and `/readyz` at the server root. Add the API with `--target=api,dashboard`. Both share `--listen` and the shutdown deadline. The page is blank. Paths outside `/dashboard` don't redirect to it.

The embedded worker serves the Temporal Web UI at `http://127.0.0.1:8080/temporal/` by default. It's a development and test facility with no authentication of its own. Changing the listen address can expose it beyond loopback. Server help describes its path configuration.

Install the web dependencies with `mise run install-web` before `mise run build`. The build compiles the production web project and embeds it in the binary. A failed web build removes any previous `bin/afakt`. Ordinary Go builds and tests don't need the web toolchain. They embed the committed stand-in page, which says that dashboard assets are unavailable. Run `mise run check-dashboard-all` for the production handler tests, binary browser check, and generated failure proofs in sequence.

The repository also has shared development tools and docs. See [AGENTS.md](AGENTS.md) for repository guidance and [docs/index.md](docs/index.md) for the docs index.
