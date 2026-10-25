// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package root_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"syscall"
	"testing"

	"github.com/tbhb-dev/afakt/internal/cmd/root"
	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

const (
	testVersion = "1.2.3-test"

	helpAbout = "Print this message or the help of the given subcommand(s)"

	wantRootHelp = `A proof of concept CLI

Usage: afakt <COMMAND>

Commands:
  status  Show the placeholder status
  help    ` + helpAbout + `

Options:
  -h, --help     Print help
  -V, --version  Print version
`

	wantStatusHelp = `Show the placeholder status

Usage: afakt status

Options:
  -h, --help  Print help
`

	wantHelpHelp = helpAbout + `

Usage: afakt help [COMMAND]...

Arguments:
  [COMMAND]...  Print help for the subcommand(s)
`

	wantStatus = "status is not implemented yet.\n"
)

var errWrite = errors.New("write failed")

// failWriter fails every write with its error.
type failWriter struct {
	err error
}

func (w failWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func newFactory(out, errOut io.Writer) *cmdutil.Factory {
	return &cmdutil.Factory{In: strings.NewReader(""), Out: out, ErrOut: errOut, Version: testVersion}
}

// execute runs the command line through the root command, as main does.
func execute(t *testing.T, args []string, out, errOut io.Writer) int {
	t.Helper()
	return root.Execute(t.Context(), newFactory(out, errOut), args)
}

// clapError is the error text clap prints: the message, the usage line and the hint.
func clapError(message, usage string) string {
	return clapErrorWithoutHint(message, usage) + "\nFor more information, try '--help'.\n"
}

// clapErrorWithoutHint is the error text of the help command, which has no hint.
func clapErrorWithoutHint(message, usage string) string {
	return "error: " + message + "\n\nUsage: " + usage + "\n"
}

func TestRun(t *testing.T) {
	t.Parallel()
	version := "afakt " + testVersion + "\n"
	cases := []struct {
		name   string
		args   []string
		code   int
		stdout string
		stderr string
	}{
		{name: "status", args: []string{"status"}, stdout: wantStatus},
		{name: "status terminator", args: []string{"status", "--"}, stdout: wantStatus},
		{name: "status help", args: []string{"status", "--help"}, stdout: wantStatusHelp},
		{name: "status short help", args: []string{"status", "-h"}, stdout: wantStatusHelp},
		{name: "status short help bundle", args: []string{"status", "-hx"}, stdout: wantStatusHelp},
		{name: "help", args: []string{"--help"}, stdout: wantRootHelp},
		{name: "short help", args: []string{"-h"}, stdout: wantRootHelp},
		{name: "short help then version", args: []string{"-hV"}, stdout: wantRootHelp},
		{name: "help before command", args: []string{"-h", "status"}, stdout: wantRootHelp},
		{name: "help with bundled value", args: []string{"-h=1"}, stdout: wantRootHelp},
		{name: "status help with bundled value", args: []string{"status", "-h=1"}, stdout: wantStatusHelp},
		{name: "version with bundled value", args: []string{"-V=1"}, stdout: version},
		{name: "version", args: []string{"--version"}, stdout: version},
		{name: "short version", args: []string{"-V"}, stdout: version},
		{name: "short version then help", args: []string{"-Vh"}, stdout: version},
		{name: "repeated help", args: []string{"--help", "--help"}, stdout: wantRootHelp},
		{name: "repeated short help", args: []string{"-h", "-h"}, stdout: wantRootHelp},
		{name: "repeated version", args: []string{"--version", "--version"}, stdout: version},
		{name: "repeated short version", args: []string{"-V", "-V"}, stdout: version},
		{name: "status repeated help", args: []string{"status", "--help", "--help"}, stdout: wantStatusHelp},
		{name: "version with argument", args: []string{"--version", "extra"}, stdout: version},
		{name: "short version with command", args: []string{"-V", "status"}, stdout: version},
		{name: "help command", args: []string{"help"}, stdout: wantRootHelp},
		{name: "help status", args: []string{"help", "status"}, stdout: wantStatusHelp},
		{name: "help of help", args: []string{"help", "help"}, stdout: wantHelpHelp},
		{
			name:   "help unknown command",
			args:   []string{"help", "nosuch"},
			code:   2,
			stderr: clapError("unrecognized subcommand 'nosuch'", "afakt <COMMAND>"),
		},
		{
			name:   "help flag",
			args:   []string{"help", "--help"},
			code:   2,
			stderr: clapError("unrecognized subcommand '--help'", "afakt <COMMAND>"),
		},
		{
			name:   "help status extra",
			args:   []string{"help", "status", "extra"},
			code:   2,
			stderr: clapError("unrecognized subcommand 'extra'", "afakt status"),
		},
		{
			name:   "help past help",
			args:   []string{"help", "help", "help"},
			code:   2,
			stderr: clapErrorWithoutHint("unrecognized subcommand 'help'", "afakt help [COMMAND]..."),
		},
		{name: "no arguments", code: 2, stderr: wantRootHelp},
		{name: "terminator only", args: []string{"--"}, code: 2, stderr: wantRootHelp},
		{
			name:   "terminator then command",
			args:   []string{"--", "status"},
			code:   2,
			stderr: clapError("unexpected argument 'status' found", "afakt <COMMAND>"),
		},
		{
			name:   "terminator then unknown",
			args:   []string{"--", "bogus"},
			code:   2,
			stderr: clapError("unrecognized subcommand 'bogus'", "afakt <COMMAND>"),
		},
		{
			name:   "unknown command",
			args:   []string{"bogus"},
			code:   2,
			stderr: clapError("unrecognized subcommand 'bogus'", "afakt <COMMAND>"),
		},
		{
			name:   "lone dash",
			args:   []string{"-"},
			code:   2,
			stderr: clapError("unrecognized subcommand '-'", "afakt <COMMAND>"),
		},
		{
			name:   "unknown long option",
			args:   []string{"--bogus"},
			code:   2,
			stderr: clapError("unexpected argument '--bogus' found", "afakt <COMMAND>"),
		},
		{
			name:   "unknown long option with value",
			args:   []string{"--bogus=x"},
			code:   2,
			stderr: clapError("unexpected argument '--bogus' found", "afakt <COMMAND>"),
		},
		{
			name:   "unknown short option",
			args:   []string{"-x"},
			code:   2,
			stderr: clapError("unexpected argument '-x' found", "afakt <COMMAND>"),
		},
		{name: "short help bundle", args: []string{"-hx"}, stdout: wantRootHelp},
		{
			name:   "three dashes",
			args:   []string{"---x"},
			code:   2,
			stderr: clapError("unexpected argument '---x' found", "afakt <COMMAND>"),
		},
		{
			name:   "value without a flag name",
			args:   []string{"--=x"},
			code:   2,
			stderr: clapError("unexpected argument '--' found", "afakt <COMMAND>"),
		},
		{
			// Cobra's hidden completion command counts as an unknown command, whatever follows it.
			name:   "completion request",
			args:   []string{"__complete", "s"},
			code:   2,
			stderr: clapError("unrecognized subcommand '__complete'", "afakt <COMMAND>"),
		},
		{
			name:   "completion request without arguments",
			args:   []string{"__complete"},
			code:   2,
			stderr: clapError("unrecognized subcommand '__complete'", "afakt <COMMAND>"),
		},
		{
			name:   "completion request without descriptions",
			args:   []string{"__completeNoDesc", "--help"},
			code:   2,
			stderr: clapError("unrecognized subcommand '__completeNoDesc'", "afakt <COMMAND>"),
		},
		{
			name:   "unknown short option before help",
			args:   []string{"-xh"},
			code:   2,
			stderr: clapError("unexpected argument '-x' found", "afakt <COMMAND>"),
		},
		{
			name: "help flag with value",
			args: []string{"--help=1"},
			code: 2,
			stderr: clapError(
				"unexpected value '1' for '--help' found; no more were expected",
				"afakt --help <COMMAND>",
			),
		},
		{
			name: "version flag with empty value",
			args: []string{"--version="},
			code: 2,
			stderr: clapError(
				"unexpected value '' for '--version' found; no more were expected",
				"afakt --version <COMMAND>",
			),
		},
		{
			name:   "status extra argument",
			args:   []string{"status", "extra"},
			code:   2,
			stderr: clapError("unexpected argument 'extra' found", "afakt status"),
		},
		{
			name:   "status after terminator",
			args:   []string{"status", "--", "extra"},
			code:   2,
			stderr: clapError("unexpected argument 'extra' found", "afakt status"),
		},
		{
			name:   "status help flag after terminator",
			args:   []string{"status", "--", "-h"},
			code:   2,
			stderr: clapError("unexpected argument '-h' found", "afakt status"),
		},
		{
			name:   "status rejects version",
			args:   []string{"status", "-V"},
			code:   2,
			stderr: clapError("unexpected argument '-V' found", "afakt status"),
		},
		{
			name:   "status rejects long version",
			args:   []string{"status", "--version"},
			code:   2,
			stderr: clapError("unexpected argument '--version' found", "afakt status"),
		},
		{
			name:   "status unknown option",
			args:   []string{"status", "--bogus"},
			code:   2,
			stderr: clapError("unexpected argument '--bogus' found", "afakt status"),
		},
		{
			name:   "status unknown short option",
			args:   []string{"status", "-xh"},
			code:   2,
			stderr: clapError("unexpected argument '-x' found", "afakt status"),
		},
		{
			name:   "status extra before help",
			args:   []string{"status", "extra", "--help"},
			code:   2,
			stderr: clapError("unexpected argument 'extra' found", "afakt status"),
		},
		{
			name:   "status lone dash",
			args:   []string{"status", "-"},
			code:   2,
			stderr: clapError("unexpected argument '-' found", "afakt status"),
		},
		{
			name:   "status help flag with value",
			args:   []string{"status", "--help=1"},
			code:   2,
			stderr: clapError("unexpected value '1' for '--help' found; no more were expected", "afakt status --help"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			if got := execute(t, tc.args, &stdout, &stderr); got != tc.code {
				t.Errorf("exit code = %d, want %d", got, tc.code)
			}
			if stdout.String() != tc.stdout {
				t.Errorf("stdout = %q, want %q", stdout.String(), tc.stdout)
			}
			if stderr.String() != tc.stderr {
				t.Errorf("stderr = %q, want %q", stderr.String(), tc.stderr)
			}
		})
	}
}

func TestExecuteHelpListsCommandsAndFlags(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if got := execute(t, []string{"--help"}, &stdout, &stderr); got != 0 {
		t.Errorf("exit code = %d, want 0", got)
	}
	for _, want := range []string{"status", "--help", "--version"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("help lacks %q: %q", want, stdout.String())
		}
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestExecuteNilArguments(t *testing.T) {
	t.Parallel()
	// Nil arguments mean no command. Cobra would read the arguments of the test process instead.
	var stdout, stderr bytes.Buffer
	if got := execute(t, nil, &stdout, &stderr); got != 2 {
		t.Errorf("exit code = %d, want 2", got)
	}
	if stdout.Len() != 0 || stderr.String() != wantRootHelp {
		t.Errorf("stdout = %q, stderr = %q; want only the root help on stderr", stdout.String(), stderr.String())
	}
}

func TestExecuteStatusWriteFailure(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		err    error
		code   int
		stderr string
	}{
		{
			name:   "full device",
			err:    syscall.ENOSPC,
			code:   1,
			stderr: "Error: Os { code: 28, kind: StorageFull, message: \"No space left on device\" }\n",
		},
		{
			name:   "closed pipe",
			err:    syscall.EPIPE,
			code:   1,
			stderr: "Error: Os { code: 32, kind: BrokenPipe, message: \"Broken pipe\" }\n",
		},
		{name: "closed standard output", err: syscall.EBADF},
		{
			name:   "not an errno",
			err:    errWrite,
			code:   1,
			stderr: "Error: Custom { kind: Other, error: \"write failed\" }\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var stderr bytes.Buffer
			if got := execute(t, []string{"status"}, failWriter{err: tc.err}, &stderr); got != tc.code {
				t.Errorf("exit code = %d, want %d", got, tc.code)
			}
			if stderr.String() != tc.stderr {
				t.Errorf("stderr = %q, want %q", stderr.String(), tc.stderr)
			}
		})
	}
}

func TestExecuteHelpAndVersionIgnoreWriteErrors(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"--help"}, {"--version"}, {"help"}, {"bogus"}, {}, {"status"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()
			writer := failWriter{err: syscall.ENOSPC}
			want := 0
			switch {
			case len(args) == 0, args[0] == "bogus":
				want = 2
			case args[0] == "status":
				want = 1
			}
			if got := execute(t, args, writer, writer); got != want {
				t.Errorf("exit code = %d, want %d", got, want)
			}
		})
	}
}

func TestNewCmdRootUsesFactoryStreams(t *testing.T) {
	t.Parallel()
	// Cobra's own streams must be the factory's, so nothing below main writes to the process
	// streams.
	in := strings.NewReader("")
	var stdout, stderr bytes.Buffer
	f := &cmdutil.Factory{In: in, Out: &stdout, ErrOut: &stderr, Version: testVersion}
	cmd := root.NewCmdRoot(f)
	if cmd.InOrStdin() != in || cmd.OutOrStdout() != &stdout || cmd.ErrOrStderr() != &stderr {
		t.Error("the root command doesn't use the factory streams")
	}
}
