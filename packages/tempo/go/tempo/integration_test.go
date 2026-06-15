package tempo

import (
	"strings"
	"testing"
	"time"
)

func assertEqual(t *testing.T, label string, got string, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %q, want %q", label, got, want)
	}
}

func TestDateCases(t *testing.T) {
	cases := []struct {
		Name               string
		Input              string
		ExpectedISO        string
		ExpectedDate       string
		AddDays            int
		ExpectedAddDaysISO string
	}{
		{
			Name:               "parse utc start of day",
			Input:              "2024-02-29T00:00:00+00:00",
			ExpectedISO:        "2024-02-29T00:00:00.000Z",
			ExpectedDate:       "2024-02-29",
			AddDays:            1,
			ExpectedAddDaysISO: "2024-03-01T00:00:00.000Z",
		},
		{
			Name:               "parse utc end of year",
			Input:              "2024-12-31T23:30:00+00:00",
			ExpectedISO:        "2024-12-31T23:30:00.000Z",
			ExpectedDate:       "2024-12-31",
			AddDays:            1,
			ExpectedAddDaysISO: "2025-01-01T23:30:00.000Z",
		},
	}

	for _, item := range cases {
		t.Run(item.Name, func(t *testing.T) {
			parsed, err := Parse(item.Input)

			if err != nil {
				t.Fatalf("parse fixture input: %v", err)
			}

			if got := parsed.ISOString(); got != item.ExpectedISO {
				t.Fatalf("ISOString() = %q, want %q", got, item.ExpectedISO)
			}

			if got := parsed.DateString(); got != item.ExpectedDate {
				t.Fatalf("DateString() = %q, want %q", got, item.ExpectedDate)
			}

			if item.ExpectedAddDaysISO != "" {
				if got := parsed.AddDays(item.AddDays).ISOString(); got != item.ExpectedAddDaysISO {
					t.Fatalf("AddDays(%d).ISOString() = %q, want %q", item.AddDays, got, item.ExpectedAddDaysISO)
				}
			}
		})
	}
}

func TestMutableAndImmutableConversions(t *testing.T) {
	immutable, err := Parse("2024-02-29T00:00:00+00:00")

	if err != nil {
		t.Fatalf("parse immutable: %v", err)
	}

	mutable := immutable.Mutable()
	mutable.AddDays(1)

	if got := immutable.ISOString(); got != "2024-02-29T00:00:00.000Z" {
		t.Fatalf("immutable after Mutable().AddDays() = %q, want unchanged", got)
	}

	if got := mutable.ISOString(); got != "2024-03-01T00:00:00.000Z" {
		t.Fatalf("mutable after AddDays() = %q, want changed", got)
	}

	converted := mutable.Immutable()
	mutable.AddDays(1)

	if got := converted.ISOString(); got != "2024-03-01T00:00:00.000Z" {
		t.Fatalf("converted immutable after mutable change = %q, want snapshot", got)
	}

	if got := mutable.ISOString(); got != "2024-03-02T00:00:00.000Z" {
		t.Fatalf("mutable after second AddDays() = %q, want changed", got)
	}
}

func TestFactoryScopedTimezoneAndTestNow(t *testing.T) {
	factory, err := NewFactory(WithTimezone("Asia/Tokyo"))

	if err != nil {
		t.Fatalf("new factory: %v", err)
	}

	parsed, err := factory.Parse("2024-01-01 09:00")

	if err != nil {
		t.Fatalf("factory parse: %v", err)
	}

	if got := parsed.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("Factory.Parse().ISOString() = %q, want scoped timezone instant", got)
	}

	formatted, err := factory.FromFormat("2024-01-01 09:30", "YYYY-MM-DD HH:mm")

	if err != nil {
		t.Fatalf("factory from format: %v", err)
	}

	if got := formatted.ISOString(); got != "2024-01-01T00:30:00.000Z" {
		t.Fatalf("Factory.FromFormat().ISOString() = %q, want scoped timezone instant", got)
	}

	if parsed, ok := factory.TryParse("2024-01-01 09:00"); !ok || parsed.ISOString() != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("Factory.TryParse(valid) = %q, %v, want scoped instant, true", parsed.ISOString(), ok)
	}

	if factory.CanParse("not a date") {
		t.Fatalf("Factory.CanParse(invalid) = true, want false")
	}

	if _, ok := factory.TryParse("not a date"); ok {
		t.Fatalf("Factory.TryParse(invalid) ok = true, want false")
	}

	if parsed, ok := factory.TryFromFormat("2024-01-01 09:30", "YYYY-MM-DD HH:mm"); !ok || parsed.ISOString() != "2024-01-01T00:30:00.000Z" {
		t.Fatalf("Factory.TryFromFormat(valid) = %q, %v, want scoped instant, true", parsed.ISOString(), ok)
	}

	if !factory.HasFormat("2024-01-01 09:30", "YYYY-MM-DD HH:mm") {
		t.Fatalf("Factory.HasFormat(valid) = false, want true")
	}

	if factory.HasFormat("2024/01/01", "YYYY-MM-DD") {
		t.Fatalf("Factory.HasFormat(invalid) = true, want false")
	}

	created, err := factory.Create(Components{Year: 2024, Month: 1, Day: 1, Hour: 9})

	if err != nil {
		t.Fatalf("factory create: %v", err)
	}

	if got := created.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("Factory.Create().ISOString() = %q, want scoped timezone instant", got)
	}

	safeCreated, err := factory.CreateSafe(Components{Year: 2024, Month: 2, Day: 29})

	if err != nil {
		t.Fatalf("factory create safe: %v", err)
	}

	if got := safeCreated.ISOString(); got != "2024-02-28T15:00:00.000Z" {
		t.Fatalf("Factory.CreateSafe().ISOString() = %q, want scoped safe instant", got)
	}

	if _, err := factory.CreateSafe(Components{Year: 2024, Month: 2, Day: 31}); err == nil {
		t.Fatalf("Factory.CreateSafe(invalid) error = nil, want error")
	}

	fromObject, err := factory.FromObject(Components{Year: 2024, Month: 1, Day: 1, Hour: 9})

	if err != nil {
		t.Fatalf("factory from object: %v", err)
	}

	if got := fromObject.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("Factory.FromObject().ISOString() = %q, want scoped timezone instant", got)
	}

	if got := factory.FromTimestamp(1704067200).DateTimeString(); got != "2024-01-01 09:00:00" {
		t.Fatalf("Factory.FromTimestamp().DateTimeString() = %q, want scoped local time", got)
	}

	if got := factory.FromTimestampMs(1704067200000).DateTimeString(); got != "2024-01-01 09:00:00" {
		t.Fatalf("Factory.FromTimestampMs().DateTimeString() = %q, want scoped local time", got)
	}

	if got := factory.CreateFromDate(2024, 1, 2).ISOString(); got != "2024-01-01T15:00:00.000Z" {
		t.Fatalf("Factory.CreateFromDate().ISOString() = %q, want scoped midnight instant", got)
	}

	if got := factory.CreateMidnightDate(2024, 1, 2).ISOString(); got != "2024-01-01T15:00:00.000Z" {
		t.Fatalf("Factory.CreateMidnightDate().ISOString() = %q, want scoped midnight instant", got)
	}

	if got := factory.CreateFromTime(9, 30, 15, 250).TimeString(MillisecondPrecision); got != "09:30:15.250" {
		t.Fatalf("Factory.CreateFromTime().TimeString(ms) = %q, want requested time", got)
	}

	fromDate, err := CreateFromDate(2024, 1, 2, WithTimezone("Asia/Tokyo"))

	if err != nil {
		t.Fatalf("create from date: %v", err)
	}

	if got := fromDate.ISOString(); got != "2024-01-01T15:00:00.000Z" {
		t.Fatalf("CreateFromDate().ISOString() = %q, want scoped midnight instant", got)
	}

	midnight, err := CreateMidnightDate(2024, 1, 2, WithTimezone("Asia/Tokyo"))

	if err != nil {
		t.Fatalf("create midnight date: %v", err)
	}

	if got := midnight.ISOString(); got != "2024-01-01T15:00:00.000Z" {
		t.Fatalf("CreateMidnightDate().ISOString() = %q, want scoped midnight instant", got)
	}

	fromTime, err := CreateFromTime(9, 30, 15, 250, WithTimezone("UTC"))

	if err != nil {
		t.Fatalf("create from time: %v", err)
	}

	if got := fromTime.TimeString(MillisecondPrecision); got != "09:30:15.250" {
		t.Fatalf("CreateFromTime().TimeString(ms) = %q, want requested time", got)
	}

	frozen, err := NewFactoryWithTestNow(parsed)

	if err != nil {
		t.Fatalf("new frozen factory: %v", err)
	}

	if got := frozen.Now().ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("Factory.Now().ISOString() = %q, want frozen instant", got)
	}

	if got := frozen.Today().DateTimeString(); got != "2024-01-01 00:00:00" {
		t.Fatalf("Factory.Today().DateTimeString() = %q, want frozen local start of day", got)
	}

	if got := frozen.Tomorrow().DateTimeString(); got != "2024-01-02 00:00:00" {
		t.Fatalf("Factory.Tomorrow().DateTimeString() = %q, want next frozen local day", got)
	}

	if got := frozen.Yesterday().DateTimeString(); got != "2023-12-31 00:00:00" {
		t.Fatalf("Factory.Yesterday().DateTimeString() = %q, want previous frozen local day", got)
	}

	if got := frozen.ImmutableNow().ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("Factory.ImmutableNow().ISOString() = %q, want frozen instant", got)
	}

	if got := frozen.MutableNow().AddHours(1).ISOString(); got != "2024-01-01T01:00:00.000Z" {
		t.Fatalf("Factory.MutableNow().AddHours().ISOString() = %q, want mutable frozen instant", got)
	}

	today, err := Today(WithTimezone("UTC"))

	if err != nil {
		t.Fatalf("today: %v", err)
	}

	tomorrow, err := Tomorrow(WithTimezone("UTC"))

	if err != nil {
		t.Fatalf("tomorrow: %v", err)
	}

	yesterday, err := Yesterday(WithTimezone("UTC"))

	if err != nil {
		t.Fatalf("yesterday: %v", err)
	}

	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 || today.Millisecond() != 0 {
		t.Fatalf("Today() = %s, want start of day", today.DateTimeString())
	}

	if got := tomorrow.DiffInDays(today); got != 1 {
		t.Fatalf("Tomorrow().DiffInDays(Today()) = %d, want 1", got)
	}

	if got := yesterday.DiffInDays(today); got != -1 {
		t.Fatalf("Yesterday().DiffInDays(Today()) = %d, want -1", got)
	}
}

func TestGlobalSettingsAffectDateBehavior(t *testing.T) {
	original := SettingsState()

	defer SettingsState(original)

	SetLocale("fr-FR")

	if got := GetLocale(); got != "fr-FR" {
		t.Fatalf("GetLocale() = %q, want fr-FR", got)
	}

	SetFallbackLocale("en-GB")

	if got := GetFallbackLocale(); got != "en-GB" {
		t.Fatalf("GetFallbackLocale() = %q, want en-GB", got)
	}

	SetWeekendDays([]time.Weekday{time.Friday, time.Saturday})
	friday, err := Parse("2024-05-17")

	if err != nil {
		t.Fatalf("parse friday: %v", err)
	}

	sunday, err := Parse("2024-05-19")

	if err != nil {
		t.Fatalf("parse sunday: %v", err)
	}

	if !friday.IsWeekend() || !sunday.IsWeekday() {
		t.Fatalf("configured weekend predicates failed")
	}

	SetMidDayAt(13)
	midday, err := Parse("2024-05-15T13:00:00Z")

	if err != nil {
		t.Fatalf("parse midday: %v", err)
	}

	if !midday.IsMidday() || midday.Midday().Hour() != 13 {
		t.Fatalf("configured midday helpers failed")
	}

	UseMonthsOverflow(false)
	january, err := Parse("2024-01-31")

	if err != nil {
		t.Fatalf("parse january: %v", err)
	}

	assertEqual(t, "configured AddMonths().DateString()", january.AddMonths(1).DateString(), "2024-02-29")
	UseYearsOverflow(false)
	leap, err := Parse("2024-02-29")

	if err != nil {
		t.Fatalf("parse leap day: %v", err)
	}

	assertEqual(t, "configured AddYears().DateString()", leap.AddYears(1).DateString(), "2025-02-28")

	UseStrictMode(false)

	if IsStrictModeEnabled() {
		t.Fatalf("IsStrictModeEnabled() = true, want false")
	}

	SetHumanDiffOptions(HumanDiffOptions{Locale: "en-US", Numeric: "auto", Style: "long"})

	if got := GetHumanDiffOptions().Numeric; got != "auto" {
		t.Fatalf("GetHumanDiffOptions().Numeric = %q, want auto", got)
	}

	frozen, err := Parse("2025-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse frozen now: %v", err)
	}

	SetTestNowAndTimezone(&frozen, "Asia/Tokyo")

	if !HasTestNow() {
		t.Fatalf("HasTestNow() = false, want true")
	}

	now, err := Now()

	if err != nil {
		t.Fatalf("Now(): %v", err)
	}

	if got := now.Timezone(); got != "Asia/Tokyo" {
		t.Fatalf("Now().Timezone() = %q, want Asia/Tokyo", got)
	}

	assertEqual(t, "configured Now().DateTimeString()", now.DateTimeString(), "2025-01-01 09:00:00")
}

func TestInjectedConfigControlsNowAndFactoryClock(t *testing.T) {
	frozen, err := Parse("2025-02-03T04:05:06Z")

	if err != nil {
		t.Fatalf("parse frozen config time: %v", err)
	}

	cfg := NewConfig(Settings{
		Locale:         "en-US",
		FallbackLocale: "en-US",
		MonthsOverflow: true,
		StrictMode:     true,
		TestNow:        &frozen,
		Timezone:       "Asia/Tokyo",
		WeekendDays:    []time.Weekday{time.Sunday, time.Saturday},
		YearsOverflow:  true,
	})

	now, err := Now(WithConfig(cfg))

	if err != nil {
		t.Fatalf("now with config: %v", err)
	}

	assertEqual(t, "Now(WithConfig).DateTimeString()", now.DateTimeString(), "2025-02-03 13:05:06")

	factory, err := NewFactory(WithConfig(cfg))

	if err != nil {
		t.Fatalf("factory with config: %v", err)
	}

	assertEqual(t, "Factory WithConfig Now", factory.Now().DateTimeString(), "2025-02-03 13:05:06")
}

func TestCreateTimezoneAwareComponents(t *testing.T) {
	tokyo, err := Create(Components{
		Year:        2024,
		Month:       1,
		Day:         1,
		Hour:        9,
		Minute:      30,
		Second:      15,
		Millisecond: 456,
		Timezone:    "Asia/Tokyo",
	})

	if err != nil {
		t.Fatalf("create tokyo tempo: %v", err)
	}

	if got := tokyo.ISOString(); got != "2024-01-01T00:30:15.456Z" {
		t.Fatalf("ISOString() = %q, want timezone converted instant", got)
	}

	if got := tokyo.DateTimeString(); got != "2024-01-01 09:30:15" {
		t.Fatalf("DateTimeString() = %q, want local datetime", got)
	}

	if got := tokyo.OffsetMinutes(); got != 540 {
		t.Fatalf("OffsetMinutes() = %d, want 540", got)
	}

	if got := tokyo.ToArray(); got != [7]int{2024, 1, 1, 9, 30, 15, 456} {
		t.Fatalf("ToArray() = %#v, want local parts", got)
	}

	object := tokyo.ToObject()

	if object.Timezone != "Asia/Tokyo" || object.Weekday != 1 {
		t.Fatalf("ToObject() = %#v, want Tokyo Monday object", object)
	}

	normalized, err := Create(Components{Year: 2024, Month: 2, Day: 31})

	if err != nil {
		t.Fatalf("create normalized date: %v", err)
	}

	if got := normalized.DateString(); got != "2024-03-02" {
		t.Fatalf("Create(invalid date).DateString() = %q, want normalized date", got)
	}

	if _, err := CreateSafe(Components{Year: 2024, Month: 2, Day: 31}); err == nil {
		t.Fatalf("CreateSafe(invalid date) error = nil, want error")
	}

	safe, err := CreateSafe(Components{Year: 2024, Month: 2, Day: 29, Timezone: "Asia/Tokyo"})

	if err != nil {
		t.Fatalf("create safe date: %v", err)
	}

	if got := safe.ISOString(); got != "2024-02-28T15:00:00.000Z" {
		t.Fatalf("CreateSafe(valid Tokyo date).ISOString() = %q, want timezone instant", got)
	}
}

func TestCompareDiffRoundAndFormat(t *testing.T) {
	base, err := Parse("2024-05-15T10:34:45.600Z")

	if err != nil {
		t.Fatalf("parse base: %v", err)
	}

	earlier, err := Parse("2024-05-14T08:00:00Z")

	if err != nil {
		t.Fatalf("parse earlier: %v", err)
	}

	sameDay, err := Parse("2024-05-15T23:59:00Z")

	if err != nil {
		t.Fatalf("parse same day: %v", err)
	}

	end, err := Parse("2024-05-16T00:00:00Z")

	if err != nil {
		t.Fatalf("parse end: %v", err)
	}

	if !base.After(earlier) {
		t.Fatalf("After() = false, want true")
	}

	if !base.Same(sameDay, Day) {
		t.Fatalf("Same(Day) = false, want true")
	}

	if !base.Is(sameDay, Day) {
		t.Fatalf("Is(Day) = false, want true")
	}

	if !base.IsCurrentUnit(Day, sameDay) {
		t.Fatalf("IsCurrentUnit(Day) = false, want true")
	}

	if !base.Between(earlier, end) {
		t.Fatalf("Between() = false, want true")
	}

	if !base.Between(base, end) {
		t.Fatalf("Between(boundary) = false, want true")
	}

	if !base.BetweenIncluded(end, base) {
		t.Fatalf("BetweenIncluded(reversed boundary) = false, want true")
	}

	if base.BetweenExcluded(base, end) {
		t.Fatalf("BetweenExcluded(boundary) = true, want false")
	}

	if got := base.DiffInHours(earlier); got != 26 {
		t.Fatalf("DiffInHours() = %d, want 26", got)
	}

	if got := base.DiffInMinutes(earlier); got != 1594 {
		t.Fatalf("DiffInMinutes() = %d, want 1594", got)
	}

	diffDuration := base.DiffAsDuration(earlier)

	if got := diffDuration.ISOString(); got != "P1DT2H34M45.600S" {
		t.Fatalf("DiffAsDuration().ISOString() = %q, want duration diff", got)
	}

	absoluteDiffDuration := base.DiffAsDuration(earlier, DiffOptions{Absolute: true})

	if got := absoluteDiffDuration.ISOString(); got != "P1DT2H34M45.600S" {
		t.Fatalf("DiffAsDuration(absolute).ISOString() = %q, want duration diff", got)
	}

	dateIntervalDuration := base.DiffAsDateInterval(earlier)

	if got := dateIntervalDuration.ISOString(); got != "P1DT2H34M45.600S" {
		t.Fatalf("DiffAsDateInterval().ISOString() = %q, want duration diff", got)
	}

	tempoIntervalDuration := base.DiffAsTempoInterval(earlier)

	if got := tempoIntervalDuration.ISOString(); got != "P1DT2H34M45.600S" {
		t.Fatalf("DiffAsTempoInterval().ISOString() = %q, want duration diff", got)
	}

	if got := base.DiffInMicroseconds(earlier); got != 95685600000 {
		t.Fatalf("DiffInMicroseconds() = %d, want microsecond diff", got)
	}

	if got := base.DiffFiltered(earlier, func(item Tempo) bool { return item.IsWeekday() }); got != 1 {
		t.Fatalf("DiffFiltered(weekday) = %d, want 1", got)
	}

	if got := base.GetPreciseTimestamp(); got != 1715769285600000 {
		t.Fatalf("GetPreciseTimestamp() = %f, want microsecond timestamp", got)
	}

	if got := base.GetPreciseTimestamp(3); got != 1715769285600 {
		t.Fatalf("GetPreciseTimestamp(3) = %f, want millisecond timestamp", got)
	}

	assertEqual(t, "Calendar(tomorrow)", base.Calendar(earlier), "Tomorrow at 10:34")
	reference, err := Parse("2024-05-20T00:00:00Z")

	if err != nil {
		t.Fatalf("parse calendar reference: %v", err)
	}

	assertEqual(t, "Calendar(last week)", base.Calendar(reference), "Last Wednesday at 10:34")

	if got := base.Floor(Hour).ISOString(); got != "2024-05-15T10:00:00.000Z" {
		t.Fatalf("Floor(Hour).ISOString() = %q, want hour floor", got)
	}

	if got := base.Ceil(Hour).ISOString(); got != "2024-05-15T11:00:00.000Z" {
		t.Fatalf("Ceil(Hour).ISOString() = %q, want hour ceil", got)
	}

	if got := base.Round(Hour).ISOString(); got != "2024-05-15T11:00:00.000Z" {
		t.Fatalf("Round(Hour).ISOString() = %q, want hour round", got)
	}

	if got := base.FloorUnit(Hour).ISOString(); got != "2024-05-15T10:00:00.000Z" {
		t.Fatalf("FloorUnit(Hour).ISOString() = %q, want hour floor", got)
	}

	if got := base.CeilUnit(Hour).ISOString(); got != "2024-05-15T11:00:00.000Z" {
		t.Fatalf("CeilUnit(Hour).ISOString() = %q, want hour ceil", got)
	}

	if got := base.RoundUnit(Hour).ISOString(); got != "2024-05-15T11:00:00.000Z" {
		t.Fatalf("RoundUnit(Hour).ISOString() = %q, want hour round", got)
	}

	if got := base.FloorWeek().DateString(); got != "2024-05-13" {
		t.Fatalf("FloorWeek().DateString() = %q, want week start", got)
	}

	if got := base.CeilWeek().DateString(); got != "2024-05-20" {
		t.Fatalf("CeilWeek().DateString() = %q, want next week start", got)
	}

	if got := base.RoundWeek().DateString(); got != "2024-05-13" {
		t.Fatalf("RoundWeek().DateString() = %q, want week start", got)
	}

	if got := base.Format("YYYY-MM-DD HH:mm:ss.SSS ZZ [Q]M"); got != "2024-05-15 10:34:45.600 +0000 Q5" {
		t.Fatalf("Format() = %q, want token output", got)
	}

	if got := base.RawFormat("YYYY-MM-DD"); got != "2024-05-15" {
		t.Fatalf("RawFormat() = %q, want date output", got)
	}

	if got := base.Format("dddd, MMMM Do YYYY"); got != "Wednesday, May 15th 2024" {
		t.Fatalf("Format() = %q, want long date output", got)
	}

	if got := base.Ordinal(Day); got != "15th" {
		t.Fatalf("Ordinal(Day) = %q, want 15th", got)
	}

	if got := base.Ordinal(Month); got != "5th" {
		t.Fatalf("Ordinal(Month) = %q, want 5th", got)
	}

	if got := base.Meridiem(false); got != "AM" {
		t.Fatalf("Meridiem(false) = %q, want AM", got)
	}

	if got := base.AddHours(2).Meridiem(true); got != "pm" {
		t.Fatalf("Meridiem(true) = %q, want pm", got)
	}

	if got := base.Week(); got != base.ISOWeekNumber() {
		t.Fatalf("Week() = %d, want ISO week number", got)
	}

	if got := base.WeekYear(); got != base.ISOWeekYear() {
		t.Fatalf("WeekYear() = %d, want ISO week year", got)
	}

	if got := base.WeeksInYear(); got != base.WeeksInISOYear() {
		t.Fatalf("WeeksInYear() = %d, want ISO weeks in year", got)
	}

	if got := base.GetDaysFromStartOfWeek(time.Monday); got != 2 {
		t.Fatalf("GetDaysFromStartOfWeek(Monday) = %d, want 2", got)
	}

	if got := base.SetDaysFromStartOfWeek(4, time.Monday).DateString(); got != "2024-05-17" {
		t.Fatalf("SetDaysFromStartOfWeek().DateString() = %q, want Friday", got)
	}

	longYear, err := Parse("2015-06-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse long year: %v", err)
	}

	if !longYear.IsLongISOYear() {
		t.Fatalf("IsLongISOYear() = false, want true")
	}

	if got := base.MonthName(); got != "May" {
		t.Fatalf("MonthName() = %q, want May", got)
	}

	if got := base.ShortMonthName(); got != "May" {
		t.Fatalf("ShortMonthName() = %q, want May", got)
	}

	if got := base.DayName(); got != "Wednesday" {
		t.Fatalf("DayName() = %q, want Wednesday", got)
	}

	if got := base.ShortDayName(); got != "Wed" {
		t.Fatalf("ShortDayName() = %q, want Wed", got)
	}
}

func TestIntervalsPeriodsAndMutableTempo(t *testing.T) {
	start, err := Parse("2024-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start: %v", err)
	}

	end, err := Parse("2024-01-03T12:00:00Z")

	if err != nil {
		t.Fatalf("parse end: %v", err)
	}

	interval := start.IntervalUntil(end)

	if got := interval.Days(); got != 2 {
		t.Fatalf("Interval.Days() = %d, want 2", got)
	}

	if got := interval.Hours(); got != 60 {
		t.Fatalf("Interval.Hours() = %d, want 60", got)
	}

	if !interval.Contains(end) {
		t.Fatalf("Interval.Contains(end) = false, want true")
	}

	if interval.Contains(end, "[)") {
		t.Fatalf("Interval.Contains(end, \"[)\") = true, want false")
	}

	inverted := end.IntervalUntil(start)

	if !inverted.Inverted() {
		t.Fatalf("Interval.Inverted() = false, want true")
	}

	if got := inverted.Hours(); got != -60 {
		t.Fatalf("inverted Interval.Hours() = %d, want -60", got)
	}

	if got := inverted.Invert().Hours(); got != 60 {
		t.Fatalf("Interval.Invert().Hours() = %d, want 60", got)
	}

	if got := inverted.Abs().Start.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("Interval.Abs().Start = %q, want normalized start", got)
	}

	if got := inverted.Abs().End.ISOString(); got != "2024-01-03T12:00:00.000Z" {
		t.Fatalf("Interval.Abs().End = %q, want normalized end", got)
	}

	calendarStart, err := Parse("2023-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse calendar start: %v", err)
	}

	calendarEnd, err := Parse("2024-03-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse calendar end: %v", err)
	}

	calendarInterval := calendarStart.IntervalUntil(calendarEnd)

	if got := calendarInterval.Weeks(); got != 60 {
		t.Fatalf("Interval.Weeks() = %d, want 60", got)
	}

	if got := calendarInterval.Months(); got != 14 {
		t.Fatalf("Interval.Months() = %d, want 14", got)
	}

	if got := calendarInterval.Years(); got != 1 {
		t.Fatalf("Interval.Years() = %d, want 1", got)
	}

	periodEnd, err := Parse("2024-01-05")

	if err != nil {
		t.Fatalf("parse period end: %v", err)
	}

	values, err := start.PeriodUntil(periodEnd, PeriodOptions{
		Step:       Duration{Days: 2},
		IncludeEnd: true,
	}).Values()

	if err != nil {
		t.Fatalf("period values: %v", err)
	}

	gotDates := make([]string, 0, len(values))

	for _, item := range values {
		gotDates = append(gotDates, item.DateString())
	}

	wantDates := []string{"2024-01-01", "2024-01-03", "2024-01-05"}

	if strings.Join(gotDates, ",") != strings.Join(wantDates, ",") {
		t.Fatalf("Period.Values() = %#v, want %#v", gotDates, wantDates)
	}

	if got, err := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).Count(); err != nil || got != 3 {
		t.Fatalf("Period.Count() = %d, %v, want 3, nil", got, err)
	}

	first, ok, err := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).First()

	if err != nil || !ok || first.DateString() != "2024-01-01" {
		t.Fatalf("Period.First() = %q, %v, %v, want 2024-01-01, true, nil", first.DateString(), ok, err)
	}

	last, ok, err := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).Last()

	if err != nil || !ok || last.DateString() != "2024-01-05" {
		t.Fatalf("Period.Last() = %q, %v, %v, want 2024-01-05, true, nil", last.DateString(), ok, err)
	}

	contains, err := Parse("2024-01-04")

	if err != nil {
		t.Fatalf("parse period contains input: %v", err)
	}

	if !start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).Contains(contains) {
		t.Fatalf("Period.Contains(inside) = false, want true")
	}

	outside, err := Parse("2024-01-06")

	if err != nil {
		t.Fatalf("parse period outside input: %v", err)
	}

	if start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).Contains(outside) {
		t.Fatalf("Period.Contains(outside) = true, want false")
	}

	if empty, err := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).IsEmpty(); err != nil || empty {
		t.Fatalf("Period.IsEmpty() = %v, %v, want false, nil", empty, err)
	}

	filtered, err := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).Filter(func(value Tempo, _ int) bool {
		return value.Day() != 3
	})

	if err != nil {
		t.Fatalf("Period.Filter(): %v", err)
	}

	filteredDates := make([]string, 0, len(filtered))

	for _, item := range filtered {
		filteredDates = append(filteredDates, item.DateString())
	}

	wantFilteredDates := []string{"2024-01-01", "2024-01-05"}

	if strings.Join(filteredDates, ",") != strings.Join(wantFilteredDates, ",") {
		t.Fatalf("Period.Filter() = %#v, want %#v", filteredDates, wantFilteredDates)
	}

	mapped, err := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).Map(func(value Tempo, index int) Tempo {
		return value.AddDays(index)
	})

	if err != nil {
		t.Fatalf("Period.Map(): %v", err)
	}

	mappedDates := make([]string, 0, len(mapped))

	for _, item := range mapped {
		mappedDates = append(mappedDates, item.DateString())
	}

	wantMappedDates := []string{"2024-01-01", "2024-01-04", "2024-01-07"}

	if strings.Join(mappedDates, ",") != strings.Join(wantMappedDates, ",") {
		t.Fatalf("Period.Map() = %#v, want %#v", mappedDates, wantMappedDates)
	}

	everyCount, err := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).Every(Duration{Days: 1}).Count()

	if err != nil || everyCount != 5 {
		t.Fatalf("Period.Every().Count() = %d, %v, want 5, nil", everyCount, err)
	}

	periodDuration := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).ToDuration()

	if got := periodDuration.ISOString(); got != "P4D" {
		t.Fatalf("Period.ToDuration().ISOString() = %q, want P4D", got)
	}

	openPeriod := start.PeriodUntil(periodEnd, PeriodOptions{
		Step:       Duration{Days: 2},
		ExcludeEnd: true,
	})
	openValues, err := openPeriod.Values()

	if err != nil {
		t.Fatalf("open period values: %v", err)
	}

	openDates := make([]string, 0, len(openValues))

	for _, item := range openValues {
		openDates = append(openDates, item.DateString())
	}

	wantOpenDates := []string{"2024-01-01", "2024-01-03"}

	if strings.Join(openDates, ",") != strings.Join(wantOpenDates, ",") {
		t.Fatalf("open Period.Values() = %#v, want %#v", openDates, wantOpenDates)
	}

	if openPeriod.Contains(periodEnd) {
		t.Fatalf("open Period.Contains(end) = true, want false")
	}

	reverseValues, err := periodEnd.PeriodUntil(start, PeriodOptions{
		Step: Duration{Days: -2},
	}).Values()

	if err != nil {
		t.Fatalf("reverse period values: %v", err)
	}

	reverseDates := make([]string, 0, len(reverseValues))

	for _, item := range reverseValues {
		reverseDates = append(reverseDates, item.DateString())
	}

	wantReverseDates := []string{"2024-01-05", "2024-01-03", "2024-01-01"}

	if strings.Join(reverseDates, ",") != strings.Join(wantReverseDates, ",") {
		t.Fatalf("reverse Period.Values() = %#v, want %#v", reverseDates, wantReverseDates)
	}

	if _, err := periodEnd.PeriodUntil(start, PeriodOptions{Step: Duration{Days: 1}}).Values(); err == nil {
		t.Fatalf("wrong-direction Period.Values() error = nil, want error")
	}

	mutable, err := ParseMutable("2024-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse mutable: %v", err)
	}

	if mutable.AddHours(5).StartOf(Day) != mutable {
		t.Fatalf("mutable chaining returned a different pointer")
	}

	if got := mutable.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("Mutable ISOString() = %q, want start of day", got)
	}

	if _, err := mutable.SetTimezone("Asia/Tokyo"); err != nil {
		t.Fatalf("mutable set timezone: %v", err)
	}

	if got := mutable.DateTimeString(); got != "2024-01-01 09:00:00" {
		t.Fatalf("Mutable DateTimeString() = %q, want Tokyo local time", got)
	}

	if mutable.SetTime(0, 0, 0, 0) != mutable {
		t.Fatalf("mutable SetTime returned a different pointer")
	}

	if got := mutable.TimeString(MillisecondPrecision); got != "00:00:00.000" {
		t.Fatalf("Mutable TimeString(ms) = %q, want midnight", got)
	}

	mutable.AddDuration(Duration{Days: 2, Hours: 3}).SubHours(1).EndOf(Day)

	if got := mutable.DateTimeString(); got != "2024-01-03 23:59:59" {
		t.Fatalf("Mutable chained mutations DateTimeString() = %q, want end of day", got)
	}

	if !mutable.IsEndOf(Day) {
		t.Fatalf("Mutable IsEndOf(Day) = false, want true")
	}

	if got := mutable.RFC3339String(); got != "2024-01-03T23:59:59+09:00" {
		t.Fatalf("Mutable RFC3339String() = %q, want local offset output", got)
	}

	compare, err := Parse("2024-01-04T00:00:00+09:00", WithTimezone("Asia/Tokyo"))

	if err != nil {
		t.Fatalf("parse mutable compare: %v", err)
	}

	if !mutable.Before(compare) {
		t.Fatalf("Mutable Before(compare) = false, want true")
	}

	if got := mutable.DiffInHours(compare); got != 0 {
		t.Fatalf("Mutable DiffInHours(compare) = %d, want truncated zero", got)
	}

	mutableInterval := mutable.IntervalUntil(compare)

	if !mutableInterval.Contains(compare) {
		t.Fatalf("Mutable IntervalUntil(compare).Contains(compare) = false, want true")
	}

	period := mutable.StartOfDay().PeriodUntil(compare, PeriodOptions{Step: Duration{Hours: 12}})
	count, err := period.Count()

	if err != nil {
		t.Fatalf("mutable period count: %v", err)
	}

	if count != 3 {
		t.Fatalf("Mutable PeriodUntil().Count() = %d, want 3", count)
	}

	minimum, err := Parse("2024-01-02T00:00:00+09:00", WithTimezone("Asia/Tokyo"))

	if err != nil {
		t.Fatalf("parse mutable minimum: %v", err)
	}

	maximum, err := Parse("2024-01-02T12:00:00+09:00", WithTimezone("Asia/Tokyo"))

	if err != nil {
		t.Fatalf("parse mutable maximum: %v", err)
	}

	if _, err := mutable.Clamp(minimum, maximum); err != nil {
		t.Fatalf("mutable clamp: %v", err)
	}

	if got := mutable.DateTimeString(); got != "2024-01-02 12:00:00" {
		t.Fatalf("Mutable Clamp() DateTimeString() = %q, want maximum", got)
	}

	mutable.Average(compare)

	if got := mutable.DateTimeString(); got != "2024-01-03 06:00:00" {
		t.Fatalf("Mutable Average() DateTimeString() = %q, want midpoint", got)
	}
}

func TestTimezoneNamesOffsetsAndDSTState(t *testing.T) {
	utc, err := Parse("2024-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse utc: %v", err)
	}

	winter, err := Parse("2024-01-15T12:00:00Z", WithTimezone("America/New_York"))

	if err != nil {
		t.Fatalf("parse winter: %v", err)
	}

	summer, err := Parse("2024-07-15T12:00:00Z", WithTimezone("America/New_York"))

	if err != nil {
		t.Fatalf("parse summer: %v", err)
	}

	if !utc.IsUTC() {
		t.Fatalf("IsUTC() = false, want true")
	}

	if got := utc.OffsetString(":"); got != "+00:00" {
		t.Fatalf("OffsetString(\":\") = %q, want +00:00", got)
	}

	if got := utc.OffsetString(""); got != "+0000" {
		t.Fatalf("OffsetString(\"\") = %q, want +0000", got)
	}

	if got := utc.GetOffsetString(":"); got != "+00:00" {
		t.Fatalf("GetOffsetString(\":\") = %q, want +00:00", got)
	}

	if got := utc.UTCOffset(); got != 0 {
		t.Fatalf("UTCOffset() = %d, want 0", got)
	}

	if got := utc.ZoneName(); got != "UTC" {
		t.Fatalf("ZoneName() = %q, want UTC", got)
	}

	if winter.IsUTC() {
		t.Fatalf("winter IsUTC() = true, want false")
	}

	if got := winter.OffsetMinutes(); got != -300 {
		t.Fatalf("winter OffsetMinutes() = %d, want -300", got)
	}

	if got := winter.OffsetString(":"); got != "-05:00" {
		t.Fatalf("winter OffsetString() = %q, want -05:00", got)
	}

	if got := winter.GetOffsetString(""); got != "-0500" {
		t.Fatalf("winter GetOffsetString() = %q, want -0500", got)
	}

	if got := winter.UTCOffset(); got != -300 {
		t.Fatalf("winter UTCOffset() = %d, want -300", got)
	}

	if got := winter.ZoneName(); got != "EST" {
		t.Fatalf("winter ZoneName() = %q, want EST", got)
	}

	if winter.IsDST() {
		t.Fatalf("winter IsDST() = true, want false")
	}

	if got := summer.OffsetMinutes(); got != -240 {
		t.Fatalf("summer OffsetMinutes() = %d, want -240", got)
	}

	if got := summer.OffsetString(":"); got != "-04:00" {
		t.Fatalf("summer OffsetString() = %q, want -04:00", got)
	}

	if got := summer.ZoneName(); got != "EDT" {
		t.Fatalf("summer ZoneName() = %q, want EDT", got)
	}

	if !summer.IsDST() {
		t.Fatalf("summer IsDST() = false, want true")
	}
}

func TestRangeClampAverageSelectionAndBoundaryPredicates(t *testing.T) {
	base, err := Parse("2024-05-15T12:00:00Z")

	if err != nil {
		t.Fatalf("parse base: %v", err)
	}

	minimum, err := Parse("2024-05-10T00:00:00Z")

	if err != nil {
		t.Fatalf("parse minimum: %v", err)
	}

	maximum, err := Parse("2024-05-20T00:00:00Z")

	if err != nil {
		t.Fatalf("parse maximum: %v", err)
	}

	before, err := Parse("2024-05-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse before: %v", err)
	}

	after, err := Parse("2024-05-30T00:00:00Z")

	if err != nil {
		t.Fatalf("parse after: %v", err)
	}

	clampedBefore, err := before.Clamp(minimum, maximum)

	if err != nil {
		t.Fatalf("clamp before: %v", err)
	}

	if got := clampedBefore.ISOString(); got != "2024-05-10T00:00:00.000Z" {
		t.Fatalf("Clamp(before).ISOString() = %q, want minimum", got)
	}

	clampedBase, err := base.Clamp(minimum, maximum)

	if err != nil {
		t.Fatalf("clamp base: %v", err)
	}

	if got := clampedBase.ISOString(); got != "2024-05-15T12:00:00.000Z" {
		t.Fatalf("Clamp(base).ISOString() = %q, want original", got)
	}

	clampedAfter, err := after.Clamp(minimum, maximum)

	if err != nil {
		t.Fatalf("clamp after: %v", err)
	}

	if got := clampedAfter.ISOString(); got != "2024-05-20T00:00:00.000Z" {
		t.Fatalf("Clamp(after).ISOString() = %q, want maximum", got)
	}

	averageEnd, err := Parse("2024-05-17T12:00:00Z")

	if err != nil {
		t.Fatalf("parse average end: %v", err)
	}

	if got := base.Average(averageEnd).ISOString(); got != "2024-05-16T12:00:00.000Z" {
		t.Fatalf("Average().ISOString() = %q, want midpoint", got)
	}

	staticStart, err := Parse("2024-05-15T00:00:00Z")

	if err != nil {
		t.Fatalf("parse static start: %v", err)
	}

	staticEnd, err := Parse("2024-05-17T00:00:00Z")

	if err != nil {
		t.Fatalf("parse static end: %v", err)
	}

	if got := Average(staticStart, staticEnd).ISOString(); got != "2024-05-16T00:00:00.000Z" {
		t.Fatalf("Average(start,end).ISOString() = %q, want midpoint", got)
	}

	if got := Minimum(staticStart, staticEnd).ISOString(); got != "2024-05-15T00:00:00.000Z" {
		t.Fatalf("Minimum(start,end).ISOString() = %q, want earliest", got)
	}

	if got := Maximum(staticStart, staticEnd).ISOString(); got != "2024-05-17T00:00:00.000Z" {
		t.Fatalf("Maximum(start,end).ISOString() = %q, want latest", got)
	}

	if got := base.Minimum(minimum).ISOString(); got != "2024-05-10T00:00:00.000Z" {
		t.Fatalf("Minimum().ISOString() = %q, want minimum", got)
	}

	if got := base.Maximum(maximum).ISOString(); got != "2024-05-20T00:00:00.000Z" {
		t.Fatalf("Maximum().ISOString() = %q, want maximum", got)
	}

	closestA, err := Parse("2024-05-10T00:00:00Z")

	if err != nil {
		t.Fatalf("parse closest A: %v", err)
	}

	closestB, err := Parse("2024-05-16T00:00:00Z")

	if err != nil {
		t.Fatalf("parse closest B: %v", err)
	}

	closestC, err := Parse("2024-05-20T00:00:00Z")

	if err != nil {
		t.Fatalf("parse closest C: %v", err)
	}

	if got := base.Closest(closestA, closestB, closestC).ISOString(); got != "2024-05-16T00:00:00.000Z" {
		t.Fatalf("Closest().ISOString() = %q, want nearest", got)
	}

	farthest, err := Parse("2024-05-22T00:00:00Z")

	if err != nil {
		t.Fatalf("parse farthest: %v", err)
	}

	if got := base.Farthest(closestA, farthest).ISOString(); got != "2024-05-22T00:00:00.000Z" {
		t.Fatalf("Farthest().ISOString() = %q, want farthest", got)
	}

	startOfDay, err := Parse("2024-05-15T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start of day: %v", err)
	}

	notStartOfDay, err := Parse("2024-05-15T00:00:01Z")

	if err != nil {
		t.Fatalf("parse not start of day: %v", err)
	}

	endOfDay, err := Parse("2024-05-15T23:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of day: %v", err)
	}

	notEndOfDay, err := Parse("2024-05-15T23:59:59.998Z")

	if err != nil {
		t.Fatalf("parse not end of day: %v", err)
	}

	if !startOfDay.IsStartOf(Day) {
		t.Fatalf("IsStartOf(Day) = false, want true")
	}

	if !startOfDay.IsStartOfDay() {
		t.Fatalf("IsStartOfDay() = false, want true")
	}

	if notStartOfDay.IsStartOf(Day) {
		t.Fatalf("IsStartOf(Day) = true, want false")
	}

	if !endOfDay.IsEndOf(Day) {
		t.Fatalf("IsEndOf(Day) = false, want true")
	}

	if !endOfDay.IsEndOfDay() {
		t.Fatalf("IsEndOfDay() = false, want true")
	}

	if notEndOfDay.IsEndOf(Day) {
		t.Fatalf("IsEndOf(Day) = true, want false")
	}

	if got := base.StartOfMillisecond().ISOString(); got != "2024-05-15T12:00:00.000Z" {
		t.Fatalf("StartOfMillisecond().ISOString() = %q, want current millisecond", got)
	}

	if got := base.EndOfMillisecond().ISOString(); got != "2024-05-15T12:00:00.000Z" {
		t.Fatalf("EndOfMillisecond().ISOString() = %q, want current millisecond", got)
	}

	if !base.IsStartOfMillisecond() || !base.IsEndOfMillisecond() {
		t.Fatalf("millisecond boundary helpers = false, want true")
	}

	secondBoundary, err := Parse("2024-05-15T12:34:56.789Z")

	if err != nil {
		t.Fatalf("parse second boundary: %v", err)
	}

	if got := secondBoundary.StartOfSecond().ISOString(); got != "2024-05-15T12:34:56.000Z" {
		t.Fatalf("StartOfSecond().ISOString() = %q, want second start", got)
	}

	if got := secondBoundary.EndOfSecond().ISOString(); got != "2024-05-15T12:34:56.999Z" {
		t.Fatalf("EndOfSecond().ISOString() = %q, want second end", got)
	}

	startOfMinute, err := Parse("2024-05-15T12:34:00Z")

	if err != nil {
		t.Fatalf("parse start of minute: %v", err)
	}

	endOfMinute, err := Parse("2024-05-15T12:34:59.999Z")

	if err != nil {
		t.Fatalf("parse end of minute: %v", err)
	}

	startOfHour, err := Parse("2024-05-15T12:00:00Z")

	if err != nil {
		t.Fatalf("parse start of hour: %v", err)
	}

	endOfHour, err := Parse("2024-05-15T12:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of hour: %v", err)
	}

	if !startOfMinute.IsStartOfMinute() || !endOfMinute.IsEndOfMinute() {
		t.Fatalf("minute boundary helpers = false, want true")
	}

	if !startOfHour.IsStartOfHour() || !endOfHour.IsEndOfHour() {
		t.Fatalf("hour boundary helpers = false, want true")
	}

	startOfWeek, err := Parse("2024-05-13T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start of week: %v", err)
	}

	endOfWeek, err := Parse("2024-05-19T23:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of week: %v", err)
	}

	startOfMonth, err := Parse("2024-05-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start of month: %v", err)
	}

	endOfMonth, err := Parse("2024-05-31T23:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of month: %v", err)
	}

	startOfQuarter, err := Parse("2024-04-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start of quarter: %v", err)
	}

	endOfQuarter, err := Parse("2024-06-30T23:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of quarter: %v", err)
	}

	startOfYear, err := Parse("2024-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start of year: %v", err)
	}

	endOfYear, err := Parse("2024-12-31T23:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of year: %v", err)
	}

	if !startOfWeek.IsStartOfWeek() || !endOfWeek.IsEndOfWeek() {
		t.Fatalf("week boundary helpers = false, want true")
	}

	if !startOfMonth.IsStartOfMonth() || !endOfMonth.IsEndOfMonth() {
		t.Fatalf("month boundary helpers = false, want true")
	}

	if !startOfQuarter.IsStartOfQuarter() || !endOfQuarter.IsEndOfQuarter() {
		t.Fatalf("quarter boundary helpers = false, want true")
	}

	if !startOfYear.IsStartOfYear() || !endOfYear.IsEndOfYear() {
		t.Fatalf("year boundary helpers = false, want true")
	}

	if got := base.StartOfDecade().DateTimeString(); got != "2020-01-01 00:00:00" {
		t.Fatalf("StartOfDecade().DateTimeString() = %q, want decade start", got)
	}

	if got := base.EndOfDecade().DateTimeString(); got != "2029-12-31 23:59:59" {
		t.Fatalf("EndOfDecade().DateTimeString() = %q, want decade end", got)
	}

	if got := base.StartOfCentury().DateTimeString(); got != "2001-01-01 00:00:00" {
		t.Fatalf("StartOfCentury().DateTimeString() = %q, want century start", got)
	}

	if got := base.EndOfCentury().DateTimeString(); got != "2100-12-31 23:59:59" {
		t.Fatalf("EndOfCentury().DateTimeString() = %q, want century end", got)
	}

	if got := base.StartOfMillennium().DateTimeString(); got != "2001-01-01 00:00:00" {
		t.Fatalf("StartOfMillennium().DateTimeString() = %q, want millennium start", got)
	}

	if got := base.EndOfMillennium().DateTimeString(); got != "3000-12-31 23:59:59" {
		t.Fatalf("EndOfMillennium().DateTimeString() = %q, want millennium end", got)
	}

	startOfDecade, err := Parse("2020-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start of decade: %v", err)
	}

	endOfDecade, err := Parse("2029-12-31T23:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of decade: %v", err)
	}

	startOfCentury, err := Parse("2001-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start of century: %v", err)
	}

	endOfCentury, err := Parse("2100-12-31T23:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of century: %v", err)
	}

	startOfMillennium, err := Parse("2001-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse start of millennium: %v", err)
	}

	endOfMillennium, err := Parse("3000-12-31T23:59:59.999Z")

	if err != nil {
		t.Fatalf("parse end of millennium: %v", err)
	}

	if !startOfDecade.IsStartOfDecade() || !endOfDecade.IsEndOfDecade() {
		t.Fatalf("decade boundary helpers = false, want true")
	}

	if !startOfDecade.IsStartOfUnit(Decade) || !endOfDecade.IsEndOfUnit(Decade) {
		t.Fatalf("decade unit boundary aliases = false, want true")
	}

	if !startOfCentury.IsStartOfCentury() || !endOfCentury.IsEndOfCentury() {
		t.Fatalf("century boundary helpers = false, want true")
	}

	if !startOfMillennium.IsStartOfMillennium() || !endOfMillennium.IsEndOfMillennium() {
		t.Fatalf("millennium boundary helpers = false, want true")
	}

	if !(Tempo{value: time.UnixMilli(-8640000000000000), location: time.UTC}).IsStartOfTime() {
		t.Fatalf("IsStartOfTime() = false, want true")
	}

	if !(Tempo{value: time.UnixMilli(8640000000000000), location: time.UTC}).IsEndOfTime() {
		t.Fatalf("IsEndOfTime() = false, want true")
	}
}

func TestParityAliases(t *testing.T) {
	monday, err := Parse("2024-01-01T00:00:00Z", WithTimezone("UTC"))

	if err != nil {
		t.Fatalf("parse monday: %v", err)
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

	if got := monday.ISOWeeksInYear(); got != 52 {
		t.Fatalf("ISOWeeksInYear() = %d, want 52", got)
	}

	if got := monday.DayOfYear(); got != 1 {
		t.Fatalf("DayOfYear() = %d, want 1", got)
	}

	assertEqual(t, "SetISOWeek().DateString()", monday.SetISOWeek(2).DateString(), "2024-01-08")

	if got := monday.SetISOWeekYear(2025).ISOWeekYear(); got != 2025 {
		t.Fatalf("SetISOWeekYear().ISOWeekYear() = %d, want 2025", got)
	}

	if got := monday.SetISOWeekday(7).ISOWeekday(); got != 7 {
		t.Fatalf("SetISOWeekday().ISOWeekday() = %d, want 7", got)
	}

	assertEqual(t, "SetDayOfYear().DateString()", monday.SetDayOfYear(32).DateString(), "2024-02-01")
	assertEqual(t, "SetTimestampFrom().ISOString()", monday.SetTimestampFrom(0).ISOString(), "1970-01-01T00:00:00.000Z")
}
