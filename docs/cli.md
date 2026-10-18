---
type: CLI Reference
title: Command reference
description: Usage, options, commands, output, and exit codes for the afakt command-line tool.
tags: [afakt, cli]
status: stable
generated: { by: "claude-code/claude-sonnet-5-5", at: "2026-10-18T14:20:44Z" }
---

# Command reference

`afakt` is a proof of concept command-line tool. It has one command, `status`, which prints a placeholder response. Run the binary with `cargo run -p afakt -- <arguments>`, or run the built `afakt` binary directly.

## Synopsis

```text
afakt <COMMAND>
afakt [OPTIONS]
```

You must give a command. The tool has no configuration files and doesn't read any environment variables of its own.

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

`status` takes no arguments and no options besides `-h` and `--help`. It writes `status is not implemented yet.` followed by a newline to standard output and exits successfully. It writes nothing to standard error.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | The command succeeded, or you asked for help or the version |
| `2` | The command line is invalid, such as a missing command, an unknown command, or an unexpected argument |

Running `afakt` with no command prints the top-level help to standard error and exits with code `2`. For an unknown command or option, clap writes an `error:` message with the matching usage line to standard error and exits with code `2`.

If writing to standard output fails, the process returns the write error as a failure.
