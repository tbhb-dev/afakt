---
type: Guide
title: Docs guide
description: What the docs directory contains, what it leaves out, and the frontmatter each page needs.
tags: [afakt, docs]
status: stable
generated: { by: "claude-code:claude-sonnet-5-5", at: "2026-10-25T00:40:00Z" }
---

# Docs guide

This directory contains the durable documentation for `afakt`: pages that describe the repository as it stands. [index.md](index.md) lists them, so start there to find a page.

## What belongs here

A page belongs in `docs/` when it remains accurate for the code on `main` and a reader wants it later. Draft design work, plans, and scratch notes don't belong. Keep them outside the repository until the code does what they describe. Then write a page about what exists.

`AGENTS.md` and `CLAUDE.md` in this directory are entry points for agents, not topic pages. The index leaves them out.

## Frontmatter

A Markdown page here begins with a YAML frontmatter block in Open Knowledge Format (OKF) v0.2. The reserved `index.md` and `log.md` are the exceptions.

- `type` must be present and non-empty. Its value is free text, such as `Guide` or `CLI Reference`.
- `title`, `description`, and `tags` belong on every page. `tags` is a list of strings.
- `status` is one of `draft`, `stable`, or `deprecated`.
- `generated` takes `{ by, at }` for content a machine wrote. `by` names the producer and model, and `at` is a timestamp in the form `2026-10-25T00:40:00Z`.
- Only a person adds `verified`. Tools and agents never do.

The checker, `scripts/okf-check.py`, enforces less than this convention. It fails a page with an empty or missing `type`, a `status` outside the three values, or a `generated` block without a `by`. A page may omit `title`, `description`, `tags`, `status`, `generated`, or `generated.at` and still pass, and the checker ignores `verified`. Writers supply the rest.

`index.md` is the reserved page. Its only frontmatter key is `okf_version`.

## Where to read next

[AGENTS.md](AGENTS.md) gives agents the rules for changing these pages, and the root `AGENTS.md` has the repository-wide rules. Run `mise run check-docs` after any change here.
