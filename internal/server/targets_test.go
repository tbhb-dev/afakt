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
	got, err := server.ParseTargets("api")
	if err != nil {
		t.Fatalf("ParseTargets(api) error = %v", err)
	}
	if want := []string{"api"}; !slices.Equal(got, want) {
		t.Errorf("ParseTargets(api) = %q, want %q", got, want)
	}
}

func TestParseTargetsRejects(t *testing.T) {
	t.Parallel()
	const choices = "choose from api, webhook, dashboard, worker, all"
	unavailable := func(name string) string {
		return "target is not available yet: '" + name + "'; only 'api' can run in this build"
	}
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
		{value: "webhook", want: server.ErrUnavailableTarget, text: unavailable("webhook")},
		{value: "dashboard", want: server.ErrUnavailableTarget, text: unavailable("dashboard")},
		{value: "worker", want: server.ErrUnavailableTarget, text: unavailable("worker")},
		{value: "all", want: server.ErrUnavailableTarget, text: unavailable("all")},
		{value: "api,all", want: server.ErrUnavailableTarget, text: unavailable("all")},
		{value: "api,webhook", want: server.ErrUnavailableTarget, text: unavailable("webhook")},
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
