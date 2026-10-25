// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// TargetAPI selects the HTTP API.
const TargetAPI = "api"

var (
	// ErrNoTarget reports an empty target list.
	ErrNoTarget = errors.New("no target given")
	// ErrEmptyTarget reports an empty entry in the target list.
	ErrEmptyTarget = errors.New("empty entry in the target list")
	// ErrUnknownTarget reports a name that isn't a target.
	ErrUnknownTarget = errors.New("unknown target")
	// ErrDuplicateTarget reports a target named more than once.
	ErrDuplicateTarget = errors.New("duplicate target")
	// ErrUnavailableTarget reports a target that this build can't run yet.
	ErrUnavailableTarget = errors.New("target is not available yet")
)

// targetNames returns every target name in the order the help text lists them. Every name
// except TargetAPI is reserved for later changes, and all stands for every target.
func targetNames() []string {
	return []string{TargetAPI, "webhook", "dashboard", "worker", "all"}
}

// ParseTargets checks a comma-separated target list and returns its names. Every name must
// be a known target that this build can run, and no name may repeat.
func ParseTargets(value string) ([]string, error) {
	choices := strings.Join(targetNames(), ", ")
	if value == "" {
		return nil, fmt.Errorf("%w; choose from %s", ErrNoTarget, choices)
	}
	names := strings.Split(value, ",")
	for i, name := range names {
		switch {
		case name == "":
			return nil, ErrEmptyTarget
		case !slices.Contains(targetNames(), name):
			return nil, fmt.Errorf("%w '%s'; choose from %s", ErrUnknownTarget, name, choices)
		case slices.Contains(names[:i], name):
			return nil, fmt.Errorf("%w '%s'", ErrDuplicateTarget, name)
		}
	}
	for _, name := range names {
		if name != TargetAPI {
			return nil, fmt.Errorf("%w: '%s'; only '%s' can run in this build", ErrUnavailableTarget, name, TargetAPI)
		}
	}
	return names, nil
}
