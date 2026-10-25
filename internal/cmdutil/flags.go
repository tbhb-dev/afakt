// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package cmdutil

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// shortKey is the flag annotation that holds the letter of the flag's short form.
	shortKey = "afakt_short"

	// noValue is what the flag parser passes to an action flag written without a value.
	// Arguments can't contain a zero byte, so any other value came after an equals sign.
	noValue = "\x00"
)

// actionFlag is a flag that takes no value, such as the help flag. Its Set returns the
// action's error, which ends parsing at the first action flag.
type actionFlag struct {
	action error
}

func (a actionFlag) String() string {
	return "false"
}

// Type reports a Boolean flag, because Cobra fails a command whose help flag isn't one.
func (a actionFlag) Type() string {
	return "bool"
}

func (a actionFlag) Set(value string) error {
	if value != noValue {
		return errUnexpectedValue
	}
	return a.action
}

// AddHelpFlag adds the short and long help flags. It also sets two parsing rules that
// every command shares. Flags keep the order they're defined in, which is the order the
// help lists them in. Parsing stops at the first positional argument, and the rest count
// as arguments too. An unexpected argument then fails before a help flag after it can
// print the help, as clap does.
func AddHelpFlag(cmd *cobra.Command) {
	addAction(cmd, "help", "h", "Print help", ErrHelp)
	cmd.Flags().SortFlags = false
	cmd.Flags().SetInterspersed(false)
}

func AddVersionFlag(cmd *cobra.Command) {
	addAction(cmd, "version", "V", "Print version", ErrVersion)
}

// addAction adds an action flag whose short form isn't a pflag shorthand. Clap acts on the
// first letter of a short flag group and ignores the rest, so `-h=1` prints the help, but
// pflag would hand the short and the long form the same value. TranslateFlagError turns
// the report of an unknown short flag into the action.
func addAction(cmd *cobra.Command, name, short, usage string, action error) {
	flag := cmd.Flags().VarPF(actionFlag{action: action}, name, "", usage)
	flag.NoOptDefVal = noValue
	flag.Annotations = map[string][]string{shortKey: {short}}
}

// shortAction returns the action of the flag of `cmd` whose short form is `letter`, or nil.
func shortAction(cmd *cobra.Command, letter string) error {
	var action error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if value, ok := flag.Value.(actionFlag); ok && flag.Annotations[shortKey][0] == letter {
			action = value.action
		}
	})
	return action
}
