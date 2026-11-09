// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package capacity

import (
	"fmt"
	"slices"
	"strings"
)

// RenderGauges returns Prometheus text exposition. Carried series express only
// a lower bound until reset. Unknown samples are omitted.
func RenderGauges(readings []Reading) string {
	var out strings.Builder

	out.WriteString("# TYPE afakt_capacity_window_utilization_ratio gauge\n")
	out.WriteString("# TYPE afakt_capacity_window_reset_time_seconds gauge\n")

	for _, reading := range readings {
		if !exportable(reading) {
			continue
		}

		labels := fmt.Sprintf(
			`afakt_plan_vendor=%q,afakt_plan_window=%q,afakt_capacity_source=%q,afakt_capacity_freshness=%q`,
			reading.Vendor, reading.Window, reading.Source, reading.Freshness,
		)
		fmt.Fprintf(&out, "afakt_capacity_window_utilization_ratio{%s} %g\n", labels, reading.Utilization)
		fmt.Fprintf(&out, "afakt_capacity_window_reset_time_seconds{%s} %d\n", labels, reading.ResetAt.Unix())
	}

	return out.String()
}

func exportable(reading Reading) bool {
	return reading.Status == StatusOK &&
		reading.Utilization >= 0 && reading.Utilization <= 1 &&
		!reading.ResetAt.IsZero() &&
		slices.Contains([]string{VendorAnthropic, VendorOpenAI, VendorGoogle}, reading.Vendor) &&
		slices.Contains([]string{Window5h, Window7d, Window7dOpus, Window7dSonnet, WindowOther}, reading.Window) &&
		slices.Contains([]string{SourceClaudeUsage, SourceCodexStatus, SourceAGYUsage}, reading.Source) &&
		slices.Contains([]string{Fresh, Carried}, reading.Freshness)
}
