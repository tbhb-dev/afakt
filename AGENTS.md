---
type: Agent Instructions
title: Agent instructions
description: Guidance for coding agents working in this repository.
tags: [afakt, agents]
generated: { by: "claude-code:claude-sonnet-5-5", at: "2026-10-20T05:55:00Z" }
---

# Agent instructions

## What this is

`afakt` is a Go command-line tool. The repository also has shared development tools, docs, and CI gates.

## Layout

- `cmd/afakt/`: the entry point and its process tests.
- `internal/cmd/`: the Cobra command tree, one package per command. `internal/cmd/root/` maps errors to exit codes.
- `internal/cmdutil/`: what the commands share.
- `internal/buildmeta/`: version metadata.
- `internal/server/`: composes the services of the selected server targets.
- `internal/dashboard/`: serves embedded assets under `/dashboard`. The `dashboard_production` build tag selects generated assets. Go-only builds use the committed stand-in page.
- `internal/lifecycle/`: starts and stops services. `internal/lifecycle/lifecycletest/` has fakes for tests.

A pnpm workspace at the repository root has one dependency lock and keeps the configuration that every web package shares. In `apps/dashboard/`, a React and Vite app renders the blank dashboard page, with its unit tests in `src/` and its Playwright tests in `e2e/`. For each web gate, `scripts/check-web-gate.sh` writes broken input into a temporary copy of the workspace and shows that the gate fails.

`docs/` holds the OKF docs bundle. `scripts/okf-check.py` checks its format and the Markdown frontmatter across the repository.

`.devtools/` and `.config/mise/conf.d/` hold the devtools payload. The `vendir` tool owns both directories. Edit repository settings in the root configuration files.

APM deploys the skills and rules from the devtools packages named in `apm.yml` to `.agents/` and `.claude/`. That includes `agent-commit`, `agent-pr`, and `agent-merge-pr`. `.agents/agent-skills.conf` holds this repository's policy for them, such as the lint commands and which changed paths go to the owner. The machine supplies the reviewer, the GitHub account, the required reviewer, and the delegation. The repository supplies none of them.

## Commands

Read `mise.toml` before trusting it. In each fresh worktree, run `mise trust`, then `mise run bootstrap` to sync the payload and install the tools and hooks. Bootstrap installs the Node and pnpm that `mise.toml` pins. The root Node pin overrides the payload's and must stay equal to it. For web work, run `mise run install-web` and `mise run install-web-browsers` next, then `mise run check-web` to run every web gate. The Go tasks run without those web packages and the browser.

| Command | Purpose |
| --- | --- |
| `mise run bootstrap` | Sync the payload, install the toolchain, and install git hooks |
| `mise run install-toolchain` | Install the pinned toolchain from `mise.toml` and `.config/mise/conf.d/` |
| `mise run install-tools` | Download the Vale style packages `.vale.ini` names |
| `mise run devtools:apm-sync` | Deploy agent tooling |
| `mise run lint` | Check source gates |
| `mise run check` | Check docs, Go tests, and Go coverage |
| `mise run build` | Build `bin/afakt` with version metadata |
| `mise run check-dashboard` | Build production assets, test embedded handlers and server contracts, and check the isolated binary in the pinned browser |
| `mise run e2e-dashboard` | Check the already-built binary with the pinned browser and accessibility scan |
| `mise run check-dashboard-gates` | Prove failed-build cleanup, entry revalidation, browser failures, and required CI result handling |
| `mise run check-dashboard-all` | Install the web dependencies and browser, then run the dashboard checks and failure proofs sequentially |
| `mise run check-go-only` | Run every Go-only gate in an environment without web tools, dependencies, or output |
| `mise run test` | Run the Go tests, as `test-go` does |
| `mise run test-go` | Run the Go tests |
| `mise run test-go-race` | Run the Go tests with the race detector |
| `mise run cover-go` | Measure coverage and enforce its floors |
| `mise run cover-go-check` | Check the existing coverage profile |
| `mise run lint-go-all` | Run the Go analysis gates |
| `mise run lint-go-fmt` | Check Go formatting |
| `mise run lint-go-arch` | Check the layering rules in `.go-arch-lint.yml` |
| `mise run lint-go-deadcode` | Fail if `deadcode` finds unreachable functions from the binary entry points |
| `mise run lint-ast-grep` | Run the `ast-grep` rules in `.ast-grep/rules` across the repository, such as the ban on `docs/` references in Go comments and strings |
| `mise run bundle-openapi` | Bundle the OpenAPI document into its single-file bundle |
| `mise run check-openapi-bundle` | Fail when the committed OpenAPI bundle is out of date |
| `mise run lint-openapi` | Lint the OpenAPI document and its bundle |
| `mise run vuln` | Scan reachable Go vulnerabilities |
| `mise run devtools:gomodscan` | Scan module dependencies |
| `mise run vendor` | Tidy `go.mod` and refresh `vendor/` |
| `mise run vendor-check` | Fail when `go.mod`, `go.sum`, or `vendor/` drift from the committed tree |
| `mise run mutate-go` | Run mutation testing |
| `mise run lint-prose` | Lint prose with Vale |
| `mise run lint-draft` | Lint one drafted document through every prose gate at once |
| `mise run fix-prose-replacements` | Apply every Vale finding that has its own replacement |
| `mise run lint-skills` | Check the prose of `.claude/rules` |
| `mise run check-docs` | Check the docs bundle and repository Markdown |
| `mise run lint-pr-description` | Check the pull request draft in `PR_AGENTDESC.md` |
| `mise run lint-commit-msg` | Check the draft in `COMMIT_AGENTMSG` |
| `mise run lint-squash-msg` | Check the squash draft in `SQUASH_AGENTMSG` |
| `mise run devtools:prek-all` | Run applicable hooks |

## Go toolchain

The shared payload pins Go 1.27.3. Use `mise exec -- go version` to inspect it. Module `github.com/tbhb-dev/afakt` declares `go 1.27` and `toolchain go1.27.3`.

Root configuration includes the shared Go task file and exposes local build and test tasks. Go tool directives declare the analyzer dependencies in the module. `DEVTOOLS_GO_MODULES_MODE=vendor` loads them from the tracked `vendor/` tree. Commit `go.sum` with those generated sources. Run `mise run vendor` to refresh them and `mise run vendor-check` to check for drift.

`.testcoverage.yml` sets 100% coverage floors for each file, each package, and the total. They skip `main.go`, whose `os.Exit` glue can't run inside a test process.

## Conventions

Change only the files the task needs and stage paths by name. Run the checks before committing. Keep private values in the ignored `mise.local.toml` file.

Write each Markdown paragraph on one line. The guard-markdown hook checks this rule. Use Open Knowledge Format (OKF) v0.2 for docs: `docs/index.md` declares `okf_version`, each doc declares its `type`, and draft stubs include `status: draft` and `generated.by`. Run `mise run check-docs` after changing docs.

The devtools payload owns `.devtools/` and `.config/mise/conf.d/`. Edit repository policy in `mise.toml`. Keep every devtools pin on the same release, including the Go tool versions in `mise.toml`, and run `mise run devtools:check-pins` after a pin change.

## Identities

Create commits as `tbhb-agent`. Route reviews through `gh-reviewer` as `tbhb-agent-reviewer`. The agent skills from the devtools packages keep these identities separate, and the accounts come from the machine configuration.

## Workflows

All scheduled times use Coordinated Universal Time (UTC).

`ci.yml` runs on pull requests, pushes to `main`, and manual dispatch. It runs bootstrap and checks lint on Linux, and it fails when bootstrap changes a tracked file. It runs the full check on Linux and macOS. Separate jobs enforce the Go coverage floors, scan reachable vulnerabilities (`vuln`), and check payload and Go vendor drift (`vendor-check`). A `web` matrix runs each web gate in its own Linux job. The `dashboard` matrix checks the release and generated failure proofs. A separate container runs Go-only gates without the web toolchain. The required `test` summary fails unless the Go, web, and dashboard matrices and the Go-only job succeed.

`security.yml` runs on pull requests and pushes to `main`. It also runs daily at 06:26 and on manual dispatch. `gitleaks` scans history through the head of the branch on pull requests. The Go scanners upload reports for reachable vulnerabilities and dependencies in `vendor/`.

`scheduled.yml` runs daily at 04:17 and on manual dispatch. It runs Go mutation testing and uploads the coverage profile.

`codeql.yml` analyzes Go on pull requests and pushes to `main`. It also runs on Sundays at 03:39 and on manual dispatch.

`lint-codeowners.yml` checks CODEOWNERS on every pull request only.

`lint-workflows.yml` runs when a pull request or a push to `main` changes a workflow. Manual dispatch runs it too.

`renovate-config-validator.yaml` validates Renovate configuration. It runs when a pull request or push to `main` changes `**renovate.json5`, and it also accepts manual dispatch.

`renovate.yaml` runs Renovate at minute 11 of every hour on weekdays and every two hours on weekends. Manual dispatch can enable debug logging. `.github/renovate.json5` runs its `mise lock` step with the mise release that `ci.yml` pins.
