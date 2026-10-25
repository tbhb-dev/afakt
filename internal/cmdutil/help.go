// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package cmdutil

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// commandOrderKey is the command annotation that lists the child commands in help order,
	// because Cobra sorts them by name.
	commandOrderKey = "afakt_command_order"

	// argumentsKey is the command annotation that holds the name of the positional arguments
	// and their description, separated by a tab.
	argumentsKey = "afakt_arguments"

	// helpHint ends a usage error of a command that has a help flag. Commands that parse no
	// flags leave it out.
	helpHint = "\nFor more information, try '--help'.\n"

	// columnGap is the space between the two columns of a help section.
	columnGap = 2
)

// AddCommands adds child commands to `parent`, and the help lists them in this order.
// Cobra itself would sort them by name.
func AddCommands(parent *cobra.Command, subcommands ...*cobra.Command) {
	names := make([]string, 0, len(subcommands))
	for _, sub := range subcommands {
		parent.AddCommand(sub)
		names = append(names, sub.Name())
	}
	setAnnotation(parent, commandOrderKey, strings.Join(names, " "))
}

// SetArguments makes the help of `cmd` describe its positional arguments.
func SetArguments(cmd *cobra.Command, name, usage string) {
	setAnnotation(cmd, argumentsKey, name+"\t"+usage)
}

func setAnnotation(cmd *cobra.Command, key, value string) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[key] = value
}

// Subcommand returns the child command of `parent` with the given name, or nil. Hidden
// commands don't count.
func Subcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, sub := range parent.Commands() {
		if sub.Name() == name && !sub.Hidden {
			return sub
		}
	}
	return nil
}

// HelpText renders the help of `cmd` in clap's layout. After the description and the usage
// line, it has a section for the child commands, the positional arguments, and the options.
func HelpText(cmd *cobra.Command) string {
	var b strings.Builder
	b.WriteString(cmd.Short + "\n\nUsage: " + UsageLine(cmd, "") + "\n")
	writeSection(&b, "Commands", commandRows(cmd))
	if name, usage, ok := strings.Cut(cmd.Annotations[argumentsKey], "\t"); ok {
		writeSection(&b, "Arguments", [][2]string{{name, usage}})
	}
	writeSection(&b, "Options", optionRows(cmd))
	return b.String()
}

// UsageLine is the command path and the rest of the command's use line, with `flag` between
// them. An empty `flag` gives the plain usage line.
func UsageLine(cmd *cobra.Command, flag string) string {
	_, rest, _ := strings.Cut(cmd.Use, " ")
	return strings.Join(strings.Fields(cmd.CommandPath()+" "+flag+" "+rest), " ")
}

// UsageText renders a flag error for the command that failed. A command that doesn't parse
// flags has no help flag. Its errors leave out the hint that points at one.
func UsageText(failed *cobra.Command, err *FlagError) string {
	cmd := failed
	if err.Cmd != nil {
		cmd = err.Cmd
	}
	text := "error: " + err.Message + "\n\nUsage: " + UsageLine(cmd, err.Flag) + "\n"
	if !cmd.DisableFlagParsing {
		text += helpHint
	}
	return text
}

func commandRows(cmd *cobra.Command) [][2]string {
	var rows [][2]string
	for name := range strings.FieldsSeq(cmd.Annotations[commandOrderKey]) {
		rows = append(rows, [2]string{name, Subcommand(cmd, name).Short})
	}
	return rows
}

func optionRows(cmd *cobra.Command) [][2]string {
	if cmd.DisableFlagParsing {
		return nil
	}
	// Each flag here is an action flag, which stores its short form in an annotation.
	var rows [][2]string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		rows = append(rows, [2]string{"-" + flag.Annotations[shortKey][0] + ", --" + flag.Name, flag.Usage})
	})
	return rows
}

func writeSection(b *strings.Builder, title string, rows [][2]string) {
	if len(rows) == 0 {
		return
	}
	// The second column starts two spaces after the widest first column. Sections without
	// rows are left out.
	width := 0
	for _, row := range rows {
		width = max(width, len(row[0]))
	}
	b.WriteString("\n" + title + ":\n")
	for _, row := range rows {
		b.WriteString("  " + row[0] + strings.Repeat(" ", width-len(row[0])+columnGap) + row[1] + "\n")
	}
}
