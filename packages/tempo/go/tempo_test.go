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
	if !base.Between(earlier, end) {
		t.Fatalf("Between() = false, want true")
	}
	if got := base.DiffInHours(earlier); got != 26 {
		t.Fatalf("DiffInHours() = %d, want 26", got)
	}
	if got := base.DiffInMinutes(earlier); got != 1594 {
		t.Fatalf("DiffInMinutes() = %d, want 1594", got)
	}
	if got := base.Floor(Hour).ISOString(); got != "2024-05-15T10:00:00.000Z" {
		t.Fatalf("Floor(Hour).ISOString() = %q, want hour floor", got)
	}
	if got := base.Ceil(Hour).ISOString(); got != "2024-05-15T11:00:00.000Z" {
		t.Fatalf("Ceil(Hour).ISOString() = %q, want hour ceil", got)
	}
	if got := base.Round(Hour).ISOString(); got != "2024-05-15T11:00:00.000Z" {
		t.Fatalf("Round(Hour).ISOString() = %q, want hour round", got)
	}
	if got := base.Format("YYYY-MM-DD HH:mm:ss.SSS ZZ [Q]M"); got != "2024-05-15 10:34:45.600 +0000 Q5" {
		t.Fatalf("Format() = %q, want token output", got)
	}
	if got := base.Format("dddd, MMMM Do YYYY"); got != "Wednesday, May 15th 2024" {
		t.Fatalf("Format() = %q, want long date output", got)
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
	if got := base.DaysInMonth(); got != 29 {
		t.Fatalf("DaysInMonth() = %d, want 29", got)
	}
	if !saturday.IsWeekend() {
		t.Fatalf("IsWeekend() = false, want true")
	}
	if !base.IsWeekday() {
		t.Fatalf("IsWeekday() = false, want true")
	}
	if !base.IsTomorrow(referenceYesterday) {
		t.Fatalf("IsTomorrow() = false, want true")
	}
	if !base.IsYesterday(referenceTomorrow) {
		t.Fatalf("IsYesterday() = false, want true")
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
	if got := friday.FirstOfYear(time.Monday).DateString(); got != "2024-01-01" {
		t.Fatalf("FirstOfYear(Monday).DateString() = %q, want first year Monday", got)
	}
	if got := friday.LastOfYear(time.Tuesday).DateString(); got != "2024-12-31" {
		t.Fatalf("LastOfYear(Tuesday).DateString() = %q, want last year Tuesday", got)
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
	if got := friday.DiffInWeekdays(wednesday); got != -3 {
		t.Fatalf("negative DiffInWeekdays() = %d, want -3", got)
	}
	if got := friday.DiffInWeekdays(wednesday, DiffOptions{Absolute: true}); got != 3 {
		t.Fatalf("absolute DiffInWeekdays() = %d, want 3", got)
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
	if !friday.Birthday(birthday) {
		t.Fatalf("Birthday() = false, want true")
	}
	if friday.Birthday(notBirthday) {
		t.Fatalf("Birthday() = true, want false")
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
	if notStartOfDay.IsStartOf(Day) {
		t.Fatalf("IsStartOf(Day) = true, want false")
	}
	if !endOfDay.IsEndOf(Day) {
		t.Fatalf("IsEndOf(Day) = false, want true")
	}
	if notEndOfDay.IsEndOf(Day) {
		t.Fatalf("IsEndOf(Day) = true, want false")
	}
}

func TestNamedSerializationAndMapConversion(t *testing.T) {
	tempo, err := Parse("2024-05-15T12:34:56.789Z", WithTimezone("Asia/Tokyo"))
	if err != nil {
		t.Fatalf("parse fixture input: %v", err)
	}

	assertEqual(t, "DateTimeLocalString()", tempo.DateTimeLocalString(), "2024-05-15T21:34:56")
	assertEqual(t, "DateTimeLocalString(ms)", tempo.DateTimeLocalString(MillisecondPrecision), "2024-05-15T21:34:56.789")
	assertEqual(t, "TimeString(ms)", tempo.TimeString(MillisecondPrecision), "21:34:56.789")
	assertEqual(t, "ISO8601String()", tempo.ISO8601String(), "2024-05-15T21:34:56+09:00")
	assertEqual(t, "RFC3339String(ms)", tempo.RFC3339String(MillisecondPrecision), "2024-05-15T21:34:56.789+09:00")
	assertEqual(t, "AtomString()", tempo.AtomString(), "2024-05-15T21:34:56+09:00")
	assertEqual(t, "RSSString()", tempo.RSSString(), "Wed, 15 May 2024 21:34:56 +0900")
	assertEqual(t, "RFC7231String()", tempo.RFC7231String(), "Wed, 15 May 2024 12:34:56 GMT")
	assertEqual(t, "CookieString()", tempo.CookieString(), "Wed, 15-May-2024 12:34:56 GMT")
	assertEqual(t, "UnixString()", tempo.UnixString(), "1715776496")

	values := tempo.ToMap()
	if values["timeZone"] != "Asia/Tokyo" {
		t.Fatalf("ToMap()[timeZone] = %v, want Asia/Tokyo", values["timeZone"])
	}
	if values["hour"] != 21 {
		t.Fatalf("ToMap()[hour] = %v, want 21", values["hour"])
	}
}

func TestExplicitSettersHandleZeroValues(t *testing.T) {
	tempo, err := Parse("2024-05-15T12:34:56.789Z")
	if err != nil {
		t.Fatalf("parse fixture input: %v", err)
	}

	assertEqual(t, "SetTime().ISOString()", tempo.SetTime(0, 0, 0, 0).ISOString(), "2024-05-15T00:00:00.000Z")
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
}
