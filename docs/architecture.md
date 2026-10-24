---
type: Architecture Overview
title: Architecture overview
description: Go layout and command structure for afakt.
tags: [afakt, architecture]
status: draft
generated: { by: "claude-code:claude-sonnet-5-5", at: "2026-10-20T05:55:00Z" }
---

# afakt architecture

`afakt` is one Go module, `github.com/tbhb-dev/afakt`, with two packages. `cmd/afakt` is the `main` package and `internal/buildmeta` holds build metadata. The layering check in `.go-arch-lint.yml` lets `cmd/afakt` depend on `internal/buildmeta` but not the reverse.

## Command package

`main.go` is the only file that touches the process. It ignores `SIGPIPE`, so a write to a closed pipe fails with an error instead of ending the process, and it passes `os.Args`, standard output, and standard error to `Run`. `Run` returns the exit code, and `main` hands that code to `os.Exit`.

`parse.go` and `command.go` hold a small hand-written parser for the command tree: the root, `status`, and `help`. The parser returns an outcome with the text for each stream and an exit code, or a request to run `status`. Help and version output go to standard output with exit code 0, and usage errors go to standard error with exit code 2.

`run.go` runs `status`, which writes `status is not implemented yet.` and a newline through a helper that takes an `io.Writer`. A failed write prints `Error:` and the described error to standard error and exits with code 1. A closed standard output (`EBADF`) counts as success, and help and version output ignore write errors. `oserror.go` formats an operating system error as `Os { code, kind, message }`, with the kind names `StorageFull` and `BrokenPipe` for `ENOSPC` and `EPIPE`, and `Uncategorized` for other codes. Any other error prints as a custom error of kind `Other`.

## Build metadata

`internal/buildmeta` exposes `Version`, `Commit`, and `Date`. `mise run build` sets them at link time with `-ldflags -X`. Without linker flags, `Version` falls back to the module version in the runtime build info, and then to the `0.1.0` default that `afakt --version` prints under `go run`.

## Tests

Table-driven tests in `cmd/afakt` call `Run` with in-memory writers to cover help, version, usage errors, and write failures, and they test the error formatting directly. Tests in `internal/buildmeta` drive each branch of the version resolution. The coverage floors in `.testcoverage.yml` are 100% for each file, each package, and the total, with `main.go` excluded.

`mise run check` runs the tests with and without the race detector and enforces the coverage floors. It also checks the docs.
