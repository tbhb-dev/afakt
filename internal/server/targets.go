// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	// TargetAPI selects the HTTP API.
	TargetAPI = "api"
	// TargetWebhook selects the webhook target. It has no receivers yet, so it adds only
	// the health endpoints.
	TargetWebhook = "webhook"
	// TargetDashboard selects the embedded web dashboard.
	TargetDashboard = "dashboard"
	// TargetWorker selects the workflow worker and its Temporal service.
	TargetWorker = "worker"
)

var (
	// ErrNoTarget reports an empty target list.
	ErrNoTarget = errors.New("no target given")
	// ErrEmptyTarget reports an empty entry in the target list.
	ErrEmptyTarget = errors.New("empty entry in the target list")
	// ErrUnknownTarget reports a name that isn't a target.
	ErrUnknownTarget = errors.New("unknown target")
	// ErrDuplicateTarget reports a target named more than once.
	ErrDuplicateTarget = errors.New("duplicate target")
)

// targetNames returns every target name in the order the help text lists them.
func targetNames() []string {
	return []string{TargetAPI, TargetWebhook, TargetDashboard, TargetWorker, "all"}
}

// availableTargets returns the targets that this build can run.
func availableTargets() []string {
	return []string{TargetAPI, TargetWebhook, TargetDashboard, TargetWorker}
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
	if slices.Contains(names, "all") {
		if len(names) != 1 {
			return nil, fmt.Errorf("%w: 'all' includes every target; use it on its own", ErrDuplicateTarget)
		}
		return availableTargets(), nil
	}
	return names, nil
}
