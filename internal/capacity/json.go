// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package capacity

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON emits epoch seconds and closed labels, omitting unknown values.
func (r Reading) MarshalJSON() ([]byte, error) {
	var (
		utilization *float64
		resetAt     *int64
	)

	if r.Status == StatusOK {
		utilization = &r.Utilization
		seconds := r.ResetAt.Unix()
		resetAt = &seconds
	}

	data, err := json.Marshal(struct {
		Vendor      string   `json:"vendor"`
		Window      string   `json:"window"`
		Utilization *float64 `json:"utilization,omitempty"`
		ResetAt     *int64   `json:"resetAt,omitempty"`
		Source      string   `json:"source"`
		Freshness   string   `json:"freshness"`
		Status      string   `json:"status"`
		CapturedAt  int64    `json:"capturedAt"`
	}{
		Vendor: r.Vendor, Window: r.Window, Utilization: utilization, ResetAt: resetAt,
		Source: r.Source, Freshness: r.Freshness, Status: r.Status, CapturedAt: r.CapturedAt.Unix(),
	})
	if err != nil {
		return nil, fmt.Errorf("encode reading: %w", err)
	}

	return data, nil
}
