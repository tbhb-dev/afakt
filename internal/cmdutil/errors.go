// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package cmdutil

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// ErrHelp reports a help flag. The command prints its help and exits with code 0.
	ErrHelp = errors.New("help requested")
	// ErrVersion reports a version flag. The program prints its version and exits with code 0.
	ErrVersion = errors.New("version requested")

	errUnexpectedValue = errors.New("the flag takes no value")
	errRepeated        = errors.New("the flag was given more than once")
)

// FlagError is a command line that a command can't accept, and it exits with code 2. It
// prints its message and the usage line of the command it names. Without one, it prints
// the usage line of the failed command. When the message cites a flag, that flag joins the
// usage line, as clap shows it.
type FlagError struct {
	Cmd     *cobra.Command
	Flag    string
	Message string
}

func (e *FlagError) Error() string {
	return e.Message
}

// NoArgs rejects every positional argument. Cobra calls it before RunE.
func NoArgs(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return UnexpectedArgument(args[0])
	}
	return nil
}

// UnexpectedArgument reports an argument that the command doesn't take.
func UnexpectedArgument(arg string) *FlagError {
	return &FlagError{Message: fmt.Sprintf("unexpected argument '%s' found", arg)}
}

// InvalidValue reports a value that failed the check of its flag.
func InvalidValue(cmd *cobra.Command, name, value string, reason error) *FlagError {
	display := flagDisplay(cmd.Flags().Lookup(name))
	return &FlagError{Message: fmt.Sprintf("invalid value '%s' for '%s': %s", value, display, reason)}
}

// MissingFlag reports a required flag that the command line left out.
func MissingFlag(cmd *cobra.Command, name string) *FlagError {
	display := flagDisplay(cmd.Flags().Lookup(name))
	return &FlagError{Message: "the following required arguments were not provided:\n  " + display}
}

// TranslateFlagError turns an error from parsing the flags of `cmd` into what clap reports
// for the same command line. Any other error comes back unchanged. It fits Cobra's
// SetFlagErrorFunc.
func TranslateFlagError(cmd *cobra.Command, err error) error {
	var (
		notExist *pflag.NotExistError
		required *pflag.ValueRequiredError
		invalid  *pflag.InvalidValueError
		syntax   *pflag.InvalidSyntaxError
	)
	switch {
	case errors.Is(err, pflag.ErrHelp):
		// Pflag reports any short flag group that starts with `h` this way when `h` isn't a
		// shorthand, as with the help flag here.
		return ErrHelp
	case errors.As(err, &notExist) && notExist.GetSpecifiedShortnames() != "":
		// Pflag names the first byte of the group, which cuts a letter of two or more bytes
		// in half. Clap names the first character.
		first, _ := utf8.DecodeRuneInString(notExist.GetSpecifiedShortnames())
		if action := shortAction(cmd, string(first)); action != nil {
			return action
		}
		return UnexpectedArgument("-" + string(first))
	case errors.As(err, &notExist):
		return UnexpectedArgument("--" + notExist.GetSpecifiedName())
	case errors.As(err, &required):
		display := flagDisplay(required.GetFlag())
		return &FlagError{Message: fmt.Sprintf("a value is required for '%s' but none was supplied", display)}
	case errors.As(err, &invalid) && errors.Is(err, errRepeated):
		display := flagDisplay(invalid.GetFlag())
		return &FlagError{Message: fmt.Sprintf("the argument '%s' cannot be used multiple times", display)}
	case errors.As(err, &invalid) && errors.Is(err, errUnexpectedValue):
		name := "--" + invalid.GetFlag().Name
		return &FlagError{
			Flag: name,
			Message: fmt.Sprintf(
				"unexpected value '%s' for '%s' found; no more were expected",
				invalid.GetValue(),
				name,
			),
		}
	case errors.As(err, &syntax):
		// With no flag name before the equals sign, clap cites the dashes alone.
		name, _, _ := strings.Cut(syntax.GetSpecifiedFlag(), "=")
		return UnexpectedArgument(name)
	}
	return err
}
