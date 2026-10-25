---
type: Agent Instructions
title: Docs agent instructions
description: The rules for agents that add or change pages in the docs directory.
tags: [afakt, agents, docs]
status: stable
generated: { by: "claude-code:claude-sonnet-5-5", at: "2026-10-25T00:40:00Z" }
---

# Docs agent instructions

These rules apply inside `docs/`. The repository-wide rules, commands, and identities are in the [root AGENTS.md](../AGENTS.md), and [README.md](README.md) says what belongs in this directory.

## Frontmatter

Start every page except `index.md` and `log.md` with a YAML block that holds these keys:

- `type`: required, non-empty text.
- `title` and `description`: one line each.
- `tags`: a list of strings.
- `status`: `draft`, `stable`, or `deprecated`.
- `generated`: `{ by, at }`, with `by` set to the producer and model (the existing pages use `claude-code:claude-sonnet-5-5`) and `at` a timestamp in the form `2026-10-25T00:40:00Z`.

Never add `verified`. Only a person does.

`scripts/okf-check.py` requires only `type`, a `status` from the three values when one is present, and a non-empty `generated.by` when `generated` is present. It doesn't check the other keys, so check them yourself before you finish.

## Writing

Put each Markdown paragraph on one line, as the root `AGENTS.md` requires. The `guard-markdown` hook checks this.

Build a filename from letters, digits, `_`, `-`, and `.`, and begin it with a letter, digit, or `_`. The checker rejects any other path segment. A relative link to another page must resolve, and the checker reads links written as `./path.md` or `name.md`.

## The index

`index.md` lists the topic pages under section headings. Each entry has the form `* [Title](link) - description`. Its only frontmatter key is `okf_version`.

When you add a topic page, add its entry. When you move or rename one, update the link. The checker validates the entry format and the links, but it doesn't notice a page missing from the index, so you must catch an omission yourself.

`README.md`, `AGENTS.md`, and `CLAUDE.md` are entry points and have no index entries.

## Checks

Run `mise run lint-draft <file>` on each page you write, then `mise run check-docs` after any change in `docs/`. The task runs `guard-markdown` on the tracked Markdown files and the OKF checker on `docs/` and on the Markdown outside it, so stage a new page before you run it.
