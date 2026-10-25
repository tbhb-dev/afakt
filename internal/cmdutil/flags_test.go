// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package cmdutil_test

import (
	"slices"
	"testing"
)

// TestHelpFlagReadsAsBool checks that Cobra can read the help flag as a Boolean. Cobra fails
// a command whose help flag isn't one.
func TestHelpFlagReadsAsBool(t *testing.T) {
	t.Parallel()
	tool, _, _ := newTool()
	if typ := tool.Flags().Lookup("help").Value.Type(); typ != "bool" {
		t.Errorf("Type() = %q, want bool", typ)
	}
	help, err := tool.Flags().GetBool("help")
	if err != nil || help {
		t.Errorf("GetBool(help) = %t, %v; want false, nil", help, err)
	}
}

// TestFirstPositionalEndsFlags checks that parsing stops at the first positional argument,
// so a help flag after it counts as an argument and not as a request for help.
func TestFirstPositionalEndsFlags(t *testing.T) {
	t.Parallel()
	_, run, _ := newTool()
	if err := parse(run, "extra", "--help"); err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	if args := run.Flags().Args(); !slices.Equal(args, []string{"extra", "--help"}) {
		t.Errorf("Args() = %q, want [extra --help]", args)
	}
}
