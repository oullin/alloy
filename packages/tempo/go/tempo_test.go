package tempo

import (
	"strings"
	"testing"
	"time"
)

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
	if got := base.EndOf(Month).ISOString(); got != "2024-01-31T23:59:59.999Z" {
		t.Fatalf("EndOf(Month).ISOString() = %q, want end of month", got)
	}
	if got := base.StartOf(Quarter).ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("StartOf(Quarter).ISOString() = %q, want quarter start", got)
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

	tokyo, err := FromFormat("2024-01-01 09:00", "YYYY-MM-DD HH:mm", WithTimezone("Asia/Tokyo"))
	if err != nil {
		t.Fatalf("from format tokyo: %v", err)
	}
	if got := tokyo.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("FromFormat timezone ISOString() = %q, want local timezone instant", got)
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
