// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/tbhb-dev/afakt/internal/server"
)

func TestParseTargetsAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		value string
		want  []string
	}{
		{value: "all", want: []string{"api", "webhook", "dashboard", "worker"}},
		{value: "api,webhook,dashboard,worker", want: []string{"api", "webhook", "dashboard", "worker"}},
		{value: "worker", want: []string{"worker"}},
		{value: "worker,api", want: []string{"worker", "api"}},
		{value: "api,webhook,worker", want: []string{"api", "webhook", "worker"}},
		{value: "api", want: []string{"api"}},
		{value: "webhook", want: []string{"webhook"}},
		{value: "api,webhook", want: []string{"api", "webhook"}},
		{value: "webhook,api", want: []string{"webhook", "api"}},
		{value: "dashboard", want: []string{"dashboard"}},
		{value: "api,dashboard", want: []string{"api", "dashboard"}},
		{value: "dashboard,webhook,api", want: []string{"dashboard", "webhook", "api"}},
	}
	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			t.Parallel()
			got, err := server.ParseTargets(tc.value)
			if err != nil {
				t.Fatalf("ParseTargets(%q) error = %v", tc.value, err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("ParseTargets(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestParseTargetsRejects(t *testing.T) {
	t.Parallel()
	const choices = "choose from api, webhook, dashboard, worker, all"
	cases := []struct {
		value string
		want  error
		text  string
	}{
		{value: "", want: server.ErrNoTarget, text: "no target given; " + choices},
		{value: ",", want: server.ErrEmptyTarget, text: "empty entry in the target list"},
		{value: "api,", want: server.ErrEmptyTarget, text: "empty entry in the target list"},
		{value: ",api", want: server.ErrEmptyTarget, text: "empty entry in the target list"},
		{value: "api,,api", want: server.ErrEmptyTarget, text: "empty entry in the target list"},
		{value: "bogus", want: server.ErrUnknownTarget, text: "unknown target 'bogus'; " + choices},
		{value: "API", want: server.ErrUnknownTarget, text: "unknown target 'API'; " + choices},
		{value: " api", want: server.ErrUnknownTarget, text: "unknown target ' api'; " + choices},
		{value: "webhook,bogus", want: server.ErrUnknownTarget, text: "unknown target 'bogus'; " + choices},
		{value: "api,api", want: server.ErrDuplicateTarget, text: "duplicate target 'api'"},
		{value: "worker,api,worker", want: server.ErrDuplicateTarget, text: "duplicate target 'worker'"},
		{value: "webhook,webhook", want: server.ErrDuplicateTarget, text: "duplicate target 'webhook'"},
		{value: "dashboard,dashboard", want: server.ErrDuplicateTarget, text: "duplicate target 'dashboard'"},
		{
			value: "worker,all",
			want:  server.ErrDuplicateTarget,
			text:  "duplicate target: 'all' includes every target; use it on its own",
		},
		{
			value: "all,api",
			want:  server.ErrDuplicateTarget,
			text:  "duplicate target: 'all' includes every target; use it on its own",
		},
		{value: "all,all", want: server.ErrDuplicateTarget, text: "duplicate target 'all'"},
	}
	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			t.Parallel()
			got, err := server.ParseTargets(tc.value)
			if !errors.Is(err, tc.want) || err.Error() != tc.text {
				t.Errorf("ParseTargets(%q) error = %v, want %q", tc.value, err, tc.text)
			}
			if got != nil {
				t.Errorf("ParseTargets(%q) = %q, want nil", tc.value, got)
			}
		})
	}
}
