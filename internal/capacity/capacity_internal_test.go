// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package capacity

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

const privateInput = "private input"

func captureTime() time.Time {
	return time.Date(2026, time.November, 8, 13, 50, 0, 0, time.UTC)
}

func weeklyPane(percent, reset string) string {
	return "Current week (all models)\n" + percent + "% used\nResets " + reset + " (America/New_York)\n"
}

func weeklyReading() Reading {
	return ParseClaudeUsage(weeklyPane("1", "Nov 14 at 10pm"), captureTime())[0]
}

func TestFixtures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		file   string
		status string
		window string
	}{
		{"claude-usage-weekly-only", StatusOK, Window7d},
		{"percent-over-100", StatusUnknown, Window7d},
		{"percent-negative", StatusUnknown, Window7d},
		{"reset-past", StatusUnknown, Window7d},
		{"reset-beyond-window", StatusUnknown, Window7d},
		{"missing-reset", StatusUnknown, Window7d},
		{"garbage", StatusUnknown, WindowOther},
		{"rate-limited", StatusOK, Window7d},
	}

	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			t.Parallel()

			pane, err := os.ReadFile("testdata/" + test.file + ".txt")
			if err != nil {
				t.Fatal(err)
			}

			got := ParseClaudeUsage(string(pane), captureTime())

			if len(got) != 1 {
				t.Fatalf("got %d readings, want one", len(got))
			}

			want := unknown(VendorAnthropic, SourceClaudeUsage, test.window, captureTime())

			if test.status == StatusOK {
				want.Status = StatusOK
				want.Utilization = 0.01
				want.ResetAt = time.Date(2026, time.November, 15, 3, 0, 0, 0, time.UTC)
			}

			assertReading(t, got[0], want)
		})
	}
}

func assertReading(t *testing.T, got, want Reading) {
	t.Helper()

	if got.Vendor != want.Vendor || got.Window != want.Window || got.Source != want.Source ||
		got.Status != want.Status || got.Freshness != want.Freshness || got.Utilization != want.Utilization ||
		!got.ResetAt.Equal(want.ResetAt) || !got.CapturedAt.Equal(want.CapturedAt) {
		t.Errorf("reading = %+v, want %+v", got, want)
	}
}

func TestWindowMapping(t *testing.T) {
	t.Parallel()

	pane := "Current session\n100% used\nResets Nov 8 at 10am (America/New_York)\n" +
		"Current week (Opus)\n85.5% used\nResets Nov 14 at 10:15pm (America/New_York)\n" +
		"Current week (Sonnet)\n0% used\nResets Nov 14 at 12am (America/New_York)\n" +
		"Current week (Fable)\n7% used\nResets Nov 14 at 10pm (America/New_York)\n" +
		"Unrecognized window\n9% used\nResets Nov 14 at 10pm (America/New_York)\n"
	got := ParseClaudeUsage(pane, captureTime())

	want := []struct {
		window string
		status string
		used   float64
	}{
		{Window5h, StatusOK, 1},
		{Window7dOpus, StatusOK, 0.855},
		{Window7dSonnet, StatusOK, 0},
		{WindowOther, StatusUnknown, 0},
		{WindowOther, StatusUnknown, 0},
	}

	if len(got) != len(want) {
		t.Fatalf("got %d readings, want %d", len(got), len(want))
	}

	for i, expected := range want {
		if got[i].Window != expected.window || got[i].Status != expected.status || got[i].Utilization != expected.used {
			t.Errorf("reading %d = %+v, want %+v", i, got[i], expected)
		}
	}

	if got[1].ResetAt.Minute() != 15 {
		t.Errorf("reset minute = %d, want 15", got[1].ResetAt.Minute())
	}
}

func TestSchemaGuard(t *testing.T) {
	t.Parallel()

	tests := []string{
		"", "Per-model breakdown unavailable (rate limited until 9:45am)",
		weeklyPane("NaN", "Nov 14 at 10pm"), weeklyPane("1e2", "Nov 14 at 10pm"),
		weeklyPane(strings.Repeat("9", 400), "Nov 14 at 10pm"),
		weeklyPane("1", "Nov 32 at 10pm"), weeklyPane("1", "Nov 14 at 13pm"),
		weeklyPane("1", "Nov 14 at 10:99pm"), weeklyPane("1", "unexpected"),
		strings.ReplaceAll(weeklyPane("1", "Nov 14 at 10pm"), "America/New_York", "Invalid/Zone"),
		strings.ReplaceAll(weeklyPane("1", "Nov 14 at 10pm"), "America/New_York", "Local"),
		"Current week (all models)\nResets Nov 14 at 10pm (America/New_York)\n",
		weeklyPane("1", "Nov 14 at 10pm") + "2% used\n",
		"Current session\n1% used\nResets Nov 8 at 2pm (America/New_York)\n",
		"Current week (all models)\n1% used\nCurrent week (Opus)\n2% used\n",
	}

	for _, pane := range tests {
		for _, reading := range ParseClaudeUsage(pane, captureTime()) {
			if reading.Status != StatusUnknown || reading.Utilization != 0 || !reading.ResetAt.IsZero() {
				t.Errorf("malformed pane yielded %+v", reading)
			}
		}
	}
}

func TestResetBoundaries(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.November, 8, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		reset  string
		now    time.Time
		status string
	}{
		{"exactly now", "Nov 8 at 10am", now, StatusUnknown},
		{"one minute future", "Nov 8 at 10:01am", now, StatusOK},
		{"weekly bound", "Nov 15 at 10am", now, StatusOK},
		{"past weekly bound", "Nov 15 at 10:01am", now, StatusUnknown},
		{"year rollover", "Jan 1 at 12am", time.Date(2026, time.December, 31, 20, 0, 0, 0, time.UTC), StatusOK},
		{"local year", "Dec 31 at 11pm", time.Date(2027, time.January, 1, 3, 0, 0, 0, time.UTC), StatusOK},
		{"old date", "Jan 1 at 12am", now, StatusUnknown},
		{
			"invalid leap rollover",
			"Feb 29 at 12am",
			time.Date(2024, time.December, 31, 20, 0, 0, 0, time.UTC),
			StatusUnknown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			reading := ParseClaudeUsage(weeklyPane("50", test.reset), test.now)[0]

			if reading.Status != test.status {
				t.Errorf("reading = %+v, want status %s", reading, test.status)
			}

			if test.name == "year rollover" && reading.ResetAt.Year() != 2027 {
				t.Errorf("reset year = %d, want 2027", reading.ResetAt.Year())
			}
		})
	}
}

func TestCarry(t *testing.T) {
	t.Parallel()

	prev := weeklyReading()

	tests := []struct {
		name      string
		now       time.Time
		freshness string
		status    string
	}{
		{"capture", prev.CapturedAt, Fresh, StatusOK},
		{"before stale", prev.CapturedAt.Add(15*time.Minute - time.Nanosecond), Fresh, StatusOK},
		{"at stale", prev.CapturedAt.Add(15 * time.Minute), Carried, StatusOK},
		{"before reset", prev.ResetAt.Add(-time.Nanosecond), Carried, StatusOK},
		{"at reset", prev.ResetAt, Fresh, StatusUnknown},
		{"after reset", prev.ResetAt.Add(time.Hour), Fresh, StatusUnknown},
		{"clock reversal", prev.CapturedAt.Add(-time.Nanosecond), Fresh, StatusUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			want := prev
			want.Freshness = test.freshness

			want.Status = test.status

			if test.status == StatusUnknown {
				want.Utilization = 0
				want.ResetAt = time.Time{}
			}

			assertReading(t, Carry(prev, test.now), want)
		})
	}

	for _, used := range []float64{0.4, 0.7, 0.85, 0.91, 1} {
		prev.Utilization = used
		carried := Carry(prev, prev.CapturedAt.Add(FreshFor))
		assertReading(t, Carry(carried, prev.CapturedAt.Add(2*FreshFor)), carried)

		if carried.Utilization != used || carried.Freshness != Carried {
			t.Errorf("carry lowered or refreshed a band floor: %+v", carried)
		}
	}

	missing := unknown(VendorGoogle, SourceAGYUsage, WindowOther, captureTime())
	assertReading(t, Carry(missing, captureTime()), missing)
	missing.ResetAt = prev.ResetAt
	missing.Utilization = 0.91
	assertReading(t, Carry(missing, captureTime()), unknown(VendorGoogle, SourceAGYUsage, WindowOther, captureTime()))
}

func TestSources(t *testing.T) {
	t.Parallel()

	for source, vendor := range map[string]string{SourceCodexStatus: VendorOpenAI, SourceAGYUsage: VendorGoogle} {
		got, err := Parse(source, weeklyPane("1", "Nov 14 at 10pm"), captureTime())

		if err != nil || len(got) != 1 {
			t.Fatalf("Parse(%s) = %+v, %v", source, got, err)
		}

		assertReading(t, got[0], unknown(vendor, source, WindowOther, captureTime()))
	}

	got, err := Parse(SourceClaudeUsage, weeklyPane("1", "Nov 14 at 10pm"), captureTime())

	if err != nil || len(got) != 1 {
		t.Fatalf("Parse(claude_usage) = %+v, %v", got, err)
	}

	assertReading(t, got[0], weeklyReading())

	readings, parseErr := Parse(privateInput, "", captureTime())

	if parseErr == nil || readings != nil {
		t.Errorf("invalid source = %+v, %v", readings, parseErr)
	}
}

func TestReadingJSON(t *testing.T) {
	t.Parallel()

	reading := weeklyReading()

	data, err := json.Marshal(reading)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"vendor":"anthropic","window":"7d","utilization":0.01,"resetAt":1794711600,` +
		`"source":"claude_usage","freshness":"fresh","status":"ok","capturedAt":1794145800}`

	if string(data) != want {
		t.Errorf("JSON = %s, want %s", data, want)
	}

	reading.Status = StatusUnknown

	data, err = json.Marshal(reading)

	if err != nil || strings.Contains(string(data), "utilization") || strings.Contains(string(data), "resetAt") {
		t.Errorf("unknown JSON = %s, %v", data, err)
	}

	reading.Status = StatusOK
	reading.Utilization = math.NaN()

	data, marshalErr := reading.MarshalJSON()
	if data != nil || marshalErr == nil {
		t.Fatalf("NaN JSON = %s, error = %v", data, marshalErr)
	}

	_, isValueErr := errors.AsType[*json.UnsupportedValueError](marshalErr)
	if !isValueErr {
		t.Errorf("JSON error lost its cause: %v", marshalErr)
	}
}

func TestRenderGauges(t *testing.T) {
	t.Parallel()

	reading := weeklyReading()
	labels := `afakt_plan_vendor="anthropic",afakt_plan_window="7d",` +
		`afakt_capacity_source="claude_usage",afakt_capacity_freshness="fresh"`

	want := "# TYPE afakt_capacity_window_utilization_ratio gauge\n" +
		"# TYPE afakt_capacity_window_reset_time_seconds gauge\n" +
		"afakt_capacity_window_utilization_ratio{" + labels + "} 0.01\n" +
		"afakt_capacity_window_reset_time_seconds{" + labels + "} 1794711600\n"

	if got := RenderGauges([]Reading{reading}); got != want {
		t.Errorf("gauges = %s, want %s", got, want)
	}

	carried := Carry(reading, captureTime().Add(FreshFor))

	if got := RenderGauges([]Reading{carried}); got != strings.ReplaceAll(want, `"fresh"`, `"carried"`) {
		t.Errorf("carried gauges = %s", got)
	}

	if got := RenderGauges([]Reading{Carry(reading, reading.ResetAt)}); got != RenderGauges(nil) {
		t.Errorf("expired gauges = %s", got)
	}
}

func TestGaugesRejectInvalidSamples(t *testing.T) {
	t.Parallel()

	tests := []struct {
		change func(*Reading)
		name   string
	}{
		{func(r *Reading) { r.Status = StatusUnknown }, "status"},
		{func(r *Reading) { r.Utilization = -0.1 }, "negative"},
		{func(r *Reading) { r.Utilization = 1.1 }, "too high"},
		{func(r *Reading) { r.Utilization = math.NaN() }, "NaN"},
		{func(r *Reading) { r.Utilization = math.Inf(1) }, "infinity"},
		{func(r *Reading) { r.ResetAt = time.Time{} }, "reset"},
		{func(r *Reading) { r.Vendor = privateInput }, "vendor"},
		{func(r *Reading) { r.Window = privateInput }, "window"},
		{func(r *Reading) { r.Source = privateInput }, "source"},
		{func(r *Reading) { r.Freshness = privateInput }, "freshness"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			reading := weeklyReading()
			test.change(&reading)

			if got := RenderGauges([]Reading{reading}); got != RenderGauges(nil) {
				t.Errorf("invalid sample exported: %s", got)
			}
		})
	}
}

func TestEmptyWindow(t *testing.T) {
	t.Parallel()

	got := ParseClaudeUsage("Current week (all models)", captureTime())
	if len(got) != 1 {
		t.Fatalf("got %d readings, want one", len(got))
	}

	assertReading(t, got[0], unknown(VendorAnthropic, SourceClaudeUsage, Window7d, captureTime()))
}

func TestGaugeRatioBounds(t *testing.T) {
	t.Parallel()

	for _, used := range []float64{0, 1} {
		reading := weeklyReading()
		reading.Utilization = used

		if got := RenderGauges([]Reading{reading}); got == RenderGauges(nil) {
			t.Errorf("valid ratio %g was omitted", used)
		}
	}
}

func TestPaneSeparators(t *testing.T) {
	t.Parallel()

	for _, separator := range []string{"\n\n", "\nPer-model breakdown unavailable (rate limited until 9:45am)\n"} {
		pane := strings.ReplaceAll(weeklyPane("1", "Nov 14 at 10pm"), "\n", separator)
		got := ParseClaudeUsage(pane, captureTime())

		if len(got) != 1 {
			t.Fatalf("got %d readings, want one", len(got))
		}

		assertReading(t, got[0], weeklyReading())
	}

	pane := weeklyPane("1", "Nov 14 at 10pm") + "Unrecognized window\n5% used\n"
	got := ParseClaudeUsage(pane, captureTime())

	if len(got) != 2 {
		t.Fatalf("got %d readings, want two", len(got))
	}

	assertReading(t, got[1], unknown(VendorAnthropic, SourceClaudeUsage, WindowOther, captureTime()))
}

func TestWindowLengths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		heading string
		window  string
		length  time.Duration
	}{
		{"Current session", Window5h, 5 * time.Hour},
		{"Current week (all models)", Window7d, 7 * 24 * time.Hour},
		{"Current week (Opus)", Window7dOpus, 7 * 24 * time.Hour},
		{"Current week (Sonnet)", Window7dSonnet, 7 * 24 * time.Hour},
		{"Unrecognized window", WindowOther, 0},
	}
	for _, test := range tests {
		window, length := windowLabel(test.heading)
		if window != test.window || length != test.length {
			t.Errorf("windowLabel(%s) = %s, %s; want %s, %s", test.heading, window, length, test.window, test.length)
		}
	}
}

func TestResetYearInference(t *testing.T) {
	t.Parallel()

	reset := time.Date(2026, time.November, 5, 15, 0, 0, 0, time.UTC)
	tests := []struct {
		age  time.Duration
		year int
	}{
		{36 * time.Hour, 2026},
		{60 * time.Hour, 2026},
		{71*time.Hour + 30*time.Minute, 2026},
		{72 * time.Hour, 2026},
		{72*time.Hour + 30*time.Minute, 2027},
		{84 * time.Hour, 2027},
	}

	for _, test := range tests {
		got := parseReset("Resets Nov 5 at 10am (America/New_York)", reset.Add(test.age))
		if got.Year() != test.year {
			t.Errorf("reset year at age %s = %d, want %d", test.age, got.Year(), test.year)
		}
	}

	now := time.Date(2027, time.February, 28, 15, 0, 0, 0, time.UTC)
	if got := parseReset("Resets Feb 29 at 10am (America/New_York)", now); !got.IsZero() {
		t.Errorf("invalid date became %s", got)
	}
}

func TestGaugesSkipInvalidAndContinue(t *testing.T) {
	t.Parallel()

	invalid := unknown(VendorAnthropic, SourceClaudeUsage, Window7d, captureTime())
	valid := weeklyReading()

	if got := RenderGauges([]Reading{invalid, valid}); got != RenderGauges([]Reading{valid}) {
		t.Errorf("invalid sample hid a later gauge: %s", got)
	}
}
