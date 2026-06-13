package tempo

import (
	"strings"
	"testing"
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
