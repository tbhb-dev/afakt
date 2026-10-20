// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

func parseRoot(args []string) outcome {
	root := rootCommand()
	switch {
	case len(args) == 0, len(args) == 1 && args[0] == "--":
		return outcome{stderr: rootHelp, code: usageExit}
	case args[0] == "--":
		// Everything after the terminator is positional, so a command name is unexpected.
		if _, ok := subcommand(args[1]); ok {
			return root.unexpected(args[1])
		}
		return root.unrecognized(args[1])
	}
	if out, ok := root.option(args[0]); ok {
		return out
	}
	switch args[0] {
	case "status":
		return parseStatus(args[1:])
	case "help":
		return parseHelp(args[1:])
	}
	return root.unrecognized(args[0])
}

func parseStatus(args []string) outcome {
	cmd, _ := subcommand("status")
	switch {
	case len(args) == 0, len(args) == 1 && args[0] == "--":
		return outcome{status: true}
	case args[0] == "--":
		return cmd.unexpected(args[1])
	}
	if out, ok := cmd.option(args[0]); ok {
		return out
	}
	return cmd.unexpected(args[0])
}

// parseHelp resolves each command in the list and prints the help of the last one.
func parseHelp(args []string) outcome {
	current := rootCommand()
	for i, arg := range args {
		next, ok := subcommand(arg)
		if !ok || i > 0 {
			return current.unrecognized(arg)
		}
		current = next
	}
	return outcome{stdout: current.help}
}
