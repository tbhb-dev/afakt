// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server_test

import (
	"slices"
	"testing"
)

func TestAllConfigurationMatchesExplicitTargets(t *testing.T) {
	t.Parallel()
	var targets []string
	for _, value := range []string{"api,webhook,dashboard,worker", "all"} {
		opts, err := parseServer([]string{"--target=" + value, "--temporal-mode=embedded"})
		if err != nil {
			t.Fatal(err)
		}
		if targets != nil && !slices.Equal(opts.Config.Targets, targets) {
			t.Errorf("targets = %q, want %q", opts.Config.Targets, targets)
		}
		targets = opts.Config.Targets
		if opts.Config.Temporal.Mode != "embedded" || opts.Config.Temporal.Namespace != "default" {
			t.Errorf("Temporal configuration = %+v", opts.Config.Temporal)
		}
	}
}
