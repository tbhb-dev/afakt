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
- `internal/lifecycle/`: starts and stops services. `internal/lifecycle/lifecycletest/` has fakes for tests.

The repository root is a pnpm workspace. `pnpm-workspace.yaml` lists `apps/*`, and `pnpm-lock.yaml` is the one lockfile for the workspace. The root also keeps the configuration that every web package shares: `tsconfig.base.json`, `biome.json`, `eslint.config.base.ts`, `vitest.shared.ts`, `.jscpd.json` and `stryker.config.json`. `apps/dashboard/` holds the web dashboard, a React and Vite app that renders a blank page, with its unit tests in `src/` and its Playwright tests in `e2e/`. `scripts/check-web-gate.sh` shows that each web gate fails on broken input it writes into a temporary copy of the workspace.

`docs/` holds the OKF docs bundle. `scripts/okf-check.py` checks its format and the Markdown frontmatter across the repository.

`.devtools/` and `.config/mise/conf.d/` hold the devtools v0.19.0 payload. The `vendir` tool owns both directories. Edit repository settings in the root configuration files.

APM deploys the skills and rules from the devtools packages named in `apm.yml` to `.agents/` and `.claude/`. That includes `agent-commit`, `agent-pr`, and `agent-merge-pr`. `.agents/agent-skills.conf` holds this repository's policy for them, such as the lint commands and which changed paths go to the owner. The machine supplies the reviewer, the GitHub account, the required reviewer, and the delegation. The repository supplies none of them.

## Commands

Read `mise.toml` before trusting it. In each fresh worktree, run `mise trust`, then `mise run bootstrap` to sync the payload and install the tools and hooks. Bootstrap installs the Node (24.21.0) and pnpm (12.11.2) that `mise.toml` pins. The root Node pin overrides the payload's and must stay equal to it. For web work, run `mise run install-web` and `mise run install-web-browsers` next. The Go tasks run without those web packages and the browser.

| Command | Purpose |
| --- | --- |
| `mise run bootstrap` | Sync the payload, install the toolchain, and install git hooks |
| `mise run install-toolchain` | Install the pinned toolchain from `mise.toml` and `.config/mise/conf.d/` |
| `mise run install-tools` | Download the Vale style packages `.vale.ini` names |
| `mise run devtools:apm-sync` | Deploy agent tooling |
| `mise run lint` | Check source gates |
| `mise run check` | Check docs, Go tests, and Go coverage |
| `mise run build` | Build `bin/afakt` with version metadata |
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
| `mise run install-web` | Install the web workspace from the root lockfile without changing it |
| `mise run install-web-browsers` | Install the browser build the pinned Playwright release runs |
| `mise run dev-web` | Serve the dashboard from the Vite development server |
| `mise run build-web` | Build the dashboard's static assets into `apps/dashboard/dist` |
| `mise run test-web` | Run the dashboard's unit tests |
| `mise run cover-web` | Run the unit tests and enforce 100% coverage per file and in total |
| `mise run e2e-web` | Load the dashboard in the pinned browser and fail on console errors or accessibility violations |
| `mise run mutate-web` | Run mutation testing on the dashboard and fail unless the tests detect every mutant |
| `mise run audit-web` | Fail on a high or critical advisory in any locked web dependency, or on an incomplete audit |
| `mise run lint-web-all` | Run Biome, ESLint, the type check, jscpd, and knip |
| `mise run check-web-gate <gate>` | Show that one web gate fails on generated broken input |
| `mise run check-web-gates` | Show that every web gate fails on generated broken input |
| `mise run check-web` | Run every web gate |

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

`ci.yml` runs on pull requests, pushes to `main`, and manual dispatch. It checks lint on Linux and runs the full check on Linux and macOS. Separate jobs enforce the Go coverage floors, scan reachable vulnerabilities (`vuln`), and check payload and Go vendor drift (`vendor-check`). A `web` matrix runs each web gate in its own Linux job: `lint-web-all`, `cover-web`, `e2e-web`, `mutate-web`, `audit-web` and `check-web-gates`. The required `test` summary fails unless every slot of the Go and web matrices succeeds.

`security.yml` runs on pull requests and pushes to `main`. It also runs daily at 06:26 and on manual dispatch. `gitleaks` scans history through the head of the branch on pull requests. The Go scanners upload reports for reachable vulnerabilities and dependencies in `vendor/`.

`scheduled.yml` runs daily at 04:17 and on manual dispatch. It runs Go mutation testing and uploads the coverage profile.

`codeql.yml` analyzes Go on pull requests and pushes to `main`. It also runs on Sundays at 03:39 and on manual dispatch.

`lint-codeowners.yml` checks CODEOWNERS on every pull request only.

`lint-workflows.yml` runs when a pull request or a push to `main` changes a workflow. Manual dispatch runs it too.

`renovate-config-validator.yaml` validates Renovate configuration. It runs when a pull request or push to `main` changes `**renovate.json5`, and it also accepts manual dispatch.

`renovate.yaml` runs Renovate at minute 11 of every hour on weekdays and every two hours on weekends. Manual dispatch can enable debug logging.
