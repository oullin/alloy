package tempo

import (
	"encoding/json"
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

type mapTranslator map[string]string

func (translator mapTranslator) Message(key string) (any, bool) {
	value, ok := translator[key]
	return value, ok
}

func (translator mapTranslator) Translate(key string, replacements map[string]string) (string, bool) {
	value, ok := translator[key]
	if !ok {
		return "", false
	}
	for name, replacement := range replacements {
		value = strings.ReplaceAll(value, ":"+name, replacement)
	}

	return value, true
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

func TestDiffInDaysUsesUnixSeconds(t *testing.T) {
	later, err := Parse("2024-01-03T00:00:00+00:00")
	if err != nil {
		t.Fatalf("parse later: %v", err)
	}

	earlier, err := Parse("2024-01-01T00:00:00+00:00")
	if err != nil {
		t.Fatalf("parse earlier: %v", err)
	}

	if got := later.DiffInDays(earlier); got != 2 {
		t.Fatalf("DiffInDays() = %d, want 2", got)
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

func TestISOStringFormatsUTC(t *testing.T) {
	parsed, err := Parse("2024-01-01T01:00:00+01:00")
	if err != nil {
		t.Fatalf("parse offset timestamp: %v", err)
	}

	if got := parsed.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("ISOString() = %q, want UTC ISO string", got)
	}
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

func TestTimezoneConversionModes(t *testing.T) {
	utc, err := Parse("2024-01-01T12:00:00Z")
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

func TestArithmeticBoundariesAndOverflowModes(t *testing.T) {
	base, err := Parse("2024-01-31T10:20:30.400Z")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}

	set, err := base.Set(Components{Day: 15, Hour: 1, Minute: 2, Second: 3, Millisecond: 4})
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

	leap, err := Parse("2024-02-29T00:00:00Z")
	if err != nil {
		t.Fatalf("parse leap: %v", err)
	}
	if got := leap.AddYearsNoOverflow(1).DateString(); got != "2025-02-28" {
		t.Fatalf("AddYearsNoOverflow(1).DateString() = %q, want clamped date", got)
	}
	if got := base.StartOf(Day).ISOString(); got != "2024-01-31T00:00:00.000Z" {
		t.Fatalf("StartOf(Day).ISOString() = %q, want start of day", got)
	}
	if got := base.StartOfWeek().DateString(); got != "2024-01-29" {
		t.Fatalf("StartOfWeek().DateString() = %q, want Monday week start", got)
	}
	if got := base.EndOfWeek(StartOfWeekOptions{WeekStartsOn: time.Sunday}).DateString(); got != "2024-02-03" {
		t.Fatalf("EndOfWeek(Sunday).DateString() = %q, want Saturday week end", got)
	}
	if got := base.EndOf(Month).ISOString(); got != "2024-01-31T23:59:59.999Z" {
		t.Fatalf("EndOf(Month).ISOString() = %q, want end of month", got)
	}
	if got := base.StartOf(Quarter).ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("StartOf(Quarter).ISOString() = %q, want quarter start", got)
	}
	if got := base.StartOfQuarter().ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("StartOfQuarter().ISOString() = %q, want quarter start", got)
	}
	if got := base.EndOfQuarter().ISOString(); got != "2024-03-31T23:59:59.999Z" {
		t.Fatalf("EndOfQuarter().ISOString() = %q, want quarter end", got)
	}
	if got := base.EndOf(Year).ISOString(); got != "2024-12-31T23:59:59.999Z" {
		t.Fatalf("EndOf(Year).ISOString() = %q, want end of year", got)
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
	if got := base.DiffAsDuration(earlier).ISOString(); got != "P1DT2H34M45.600S" {
		t.Fatalf("DiffAsDuration().ISOString() = %q, want duration diff", got)
	}
	if got := base.DiffAsDuration(earlier, DiffOptions{Absolute: true}).ISOString(); got != "P1DT2H34M45.600S" {
		t.Fatalf("DiffAsDuration(absolute).ISOString() = %q, want duration diff", got)
	}
	if got := base.DiffAsDateInterval(earlier).ISOString(); got != "P1DT2H34M45.600S" {
		t.Fatalf("DiffAsDateInterval().ISOString() = %q, want duration diff", got)
	}
	if got := base.DiffAsTempoInterval(earlier).ISOString(); got != "P1DT2H34M45.600S" {
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
	if got := start.PeriodUntil(periodEnd, PeriodOptions{Step: Duration{Days: 2}}).ToDuration().ISOString(); got != "P4D" {
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

func TestFromFormatPredicatesAndHumanDiffs(t *testing.T) {
	withOffset, err := FromFormat("2024/05/15 10:34:45.600 +0900", "YYYY/MM/DD HH:mm:ss.SSS ZZ")
	if err != nil {
		t.Fatalf("from format offset: %v", err)
	}
	if got := withOffset.ISOString(); got != "2024-05-15T01:34:45.600Z" {
		t.Fatalf("FromFormat offset ISOString() = %q, want parsed offset instant", got)
	}

	meridiem, err := FromFormat("05-15-24 10:34 PM", "MM-DD-YY hh:mm A")
	if err != nil {
		t.Fatalf("from format meridiem: %v", err)
	}
	if got := meridiem.ISOString(); got != "2024-05-15T22:34:00.000Z" {
		t.Fatalf("FromFormat meridiem ISOString() = %q, want parsed time", got)
	}

	named, err := FromFormat("Wednesday, May 15th 2024 10:34 PM", "dddd, MMMM Do YYYY hh:mm A")
	if err != nil {
		t.Fatalf("from format named month: %v", err)
	}
	if got := named.ISOString(); got != "2024-05-15T22:34:00.000Z" {
		t.Fatalf("FromFormat named month ISOString() = %q, want parsed time", got)
	}

	shortNamed, err := FromFormat("Wed, May 15 2024", "ddd, MMM D YYYY")
	if err != nil {
		t.Fatalf("from format short named month: %v", err)
	}
	if got := shortNamed.DateString(); got != "2024-05-15" {
		t.Fatalf("FromFormat short named DateString() = %q, want parsed date", got)
	}

	tokyo, err := FromFormat("2024-01-01 09:00", "YYYY-MM-DD HH:mm", WithTimezone("Asia/Tokyo"))
	if err != nil {
		t.Fatalf("from format tokyo: %v", err)
	}
	if got := tokyo.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("FromFormat timezone ISOString() = %q, want local timezone instant", got)
	}

	if !CanParse("2024-05-15T10:34:45Z") {
		t.Fatalf("CanParse(valid) = false, want true")
	}
	if CanParse("not a date") {
		t.Fatalf("CanParse(invalid) = true, want false")
	}
	if parsed, ok := TryParse("2024-05-15"); !ok || parsed.DateString() != "2024-05-15" {
		t.Fatalf("TryParse(valid) = %q, %v, want date, true", parsed.DateString(), ok)
	}
	if _, ok := TryParse("not a date"); ok {
		t.Fatalf("TryParse(invalid) ok = true, want false")
	}
	if !HasFormat("2024/05/15 10:34", "YYYY/MM/DD HH:mm") {
		t.Fatalf("HasFormat(valid) = false, want true")
	}
	if HasFormat("2024-05-15", "YYYY/MM/DD") {
		t.Fatalf("HasFormat(invalid) = true, want false")
	}
	if parsed, ok := TryFromFormat("2024/05/15 10:34", "YYYY/MM/DD HH:mm"); !ok || parsed.ISOString() != "2024-05-15T10:34:00.000Z" {
		t.Fatalf("TryFromFormat(valid) = %q, %v, want parsed instant, true", parsed.ISOString(), ok)
	}
	if _, ok := TryFromFormat("2024-05-15", "YYYY/MM/DD"); ok {
		t.Fatalf("TryFromFormat(invalid) ok = true, want false")
	}

	base, err := Parse("2024-02-29T00:00:00Z")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	saturday, err := Parse("2024-03-02T00:00:00Z")
	if err != nil {
		t.Fatalf("parse saturday: %v", err)
	}
	referenceYesterday, err := Parse("2024-02-28T12:00:00Z")
	if err != nil {
		t.Fatalf("parse yesterday reference: %v", err)
	}
	referenceTomorrow, err := Parse("2024-03-01T12:00:00Z")
	if err != nil {
		t.Fatalf("parse tomorrow reference: %v", err)
	}

	if !base.IsLeapYear() {
		t.Fatalf("IsLeapYear() = false, want true")
	}
	if got := base.DaysInYear(); got != 366 {
		t.Fatalf("DaysInYear() = %d, want 366", got)
	}
	if base.IsLongYear() {
		t.Fatalf("IsLongYear() = true, want false")
	}
	longYear, err := Parse("2020-12-31T00:00:00Z")
	if err != nil {
		t.Fatalf("parse long year: %v", err)
	}
	if !longYear.IsLongYear() {
		t.Fatalf("IsLongYear() = false, want true")
	}
	if got := base.DaysInMonth(); got != 29 {
		t.Fatalf("DaysInMonth() = %d, want 29", got)
	}
	if !base.IsLastOfMonth() {
		t.Fatalf("IsLastOfMonth() = false, want true")
	}
	if base.SubDays(1).IsLastOfMonth() {
		t.Fatalf("IsLastOfMonth(previous day) = true, want false")
	}
	if !saturday.IsWeekend() {
		t.Fatalf("IsWeekend() = false, want true")
	}
	if !base.IsWeekday() {
		t.Fatalf("IsWeekday() = false, want true")
	}
	if !base.IsDayOfWeek(time.Thursday) {
		t.Fatalf("IsDayOfWeek(Thursday) = false, want true")
	}
	if base.IsDayOfWeek(time.Friday) {
		t.Fatalf("IsDayOfWeek(Friday) = true, want false")
	}
	if !base.IsTomorrow(referenceYesterday) {
		t.Fatalf("IsTomorrow() = false, want true")
	}
	if !base.IsYesterday(referenceTomorrow) {
		t.Fatalf("IsYesterday() = false, want true")
	}
	if !base.IsPast(referenceTomorrow) {
		t.Fatalf("IsPast() = false, want true")
	}
	if !base.IsFuture(referenceYesterday) {
		t.Fatalf("IsFuture() = false, want true")
	}
	farFuture, err := Parse("3000-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("parse far future: %v", err)
	}
	if !base.IsNowOrPast() || !farFuture.IsNowOrFuture() {
		t.Fatalf("now-or-past/future predicates did not match expected values")
	}
	if got := base.AddDays(2).DiffForHumans(base); got != "in 2 days" {
		t.Fatalf("DiffForHumans() = %q, want future diff", got)
	}
	if got := base.DiffForHumans(base.AddHours(3)); got != "3 hours ago" {
		t.Fatalf("DiffForHumans() = %q, want past diff", got)
	}
}

func TestISOWeekMetadataWeekdayNavigationAndAge(t *testing.T) {
	monday, err := Parse("2024-01-01T12:30:00Z")
	if err != nil {
		t.Fatalf("parse monday: %v", err)
	}
	week53, err := Parse("2020-12-31T12:30:00Z")
	if err != nil {
		t.Fatalf("parse week53: %v", err)
	}
	friday, err := Parse("2024-05-17T12:30:00Z")
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

	birthday, err := Parse("2000-06-15T00:00:00Z")
	if err != nil {
		t.Fatalf("parse birthday: %v", err)
	}
	beforeBirthday, err := Parse("2024-06-14T23:59:59Z")
	if err != nil {
		t.Fatalf("parse before birthday: %v", err)
	}
	onBirthday, err := Parse("2024-06-15T00:00:00Z")
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

func TestDurationsParseNormalizeSerializeAndApply(t *testing.T) {
	parsed, err := ParseDuration("P1Y2M3DT4H5M6.007S")
	if err != nil {
		t.Fatalf("parse duration: %v", err)
	}
	normalized, err := ParseDuration("P1Y14M8DT25H61M61.250S")
	if err != nil {
		t.Fatalf("parse normalized duration: %v", err)
	}

	if parsed.Years != 1 || parsed.Months != 2 || parsed.Days != 3 || parsed.Hours != 4 || parsed.Minutes != 5 || parsed.Seconds != 6 || parsed.Milliseconds != 7 {
		t.Fatalf("ParseDuration() = %#v, want parsed components", parsed)
	}
	if got := parsed.ToMap()["hours"]; got != 4 {
		t.Fatalf("Duration.ToMap()[hours] = %d, want 4", got)
	}
	if got := parsed.ToArray(); got != [9]int{1, 0, 2, 0, 3, 4, 5, 6, 7} {
		t.Fatalf("Duration.ToArray() = %#v, want component array", got)
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
	zero, err := ParseDuration("PT0S")
	if err != nil {
		t.Fatalf("parse zero duration: %v", err)
	}
	if !zero.IsZero() {
		t.Fatalf("IsZero() = false, want true")
	}
	if !parsed.IsPositive() {
		t.Fatalf("IsPositive() = false, want true")
	}
	negative, err := ParseDuration("-P1D")
	if err != nil {
		t.Fatalf("parse negative duration: %v", err)
	}
	if !negative.IsNegative() {
		t.Fatalf("IsNegative() = false, want true")
	}
	if zero.IsPositive() || zero.IsNegative() {
		t.Fatalf("zero duration sign predicates = positive:%t negative:%t, want false/false", zero.IsPositive(), zero.IsNegative())
	}
	weeks, err := ParseDuration("P2W")
	if err != nil {
		t.Fatalf("parse week duration: %v", err)
	}
	if got := weeks.Normalize().ISOString(); got != "P14D" {
		t.Fatalf("Normalize().ISOString() = %q, want week carried to days", got)
	}

	base, err := Parse("2024-01-31T00:00:00Z")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	oneMonth, err := ParseDuration("P1M")
	if err != nil {
		t.Fatalf("parse one month: %v", err)
	}
	if got := base.AddDuration(oneMonth).DateString(); got != "2024-03-02" {
		t.Fatalf("AddDuration(P1M).DateString() = %q, want overflowed date", got)
	}
	if got := base.SubDuration(Duration{Days: 2, Hours: 3}).ISOString(); got != "2024-01-28T21:00:00.000Z" {
		t.Fatalf("SubDuration().ISOString() = %q, want shifted instant", got)
	}

	start, err := Parse("2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("parse interval start: %v", err)
	}
	end, err := Parse("2024-01-03T12:00:00Z")
	if err != nil {
		t.Fatalf("parse interval end: %v", err)
	}
	if got := start.IntervalUntil(end).ToDuration().ISOString(); got != "P2DT12H" {
		t.Fatalf("Interval.ToDuration().ISOString() = %q, want normalized interval", got)
	}
	periodStart, err := Parse("2024-01-01")
	if err != nil {
		t.Fatalf("parse period start: %v", err)
	}
	periodEnd, err := Parse("2024-01-03")
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

func TestWeekdayArithmeticAndSameUnitComparisons(t *testing.T) {
	friday, err := Parse("2024-05-17T10:00:00Z")
	if err != nil {
		t.Fatalf("parse friday: %v", err)
	}
	wednesday, err := Parse("2024-05-22T10:00:00Z")
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
	if got := wednesday.DiffInUnit(Day, friday); got != 5 {
		t.Fatalf("DiffInUnit(Day) = %d, want 5", got)
	}
	if got := wednesday.DiffInDaysFiltered(friday, func(item Tempo) bool { return item.IsMonday() }); got != 1 {
		t.Fatalf("DiffInDaysFiltered(Monday) = %d, want 1", got)
	}
	if got := wednesday.DiffInHoursFiltered(friday, func(item Tempo) bool { return item.Hour() == 12 }); got != 5 {
		t.Fatalf("DiffInHoursFiltered(hour 12) = %d, want 5", got)
	}
	if got := friday.DiffInWeekdays(wednesday); got != -3 {
		t.Fatalf("negative DiffInWeekdays() = %d, want -3", got)
	}
	if got := friday.DiffInWeekdays(wednesday, DiffOptions{Absolute: true}); got != 3 {
		t.Fatalf("absolute DiffInWeekdays() = %d, want 3", got)
	}
	october, err := Parse("2024-10-01T00:00:00Z")
	if err != nil {
		t.Fatalf("parse october: %v", err)
	}
	if got := october.DiffInQuarters(friday); got != 1 {
		t.Fatalf("DiffInQuarters() = %d, want 1", got)
	}
	if got := friday.DiffInQuarters(october); got != -1 {
		t.Fatalf("negative DiffInQuarters() = %d, want -1", got)
	}
	november, err := Parse("2024-11-17T10:00:00Z")
	if err != nil {
		t.Fatalf("parse november: %v", err)
	}
	if got := friday.IntervalUntil(november).Quarters(); got != 2 {
		t.Fatalf("Interval.Quarters() = %d, want 2", got)
	}

	sameSecond, err := Parse("2024-05-17T10:00:00.999Z")
	if err != nil {
		t.Fatalf("parse same second: %v", err)
	}
	sameMinute, err := Parse("2024-05-17T10:00:59Z")
	if err != nil {
		t.Fatalf("parse same minute: %v", err)
	}
	sameHour, err := Parse("2024-05-17T10:59:59Z")
	if err != nil {
		t.Fatalf("parse same hour: %v", err)
	}
	sameDay, err := Parse("2024-05-17T23:59:59Z")
	if err != nil {
		t.Fatalf("parse same day: %v", err)
	}
	sameWeek, err := Parse("2024-05-13T00:00:00Z")
	if err != nil {
		t.Fatalf("parse same week: %v", err)
	}
	sameMonth, err := Parse("2024-05-01T00:00:00Z")
	if err != nil {
		t.Fatalf("parse same month: %v", err)
	}
	sameQuarter, err := Parse("2024-04-01T00:00:00Z")
	if err != nil {
		t.Fatalf("parse same quarter: %v", err)
	}
	sameYear, err := Parse("2024-12-31T23:59:59Z")
	if err != nil {
		t.Fatalf("parse same year: %v", err)
	}
	birthday, err := Parse("1990-05-17T00:00:00Z")
	if err != nil {
		t.Fatalf("parse birthday: %v", err)
	}
	notBirthday, err := Parse("1990-05-18T00:00:00Z")
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
	if !friday.IsSameUnit(Day, sameDay) {
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
	if got := friday.SetTime(0, 0, 42, 0).SecondsSinceMidnight(); got != 42 {
		t.Fatalf("SecondsSinceMidnight() = %d, want 42", got)
	}
	if got := friday.SetTime(23, 59, 17, 0).SecondsUntilEndOfDay(); got != 42 {
		t.Fatalf("SecondsUntilEndOfDay() = %d, want 42", got)
	}
	if got := friday.MidDay().TimeString(); got != "12:00:00" {
		t.Fatalf("MidDay().TimeString() = %q, want noon", got)
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

func TestNamedSerializationAndMapConversion(t *testing.T) {
	tempo, err := Parse("2024-05-15T12:34:56.789Z", WithTimezone("Asia/Tokyo"))
	if err != nil {
		t.Fatalf("parse fixture input: %v", err)
	}

	assertEqual(t, "DateTimeLocalString()", tempo.DateTimeLocalString(), "2024-05-15T21:34:56")
	assertEqual(t, "DateTimeLocalString(ms)", tempo.DateTimeLocalString(MillisecondPrecision), "2024-05-15T21:34:56.789")
	assertEqual(t, "ISOFormat()", tempo.ISOFormat("YYYY-MM-DD HH:mm:ss"), "2024-05-15 21:34:56")
	assertEqual(t, "TranslatedFormat()", tempo.TranslatedFormat("YYYY-MM-DD HH:mm:ss"), "2024-05-15 21:34:56")
	assertEqual(t, "FormattedDateString()", tempo.FormattedDateString(), "May 15, 2024")
	assertEqual(t, "FormattedDayDateString()", tempo.FormattedDayDateString(), "Wed, May 15, 2024")
	assertEqual(t, "DayDateTimeString()", tempo.DayDateTimeString(), "Wed, May 15, 2024 9:34 PM")
	assertEqual(t, "TimeString(ms)", tempo.TimeString(MillisecondPrecision), "21:34:56.789")
	assertEqual(t, "ISO8601String()", tempo.ISO8601String(), "2024-05-15T21:34:56+09:00")
	assertEqual(t, "ISO8601ZuluString()", tempo.ISO8601ZuluString(), "2024-05-15T12:34:56Z")
	assertEqual(t, "ISO8601ZuluString(ms)", tempo.ISO8601ZuluString(MillisecondPrecision), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "RFC3339String(ms)", tempo.RFC3339String(MillisecondPrecision), "2024-05-15T21:34:56.789+09:00")
	assertEqual(t, "RFC822String()", tempo.RFC822String(), "Wed, 15 May 24 21:34:56 +0900")
	assertEqual(t, "RFC850String()", tempo.RFC850String(), "Wednesday, 15-May-24 21:34:56 +0900")
	assertEqual(t, "RFC1036String()", tempo.RFC1036String(), "Wed, 15 May 24 21:34:56 +0900")
	assertEqual(t, "RFC1123String()", tempo.RFC1123String(), "Wed, 15 May 2024 21:34:56 +0900")
	assertEqual(t, "RFC2822String()", tempo.RFC2822String(), "Wed, 15 May 2024 21:34:56 +0900")
	assertEqual(t, "W3CString()", tempo.W3CString(), "2024-05-15T21:34:56+09:00")
	assertEqual(t, "AtomString()", tempo.AtomString(), "2024-05-15T21:34:56+09:00")
	assertEqual(t, "RSSString()", tempo.RSSString(), "Wed, 15 May 2024 21:34:56 +0900")
	assertEqual(t, "RFC7231String()", tempo.RFC7231String(), "Wed, 15 May 2024 12:34:56 GMT")
	assertEqual(t, "CookieString()", tempo.CookieString(), "Wed, 15-May-2024 12:34:56 GMT")
	assertEqual(t, "UnixString()", tempo.UnixString(), "1715776496")
	if got := tempo.Unix(); got != 1715776496 {
		t.Fatalf("Unix() = %d, want 1715776496", got)
	}
	if got := tempo.GetTimestampMs(); got != 1715776496789 {
		t.Fatalf("GetTimestampMs() = %d, want 1715776496789", got)
	}
	assertEqual(t, "JSONSerialize()", tempo.JSONSerialize(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "Serialize()", tempo.Serialize(), "2024-05-15T12:34:56.789Z")
	tempoJSON, err := json.Marshal(tempo)
	if err != nil {
		t.Fatalf("marshal tempo: %v", err)
	}
	assertEqual(t, "json.Marshal(Tempo)", string(tempoJSON), `"2024-05-15T12:34:56.789Z"`)
	mutableJSON, err := json.Marshal(NewMutable(tempo))
	if err != nil {
		t.Fatalf("marshal mutable tempo: %v", err)
	}
	assertEqual(t, "json.Marshal(MutableTempo)", string(mutableJSON), `"2024-05-15T12:34:56.789Z"`)
	durationJSON, err := json.Marshal(Duration{Days: 1, Hours: 2})
	if err != nil {
		t.Fatalf("marshal duration: %v", err)
	}
	assertEqual(t, "json.Marshal(Duration)", string(durationJSON), `"P1DT2H"`)
	var decoded Tempo
	if err := json.Unmarshal(tempoJSON, &decoded); err != nil {
		t.Fatalf("unmarshal tempo: %v", err)
	}
	assertEqual(t, "json.Unmarshal(Tempo)", decoded.ISOString(), "2024-05-15T12:34:56.789Z")
	fromSerialized, err := FromSerialized(`"2024-05-15T12:34:56.789Z"`)
	if err != nil {
		t.Fatalf("FromSerialized(): %v", err)
	}
	assertEqual(t, "FromSerialized().ISOString()", fromSerialized.ISOString(), "2024-05-15T12:34:56.789Z")
	var decodedMutable MutableTempo
	if err := json.Unmarshal(tempoJSON, &decodedMutable); err != nil {
		t.Fatalf("unmarshal mutable tempo: %v", err)
	}
	assertEqual(t, "json.Unmarshal(MutableTempo)", decodedMutable.ISOString(), "2024-05-15T12:34:56.789Z")
	var decodedDuration Duration
	if err := json.Unmarshal(durationJSON, &decodedDuration); err != nil {
		t.Fatalf("unmarshal duration: %v", err)
	}
	assertEqual(t, "json.Unmarshal(Duration)", decodedDuration.ISOString(), "P1DT2H")

	values := tempo.ToMap()
	if values["timeZone"] != "Asia/Tokyo" {
		t.Fatalf("ToMap()[timeZone] = %v, want Asia/Tokyo", values["timeZone"])
	}
	if values["hour"] != 21 {
		t.Fatalf("ToMap()[hour] = %v, want 21", values["hour"])
	}
	if value, ok := tempo.Get("year"); !ok || value != 2024 {
		t.Fatalf("Get(year) = %v, %v, want 2024, true", value, ok)
	}
	if value, ok := tempo.GetPaddedUnit("month", 2); !ok || value != "05" {
		t.Fatalf("GetPaddedUnit(month) = %q, %v, want 05, true", value, ok)
	}
	assertEqual(t, "GetTranslatedDayName()", tempo.GetTranslatedDayName(), "Wednesday")
	assertEqual(t, "GetTranslatedShortMonthName()", tempo.GetTranslatedShortMonthName(), "May")
	assertEqual(t, "TranslateNumber()", tempo.TranslateNumber(1234), "1234")
	assertEqual(t, "Translate()", tempo.Translate("Hello :name", map[string]string{"name": "Tempo"}), "Hello Tempo")
	if _, ok := TryParse("not a date"); ok || GetLastErrors() == nil {
		t.Fatalf("TryParse(invalid), GetLastErrors() = %v, %v, want false and error", ok, GetLastErrors())
	}
	if got := ExecuteWithLocale("fr-FR", GetLocale); got != "fr-FR" {
		t.Fatalf("ExecuteWithLocale() = %q, want fr-FR", got)
	}
	if got := GetLocale(); got != "en-US" {
		t.Fatalf("GetLocale() after ExecuteWithLocale = %q, want restored default", got)
	}
	SetToStringFormat("YYYY-MM-DD")
	assertEqual(t, "String() formatted", tempo.String(), "2024-05-15")
	ResetToStringFormat()
	SerializeUsing(func(value Tempo) string { return value.DateString() })
	hookedJSON, err := json.Marshal(tempo)
	if err != nil {
		t.Fatalf("marshal hooked tempo: %v", err)
	}
	assertEqual(t, "json.Marshal(hooked Tempo)", string(hookedJSON), `"2024-05-15"`)
	SerializeUsing(nil)
	dateOnly := func(value Tempo) string { return value.DateString() }
	assertEqual(t, "composable dateOnly", dateOnly(tempo), "2024-05-15")
}

func TestExplicitSettersHandleZeroValues(t *testing.T) {
	tempo, err := Parse("2024-05-15T12:34:56.789Z")
	if err != nil {
		t.Fatalf("parse fixture input: %v", err)
	}

	assertEqual(t, "SetTime().ISOString()", tempo.SetTime(0, 0, 0, 0).ISOString(), "2024-05-15T00:00:00.000Z")
	if !tempo.SetTime(0, 0, 0, 0).IsMidnight() {
		t.Fatalf("SetTime(0,0,0,0).IsMidnight() = false, want true")
	}
	assertEqual(t, "Midday().ISOString()", tempo.Midday().ISOString(), "2024-05-15T12:00:00.000Z")
	if !tempo.Midday().IsMidday() {
		t.Fatalf("Midday().IsMidday() = false, want true")
	}
	if got := tempo.SetHour(0).Hour(); got != 0 {
		t.Fatalf("SetHour(0).Hour() = %d, want 0", got)
	}
	if got := tempo.SetMinute(0).Minute(); got != 0 {
		t.Fatalf("SetMinute(0).Minute() = %d, want 0", got)
	}
	if got := tempo.SetSecond(0).Second(); got != 0 {
		t.Fatalf("SetSecond(0).Second() = %d, want 0", got)
	}
	if got := tempo.SetMillisecond(0).Millisecond(); got != 0 {
		t.Fatalf("SetMillisecond(0).Millisecond() = %d, want 0", got)
	}
	assertEqual(t, "SetDate().DateString()", tempo.SetDate(2025, 1, 2).DateString(), "2025-01-02")
	setUnit, err := tempo.SetUnit(Day, 2)
	if err != nil {
		t.Fatalf("SetUnit(Day, 2): %v", err)
	}
	assertEqual(t, "SetUnit(Day, 2).DateString()", setUnit.DateString(), "2024-05-02")
	assertEqual(t, "SetDateTime().ISOString()", tempo.SetDateTime(2025, 1, 2, 3, 4, 5, 6).ISOString(), "2025-01-02T03:04:05.006Z")

	source, err := Parse("2025-01-02T03:04:05.006Z")
	if err != nil {
		t.Fatalf("parse setter source: %v", err)
	}
	assertEqual(t, "SetDateFrom().DateTimeString()", tempo.SetDateFrom(source).DateTimeString(), "2025-01-02 12:34:56")
	assertEqual(t, "SetTimeFrom().DateTimeString()", tempo.SetTimeFrom(source).DateTimeString(), "2024-05-15 03:04:05")
	assertEqual(t, "SetDateTimeFrom().ISOString()", tempo.SetDateTimeFrom(source).ISOString(), "2025-01-02T03:04:05.006Z")

	fromTimeString, err := tempo.SetTimeFromTimeString("03:04:05.006")
	if err != nil {
		t.Fatalf("SetTimeFromTimeString(): %v", err)
	}
	assertEqual(t, "SetTimeFromTimeString().ISOString()", fromTimeString.ISOString(), "2024-05-15T03:04:05.006Z")
	assertEqual(t, "SetTimestamp(0).ISOString()", tempo.SetTimestamp(0).ISOString(), "1970-01-01T00:00:00.000Z")
	assertEqual(t, "SetISODate().DateString()", tempo.SetISODate(2024, 20, 3).DateString(), "2024-05-15")
	if got := tempo.Weekday(); got != 3 {
		t.Fatalf("Weekday() = %d, want 3", got)
	}
	assertEqual(t, "SetWeekday().DateString()", tempo.SetWeekday(time.Friday).DateString(), "2024-05-17")
	assertEqual(t, "SetDayOfYear().DateString()", tempo.SetDayOfYear(60).DateString(), "2024-02-29")
	modified, err := tempo.Modify("+2 days")
	if err != nil {
		t.Fatalf("Modify(+2 days): %v", err)
	}
	assertEqual(t, "Modify(+2 days).DateString()", modified.DateString(), "2024-05-17")
	modified, err = tempo.Modify("previous day")
	if err != nil {
		t.Fatalf("Modify(previous day): %v", err)
	}
	assertEqual(t, "Modify(previous day).DateString()", modified.DateString(), "2024-05-14")
	modified, err = tempo.Change("next week")
	if err != nil {
		t.Fatalf("Change(next week): %v", err)
	}
	assertEqual(t, "Change(next week).DateString()", modified.DateString(), "2024-05-22")
	assertEqual(t, "Subtract(2, Day).DateString()", tempo.Subtract(2, Day).DateString(), "2024-05-13")
	assertEqual(t, "AddUnit(Day, 2).DateString()", tempo.AddUnit(Day, 2).DateString(), "2024-05-17")
	assertEqual(t, "AddRealUnit(Day, 2).DateString()", tempo.AddRealUnit(Day, 2).DateString(), "2024-05-17")
	assertEqual(t, "AddUTCUnit(Day, 2).DateString()", tempo.AddUTCUnit(Day, 2).DateString(), "2024-05-17")
	assertEqual(t, "RawAdd(2, Day).DateString()", tempo.RawAdd(2, Day).DateString(), "2024-05-17")
	assertEqual(t, "AddUnitNoOverflow(Day, 20, Month).DateString()", tempo.AddUnitNoOverflow(Day, 20, Month).DateString(), "2024-05-31")
	assertEqual(t, "SubUnit(Day, 2).DateString()", tempo.SubUnit(Day, 2).DateString(), "2024-05-13")
	assertEqual(t, "SubRealUnit(Day, 2).DateString()", tempo.SubRealUnit(Day, 2).DateString(), "2024-05-13")
	assertEqual(t, "SubUTCUnit(Day, 2).DateString()", tempo.SubUTCUnit(Day, 2).DateString(), "2024-05-13")
	assertEqual(t, "RawSub(2, Day).DateString()", tempo.RawSub(2, Day).DateString(), "2024-05-13")
	assertEqual(t, "SubUnitNoOverflow(Day, 20, Month).DateString()", tempo.SubUnitNoOverflow(Day, 20, Month).DateString(), "2024-05-01")
	assertEqual(t, "SetUnitNoOverflow(Day, 40, Month).DateString()", tempo.SetUnitNoOverflow(Day, 40, Month).DateString(), "2024-05-31")

	friday, err := Parse("2024-05-17T12:00:00Z")
	if err != nil {
		t.Fatalf("parse friday: %v", err)
	}
	monday, err := Parse("2024-05-20T12:00:00Z")
	if err != nil {
		t.Fatalf("parse monday: %v", err)
	}
	if got := friday.NextWeekendDay().DayOfWeek(); got != 6 {
		t.Fatalf("NextWeekendDay().DayOfWeek() = %d, want 6", got)
	}
	if got := monday.PreviousWeekendDay().DayOfWeek(); got != 0 {
		t.Fatalf("PreviousWeekendDay().DayOfWeek() = %d, want 0", got)
	}

	shifted, err := tempo.ShiftTimezone("Asia/Tokyo")
	if err != nil {
		t.Fatalf("ShiftTimezone(): %v", err)
	}
	assertEqual(t, "ShiftTimezone().DateTimeString()", shifted.DateTimeString(), "2024-05-15 12:34:56")

	createdFromTimeString, err := CreateFromTimeString("03:04:05.006")
	if err != nil {
		t.Fatalf("CreateFromTimeString(): %v", err)
	}
	assertEqual(t, "CreateFromTimeString().TimeString(ms)", createdFromTimeString.TimeString(MillisecondPrecision), "03:04:05.006")

	createdFromFormat, err := CreateFromFormat("2024/05/15", "YYYY/MM/DD")
	if err != nil {
		t.Fatalf("CreateFromFormat(): %v", err)
	}
	assertEqual(t, "CreateFromFormat().DateString()", createdFromFormat.DateString(), "2024-05-15")
	rawCreatedFromFormat, err := RawCreateFromFormat("2024/05/15", "YYYY/MM/DD")
	if err != nil {
		t.Fatalf("RawCreateFromFormat(): %v", err)
	}
	assertEqual(t, "RawCreateFromFormat().DateString()", rawCreatedFromFormat.DateString(), "2024-05-15")
	rawParsed, err := RawParse("2024-05-15")
	if err != nil {
		t.Fatalf("RawParse(): %v", err)
	}
	assertEqual(t, "RawParse().DateString()", rawParsed.DateString(), "2024-05-15")
	made, err := Make("2024-05-15")
	if err != nil {
		t.Fatalf("Make(): %v", err)
	}
	assertEqual(t, "Make().DateString()", made.DateString(), "2024-05-15")
	parsedFromLocale, err := ParseFromLocale("2024-05-15", "en-US")
	if err != nil {
		t.Fatalf("ParseFromLocale(): %v", err)
	}
	assertEqual(t, "ParseFromLocale().DateString()", parsedFromLocale.DateString(), "2024-05-15")
	days := GetDays()
	if len(days) < 2 || days[0] != "Sunday" || days[1] != "Monday" {
		t.Fatalf("GetDays()[0:2] = %v, want Sunday/Monday", days[:2])
	}
	if got := GetCalendarFormats()["sameDay"]; got != "[Today at] HH:mm" {
		t.Fatalf("GetCalendarFormats()[sameDay] = %q, want default format", got)
	}
	if got := GetIsoFormats()["date"]; got != "YYYY-MM-DD" {
		t.Fatalf("GetIsoFormats()[date] = %q, want date format", got)
	}
	if units := GetIsoUnits(); len(units) == 0 || units[0] != Millisecond {
		t.Fatalf("GetIsoUnits()[0] = %v, want millisecond", units)
	}
	if got := GetTimeFormatByPrecision(MillisecondPrecision); got != "HH:mm:ss.SSS" {
		t.Fatalf("GetTimeFormatByPrecision(ms) = %q, want millisecond format", got)
	}
	if GetWeekStartsAt() != time.Monday || GetWeekEndsAt() != time.Sunday {
		t.Fatalf("week start/end settings = %v/%v, want Monday/Sunday", GetWeekStartsAt(), GetWeekEndsAt())
	}
	if !LocaleHasDiffSyntax("en-US") || !LocaleHasPeriodSyntax("en-US") {
		t.Fatalf("locale capability helpers = false, want true")
	}
	if !HasFormatWithModifiers("2024/05/15", "YYYY/MM/DD") {
		t.Fatalf("HasFormatWithModifiers() = false, want true")
	}
	if !CanBeCreatedFromFormat("2024/05/15", "YYYY/MM/DD") {
		t.Fatalf("CanBeCreatedFromFormat() = false, want true")
	}

	createdFromTimestamp, err := CreateFromTimestamp(0)
	if err != nil {
		t.Fatalf("CreateFromTimestamp(): %v", err)
	}
	assertEqual(t, "CreateFromTimestamp(0).ISOString()", createdFromTimestamp.ISOString(), "1970-01-01T00:00:00.000Z")

	createdFromTimestampMs, err := CreateFromTimestampMs(1)
	if err != nil {
		t.Fatalf("CreateFromTimestampMs(): %v", err)
	}
	assertEqual(t, "CreateFromTimestampMs(1).ISOString()", createdFromTimestampMs.ISOString(), "1970-01-01T00:00:00.001Z")

	fromTimestampUTC, err := FromTimestampUTC(0)
	if err != nil {
		t.Fatalf("FromTimestampUTC(): %v", err)
	}
	assertEqual(t, "FromTimestampUTC(0).ISOString()", fromTimestampUTC.ISOString(), "1970-01-01T00:00:00.000Z")

	fromTimestampMsUTC, err := FromTimestampMsUTC(1)
	if err != nil {
		t.Fatalf("FromTimestampMsUTC(): %v", err)
	}
	assertEqual(t, "FromTimestampMsUTC(1).ISOString()", fromTimestampMsUTC.ISOString(), "1970-01-01T00:00:00.001Z")

	createdFromTimestampUTC, err := CreateFromTimestampUTC(0)
	if err != nil {
		t.Fatalf("CreateFromTimestampUTC(): %v", err)
	}
	assertEqual(t, "CreateFromTimestampUTC(0).ISOString()", createdFromTimestampUTC.ISOString(), "1970-01-01T00:00:00.000Z")

	createdFromTimestampMsUTC, err := CreateFromTimestampMsUTC(1)
	if err != nil {
		t.Fatalf("CreateFromTimestampMsUTC(): %v", err)
	}
	assertEqual(t, "CreateFromTimestampMsUTC(1).ISOString()", createdFromTimestampMsUTC.ISOString(), "1970-01-01T00:00:00.001Z")

	assertEqual(t, "From()", tempo.AddDays(2).From(tempo), "in 2 days")
	assertEqual(t, "Since()", tempo.AddDays(2).Since(tempo), "in 2 days")
	assertEqual(t, "To()", tempo.To(tempo.AddDays(2)), "in 2 days")
	assertEqual(t, "Timespan()", tempo.AddDays(2).Timespan(tempo), "in 2 days")
	if !tempo.IsImmutable() || tempo.IsMutable() {
		t.Fatalf("Tempo mutability predicates = immutable:%t mutable:%t, want true/false", tempo.IsImmutable(), tempo.IsMutable())
	}
	assertEqual(t, "AvoidMutation().ISOString()", tempo.AvoidMutation().ISOString(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "Cast().ISOString()", tempo.Cast().ISOString(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "Tempoize().DateString()", tempo.Tempoize(tempo.AddDays(1)).DateString(), "2024-05-16")
	if got := tempo.NowWithSameTz().Timezone(); got != "UTC" {
		t.Fatalf("NowWithSameTz().Timezone() = %q, want UTC", got)
	}
	mutable := NewMutable(tempo)
	if mutable.IsImmutable() || !mutable.IsMutable() {
		t.Fatalf("MutableTempo mutability predicates = immutable:%t mutable:%t, want false/true", mutable.IsImmutable(), mutable.IsMutable())
	}
	assertEqual(t, "MutableTempo.AvoidMutation().ISOString()", mutable.AvoidMutation().ISOString(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "MutableTempo.Cast().ISOString()", mutable.Cast().ISOString(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "MutableTempo.Tempoize().DateString()", mutable.Tempoize(tempo.AddDays(1)).DateString(), "2024-05-16")
	if got := mutable.NowWithSameTz().Timezone(); got != "UTC" {
		t.Fatalf("MutableTempo.NowWithSameTz().Timezone() = %q, want UTC", got)
	}
}

func TestRuntimeScopedTranslator(t *testing.T) {
	firstRuntime := NewRuntime(
		RuntimeLocale("en-US"),
		RuntimeTranslator(mapTranslator{"greeting": "Hello :name"}),
	)
	secondRuntime := NewRuntime(
		RuntimeLocale("es-ES"),
		RuntimeTranslator(mapTranslator{"greeting": "Hola :name"}),
	)

	firstFactory, err := NewFactory(WithTimezone("UTC"), WithRuntime(firstRuntime))
	if err != nil {
		t.Fatalf("new first factory: %v", err)
	}
	secondFactory, err := NewFactory(WithTimezone("UTC"), WithRuntime(secondRuntime))
	if err != nil {
		t.Fatalf("new second factory: %v", err)
	}

	first, err := firstFactory.Parse("2024-05-15")
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}
	second, err := secondFactory.Parse("2024-05-15")
	if err != nil {
		t.Fatalf("second parse: %v", err)
	}

	assertEqual(t, "first Translate()", first.Translate("greeting", map[string]string{"name": "Tempo"}), "Hello Tempo")
	assertEqual(t, "second Translate()", second.Translate("greeting", map[string]string{"name": "Tempo"}), "Hola Tempo")
	if value, ok := first.GetTranslationMessage("locale"); !ok || value != "en-US" {
		t.Fatalf("first locale message = %v, %v, want en-US, true", value, ok)
	}
	if value, ok := second.GetTranslationMessage("locale"); !ok || value != "es-ES" {
		t.Fatalf("second locale message = %v, %v, want es-ES, true", value, ok)
	}
	if !first.HasLocalTranslator() {
		t.Fatalf("HasLocalTranslator() = false, want true")
	}
	assertEqual(t, "Clone().Translate()", first.Clone().Translate("greeting", map[string]string{"name": "Tempo"}), "Hello Tempo")
	assertEqual(t, "AddDays().Translate()", first.AddDays(1).Translate("greeting", map[string]string{"name": "Tempo"}), "Hello Tempo")
	assertEqual(t, "Mutable AddDays().Translate()", first.Mutable().AddDays(1).Translate("greeting", map[string]string{"name": "Tempo"}), "Hello Tempo")

	replaced := first.SetLocalTranslator(mapTranslator{"greeting": "Salut :name"})
	assertEqual(t, "replaced Translate()", replaced.Translate("greeting", map[string]string{"name": "Tempo"}), "Salut Tempo")
	assertEqual(t, "original Translate()", first.Translate("greeting", map[string]string{"name": "Tempo"}), "Hello Tempo")
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
