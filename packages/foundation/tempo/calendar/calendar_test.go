package calendar_test

import (
	"testing"

	"alloy.dev/foundation/tempo/calendar"
)

func TestMonthAndDayNames(t *testing.T) {
	if got := calendar.MonthName(1); got != "January" {
		t.Fatalf("MonthName(1) = %q, want January", got)
	}

	if got := calendar.MonthName(12); got != "December" {
		t.Fatalf("MonthName(12) = %q, want December", got)
	}

	if got := calendar.ShortMonthName(3); got != "Mar" {
		t.Fatalf("ShortMonthName(3) = %q, want Mar", got)
	}

	if got := calendar.DayName(0); got != "Sunday" {
		t.Fatalf("DayName(0) = %q, want Sunday", got)
	}

	if got := calendar.ShortDayName(6); got != "Sat" {
		t.Fatalf("ShortDayName(6) = %q, want Sat", got)
	}

	days := calendar.Days()

	if len(days) != 7 || days[0] != "Sunday" || days[6] != "Saturday" {
		t.Fatalf("Days() = %v, want 7 names Sunday..Saturday", days)
	}

	days[0] = "mutated"

	if calendar.Days()[0] != "Sunday" {
		t.Fatalf("Days() returned a shared slice — mutation leaked back into the package")
	}
}

func TestDaysInMonth(t *testing.T) {
	cases := []struct {
		year, month, want int
	}{
		{2024, 1, 31},
		{2024, 2, 29},
		{2023, 2, 28},
		{2024, 4, 30},
		{2000, 2, 29},
		{1900, 2, 28},
	}

	for _, c := range cases {
		if got := calendar.DaysInMonth(c.year, c.month); got != c.want {
			t.Fatalf("DaysInMonth(%d, %d) = %d, want %d", c.year, c.month, got, c.want)
		}
	}
}

func TestMonthDiff(t *testing.T) {
	if got := calendar.MonthDiff(2024, 1, 15, 2024, 3, 15); got != 2 {
		t.Fatalf("MonthDiff Jan 15 → Mar 15 = %d, want 2", got)
	}

	if got := calendar.MonthDiff(2024, 1, 20, 2024, 3, 15); got != 1 {
		t.Fatalf("MonthDiff Jan 20 → Mar 15 = %d, want 1 (end-day < start-day rolls back)", got)
	}

	if got := calendar.MonthDiff(2023, 6, 1, 2024, 6, 1); got != 12 {
		t.Fatalf("MonthDiff one year = %d, want 12", got)
	}

	if got := calendar.MonthDiff(2024, 6, 1, 2023, 6, 1); got != -12 {
		t.Fatalf("MonthDiff backwards one year = %d, want -12", got)
	}
}
