package setters_test

import (
	"testing"
	"time"

	"hara.sh/alloy/tempo"
)

func assertEqual(t *testing.T, label string, got string, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %q, want %q", label, got, want)
	}
}

func mustTempo(t *testing.T, value tempo.Time, err error) tempo.Time {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected tempo error: %v", err)
	}

	return value
}

func TestTimezoneConversionModes(t *testing.T) {
	utc, err := tempo.Parse("2024-01-01T12:00:00Z")

	if err != nil {
		t.Fatalf("parse utc: %v", err)
	}

	tokyo, err := utc.SetTimezone("Asia/Tokyo")

	if err != nil {
		t.Fatalf("set timezone: %v", err)
	}

	preserved, err := utc.SetTimezoneKeepLocal("Asia/Tokyo")

	if err != nil {
		t.Fatalf("set timezone keep local: %v", err)
	}

	if got := tokyo.DateTimeString(); got != "2024-01-01 21:00:00" {
		t.Fatalf("DateTimeString() = %q, want Tokyo local time", got)
	}

	if got := tokyo.ISOString(); got != "2024-01-01T12:00:00.000Z" {
		t.Fatalf("ISOString() = %q, want same instant", got)
	}

	if got := preserved.DateTimeString(); got != "2024-01-01 12:00:00" {
		t.Fatalf("DateTimeString() = %q, want preserved local time", got)
	}

	if got := preserved.ISOString(); got != "2024-01-01T03:00:00.000Z" {
		t.Fatalf("ISOString() = %q, want shifted instant", got)
	}
}

func TestISOWeekMetadataWeekdayNavigationAndAge(t *testing.T) {
	monday, err := tempo.Parse("2024-01-01T12:30:00Z")

	if err != nil {
		t.Fatalf("parse monday: %v", err)
	}

	week53, err := tempo.Parse("2020-12-31T12:30:00Z")

	if err != nil {
		t.Fatalf("parse week53: %v", err)
	}

	friday, err := tempo.Parse("2024-05-17T12:30:00Z")

	if err != nil {
		t.Fatalf("parse friday: %v", err)
	}

	if got := monday.ISOWeekday(); got != 1 {
		t.Fatalf("ISOWeekday() = %d, want 1", got)
	}

	if got := monday.ISOWeekNumber(); got != 1 {
		t.Fatalf("ISOWeekNumber() = %d, want 1", got)
	}

	if got := monday.ISOWeekYear(); got != 2024 {
		t.Fatalf("ISOWeekYear() = %d, want 2024", got)
	}

	if got := monday.WeeksInISOYear(); got != 52 {
		t.Fatalf("WeeksInISOYear() = %d, want 52", got)
	}

	if got := week53.ISOWeekNumber(); got != 53 {
		t.Fatalf("ISOWeekNumber() = %d, want 53", got)
	}

	if got := week53.ISOWeekYear(); got != 2020 {
		t.Fatalf("ISOWeekYear() = %d, want 2020", got)
	}

	if got := week53.WeeksInISOYear(); got != 53 {
		t.Fatalf("WeeksInISOYear() = %d, want 53", got)
	}

	if !monday.IsMonday() {
		t.Fatalf("IsMonday() = false, want true")
	}

	if got := monday.Next(time.Friday).DateTimeString(); got != "2024-01-05 12:30:00" {
		t.Fatalf("Next(Friday).DateTimeString() = %q, want next Friday", got)
	}

	if got := monday.Previous(time.Friday).DateTimeString(); got != "2023-12-29 12:30:00" {
		t.Fatalf("Previous(Friday).DateTimeString() = %q, want previous Friday", got)
	}

	if got := monday.NextOrSame(time.Monday).DateTimeString(); got != "2024-01-01 12:30:00" {
		t.Fatalf("NextOrSame(Monday).DateTimeString() = %q, want same Monday", got)
	}

	if got := monday.PreviousOrSame(time.Monday).DateTimeString(); got != "2024-01-01 12:30:00" {
		t.Fatalf("PreviousOrSame(Monday).DateTimeString() = %q, want same Monday", got)
	}

	if got := monday.NextOrSame(time.Friday).DateTimeString(); got != "2024-01-05 12:30:00" {
		t.Fatalf("NextOrSame(Friday).DateTimeString() = %q, want next Friday", got)
	}

	if got := monday.PreviousOrSame(time.Friday).DateTimeString(); got != "2023-12-29 12:30:00" {
		t.Fatalf("PreviousOrSame(Friday).DateTimeString() = %q, want previous Friday", got)
	}

	if got := friday.NextWeekday().DateString(); got != "2024-05-20" {
		t.Fatalf("NextWeekday().DateString() = %q, want Monday", got)
	}

	if got := monday.PreviousWeekday().DateString(); got != "2023-12-29" {
		t.Fatalf("PreviousWeekday().DateString() = %q, want Friday", got)
	}

	if got := friday.FirstOfMonth().DateString(); got != "2024-05-01" {
		t.Fatalf("FirstOfMonth().DateString() = %q, want month start", got)
	}

	if got := friday.FirstOfMonth(time.Monday).DateString(); got != "2024-05-06" {
		t.Fatalf("FirstOfMonth(Monday).DateString() = %q, want first Monday", got)
	}

	if got := friday.LastOfMonth().DateString(); got != "2024-05-31" {
		t.Fatalf("LastOfMonth().DateString() = %q, want month end date", got)
	}

	if got := friday.LastOfMonth(time.Friday).DateString(); got != "2024-05-31" {
		t.Fatalf("LastOfMonth(Friday).DateString() = %q, want last Friday", got)
	}

	if got, ok := friday.NthOfMonth(3, time.Monday); !ok || got.DateString() != "2024-05-20" {
		t.Fatalf("NthOfMonth(3, Monday) = %q, %v, want 2024-05-20, true", got.DateString(), ok)
	}

	if got, ok := friday.NthOfMonth(-1, time.Monday); !ok || got.DateString() != "2024-05-27" {
		t.Fatalf("NthOfMonth(-1, Monday) = %q, %v, want 2024-05-27, true", got.DateString(), ok)
	}

	if _, ok := friday.NthOfMonth(5, time.Monday); ok {
		t.Fatalf("NthOfMonth(5, Monday) ok = true, want false")
	}

	if got := friday.FirstOfQuarter(time.Monday).DateString(); got != "2024-04-01" {
		t.Fatalf("FirstOfQuarter(Monday).DateString() = %q, want first quarter Monday", got)
	}

	if got := friday.LastOfQuarter(time.Friday).DateString(); got != "2024-06-28" {
		t.Fatalf("LastOfQuarter(Friday).DateString() = %q, want last quarter Friday", got)
	}

	if got, ok := friday.NthOfQuarter(2, time.Monday); !ok || got.DateString() != "2024-04-08" {
		t.Fatalf("NthOfQuarter(2, Monday) = %q, %v, want 2024-04-08, true", got.DateString(), ok)
	}

	if got, ok := friday.NthOfQuarter(-1, time.Friday); !ok || got.DateString() != "2024-06-28" {
		t.Fatalf("NthOfQuarter(-1, Friday) = %q, %v, want 2024-06-28, true", got.DateString(), ok)
	}

	if _, ok := friday.NthOfQuarter(14, time.Monday); ok {
		t.Fatalf("NthOfQuarter(14, Monday) ok = true, want false")
	}

	if got := friday.FirstOfYear(time.Monday).DateString(); got != "2024-01-01" {
		t.Fatalf("FirstOfYear(Monday).DateString() = %q, want first year Monday", got)
	}

	if got := friday.LastOfYear(time.Tuesday).DateString(); got != "2024-12-31" {
		t.Fatalf("LastOfYear(Tuesday).DateString() = %q, want last year Tuesday", got)
	}

	if got, ok := friday.NthOfYear(20, time.Monday); !ok || got.DateString() != "2024-05-13" {
		t.Fatalf("NthOfYear(20, Monday) = %q, %v, want 2024-05-13, true", got.DateString(), ok)
	}

	if got, ok := friday.NthOfYear(-1, time.Tuesday); !ok || got.DateString() != "2024-12-31" {
		t.Fatalf("NthOfYear(-1, Tuesday) = %q, %v, want 2024-12-31, true", got.DateString(), ok)
	}

	if _, ok := friday.NthOfYear(54, time.Monday); ok {
		t.Fatalf("NthOfYear(54, Monday) ok = true, want false")
	}

	birthday, err := tempo.Parse("2000-06-15T00:00:00Z")

	if err != nil {
		t.Fatalf("parse birthday: %v", err)
	}

	beforeBirthday, err := tempo.Parse("2024-06-14T23:59:59Z")

	if err != nil {
		t.Fatalf("parse before birthday: %v", err)
	}

	onBirthday, err := tempo.Parse("2024-06-15T00:00:00Z")

	if err != nil {
		t.Fatalf("parse on birthday: %v", err)
	}

	if got := birthday.Age(beforeBirthday); got != 23 {
		t.Fatalf("Age(before birthday) = %d, want 23", got)
	}

	if got := birthday.Age(onBirthday); got != 24 {
		t.Fatalf("Age(on birthday) = %d, want 24", got)
	}
}

func TestExplicitSettersHandleZeroValues(t *testing.T) {
	instance, err := tempo.Parse("2024-05-15T12:34:56.789Z")

	if err != nil {
		t.Fatalf("parse fixture input: %v", err)
	}

	midnight, err := instance.SetTime(0, 0, 0, 0)

	if err != nil {
		t.Fatalf("SetTime(0,0,0,0): %v", err)
	}

	assertEqual(t, "SetTime().ISOString()", midnight.ISOString(), "2024-05-15T00:00:00.000Z")

	if !midnight.IsMidnight() {
		t.Fatalf("SetTime(0,0,0,0).IsMidnight() = false, want true")
	}

	assertEqual(t, "Midday().ISOString()", instance.Midday().ISOString(), "2024-05-15T12:00:00.000Z")

	if !instance.Midday().IsMidday() {
		t.Fatalf("Midday().IsMidday() = false, want true")
	}

	hourZero, err := instance.SetHour(0)

	if err != nil {
		t.Fatalf("SetHour(0): %v", err)
	}

	if got := hourZero.Hour(); got != 0 {
		t.Fatalf("SetHour(0).Hour() = %d, want 0", got)
	}

	minuteZero, err := instance.SetMinute(0)

	if err != nil {
		t.Fatalf("SetMinute(0): %v", err)
	}

	if got := minuteZero.Minute(); got != 0 {
		t.Fatalf("SetMinute(0).Minute() = %d, want 0", got)
	}

	secondZero, err := instance.SetSecond(0)

	if err != nil {
		t.Fatalf("SetSecond(0): %v", err)
	}

	if got := secondZero.Second(); got != 0 {
		t.Fatalf("SetSecond(0).Second() = %d, want 0", got)
	}

	millisecondZero, err := instance.SetMillisecond(0)

	if err != nil {
		t.Fatalf("SetMillisecond(0): %v", err)
	}

	if got := millisecondZero.Millisecond(); got != 0 {
		t.Fatalf("SetMillisecond(0).Millisecond() = %d, want 0", got)
	}

	setDate, err := instance.SetDate(2025, 1, 2)

	if err != nil {
		t.Fatalf("SetDate(2025,1,2): %v", err)
	}

	assertEqual(t, "SetDate().DateString()", setDate.DateString(), "2025-01-02")
	setUnit, err := instance.SetUnit(tempo.Day, 2)

	if err != nil {
		t.Fatalf("SetUnit(Day, 2): %v", err)
	}

	assertEqual(t, "SetUnit(Day, 2).DateString()", setUnit.DateString(), "2024-05-02")
	setDateTime, err := instance.SetDateTime(2025, 1, 2, 3, 4, 5, 6)

	if err != nil {
		t.Fatalf("SetDateTime(): %v", err)
	}

	assertEqual(t, "SetDateTime().ISOString()", setDateTime.ISOString(), "2025-01-02T03:04:05.006Z")

	source, err := tempo.Parse("2025-01-02T03:04:05.006Z")

	if err != nil {
		t.Fatalf("parse setter source: %v", err)
	}

	setDateFrom, err := instance.SetDateFrom(source)

	if err != nil {
		t.Fatalf("SetDateFrom(): %v", err)
	}

	assertEqual(t, "SetDateFrom().DateTimeString()", setDateFrom.DateTimeString(), "2025-01-02 12:34:56")
	setTimeFrom, err := instance.SetTimeFrom(source)

	if err != nil {
		t.Fatalf("SetTimeFrom(): %v", err)
	}

	assertEqual(t, "SetTimeFrom().DateTimeString()", setTimeFrom.DateTimeString(), "2024-05-15 03:04:05")
	setDateTimeFrom, err := instance.SetDateTimeFrom(source)

	if err != nil {
		t.Fatalf("SetDateTimeFrom(): %v", err)
	}

	assertEqual(t, "SetDateTimeFrom().ISOString()", setDateTimeFrom.ISOString(), "2025-01-02T03:04:05.006Z")

	fromTimeString, err := instance.SetTimeFromTimeString("03:04:05.006")

	if err != nil {
		t.Fatalf("SetTimeFromTimeString(): %v", err)
	}

	assertEqual(t, "SetTimeFromTimeString().ISOString()", fromTimeString.ISOString(), "2024-05-15T03:04:05.006Z")
	assertEqual(t, "SetTimestamp(0).ISOString()", instance.SetTimestamp(0).ISOString(), "1970-01-01T00:00:00.000Z")
	assertEqual(t, "SetISODate().DateString()", instance.SetISODate(2024, 20, 3).DateString(), "2024-05-15")

	if got := instance.Weekday(); got != 3 {
		t.Fatalf("Weekday() = %d, want 3", got)
	}

	assertEqual(t, "SetWeekday().DateString()", instance.SetWeekday(time.Friday).DateString(), "2024-05-17")
	assertEqual(t, "SetDayOfYear().DateString()", instance.SetDayOfYear(60).DateString(), "2024-02-29")
	modified, err := instance.Modify("+2 days")

	if err != nil {
		t.Fatalf("Modify(+2 days): %v", err)
	}

	assertEqual(t, "Modify(+2 days).DateString()", modified.DateString(), "2024-05-17")
	modified, err = instance.Modify("previous day")

	if err != nil {
		t.Fatalf("Modify(previous day): %v", err)
	}

	assertEqual(t, "Modify(previous day).DateString()", modified.DateString(), "2024-05-14")
	modified, err = instance.Change("next week")

	if err != nil {
		t.Fatalf("Change(next week): %v", err)
	}

	assertEqual(t, "Change(next week).DateString()", modified.DateString(), "2024-05-22")
	assertEqual(t, "Sub(2, Day).DateString()", instance.Sub(2, tempo.Day).DateString(), "2024-05-13")
	assertEqual(t, "Add(2, Day).DateString()", instance.Add(2, tempo.Day).DateString(), "2024-05-17")
	assertEqual(t, "AddNoOverflow(20, Day, Month).DateString()", instance.AddNoOverflow(20, tempo.Day, tempo.Month).DateString(), "2024-05-31")
	assertEqual(t, "SubNoOverflow(20, Day, Month).DateString()", instance.SubNoOverflow(20, tempo.Day, tempo.Month).DateString(), "2024-05-01")
	assertEqual(t, "SetUnitNoOverflow(Day, 40, Month).DateString()", instance.SetUnitNoOverflow(tempo.Day, 40, tempo.Month).DateString(), "2024-05-31")

	friday, err := tempo.Parse("2024-05-17T12:00:00Z")

	if err != nil {
		t.Fatalf("parse friday: %v", err)
	}

	monday, err := tempo.Parse("2024-05-20T12:00:00Z")

	if err != nil {
		t.Fatalf("parse monday: %v", err)
	}

	if got := friday.NextWeekendDay().DayOfWeek(); got != 6 {
		t.Fatalf("NextWeekendDay().DayOfWeek() = %d, want 6", got)
	}

	if got := monday.PreviousWeekendDay().DayOfWeek(); got != 0 {
		t.Fatalf("PreviousWeekendDay().DayOfWeek() = %d, want 0", got)
	}

	shifted, err := instance.ShiftTimezone("Asia/Tokyo")

	if err != nil {
		t.Fatalf("ShiftTimezone(): %v", err)
	}

	assertEqual(t, "ShiftTimezone().DateTimeString()", shifted.DateTimeString(), "2024-05-15 12:34:56")

	createdFromTimeString, err := tempo.CreateFromTimeString("03:04:05.006")

	if err != nil {
		t.Fatalf("CreateFromTimeString(): %v", err)
	}

	assertEqual(t, "CreateFromTimeString().TimeString(ms)", createdFromTimeString.TimeString(tempo.MillisecondPrecision), "03:04:05.006")

	createdFromFormat, err := tempo.FromFormat("2024/05/15", "YYYY/MM/DD")

	if err != nil {
		t.Fatalf("FromFormat(): %v", err)
	}

	assertEqual(t, "FromFormat().DateString()", createdFromFormat.DateString(), "2024-05-15")
	parsed, err := tempo.Parse("2024-05-15")

	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	assertEqual(t, "Parse().DateString()", parsed.DateString(), "2024-05-15")
	fromLocale, err := tempo.Parse("2024-05-15", tempo.WithLocale("en-US"))

	if err != nil {
		t.Fatalf("Parse with locale: %v", err)
	}

	assertEqual(t, "Parse(locale).DateString()", fromLocale.DateString(), "2024-05-15")
	days := tempo.Days()

	if len(days) < 2 || days[0] != "Sunday" || days[1] != "Monday" {
		t.Fatalf("Days()[0:2] = %v, want Sunday/Monday", days[:2])
	}

	if got := tempo.CalendarFormats()["sameDay"]; got != "[Today at] HH:mm" {
		t.Fatalf("CalendarFormats()[sameDay] = %q, want default format", got)
	}

	if got := tempo.ISOFormats()["date"]; got != "YYYY-MM-DD" {
		t.Fatalf("ISOFormats()[date] = %q, want date format", got)
	}

	if units := tempo.ISOUnits(); len(units) == 0 || units[0] != tempo.Millisecond {
		t.Fatalf("ISOUnits()[0] = %v, want millisecond", units)
	}

	if got := tempo.TimeFormatByPrecision(tempo.MillisecondPrecision); got != "HH:mm:ss.SSS" {
		t.Fatalf("TimeFormatByPrecision(ms) = %q, want millisecond format", got)
	}

	if tempo.WeekStartsAt() != time.Monday || tempo.WeekEndsAt() != time.Sunday {
		t.Fatalf("week start/end settings = %v/%v, want Monday/Sunday", tempo.WeekStartsAt(), tempo.WeekEndsAt())
	}

	if !tempo.LocaleHasDiffSyntax("en-US") || !tempo.LocaleHasPeriodSyntax("en-US") {
		t.Fatalf("locale capability helpers = false, want true")
	}

	createdFromTimestamp, err := tempo.FromTimestamp(0)

	if err != nil {
		t.Fatalf("FromTimestamp(): %v", err)
	}

	assertEqual(t, "FromTimestamp(0).ISOString()", createdFromTimestamp.ISOString(), "1970-01-01T00:00:00.000Z")

	createdFromTimestampMs, err := tempo.FromTimestampMs(1)

	if err != nil {
		t.Fatalf("FromTimestampMs(): %v", err)
	}

	assertEqual(t, "FromTimestampMs(1).ISOString()", createdFromTimestampMs.ISOString(), "1970-01-01T00:00:00.001Z")

	fromTimestampUTC, err := tempo.FromTimestampUTC(0)

	if err != nil {
		t.Fatalf("FromTimestampUTC(): %v", err)
	}

	assertEqual(t, "FromTimestampUTC(0).ISOString()", fromTimestampUTC.ISOString(), "1970-01-01T00:00:00.000Z")

	fromTimestampMsUTC, err := tempo.FromTimestampMsUTC(1)

	if err != nil {
		t.Fatalf("FromTimestampMsUTC(): %v", err)
	}

	assertEqual(t, "FromTimestampMsUTC(1).ISOString()", fromTimestampMsUTC.ISOString(), "1970-01-01T00:00:00.001Z")

	createdFromTimestampUTC, err := tempo.FromTimestampUTC(0)

	if err != nil {
		t.Fatalf("FromTimestampUTC(): %v", err)
	}

	assertEqual(t, "FromTimestampUTC(0).ISOString()", createdFromTimestampUTC.ISOString(), "1970-01-01T00:00:00.000Z")

	createdFromTimestampMsUTC, err := tempo.FromTimestampMsUTC(1)

	if err != nil {
		t.Fatalf("FromTimestampMsUTC(): %v", err)
	}

	assertEqual(t, "FromTimestampMsUTC(1).ISOString()", createdFromTimestampMsUTC.ISOString(), "1970-01-01T00:00:00.001Z")

	assertEqual(t, "From()", instance.AddDays(2).From(instance), "in 2 days")
	assertEqual(t, "Since()", instance.AddDays(2).Since(instance), "in 2 days")
	assertEqual(t, "To()", instance.To(instance.AddDays(2)), "in 2 days")
	assertEqual(t, "Timespan()", instance.AddDays(2).Timespan(instance), "in 2 days")

	if !instance.IsImmutable() || instance.IsMutable() {
		t.Fatalf("Time mutability predicates = immutable:%t mutable:%t, want true/false", instance.IsImmutable(), instance.IsMutable())
	}

	assertEqual(t, "AvoidMutation().ISOString()", instance.AvoidMutation().ISOString(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "Cast().ISOString()", instance.Cast().ISOString(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "Tempoize().DateString()", instance.Tempoize(instance.AddDays(1)).DateString(), "2024-05-16")

	if got := instance.NowWithSameTz().Timezone(); got != "UTC" {
		t.Fatalf("NowWithSameTz().Timezone() = %q, want UTC", got)
	}

	mutable := tempo.NewMutable(instance)

	if mutable.IsImmutable() || !mutable.IsMutable() {
		t.Fatalf("MutableTime mutability predicates = immutable:%t mutable:%t, want false/true", mutable.IsImmutable(), mutable.IsMutable())
	}

	assertEqual(t, "MutableTime.AvoidMutation().ISOString()", mutable.AvoidMutation().ISOString(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "MutableTime.Cast().ISOString()", mutable.Cast().ISOString(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "MutableTime.Tempoize().DateString()", mutable.Tempoize(instance.AddDays(1)).DateString(), "2024-05-16")

	if got := mutable.NowWithSameTz().Timezone(); got != "UTC" {
		t.Fatalf("MutableTime.NowWithSameTz().Timezone() = %q, want UTC", got)
	}
}
