// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package buildmeta

import (
	"runtime/debug"
	"testing"
)

// TestResolveVersion drives each branch of the runtime-fallback
// chain so mutation testing doesn't find a surviving conditional flip
// in the init-time logic. The cases pair a linker flag input with a
// hand-rolled readBuildInfo result and assert the resolved version.
func TestResolveVersion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		ldflagsVersion string
		readBuildInfo  func() (*debug.BuildInfo, bool)
		want           string
	}{
		{
			name:           "ldflags version wins over runtime fallback",
			ldflagsVersion: "v1.2.3",
			readBuildInfo: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, true
			},
			want: "v1.2.3",
		},
		{
			name:           "build info unavailable keeps default version",
			ldflagsVersion: "0.1.0",
			readBuildInfo: func() (*debug.BuildInfo, bool) {
				return nil, false
			},
			want: "0.1.0",
		},
		{
			name:           "build info devel placeholder keeps default version",
			ldflagsVersion: "0.1.0",
			readBuildInfo: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
			},
			want: "0.1.0",
		},
		{
			name:           "build info concrete version replaces default version",
			ldflagsVersion: "0.1.0",
			readBuildInfo: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{Main: debug.Module{Version: "v0.4.2"}}, true
			},
			want: "v0.4.2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := resolveVersion(tc.ldflagsVersion, tc.readBuildInfo)
			if got != tc.want {
				t.Errorf("resolveVersion() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestDefaultVersion pins the no-linker-flags version.
func TestDefaultVersion(t *testing.T) {
	t.Parallel()
	if devFallback != "0.1.0" {
		t.Errorf("devFallback = %q, want %q", devFallback, "0.1.0")
	}
}
