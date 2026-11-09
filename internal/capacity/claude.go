// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package capacity

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	// Bundle the time zone database so zone names in the usage pane resolve on
	// hosts that lack one, such as Windows runners.
	_ "time/tzdata"
)

// ParseClaudeUsage parses complete percent/reset pairs under window headings.
// Unrecognized windows and malformed sections yield unknown readings.
func ParseClaudeUsage(pane string, now time.Time) []Reading {
	readings := make([]Reading, 0)
	heading := ""

	var fields []string

	flush := func() {
		if strings.HasPrefix(heading, "Current ") || len(fields) > 0 {
			readings = append(readings, parseWindow(heading, fields, now))
		}
	}

	for line := range strings.SplitSeq(pane, "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "Per-model breakdown unavailable (rate limited until ") {
			continue
		}

		if strings.Contains(line, "%") || strings.HasPrefix(line, "Resets ") {
			fields = append(fields, line)

			continue
		}

		flush()

		heading = line
		fields = nil
	}

	flush()

	if len(readings) == 0 {
		readings = append(readings, unknown(VendorAnthropic, SourceClaudeUsage, WindowOther, now))
	}

	return readings
}

func windowLabel(heading string) (string, time.Duration) {
	switch heading {
	case "Current session":
		const sessionWindow = 5 * time.Hour

		return Window5h, sessionWindow
	case "Current week (all models)":
		return Window7d, weeklyWindow
	case "Current week (Opus)":
		return Window7dOpus, weeklyWindow
	case "Current week (Sonnet)":
		return Window7dSonnet, weeklyWindow
	default:
		return WindowOther, 0
	}
}

func parseWindow(heading string, fields []string, now time.Time) Reading {
	window, length := windowLabel(heading)
	reading := unknown(VendorAnthropic, SourceClaudeUsage, window, now)

	const fieldCount = 2

	if len(fields) != fieldCount {
		return reading
	}

	pattern := regexp.MustCompile(`(?:^|\s)([+-]?[0-9]+(?:\.[0-9]+)?)% used$`)

	percent := pattern.FindStringSubmatch(fields[0])

	if percent == nil {
		return reading
	}

	// gomutants:disable-next-line INTEGER_INCREMENT,INTEGER_DECREMENT reason="63, 64, and 65 select float64."
	used, err := strconv.ParseFloat(percent[1], 64)

	const percentScale = 100

	// gomutants:disable-next-line EXPRESSION_REMOVE reason="Numeric parse errors also fail the range check."
	if err != nil || used < 0 || used > percentScale {
		return reading
	}

	reset := parseReset(fields[1], now)

	if !reset.After(now) || reset.After(now.Add(length)) {
		return reading
	}

	reading.Status = StatusOK
	reading.Utilization = used / percentScale
	reading.ResetAt = reset

	return reading
}

func parseReset(line string, now time.Time) time.Time {
	pattern := regexp.MustCompile(
		`^Resets ([A-Z][a-z]{2} [0-9]{1,2}) at ([0-9]{1,2}(?::[0-9]{2})?(?:am|pm)) \(([^()]+)\)$`,
	)

	parts := pattern.FindStringSubmatch(line)

	if parts == nil || parts[3] == "Local" {
		return time.Time{}
	}

	zone, err := time.LoadLocation(parts[3])
	if err != nil {
		return time.Time{}
	}

	layout := "2006 Jan 2 3pm"

	if strings.Contains(parts[2], ":") {
		layout = "2006 Jan 2 3:04pm"
	}

	year := now.In(zone).Year()

	reset, err := time.ParseInLocation(layout, fmt.Sprintf("%d %s %s", year, parts[1], parts[2]), zone)
	if err != nil {
		return time.Time{}
	}

	const rolloverGrace = 3 * 24 * time.Hour

	if reset.Before(now.Add(-rolloverGrace)) {
		reset, err = time.ParseInLocation(layout, fmt.Sprintf("%d %s %s", year+1, parts[1], parts[2]), zone)
		// gomutants:disable-next-line BRANCH_IF reason="Parse errors already return zero."
		if err != nil {
			return time.Time{}
		}
	}

	return reset
}
