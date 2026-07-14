package duration_test

import (
	"reflect"
	"testing"

	"github.com/oullin/alloy/pkg/hub/tempo"
)

func TestDurationsParseNormalizeSerializeAndApply(t *testing.T) {
	parsed, err := tempo.ParseDuration("P1Y2M3DT4H5M6.007S")

	if err != nil {
		t.Fatalf("parse duration: %v", err)
	}

	normalized, err := tempo.ParseDuration("P1Y14M8DT25H61M61.250S")

	if err != nil {
		t.Fatalf("parse normalized duration: %v", err)
	}

	if parsed.Years != 1 || parsed.Months != 2 || parsed.Days != 3 || parsed.Hours != 4 || parsed.Minutes != 5 || parsed.Seconds != 6 || parsed.Milliseconds != 7 {
		t.Fatalf("ParseDuration() = %#v, want parsed components", parsed)
	}

	if got := parsed.ToMap()["hours"]; got != 4 {
		t.Fatalf("Span.ToMap()[hours] = %d, want 4", got)
	}

	if got := parsed.ToSlice(); !reflect.DeepEqual(got, []int{1, 0, 2, 0, 3, 4, 5, 6, 7}) {
		t.Fatalf("Span.ToSlice() = %#v, want component slice", got)
	}

	if got := parsed.ISOString(); got != "P1Y2M3DT4H5M6.007S" {
		t.Fatalf("ISOString() = %q, want parsed duration string", got)
	}

	if normalized.Years != 2 || normalized.Months != 2 || normalized.Days != 9 || normalized.Hours != 2 || normalized.Minutes != 2 || normalized.Seconds != 1 || normalized.Milliseconds != 250 {
		t.Fatalf("Normalize() = %#v, want carried components", normalized)
	}

	if got := normalized.ISOString(); got != "P2Y2M9DT2H2M1.250S" {
		t.Fatalf("ISOString() = %q, want normalized duration string", got)
	}

	zero, err := tempo.ParseDuration("PT0S")

	if err != nil {
		t.Fatalf("parse zero duration: %v", err)
	}

	if !zero.IsZero() {
		t.Fatalf("IsZero() = false, want true")
	}

	if !parsed.IsPositive() {
		t.Fatalf("IsPositive() = false, want true")
	}

	negative, err := tempo.ParseDuration("-P1D")

	if err != nil {
		t.Fatalf("parse negative duration: %v", err)
	}

	if !negative.IsNegative() {
		t.Fatalf("IsNegative() = false, want true")
	}

	if zero.IsPositive() || zero.IsNegative() {
		t.Fatalf("zero duration sign predicates = positive:%t negative:%t, want false/false", zero.IsPositive(), zero.IsNegative())
	}

	weeks, err := tempo.ParseDuration("P2W")

	if err != nil {
		t.Fatalf("parse week duration: %v", err)
	}

	normalizedWeeks := weeks.Normalize()

	if got := normalizedWeeks.ISOString(); got != "P14D" {
		t.Fatalf("Normalize().ISOString() = %q, want week carried to days", got)
	}

	base, err := tempo.Parse("2024-01-31T00:00:00Z")

	if err != nil {
		t.Fatalf("parse base: %v", err)
	}

	oneMonth, err := tempo.ParseDuration("P1M")

	if err != nil {
		t.Fatalf("parse one month: %v", err)
	}

	if got := base.AddDuration(oneMonth).DateString(); got != "2024-03-02" {
		t.Fatalf("AddDuration(P1M).DateString() = %q, want overflowed date", got)
	}

	if got := base.SubDuration(tempo.Span{Days: 2, Hours: 3}).ISOString(); got != "2024-01-28T21:00:00.000Z" {
		t.Fatalf("SubDuration().ISOString() = %q, want shifted instant", got)
	}

	start, err := tempo.Parse("2024-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse interval start: %v", err)
	}

	end, err := tempo.Parse("2024-01-03T12:00:00Z")

	if err != nil {
		t.Fatalf("parse interval end: %v", err)
	}

	intervalDuration := start.IntervalUntil(end).ToDuration()

	if got := intervalDuration.ISOString(); got != "P2DT12H" {
		t.Fatalf("Interval.ToDuration().ISOString() = %q, want normalized interval", got)
	}

	periodStart, err := tempo.Parse("2024-01-01")

	if err != nil {
		t.Fatalf("parse period start: %v", err)
	}

	periodEnd, err := tempo.Parse("2024-01-03")

	if err != nil {
		t.Fatalf("parse period end: %v", err)
	}

	if got, err := periodStart.ToPeriod(periodEnd).Count(); err != nil || got != 3 {
		t.Fatalf("ToPeriod().Count() = %d, %v, want 3, nil", got, err)
	}

	if got, err := periodStart.Until(periodEnd).Count(); err != nil || got != 3 {
		t.Fatalf("Until().Count() = %d, %v, want 3, nil", got, err)
	}

	if got, err := periodStart.Range(periodEnd).Count(); err != nil || got != 3 {
		t.Fatalf("Range().Count() = %d, %v, want 3, nil", got, err)
	}
}
