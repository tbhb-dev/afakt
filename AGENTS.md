---
type: Agent Instructions
title: Agent instructions
description: Guidance for coding agents working in this repository.
tags: [afakt, agents]
generated: { by: "claude-code:claude-sonnet-5-5", at: "2026-10-20T05:55:00Z" }
---

# Agent instructions

## What this is

`afakt` is a proof of concept command-line tool written in Go. It has one command, `status`, which prints a placeholder line. The repository also has shared development tools, docs, and CI gates. It uses devtools v0.12.0.

## Layout

`cmd/afakt/` holds the command-line entry point, the argument parser, the `status` command, and their tests. `internal/buildmeta/` resolves version metadata from linker flags or runtime build info.

`docs/` holds the OKF docs bundle. `scripts/okf-check.py` checks its format and the Markdown frontmatter across the repository.

`.devtools/` and `.config/mise/conf.d/` hold the devtools v0.12.0 payload. The `vendir` tool owns both directories. Edit repository settings in the root configuration files.

`.agents/skills/agent-*` holds the local agent skills. APM deploys skills and rules to `.agents/` and `.claude/` from `apm.yml`.

## Commands

Read `mise.toml` before trusting it. In each fresh worktree, run `mise trust`, then `mise run bootstrap` to sync the payload and install the tools and hooks.

| Command | Purpose |
| --- | --- |
| `mise run devtools:apm-sync` | Deploy agent tooling |
| `mise run lint` | Check source gates |
| `mise run check` | Check docs, local skills, and Go coverage |
| `mise run build` | Build `bin/afakt` with version metadata |
| `mise run test` | Run the Go tests, as `test-go` does |
| `mise run test-go` | Run the Go tests |
| `mise run test-go-race` | Run the Go tests with the race detector |
| `mise run cover-go` | Measure coverage and enforce its floors |
| `mise run cover-go-check` | Check the existing coverage profile |
| `mise run lint-go-all` | Run the Go analysis gates |
| `mise run lint-go-fmt` | Check Go formatting |
| `mise run lint-go-arch` | Check the layering rules in `.go-arch-lint.yml` |
| `mise run vuln` | Scan reachable Go vulnerabilities |
| `mise run devtools:gomodscan` | Scan module dependencies |
| `mise run vendor` | Tidy `go.mod` and refresh `vendor/` |
| `mise run vendor-check` | Fail when `go.mod`, `go.sum`, or `vendor/` drift from the committed tree |
| `mise run mutate-go` | Run mutation testing |
| `mise run test-skills` | Run the local skills' bats suites |
| `mise run lint-skills` | Check skill prose and scripts |
| `mise run check-docs` | Check the docs bundle and repository Markdown |
| `mise run lint-commit-msg` | Check the draft in `COMMIT_AGENTMSG` |
| `mise run devtools:prek-all` | Run applicable hooks |

## Go toolchain

The shared payload pins Go 1.27.2. Use `mise exec -- go version` to inspect it. Module `github.com/tbhb-dev/afakt` declares `go 1.27` and `toolchain go1.27.2`. Its packages import only the standard library.

Root configuration includes the shared Go task file and exposes local build and test tasks. Go tool directives declare the analyzer dependencies in the module. `DEVTOOLS_GO_MODULES_MODE=vendor` loads them from the tracked `vendor/` tree. Commit `go.sum` with those generated sources. Run `mise run vendor` to refresh them and `mise run vendor-check` to check for drift.

`.testcoverage.yml` sets the coverage floors: 100% for each file, each package, and the total. The measured total is 100% (88 of 88 statements). The floors skip `main.go`, whose `os.Exit` glue can't run inside a test process. `mise run build` stamps the version from `git describe`, so `afakt --version` prints a tag or a short commit. A build without stamping, such as `go run`, prints `0.1.0`.

## Conventions

Change only the files the task needs and stage paths by name. Run the checks before committing. Keep private values in the ignored `mise.local.toml` file.

Write each Markdown paragraph on one line. The guard-markdown hook checks this rule. Use Open Knowledge Format (OKF) v0.2 for docs: `docs/index.md` declares `okf_version`, each doc declares its `type`, and draft stubs include `status: draft` and `generated.by`. Run `mise run check-docs` after changing docs.

The devtools payload owns `.devtools/` and `.config/mise/conf.d/`. Edit repository policy in `mise.toml`. Keep every devtools pin on the same release and run `mise run devtools:check-pins` after a pin change.

## Identities

Create commits as `tbhb-agent`. Route reviews through `gh-reviewer` as `tbhb-agent-reviewer`. The local agent skills keep these identities separate.

## Workflows

All scheduled times use Coordinated Universal Time (UTC).

`ci.yml` runs on pull requests, pushes to `main`, and manual dispatch. It checks lint on Linux and runs the full check on Linux and macOS. Separate jobs enforce the Go coverage floors, scan reachable vulnerabilities (`vuln`), and check payload and Go vendor drift (`vendor-check`). The required `test` summary fails unless every matrix slot succeeds.

`security.yml` runs on pull requests and pushes to `main`. It also runs daily at 06:26 and on manual dispatch. `gitleaks` scans history through the head of the branch on pull requests. The Go scanners upload reports for reachable vulnerabilities and dependencies in `vendor/`.

`scheduled.yml` runs daily at 04:17 and on manual dispatch. It runs Go mutation testing and uploads the coverage profile.

`codeql.yml` analyzes Go on pull requests and pushes to `main`. It also runs on Sundays at 03:39 and on manual dispatch.

`lint-codeowners.yml` checks CODEOWNERS on every pull request only.

`lint-workflows.yml` runs when a pull request or a push to `main` changes a workflow. Manual dispatch runs it too.

`renovate-config-validator.yaml` validates Renovate configuration. It runs when a pull request or push to `main` changes `**renovate.json5`, and it also accepts manual dispatch.

`renovate.yaml` runs Renovate at minute 11 of every hour on weekdays and every two hours on weekends. Manual dispatch can enable debug logging.
