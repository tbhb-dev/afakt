// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server_test

import (
	"strings"
	"testing"
)

func TestTemporalUIFlags(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"default", nil, "/temporal"},
		{"override", []string{"--temporal-ui-path=/tools/temporal"}, "/tools/temporal"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := append([]string{"--target=worker", "--temporal-mode=embedded"}, tc.args...)
			opts, err := parseServer(args)
			if err != nil {
				t.Fatal(err)
			}
			if opts.Config.TemporalUIPath != tc.want || opts.Config.Listen != "127.0.0.1:8080" {
				t.Errorf("configuration = %+v, want loopback and path %s", opts.Config, tc.want)
			}
		})
	}
	for _, args := range [][]string{
		{"--target=api", "--temporal-ui-path=/tools"},
		{"--target=worker", "--temporal-mode=external", "--temporal-address=127.0.0.1:7233", "--temporal-ui-path=/tools"},
		{"--target=worker", "--temporal-mode=embedded", "--temporal-ui-path=/api/v1"},
		{"--target=worker,dashboard", "--temporal-mode=embedded", "--temporal-ui-path=/dashboard"},
	} {
		if opts, err := parseServer(
			args,
		); err == nil || opts != nil ||
			!strings.Contains(err.Error(), "--temporal-ui-path") {
			t.Errorf("parseServer(%q) = %v, %v, want a UI flag error before startup", args, opts, err)
		}
	}
}
