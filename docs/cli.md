---
type: CLI Reference
title: Command reference
description: Usage, options, commands, output, and exit codes for the afakt command-line tool.
tags: [afakt, cli]
status: stable
generated: { by: "claude-code:claude-sonnet-5-5", at: "2026-10-18T19:27:15Z" }
---

# Command reference

`afakt` is a proof of concept command-line tool. It has one command, `status`, which prints a placeholder response. Run the binary with `cargo run -p afakt -- <arguments>`, or run the built `afakt` binary directly.

## Synopsis

```text
afakt <COMMAND>
afakt -h | --help
afakt -V | --version
```

You must give a command. The only exceptions are `-h`/`--help` and `-V`/`--version`, which print and exit. The tool has no configuration files and doesn't read any environment variables of its own.

## Global options

| Option | Behavior |
| --- | --- |
| `-h`, `--help` | Print usage, commands, and options to standard output, then exit successfully |
| `-V`, `--version` | Print `afakt 0.1.0` to standard output, then exit successfully |

## Commands

| Command | Behavior |
| --- | --- |
| `status` | Print the placeholder status |
| `help` | Print the top-level help, or the help of a named command, such as `afakt help status` |

### The status command

```text
afakt status
```

`status` takes no arguments and no options besides `-h` and `--help`. It writes `status is not implemented yet.` followed by a newline to standard output and exits successfully. When the write succeeds, it writes nothing to standard error.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | The command succeeded, or you asked for help or the version |
| `1` | `status` couldn't write to standard output, such as when standard output is a closed pipe |
| `2` | The command line is invalid, such as a missing command, an unknown command, or an unexpected argument |

Running `afakt` with no command prints the top-level help to standard error and exits with code `2`. For an unknown command or option, clap writes an `error:` message with the matching usage line to standard error and exits with code `2`.

If `status` fails to write to standard output, the process prints `Error:` and the I/O error to standard error, then exits with code `1`. Help and version output differ. Clap discards write errors there, and the process exits with code `0`.

On Linux, every write to `/dev/full` fails, so this reproduces the failure every time:

```bash
afakt status > /dev/full
```

The command prints `Error: Os { code: 28, kind: StorageFull, message: "No space left on device" }` to standard error and exits with code `1`.

A pipe whose reader has already exited fails the same way, with `Error: Os { code: 32, kind: BrokenPipe, message: "Broken pipe" }`, because Rust ignores the broken pipe signal. Whether it fails depends on timing. If the reader closes its end before `afakt` writes, the exit code is `1`. If `afakt` writes first, the output goes into the pipe buffer and the exit code is `0`. A command such as `afakt status | true` can produce either result.

A shell reports the status of the last command in a pipeline, not the status of `afakt`. In bash, run `set -o pipefail` before the pipeline to make it fail when any command fails, or read `${PIPESTATUS[0]}` right after it for the status of `afakt` alone.
