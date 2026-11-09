// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

// Package capacity parses plan windows without retaining pane text.
package capacity

import "time"

// Closed codes bound the labels in JSON and metric output.
const (
	VendorAnthropic   = "anthropic"
	VendorOpenAI      = "openai"
	VendorGoogle      = "google"
	Window5h          = "5h"
	Window7d          = "7d"
	Window7dOpus      = "7d_opus"
	Window7dSonnet    = "7d_sonnet"
	WindowOther       = "other"
	SourceClaudeUsage = "claude_usage"
	SourceCodexStatus = "codex_status"
	SourceAGYUsage    = "agy_usage"
	Fresh             = "fresh"
	Carried           = "carried"
	StatusOK          = "ok"
	StatusUnknown     = "unknown"

	// FreshFor is the age at which a capture becomes stale.
	FreshFor     = 15 * time.Minute
	weeklyWindow = 7 * 24 * time.Hour
)

// Reading contains closed labels and a validated window sample.
// Its ratio and reset have no value when Status is unknown.
type Reading struct {
	ResetAt     time.Time
	CapturedAt  time.Time
	Vendor      string
	Window      string
	Source      string
	Freshness   string
	Status      string
	Utilization float64
}

// Carry preserves the sample's capture time. After FreshFor, its ratio is only
// a lower bound for raising a band. It can't establish an all-clear. At reset,
// the sample becomes unknown rather than implying zero use in the next window.
func Carry(prev Reading, now time.Time) Reading {
	if prev.Status != StatusOK || now.Before(prev.CapturedAt) || !prev.ResetAt.After(now) {
		prev.Status = StatusUnknown
		prev.Utilization = 0
		prev.ResetAt = time.Time{}

		return prev
	}

	if !now.Before(prev.CapturedAt.Add(FreshFor)) {
		prev.Freshness = Carried
	}

	return prev
}

func unknown(vendor, source, window string, now time.Time) Reading {
	return Reading{
		Vendor: vendor, Window: window, Source: source, Freshness: Fresh, Status: StatusUnknown,
		CapturedAt: now, Utilization: 0, ResetAt: time.Time{},
	}
}
