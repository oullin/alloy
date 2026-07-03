package arithmetic_test

import (
	"testing"
	"time"

	"github.com/oullin/alloy/packages/foundation/tempo"
)

func mustTempo(t *testing.T, value tempo.Time, err error) tempo.Time {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected tempo error: %v", err)
	}

	return value
}

func TestArithmeticBoundariesAndOverflowModes(t *testing.T) {
	base, err := tempo.Parse("2024-01-31T10:20:30.400Z")

	if err != nil {
		t.Fatalf("parse base: %v", err)
	}

	set, err := base.Set(tempo.Components{Day: 15, Hour: 1, Minute: 2, Second: 3, Millisecond: 4})

	if err != nil {
		t.Fatalf("set components: %v", err)
	}

	if got := set.ISOString(); got != "2024-01-15T01:02:03.004Z" {
		t.Fatalf("Set().ISOString() = %q, want updated components", got)
	}

	if got := base.AddMonths(1).DateString(); got != "2024-03-02" {
		t.Fatalf("AddMonths(1).DateString() = %q, want overflowed date", got)
	}

	if got := base.AddMonthsNoOverflow(1).DateString(); got != "2024-02-29" {
		t.Fatalf("AddMonthsNoOverflow(1).DateString() = %q, want clamped date", got)
	}

	leap, err := tempo.Parse("2024-02-29T00:00:00Z")

	if err != nil {
		t.Fatalf("parse leap: %v", err)
	}

	if got := leap.AddYearsNoOverflow(1).DateString(); got != "2025-02-28" {
		t.Fatalf("AddYearsNoOverflow(1).DateString() = %q, want clamped date", got)
	}

	if got := base.StartOf(tempo.Day).ISOString(); got != "2024-01-31T00:00:00.000Z" {
		t.Fatalf("StartOf(Day).ISOString() = %q, want start of day", got)
	}

	if got := base.StartOfWeek().DateString(); got != "2024-01-29" {
		t.Fatalf("StartOfWeek().DateString() = %q, want Monday week start", got)
	}

	if got := base.EndOfWeek(tempo.StartOfWeekOptions{WeekStartsOn: time.Sunday}).DateString(); got != "2024-02-03" {
		t.Fatalf("EndOfWeek(Sunday).DateString() = %q, want Saturday week end", got)
	}

	if got := base.EndOf(tempo.Month).ISOString(); got != "2024-01-31T23:59:59.999Z" {
		t.Fatalf("EndOf(Month).ISOString() = %q, want end of month", got)
	}

	if got := base.StartOf(tempo.Quarter).ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("StartOf(Quarter).ISOString() = %q, want quarter start", got)
	}

	if got := base.StartOfQuarter().ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("StartOfQuarter().ISOString() = %q, want quarter start", got)
	}

	if got := base.EndOfQuarter().ISOString(); got != "2024-03-31T23:59:59.999Z" {
		t.Fatalf("EndOfQuarter().ISOString() = %q, want quarter end", got)
	}

	if got := base.EndOf(tempo.Year).ISOString(); got != "2024-12-31T23:59:59.999Z" {
		t.Fatalf("EndOf(Year).ISOString() = %q, want end of year", got)
	}
}

func TestWeekdayArithmeticAndSameUnitComparisons(t *testing.T) {
	friday, err := tempo.Parse("2024-05-17T10:00:00Z")

	if err != nil {
		t.Fatalf("parse friday: %v", err)
	}

	wednesday, err := tempo.Parse("2024-05-22T10:00:00Z")

	if err != nil {
		t.Fatalf("parse wednesday: %v", err)
	}

	if got := friday.AddWeekdays(1).DateTimeString(); got != "2024-05-20 10:00:00" {
		t.Fatalf("AddWeekdays(1).DateTimeString() = %q, want next Monday", got)
	}

	if got := friday.AddWeekdays(3).DateTimeString(); got != "2024-05-22 10:00:00" {
		t.Fatalf("AddWeekdays(3).DateTimeString() = %q, want Wednesday", got)
	}

	if got := wednesday.SubWeekdays(3).DateTimeString(); got != "2024-05-17 10:00:00" {
		t.Fatalf("SubWeekdays(3).DateTimeString() = %q, want Friday", got)
	}

	if got := wednesday.DiffInWeekdays(friday); got != 3 {
		t.Fatalf("DiffInWeekdays() = %d, want 3", got)
	}

	if got := wednesday.DiffInWeekendDays(friday); got != 2 {
		t.Fatalf("DiffInWeekendDays() = %d, want 2", got)
	}

	if got := wednesday.DiffInUnit(tempo.Day, friday); got != 5 {
		t.Fatalf("DiffInUnit(Day) = %d, want 5", got)
	}

	if got := wednesday.DiffInDaysFiltered(friday, func(item tempo.Time) bool { return item.IsMonday() }); got != 1 {
		t.Fatalf("DiffInDaysFiltered(Monday) = %d, want 1", got)
	}

	if got := wednesday.DiffInHoursFiltered(friday, func(item tempo.Time) bool { return item.Hour() == 12 }); got != 5 {
		t.Fatalf("DiffInHoursFiltered(hour 12) = %d, want 5", got)
	}

	if got := friday.DiffInWeekdays(wednesday); got != -3 {
		t.Fatalf("negative DiffInWeekdays() = %d, want -3", got)
	}

	if got := friday.DiffInWeekdays(wednesday, tempo.DiffOptions{Absolute: true}); got != 3 {
		t.Fatalf("absolute DiffInWeekdays() = %d, want 3", got)
	}

	october, err := tempo.Parse("2024-10-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse october: %v", err)
	}

	if got := october.DiffInQuarters(friday); got != 1 {
		t.Fatalf("DiffInQuarters() = %d, want 1", got)
	}

	if got := friday.DiffInQuarters(october); got != -1 {
		t.Fatalf("negative DiffInQuarters() = %d, want -1", got)
	}

	november, err := tempo.Parse("2024-11-17T10:00:00Z")

	if err != nil {
		t.Fatalf("parse november: %v", err)
	}

	if got := friday.IntervalUntil(november).Quarters(); got != 2 {
		t.Fatalf("Interval.Quarters() = %d, want 2", got)
	}

	sameSecond, err := tempo.Parse("2024-05-17T10:00:00.999Z")

	if err != nil {
		t.Fatalf("parse same second: %v", err)
	}

	sameMinute, err := tempo.Parse("2024-05-17T10:00:59Z")

	if err != nil {
		t.Fatalf("parse same minute: %v", err)
	}

	sameHour, err := tempo.Parse("2024-05-17T10:59:59Z")

	if err != nil {
		t.Fatalf("parse same hour: %v", err)
	}

	sameDay, err := tempo.Parse("2024-05-17T23:59:59Z")

	if err != nil {
		t.Fatalf("parse same day: %v", err)
	}

	sameWeek, err := tempo.Parse("2024-05-13T00:00:00Z")

	if err != nil {
		t.Fatalf("parse same week: %v", err)
	}

	sameMonth, err := tempo.Parse("2024-05-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse same month: %v", err)
	}

	sameQuarter, err := tempo.Parse("2024-04-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse same quarter: %v", err)
	}

	sameYear, err := tempo.Parse("2024-12-31T23:59:59Z")

	if err != nil {
		t.Fatalf("parse same year: %v", err)
	}

	birthday, err := tempo.Parse("1990-05-17T00:00:00Z")

	if err != nil {
		t.Fatalf("parse birthday: %v", err)
	}

	notBirthday, err := tempo.Parse("1990-05-18T00:00:00Z")

	if err != nil {
		t.Fatalf("parse not birthday: %v", err)
	}

	if !friday.SameSecond(sameSecond) || !friday.SameMinute(sameMinute) || !friday.SameHour(sameHour) || !friday.SameDay(sameDay) || !friday.SameWeek(sameWeek) || !friday.SameMonth(sameMonth) || !friday.SameQuarter(sameQuarter) || !friday.SameYear(sameYear) {
		t.Fatalf("same-unit comparisons did not all match expected true values")
	}

	if !friday.EqualTo(friday.Clone()) || !friday.Eq(friday.Clone()) {
		t.Fatalf("equality aliases = false, want true")
	}

	if !friday.NotEqualTo(wednesday) || !friday.Ne(wednesday) {
		t.Fatalf("inequality aliases = false, want true")
	}

	if !wednesday.GreaterThan(friday) || !wednesday.Gt(friday) {
		t.Fatalf("greater-than aliases = false, want true")
	}

	if !friday.GreaterThanOrEqualTo(friday.Clone()) || !friday.Gte(friday.Clone()) {
		t.Fatalf("greater-than-or-equal aliases = false, want true")
	}

	if !friday.LessThan(wednesday) || !friday.Lt(wednesday) {
		t.Fatalf("less-than aliases = false, want true")
	}

	if !friday.LessThanOrEqualTo(friday.Clone()) || !friday.Lte(friday.Clone()) {
		t.Fatalf("less-than-or-equal aliases = false, want true")
	}

	if !friday.IsBetween(friday, wednesday) {
		t.Fatalf("IsBetween(inclusive) = false, want true")
	}

	if friday.IsBetween(friday, wednesday, "()") {
		t.Fatalf("IsBetween(exclusive) = true, want false")
	}

	if !friday.SameAs("YYYY-MM-DD", sameDay) {
		t.Fatalf("SameAs(date pattern) = false, want true")
	}

	if !friday.IsSameUnit(tempo.Day, sameDay) {
		t.Fatalf("IsSameUnit(Day) = false, want true")
	}

	if friday.SameAs("YYYY-MM-DD HH:mm", sameDay) {
		t.Fatalf("SameAs(datetime pattern) = true, want false")
	}

	if !friday.Birthday(birthday) {
		t.Fatalf("Birthday() = false, want true")
	}

	if friday.Birthday(notBirthday) {
		t.Fatalf("Birthday() = true, want false")
	}

	secondsSince, err := friday.SetTime(0, 0, 42, 0)

	if err != nil {
		t.Fatalf("SetTime seconds since: %v", err)
	}

	if got := secondsSince.SecondsSinceMidnight(); got != 42 {
		t.Fatalf("SecondsSinceMidnight() = %d, want 42", got)
	}

	secondsUntil, err := friday.SetTime(23, 59, 17, 0)

	if err != nil {
		t.Fatalf("SetTime seconds until: %v", err)
	}

	if got := secondsUntil.SecondsUntilEndOfDay(); got != 42 {
		t.Fatalf("SecondsUntilEndOfDay() = %d, want 42", got)
	}

	if got := friday.MidDay().TimeString(); got != "12:00:00" {
		t.Fatalf("MidDay().TimeString() = %q, want noon", got)
	}
}
