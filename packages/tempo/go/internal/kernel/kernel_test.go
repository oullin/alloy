package kernel_test

import (
	"testing"
	"time"

	"github.com/oullin/alloy/tempo/duration"
	"github.com/oullin/alloy/tempo/internal/kernel"
)

func parseUTC(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339Nano, value)

	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}

	return parsed.UTC()
}

func TestAddAcrossUnits(t *testing.T) {
	base := parseUTC(t, "2024-01-31T10:20:30.400Z")

	if got := kernel.Add(base, time.UTC, 90, duration.Minute, true, true); got.Format(time.RFC3339Nano) != "2024-01-31T11:50:30.4Z" {
		t.Fatalf("Add(Minute) = %q, want shifted instant", got.Format(time.RFC3339Nano))
	}

	if got := kernel.Add(base, time.UTC, 1, duration.Month, true, true); got.Format("2006-01-02") != "2024-03-02" {
		t.Fatalf("Add(Month, overflow=true) = %q, want overflowed date", got.Format("2006-01-02"))
	}

	if got := kernel.Add(base, time.UTC, 1, duration.Month, false, true); got.Format("2006-01-02") != "2024-02-29" {
		t.Fatalf("Add(Month, overflow=false) = %q, want clamped date", got.Format("2006-01-02"))
	}

	leap := parseUTC(t, "2024-02-29T00:00:00Z")

	if got := kernel.Add(leap, time.UTC, 1, duration.Year, true, false); got.Format("2006-01-02") != "2025-02-28" {
		t.Fatalf("Add(Year, yearsOverflow=false) = %q, want clamped date", got.Format("2006-01-02"))
	}

	if got := kernel.Add(base, time.UTC, 1, duration.Quarter, true, true); got.Format("2006-01-02") != "2024-05-01" {
		t.Fatalf("Add(Quarter) = %q, want quarter advanced", got.Format("2006-01-02"))
	}
}

func TestStartOfAndEndOf(t *testing.T) {
	base := parseUTC(t, "2024-05-15T10:34:45.600Z")

	if got := kernel.StartOf(base, time.UTC, duration.Day); got.Format(time.RFC3339Nano) != "2024-05-15T00:00:00Z" {
		t.Fatalf("StartOf(Day) = %q", got.Format(time.RFC3339Nano))
	}

	if got := kernel.StartOf(base, time.UTC, duration.Month); got.Format(time.RFC3339Nano) != "2024-05-01T00:00:00Z" {
		t.Fatalf("StartOf(Month) = %q", got.Format(time.RFC3339Nano))
	}

	if got := kernel.StartOf(base, time.UTC, duration.Quarter); got.Format(time.RFC3339Nano) != "2024-04-01T00:00:00Z" {
		t.Fatalf("StartOf(Quarter) = %q", got.Format(time.RFC3339Nano))
	}

	if got := kernel.EndOf(base, time.UTC, duration.Day); got.Format("2006-01-02 15:04:05.000") != "2024-05-15 23:59:59.999" {
		t.Fatalf("EndOf(Day) = %q", got.Format("2006-01-02 15:04:05.000"))
	}

	if got := kernel.EndOf(base, time.UTC, duration.Year); got.Format("2006-01-02 15:04:05.000") != "2024-12-31 23:59:59.999" {
		t.Fatalf("EndOf(Year) = %q", got.Format("2006-01-02 15:04:05.000"))
	}

	wednesday := parseUTC(t, "2024-05-15T10:34:45.600Z")

	if got := kernel.StartOf(wednesday, time.UTC, duration.Week); got.Format("2006-01-02") != "2024-05-13" {
		t.Fatalf("StartOf(Week) default Monday = %q, want Monday", got.Format("2006-01-02"))
	}

	if got := kernel.StartOf(wednesday, time.UTC, duration.Week, kernel.WeekOptions{WeekStartsOn: time.Sunday}); got.Format("2006-01-02") != "2024-05-12" {
		t.Fatalf("StartOf(Week, Sunday) = %q, want Sunday", got.Format("2006-01-02"))
	}
}

func TestCompareValueOrdersByUnit(t *testing.T) {
	left := parseUTC(t, "2024-05-15T00:00:00.000Z")
	right := parseUTC(t, "2024-05-15T23:59:59.999Z")

	if kernel.CompareValue(left, time.UTC, duration.Day) != kernel.CompareValue(right, time.UTC, duration.Day) {
		t.Fatalf("CompareValue(Day) should collapse same-day instants")
	}

	if kernel.CompareValue(left, time.UTC, duration.Millisecond) == kernel.CompareValue(right, time.UTC, duration.Millisecond) {
		t.Fatalf("CompareValue(Millisecond) should distinguish instants")
	}

	zeroUnit := kernel.CompareValue(left, time.UTC, "")

	if zeroUnit != left.UnixNano() {
		t.Fatalf("CompareValue(\"\") = %d, want raw nanoseconds %d", zeroUnit, left.UnixNano())
	}
}

func TestAddMonthsAndYearsNoOverflowClampsDay(t *testing.T) {
	march := parseUTC(t, "2024-03-31T08:00:00Z")

	if got := kernel.AddMonthsNoOverflow(march, time.UTC, 1); got.Format("2006-01-02") != "2024-04-30" {
		t.Fatalf("AddMonthsNoOverflow = %q, want clamped", got.Format("2006-01-02"))
	}

	leap := parseUTC(t, "2024-02-29T00:00:00Z")

	if got := kernel.AddYearsNoOverflow(leap, time.UTC, 1); got.Format("2006-01-02") != "2025-02-28" {
		t.Fatalf("AddYearsNoOverflow = %q, want clamped", got.Format("2006-01-02"))
	}
}
