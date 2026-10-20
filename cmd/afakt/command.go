// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/tbhb-dev/afakt/internal/buildmeta"
)

const (
	helpAbout = "Print this message or the help of the given subcommand(s)"

	rootHelp = `A proof of concept CLI

Usage: afakt <COMMAND>

Commands:
  status  Show the placeholder status
  help    ` + helpAbout + `

Options:
  -h, --help     Print help
  -V, --version  Print version
`

	statusHelp = `Show the placeholder status

Usage: afakt status

Options:
  -h, --help  Print help
`

	helpHelp = helpAbout + `

Usage: afakt help [COMMAND]...

Arguments:
  [COMMAND]...  Print help for the subcommand(s)
`
)

// command is one level of the command tree. The usage line is the prefix and the
// suffix, with the flag from an error message between them. The help command has
// no help flag of its own and omits the hint that points at one.
type command struct {
	prefix  string
	suffix  string
	help    string
	version bool
	noHint  bool
}

// outcome is the result of parsing a command line: text for each stream and an exit
// code, or a request to run the status command.
type outcome struct {
	stdout string
	stderr string
	code   int
	status bool
}

func rootCommand() command {
	return command{prefix: "afakt", suffix: "<COMMAND>", help: rootHelp, version: true}
}

func subcommand(name string) (command, bool) {
	switch name {
	case "status":
		return command{prefix: "afakt status", help: statusHelp}, true
	case "help":
		return command{prefix: "afakt help", suffix: "[COMMAND]...", help: helpHelp, noHint: true}, true
	}
	return command{}, false
}

// option handles an argument that starts with a dash. Only the first short flag of
// a bundle counts, because a help or version flag ends parsing and an unknown one
// is an error. It reports false for a positional argument.
func (c command) option(arg string) (outcome, bool) {
	if arg == "-" || !strings.HasPrefix(arg, "-") {
		return outcome{}, false
	}
	var (
		name, value string
		hasValue    bool
	)
	if strings.HasPrefix(arg, "--") {
		name, value, hasValue = strings.Cut(arg, "=")
	} else {
		first, _ := utf8.DecodeRuneInString(arg[1:])
		name = "-" + string(first)
	}
	out, known := c.action(name)
	switch {
	case !known:
		return c.unexpected(name), true
	case hasValue:
		return c.fail(
			fmt.Sprintf("unexpected value '%s' for '%s' found; no more were expected", value, name),
			name,
		), true
	}
	return out, true
}

// action maps a help or version flag to the text it prints.
func (c command) action(name string) (outcome, bool) {
	switch name {
	case "-h", "--help":
		return outcome{stdout: c.help}, true
	case "-V", "--version":
		if c.version {
			return outcome{stdout: fmt.Sprintf("afakt %s\n", buildmeta.Get().Version)}, true
		}
	}
	return outcome{}, false
}

func (c command) unexpected(arg string) outcome {
	return c.fail(fmt.Sprintf("unexpected argument '%s' found", arg), "")
}

func (c command) unrecognized(arg string) outcome {
	return c.fail(fmt.Sprintf("unrecognized subcommand '%s'", arg), "")
}

// fail reports a usage error. A non-empty flag joins the usage line, as clap shows
// the flag that the message cites.
func (c command) fail(message, flag string) outcome {
	usage := strings.Join(strings.Fields(c.prefix+" "+flag+" "+c.suffix), " ")
	text := fmt.Sprintf("error: %s\n\nUsage: %s\n", message, usage)
	if !c.noHint {
		text += "\nFor more information, try '--help'.\n"
	}
	return outcome{stderr: text, code: usageExit}
}
