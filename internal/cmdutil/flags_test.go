// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package cmdutil_test

import (
	"slices"
	"testing"

	"github.com/tbhb-dev/afakt/internal/cmdutil"
)

func TestStringFlagForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "default", want: "fast"},
		{name: "equals form", args: []string{"--mode=slow"}, want: "slow"},
		{name: "separate form", args: []string{"--mode", "slow"}, want: "slow"},
		{name: "empty value", args: []string{"--mode="}, want: ""},
	}
	for _, tc := range cases {
		_, run, _ := newTool()
		if err := cmdutil.TranslateFlagError(run, run.Flags().Parse(tc.args)); err != nil {
			t.Fatalf("parse(%q) error = %v", tc.args, err)
		}
		value := run.Flags().Lookup("mode").Value
		if value.String() != tc.want {
			t.Errorf("%s: mode = %q, want %q", tc.name, value.String(), tc.want)
		}
		if value.Type() != "string" {
			t.Errorf("Type() = %q, want string", value.Type())
		}
	}
}

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
	if err := cmdutil.TranslateFlagError(run, run.Flags().Parse([]string{"--name=a", "extra", "--help"})); err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	if args := run.Flags().Args(); !slices.Equal(args, []string{"extra", "--help"}) {
		t.Errorf("Args() = %q, want [extra --help]", args)
	}
}
