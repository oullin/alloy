package boundaries_test

import (
	"testing"
	"time"

	"alloy.dev/api/tempo"
	"alloy.dev/api/tempo/boundaries"
	"alloy.dev/api/tempo/duration"
	"alloy.dev/api/tempo/internal/kernel"
)

func mustParse(t *testing.T, value string) tempo.Time {
	t.Helper()

	parsed, err := tempo.Parse(value)

	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}

	return parsed
}

func TestStartOfAndEndOfAcrossUnits(t *testing.T) {
	base := mustParse(t, "2024-05-15T10:34:45.600Z")

	if got := boundaries.StartOf(base, duration.Day).ISOString(); got != "2024-05-15T00:00:00.000Z" {
		t.Fatalf("StartOf(Day) = %q", got)
	}

	if got := boundaries.EndOf(base, duration.Day).ISOString(); got != "2024-05-15T23:59:59.999Z" {
		t.Fatalf("EndOf(Day) = %q", got)
	}

	if got := boundaries.StartOf(base, duration.Month).ISOString(); got != "2024-05-01T00:00:00.000Z" {
		t.Fatalf("StartOf(Month) = %q", got)
	}

	if got := boundaries.EndOf(base, duration.Year).ISOString(); got != "2024-12-31T23:59:59.999Z" {
		t.Fatalf("EndOf(Year) = %q", got)
	}

	if got := boundaries.StartOf(base, duration.Week).DateString(); got != "2024-05-13" {
		t.Fatalf("StartOf(Week) default Monday = %q", got)
	}

	if got := boundaries.StartOf(base, duration.Week, kernel.WeekOptions{WeekStartsOn: time.Sunday}).DateString(); got != "2024-05-12" {
		t.Fatalf("StartOf(Week, Sunday) = %q", got)
	}
}

func TestFloorCeilRoundOnHour(t *testing.T) {
	base := mustParse(t, "2024-05-15T10:34:45.600Z")

	if got := boundaries.Floor(base, duration.Hour).ISOString(); got != "2024-05-15T10:00:00.000Z" {
		t.Fatalf("Floor(Hour) = %q", got)
	}

	if got := boundaries.Ceil(base, duration.Hour).ISOString(); got != "2024-05-15T11:00:00.000Z" {
		t.Fatalf("Ceil(Hour) = %q", got)
	}

	if got := boundaries.Round(base, duration.Hour).ISOString(); got != "2024-05-15T11:00:00.000Z" {
		t.Fatalf("Round(Hour) = %q", got)
	}

	exact := mustParse(t, "2024-05-15T10:00:00.000Z")

	if got := boundaries.Ceil(exact, duration.Hour).ISOString(); got != "2024-05-15T10:00:00.000Z" {
		t.Fatalf("Ceil(Hour) on aligned = %q, want unchanged", got)
	}
}

func TestFirstAndLastOfMonth(t *testing.T) {
	base := mustParse(t, "2024-05-15T10:00:00Z")

	if got := boundaries.FirstOfMonth(base).DateString(); got != "2024-05-01" {
		t.Fatalf("FirstOfMonth() = %q", got)
	}

	if got := boundaries.FirstOfMonth(base, time.Friday).DateString(); got != "2024-05-03" {
		t.Fatalf("FirstOfMonth(Friday) = %q", got)
	}

	if got := boundaries.LastOfMonth(base).DateString(); got != "2024-05-31" {
		t.Fatalf("LastOfMonth() = %q", got)
	}

	if got := boundaries.LastOfMonth(base, time.Friday).DateString(); got != "2024-05-31" {
		t.Fatalf("LastOfMonth(Friday) = %q", got)
	}
}

func TestNthOfMonthHappyAndOverflow(t *testing.T) {
	base := mustParse(t, "2024-05-15T10:00:00Z")

	second, ok := boundaries.NthOfMonth(base, 2, time.Wednesday)

	if !ok {
		t.Fatalf("NthOfMonth(2, Wednesday) ok = false, want true")
	}

	if got := second.DateString(); got != "2024-05-08" {
		t.Fatalf("NthOfMonth(2, Wednesday) = %q, want second Wednesday", got)
	}

	if _, ok := boundaries.NthOfMonth(base, 6, time.Wednesday); ok {
		t.Fatalf("NthOfMonth(6, Wednesday) ok = true, want false (overflow)")
	}
}

func TestIsStartOfAndIsEndOf(t *testing.T) {
	startOfDay := mustParse(t, "2024-05-15T00:00:00Z")

	if !boundaries.IsStartOf(startOfDay, duration.Day) {
		t.Fatalf("IsStartOf(Day) = false, want true")
	}

	notStart := mustParse(t, "2024-05-15T00:00:01Z")

	if boundaries.IsStartOf(notStart, duration.Day) {
		t.Fatalf("IsStartOf(Day) on off-boundary = true, want false")
	}

	endOfDay := mustParse(t, "2024-05-15T23:59:59.999Z")

	if !boundaries.IsEndOf(endOfDay, duration.Day) {
		t.Fatalf("IsEndOf(Day) = false, want true")
	}
}
