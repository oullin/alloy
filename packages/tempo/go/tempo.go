package tempo

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Unit string

const (
	Millisecond Unit = "millisecond"
	Second      Unit = "second"
	Minute      Unit = "minute"
	Hour        Unit = "hour"
	Day         Unit = "day"
	Week        Unit = "week"
	Month       Unit = "month"
	Quarter     Unit = "quarter"
	Year        Unit = "year"
)

type Components struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Millisecond int
	Timezone    string
}

type Object struct {
	Year          int
	Month         int
	Day           int
	Hour          int
	Minute        int
	Second        int
	Millisecond   int
	Timezone      string
	OffsetMinutes int
	Weekday       int
}

type Duration struct {
	Years        int
	Quarters     int
	Months       int
	Weeks        int
	Days         int
	Hours        int
	Minutes      int
	Seconds      int
	Milliseconds int
}

type DiffOptions struct {
	Absolute bool
	Float    bool
}

type HumanDiffOptions struct {
	Absolute bool
	Unit     Unit
}

type StartOfWeekOptions struct {
	WeekStartsOn time.Weekday
}

type PeriodOptions struct {
	Step       Duration
	IncludeEnd bool
}

type Option func(*config) error

type config struct {
	location *time.Location
}

type Tempo struct {
	value    time.Time
	location *time.Location
}

type MutableTempo struct {
	value    time.Time
	location *time.Location
}

type Interval struct {
	Start Tempo
	End   Tempo
}

type Period struct {
	Start      Tempo
	End        Tempo
	Step       Duration
	IncludeEnd bool
}

var (
	defaultLocation = time.UTC
	dateOnlyPattern = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
	localPattern    = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})(?:[T\s](\d{2})(?::?(\d{2}))?(?::?(\d{2})(?:\.(\d{1,9}))?)?)?$`)
	zonePattern     = regexp.MustCompile(`(?:Z|[+-]\d{2}:?\d{2})$`)
)

func WithTimezone(name string) Option {
	return func(cfg *config) error {
		location, err := loadLocation(name)
		if err != nil {
			return err
		}

		cfg.location = location
		return nil
	}
}

func Now(options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: time.Now().UTC(), location: cfg.location}, nil
}

func FromTime(value time.Time, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: value.UTC(), location: cfg.location}, nil
}

func Parse(input string, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	parsed, err := parseInLocation(input, cfg.location)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: parsed.UTC(), location: cfg.location}, nil
}

func FromFormat(input string, pattern string, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	parsed, err := parseFromPattern(input, pattern, cfg.location)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: parsed.UTC(), location: cfg.location}, nil
}

func Create(components Components) (Tempo, error) {
	location, err := loadLocation(components.Timezone)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: timeFromComponents(components, location).UTC(), location: location}, nil
}

func FromObject(components Components) (Tempo, error) {
	return Create(components)
}

func FromTimestamp(timestamp int64, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: time.Unix(timestamp, 0).UTC(), location: cfg.location}, nil
}

func FromTimestampMs(timestamp int64, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: time.UnixMilli(timestamp).UTC(), location: cfg.location}, nil
}

func Min(first Tempo, rest ...Tempo) Tempo {
	result := first
	for _, item := range rest {
		if item.Before(result) {
			result = item
		}
	}

	return result
}

func Max(first Tempo, rest ...Tempo) Tempo {
	result := first
	for _, item := range rest {
		if item.After(result) {
			result = item
		}
	}

	return result
}

func NewMutable(input Tempo) *MutableTempo {
	return &MutableTempo{value: input.value, location: input.location}
}

func ParseMutable(input string, options ...Option) (*MutableTempo, error) {
	parsed, err := Parse(input, options...)
	if err != nil {
		return nil, err
	}

	return NewMutable(parsed), nil
}

func (tempo Tempo) Clone() Tempo {
	return tempo
}

func (tempo Tempo) Timezone() string {
	return tempo.location.String()
}

func (tempo Tempo) Timestamp() int64 {
	return tempo.value.Unix()
}

func (tempo Tempo) TimestampMs() int64 {
	return tempo.value.UnixMilli()
}

func (tempo Tempo) Year() int {
	return tempo.local().Year()
}

func (tempo Tempo) Month() int {
	return int(tempo.local().Month())
}

func (tempo Tempo) Quarter() int {
	return (tempo.Month()-1)/3 + 1
}

func (tempo Tempo) Day() int {
	return tempo.local().Day()
}

func (tempo Tempo) DayOfWeek() int {
	return int(tempo.local().Weekday())
}

func (tempo Tempo) ISOWeekday() int {
	weekday := tempo.local().Weekday()
	if weekday == time.Sunday {
		return 7
	}

	return int(weekday)
}

func (tempo Tempo) ISOWeek() (int, int) {
	year, week := tempo.local().ISOWeek()
	return year, week
}

func (tempo Tempo) ISOWeekYear() int {
	year, _ := tempo.ISOWeek()
	return year
}

func (tempo Tempo) ISOWeekNumber() int {
	_, week := tempo.ISOWeek()
	return week
}

func (tempo Tempo) WeeksInISOYear() int {
	_, week := time.Date(tempo.ISOWeekYear(), time.December, 28, 0, 0, 0, 0, tempo.location).ISOWeek()
	return week
}

func (tempo Tempo) DayOfYear() int {
	return tempo.local().YearDay()
}

func (tempo Tempo) Hour() int {
	return tempo.local().Hour()
}

func (tempo Tempo) Minute() int {
	return tempo.local().Minute()
}

func (tempo Tempo) Second() int {
	return tempo.local().Second()
}

func (tempo Tempo) Millisecond() int {
	return tempo.local().Nanosecond() / int(time.Millisecond)
}

func (tempo Tempo) OffsetMinutes() int {
	_, offset := tempo.local().Zone()
	return offset / 60
}

func (tempo Tempo) IsLeapYear() bool {
	year := tempo.Year()
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func (tempo Tempo) DaysInMonth() int {
	return daysInMonth(tempo.Year(), tempo.Month())
}

func (tempo Tempo) IsWeekend() bool {
	weekday := tempo.local().Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func (tempo Tempo) IsSunday() bool {
	return tempo.local().Weekday() == time.Sunday
}

func (tempo Tempo) IsMonday() bool {
	return tempo.local().Weekday() == time.Monday
}

func (tempo Tempo) IsTuesday() bool {
	return tempo.local().Weekday() == time.Tuesday
}

func (tempo Tempo) IsWednesday() bool {
	return tempo.local().Weekday() == time.Wednesday
}

func (tempo Tempo) IsThursday() bool {
	return tempo.local().Weekday() == time.Thursday
}

func (tempo Tempo) IsFriday() bool {
	return tempo.local().Weekday() == time.Friday
}

func (tempo Tempo) IsSaturday() bool {
	return tempo.local().Weekday() == time.Saturday
}

func (tempo Tempo) IsWeekday() bool {
	return !tempo.IsWeekend()
}

func (tempo Tempo) IsPast(reference Tempo) bool {
	return tempo.Before(reference)
}

func (tempo Tempo) IsFuture(reference Tempo) bool {
	return tempo.After(reference)
}

func (tempo Tempo) IsToday(reference Tempo) bool {
	return tempo.Same(reference, Day)
}

func (tempo Tempo) IsTomorrow(reference Tempo) bool {
	return tempo.Same(reference.AddDays(1), Day)
}

func (tempo Tempo) IsYesterday(reference Tempo) bool {
	return tempo.Same(reference.SubDays(1), Day)
}

func (tempo Tempo) SetTimezone(name string) (Tempo, error) {
	location, err := loadLocation(name)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: tempo.value, location: location}, nil
}

func (tempo Tempo) SetTimezoneKeepLocal(name string) (Tempo, error) {
	location, err := loadLocation(name)
	if err != nil {
		return Tempo{}, err
	}

	object := tempo.ToObject()
	next := time.Date(
		object.Year,
		time.Month(object.Month),
		object.Day,
		object.Hour,
		object.Minute,
		object.Second,
		object.Millisecond*int(time.Millisecond),
		location,
	)

	return Tempo{value: next.UTC(), location: location}, nil
}

func (tempo Tempo) UTC() Tempo {
	return Tempo{value: tempo.value, location: time.UTC}
}

func (tempo Tempo) Set(components Components) (Tempo, error) {
	object := tempo.ToObject()
	location := tempo.location
	if components.Timezone != "" {
		nextLocation, err := loadLocation(components.Timezone)
		if err != nil {
			return Tempo{}, err
		}
		location = nextLocation
	}

	if components.Year != 0 {
		object.Year = components.Year
	}
	if components.Month != 0 {
		object.Month = components.Month
	}
	if components.Day != 0 {
		object.Day = components.Day
	}
	if components.Hour != 0 {
		object.Hour = components.Hour
	}
	if components.Minute != 0 {
		object.Minute = components.Minute
	}
	if components.Second != 0 {
		object.Second = components.Second
	}
	if components.Millisecond != 0 {
		object.Millisecond = components.Millisecond
	}

	next := time.Date(
		object.Year,
		time.Month(object.Month),
		object.Day,
		object.Hour,
		object.Minute,
		object.Second,
		object.Millisecond*int(time.Millisecond),
		location,
	)

	return Tempo{value: next.UTC(), location: location}, nil
}

func (tempo Tempo) Add(value int, unit Unit) Tempo {
	switch normalizeUnit(unit) {
	case Millisecond:
		return tempo.addDuration(time.Duration(value) * time.Millisecond)
	case Second:
		return tempo.addDuration(time.Duration(value) * time.Second)
	case Minute:
		return tempo.addDuration(time.Duration(value) * time.Minute)
	case Hour:
		return tempo.addDuration(time.Duration(value) * time.Hour)
	case Day:
		return tempo.addDurationDate(0, 0, value)
	case Week:
		return tempo.addDurationDate(0, 0, value*7)
	case Month:
		return tempo.AddMonths(value)
	case Quarter:
		return tempo.AddMonths(value * 3)
	case Year:
		return tempo.AddYears(value)
	default:
		return tempo
	}
}

func (tempo Tempo) Sub(value int, unit Unit) Tempo {
	return tempo.Add(-value, unit)
}

func (tempo Tempo) AddDuration(duration Duration) Tempo {
	return tempo.
		AddYears(duration.Years).
		AddMonths(duration.Quarters*3 + duration.Months).
		AddWeeks(duration.Weeks).
		AddDays(duration.Days).
		AddHours(duration.Hours).
		AddMinutes(duration.Minutes).
		AddSeconds(duration.Seconds).
		AddMilliseconds(duration.Milliseconds)
}

func (tempo Tempo) SubDuration(duration Duration) Tempo {
	return tempo.AddDuration(Duration{
		Years:        -duration.Years,
		Quarters:     -duration.Quarters,
		Months:       -duration.Months,
		Weeks:        -duration.Weeks,
		Days:         -duration.Days,
		Hours:        -duration.Hours,
		Minutes:      -duration.Minutes,
		Seconds:      -duration.Seconds,
		Milliseconds: -duration.Milliseconds,
	})
}

func (tempo Tempo) AddMilliseconds(milliseconds int) Tempo {
	return tempo.Add(milliseconds, Millisecond)
}

func (tempo Tempo) SubMilliseconds(milliseconds int) Tempo {
	return tempo.Sub(milliseconds, Millisecond)
}

func (tempo Tempo) AddSeconds(seconds int) Tempo {
	return tempo.Add(seconds, Second)
}

func (tempo Tempo) SubSeconds(seconds int) Tempo {
	return tempo.Sub(seconds, Second)
}

func (tempo Tempo) AddMinutes(minutes int) Tempo {
	return tempo.Add(minutes, Minute)
}

func (tempo Tempo) SubMinutes(minutes int) Tempo {
	return tempo.Sub(minutes, Minute)
}

func (tempo Tempo) AddHours(hours int) Tempo {
	return tempo.Add(hours, Hour)
}

func (tempo Tempo) SubHours(hours int) Tempo {
	return tempo.Sub(hours, Hour)
}

func (tempo Tempo) AddDays(days int) Tempo {
	return tempo.Add(days, Day)
}

func (tempo Tempo) SubDays(days int) Tempo {
	return tempo.Sub(days, Day)
}

func (tempo Tempo) AddWeeks(weeks int) Tempo {
	return tempo.Add(weeks, Week)
}

func (tempo Tempo) SubWeeks(weeks int) Tempo {
	return tempo.Sub(weeks, Week)
}

func (tempo Tempo) AddMonths(months int) Tempo {
	return tempo.addDurationDate(0, months, 0)
}

func (tempo Tempo) SubMonths(months int) Tempo {
	return tempo.AddMonths(-months)
}

func (tempo Tempo) AddMonthsNoOverflow(months int) Tempo {
	local := tempo.local()
	target := time.Date(local.Year(), local.Month()+time.Month(months), 1, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), tempo.location)
	day := min(local.Day(), daysInMonth(target.Year(), int(target.Month())))
	next := time.Date(
		target.Year(),
		target.Month(),
		day,
		local.Hour(),
		local.Minute(),
		local.Second(),
		local.Nanosecond(),
		tempo.location,
	)

	return Tempo{value: next.UTC(), location: tempo.location}
}

func (tempo Tempo) SubMonthsNoOverflow(months int) Tempo {
	return tempo.AddMonthsNoOverflow(-months)
}

func (tempo Tempo) AddQuarters(quarters int) Tempo {
	return tempo.AddMonths(quarters * 3)
}

func (tempo Tempo) SubQuarters(quarters int) Tempo {
	return tempo.AddQuarters(-quarters)
}

func (tempo Tempo) AddYears(years int) Tempo {
	return tempo.addDurationDate(years, 0, 0)
}

func (tempo Tempo) SubYears(years int) Tempo {
	return tempo.AddYears(-years)
}

func (tempo Tempo) AddYearsNoOverflow(years int) Tempo {
	local := tempo.local()
	year := local.Year() + years
	day := min(local.Day(), daysInMonth(year, int(local.Month())))
	next := time.Date(
		year,
		local.Month(),
		day,
		local.Hour(),
		local.Minute(),
		local.Second(),
		local.Nanosecond(),
		tempo.location,
	)

	return Tempo{value: next.UTC(), location: tempo.location}
}

func (tempo Tempo) Age(reference Tempo) int {
	return reference.DiffInYears(tempo)
}

func (tempo Tempo) SubYearsNoOverflow(years int) Tempo {
	return tempo.AddYearsNoOverflow(-years)
}

func (tempo Tempo) StartOf(unit Unit, options ...StartOfWeekOptions) Tempo {
	local := tempo.local()

	switch normalizeUnit(unit) {
	case Second:
		return tempo.fromLocal(time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second(), 0, tempo.location))
	case Minute:
		return tempo.fromLocal(time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), 0, 0, tempo.location))
	case Hour:
		return tempo.fromLocal(time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), 0, 0, 0, tempo.location))
	case Day:
		return tempo.fromLocal(time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, tempo.location))
	case Week:
		weekStartsOn := time.Monday
		if len(options) > 0 {
			weekStartsOn = options[0].WeekStartsOn
		}
		delta := (int(local.Weekday()) - int(weekStartsOn) + 7) % 7
		return tempo.StartOf(Day).SubDays(delta)
	case Month:
		return tempo.fromLocal(time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, tempo.location))
	case Quarter:
		month := time.Month((tempo.Quarter()-1)*3 + 1)
		return tempo.fromLocal(time.Date(local.Year(), month, 1, 0, 0, 0, 0, tempo.location))
	case Year:
		return tempo.fromLocal(time.Date(local.Year(), time.January, 1, 0, 0, 0, 0, tempo.location))
	default:
		return tempo
	}
}

func (tempo Tempo) EndOf(unit Unit, options ...StartOfWeekOptions) Tempo {
	switch normalizeUnit(unit) {
	case Second:
		return tempo.StartOf(Second).AddSeconds(1).SubMilliseconds(1)
	case Minute:
		return tempo.StartOf(Minute).AddMinutes(1).SubMilliseconds(1)
	case Hour:
		return tempo.StartOf(Hour).AddHours(1).SubMilliseconds(1)
	case Day:
		return tempo.StartOf(Day).AddDays(1).SubMilliseconds(1)
	case Week:
		return tempo.StartOf(Week, options...).AddWeeks(1).SubMilliseconds(1)
	case Month:
		return tempo.StartOf(Month).AddMonths(1).SubMilliseconds(1)
	case Quarter:
		return tempo.StartOf(Quarter).AddQuarters(1).SubMilliseconds(1)
	case Year:
		return tempo.StartOf(Year).AddYears(1).SubMilliseconds(1)
	default:
		return tempo
	}
}

func (tempo Tempo) StartOfDay() Tempo {
	return tempo.StartOf(Day)
}

func (tempo Tempo) EndOfDay() Tempo {
	return tempo.EndOf(Day)
}

func (tempo Tempo) StartOfMonth() Tempo {
	return tempo.StartOf(Month)
}

func (tempo Tempo) EndOfMonth() Tempo {
	return tempo.EndOf(Month)
}

func (tempo Tempo) FirstOfMonth(weekdays ...time.Weekday) Tempo {
	first := tempo.StartOf(Month)
	if len(weekdays) == 0 {
		return first
	}

	target := weekdays[0]
	delta := (int(target) - int(first.local().Weekday()) + 7) % 7
	return first.AddDays(delta)
}

func (tempo Tempo) LastOfMonth(weekdays ...time.Weekday) Tempo {
	last := tempo.EndOf(Month).StartOf(Day)
	if len(weekdays) == 0 {
		return last
	}

	target := weekdays[0]
	delta := (int(last.local().Weekday()) - int(target) + 7) % 7
	return last.SubDays(delta)
}

func (tempo Tempo) StartOfYear() Tempo {
	return tempo.StartOf(Year)
}

func (tempo Tempo) EndOfYear() Tempo {
	return tempo.EndOf(Year)
}

func (tempo Tempo) Floor(unit Unit) Tempo {
	fixed, ok := fixedUnitDuration(unit)
	if !ok {
		return tempo.StartOf(unit)
	}

	unixNano := tempo.value.UnixNano()
	fixedNano := int64(fixed)

	return Tempo{value: time.Unix(0, unixNano/fixedNano*fixedNano).UTC(), location: tempo.location}
}

func (tempo Tempo) Ceil(unit Unit) Tempo {
	floored := tempo.Floor(unit)
	if floored.Same(tempo) {
		return floored
	}

	return floored.Add(1, unit)
}

func (tempo Tempo) Round(unit Unit) Tempo {
	fixed, ok := fixedUnitDuration(unit)
	if !ok {
		start := tempo.StartOf(unit)
		end := tempo.EndOf(unit)
		midpoint := start.TimestampMs() + (end.TimestampMs()-start.TimestampMs())/2
		if tempo.TimestampMs() >= midpoint {
			return tempo.Ceil(unit)
		}

		return start
	}

	return Tempo{value: tempo.value.Round(fixed).UTC(), location: tempo.location}
}

func (tempo Tempo) Next(weekday time.Weekday) Tempo {
	delta := (int(weekday) - int(tempo.local().Weekday()) + 7) % 7
	if delta == 0 {
		delta = 7
	}

	return tempo.AddDays(delta)
}

func (tempo Tempo) Previous(weekday time.Weekday) Tempo {
	delta := (int(tempo.local().Weekday()) - int(weekday) + 7) % 7
	if delta == 0 {
		delta = 7
	}

	return tempo.SubDays(delta)
}

func (tempo Tempo) NextWeekday() Tempo {
	next := tempo.AddDays(1)
	for next.IsWeekend() {
		next = next.AddDays(1)
	}

	return next
}

func (tempo Tempo) PreviousWeekday() Tempo {
	previous := tempo.SubDays(1)
	for previous.IsWeekend() {
		previous = previous.SubDays(1)
	}

	return previous
}

func (tempo Tempo) Diff(other Tempo, unit Unit, options ...DiffOptions) float64 {
	opts := DiffOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	duration := tempo.value.Sub(other.value)
	value := 0.0
	switch normalizeUnit(unit) {
	case Millisecond:
		value = float64(duration.Milliseconds())
	case Second:
		value = duration.Seconds()
	case Minute:
		value = duration.Minutes()
	case Hour:
		value = duration.Hours()
	case Day:
		value = duration.Hours() / 24
	case Week:
		value = duration.Hours() / (24 * 7)
	case Month, Quarter, Year:
		value = monthDiff(tempo, other, unit)
	}

	if opts.Absolute {
		value = math.Abs(value)
	}

	if opts.Float {
		return value
	}

	return math.Trunc(value)
}

func (tempo Tempo) DiffInMilliseconds(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Millisecond, options...))
}

func (tempo Tempo) DiffInSeconds(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Second, options...))
}

func (tempo Tempo) DiffInMinutes(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Minute, options...))
}

func (tempo Tempo) DiffInHours(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Hour, options...))
}

func (tempo Tempo) DiffInDays(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Day, options...))
}

func (tempo Tempo) DiffInWeeks(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Week, options...))
}

func (tempo Tempo) DiffInMonths(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Month, options...))
}

func (tempo Tempo) DiffInYears(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Year, options...))
}

func (tempo Tempo) DiffForHumans(other Tempo, options ...HumanDiffOptions) string {
	opts := HumanDiffOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	milliseconds := tempo.TimestampMs() - other.TimestampMs()
	unit := opts.Unit
	if unit == "" {
		unit = bestRelativeUnit(milliseconds)
	}

	value := int(math.Round(float64(milliseconds) / float64(unitDuration(unit).Milliseconds())))
	if opts.Absolute && value < 0 {
		value = -value
	}

	unitName := string(normalizeUnit(unit))
	if value == 1 || value == -1 {
		unitName = strings.TrimSuffix(unitName, "s")
	} else {
		unitName += "s"
	}

	if value < 0 {
		return fmt.Sprintf("%d %s ago", -value, unitName)
	}

	return fmt.Sprintf("in %d %s", value, unitName)
}

func (tempo Tempo) Before(other Tempo, units ...Unit) bool {
	return tempo.compareValue(units...) < other.compareValue(units...)
}

func (tempo Tempo) After(other Tempo, units ...Unit) bool {
	return tempo.compareValue(units...) > other.compareValue(units...)
}

func (tempo Tempo) Same(other Tempo, units ...Unit) bool {
	return tempo.compareValue(units...) == other.compareValue(units...)
}

func (tempo Tempo) SameOrBefore(other Tempo, units ...Unit) bool {
	return tempo.Same(other, units...) || tempo.Before(other, units...)
}

func (tempo Tempo) SameOrAfter(other Tempo, units ...Unit) bool {
	return tempo.Same(other, units...) || tempo.After(other, units...)
}

func (tempo Tempo) Between(start Tempo, end Tempo, inclusivity ...string) bool {
	mode := "()"
	if len(inclusivity) > 0 {
		mode = inclusivity[0]
	}

	afterStart := tempo.After(start)
	if strings.HasPrefix(mode, "[") {
		afterStart = tempo.SameOrAfter(start)
	}

	beforeEnd := tempo.Before(end)
	if strings.HasSuffix(mode, "]") {
		beforeEnd = tempo.SameOrBefore(end)
	}

	return afterStart && beforeEnd
}

func (tempo Tempo) Format(pattern string) string {
	local := tempo.local()
	offset := tempo.OffsetMinutes()
	hour12 := local.Hour() % 12
	if hour12 == 0 {
		hour12 = 12
	}

	values := map[string]string{
		"A":    ternary(local.Hour() < 12, "AM", "PM"),
		"a":    ternary(local.Hour() < 12, "am", "pm"),
		"D":    strconv.Itoa(local.Day()),
		"DD":   pad(local.Day(), 2),
		"Do":   ordinal(local.Day()),
		"d":    strconv.Itoa(int(local.Weekday())),
		"ddd":  local.Weekday().String()[:3],
		"dddd": local.Weekday().String(),
		"H":    strconv.Itoa(local.Hour()),
		"HH":   pad(local.Hour(), 2),
		"h":    strconv.Itoa(hour12),
		"hh":   pad(hour12, 2),
		"M":    strconv.Itoa(int(local.Month())),
		"MM":   pad(int(local.Month()), 2),
		"MMM":  local.Month().String()[:3],
		"MMMM": local.Month().String(),
		"m":    strconv.Itoa(local.Minute()),
		"mm":   pad(local.Minute(), 2),
		"S":    strconv.Itoa(tempo.Millisecond() / 100),
		"SSS":  pad(tempo.Millisecond(), 3),
		"s":    strconv.Itoa(local.Second()),
		"ss":   pad(local.Second(), 2),
		"X":    strconv.FormatInt(tempo.Timestamp(), 10),
		"x":    strconv.FormatInt(tempo.TimestampMs(), 10),
		"Y":    strconv.Itoa(local.Year()),
		"YY":   pad(local.Year()%100, 2),
		"YYYY": pad(local.Year(), 4),
		"Z":    formatOffset(offset, ":"),
		"ZZ":   formatOffset(offset, ""),
	}

	tokens := []string{"YYYY", "MMMM", "dddd", "MMM", "ddd", "SSS", "Do", "YY", "ZZ", "MM", "DD", "HH", "hh", "mm", "ss", "Z", "X", "x", "Y", "M", "D", "H", "h", "m", "s", "A", "a", "d"}
	var builder strings.Builder

	for index := 0; index < len(pattern); {
		if pattern[index] == '[' {
			end := strings.IndexByte(pattern[index:], ']')
			if end >= 0 {
				builder.WriteString(pattern[index+1 : index+end])
				index += end + 1
				continue
			}
		}

		matched := false
		for _, token := range tokens {
			if strings.HasPrefix(pattern[index:], token) {
				builder.WriteString(values[token])
				index += len(token)
				matched = true
				break
			}
		}

		if !matched {
			builder.WriteByte(pattern[index])
			index++
		}
	}

	return builder.String()
}

func (tempo Tempo) DateString() string {
	return tempo.Format("YYYY-MM-DD")
}

func (tempo Tempo) TimeString() string {
	return tempo.Format("HH:mm:ss")
}

func (tempo Tempo) DateTimeString() string {
	return tempo.Format("YYYY-MM-DD HH:mm:ss")
}

func (tempo Tempo) ISOString() string {
	return tempo.value.UTC().Format("2006-01-02T15:04:05.000Z")
}

func (tempo Tempo) Time() time.Time {
	return tempo.value
}

func (tempo Tempo) ToObject() Object {
	local := tempo.local()

	return Object{
		Year:          local.Year(),
		Month:         int(local.Month()),
		Day:           local.Day(),
		Hour:          local.Hour(),
		Minute:        local.Minute(),
		Second:        local.Second(),
		Millisecond:   local.Nanosecond() / int(time.Millisecond),
		Timezone:      tempo.Timezone(),
		OffsetMinutes: tempo.OffsetMinutes(),
		Weekday:       int(local.Weekday()),
	}
}

func (tempo Tempo) ToArray() [7]int {
	object := tempo.ToObject()
	return [7]int{
		object.Year,
		object.Month,
		object.Day,
		object.Hour,
		object.Minute,
		object.Second,
		object.Millisecond,
	}
}

func (tempo Tempo) IntervalUntil(end Tempo) Interval {
	return Interval{Start: tempo, End: end}
}

func (tempo Tempo) PeriodUntil(end Tempo, options ...PeriodOptions) Period {
	step := Duration{Days: 1}
	includeEnd := true
	if len(options) > 0 {
		if options[0].Step != (Duration{}) {
			step = options[0].Step
		}
		includeEnd = options[0].IncludeEnd
	}

	return Period{Start: tempo, End: end, Step: step, IncludeEnd: includeEnd}
}

func (tempo Tempo) local() time.Time {
	return tempo.value.In(tempo.location)
}

func (tempo Tempo) addDuration(duration time.Duration) Tempo {
	return Tempo{value: tempo.value.Add(duration).UTC(), location: tempo.location}
}

func (tempo Tempo) addDurationDate(years int, months int, days int) Tempo {
	return Tempo{value: tempo.local().AddDate(years, months, days).UTC(), location: tempo.location}
}

func (tempo Tempo) fromLocal(local time.Time) Tempo {
	return Tempo{value: local.UTC(), location: tempo.location}
}

func (tempo Tempo) compareValue(units ...Unit) int64 {
	unit := Millisecond
	if len(units) > 0 {
		unit = units[0]
	}

	if normalizeUnit(unit) == Millisecond {
		return tempo.TimestampMs()
	}

	return tempo.StartOf(unit).TimestampMs()
}

func (mutable *MutableTempo) Tempo() Tempo {
	return Tempo{value: mutable.value, location: mutable.location}
}

func (mutable *MutableTempo) ISOString() string {
	return mutable.Tempo().ISOString()
}

func (mutable *MutableTempo) DateTimeString() string {
	return mutable.Tempo().DateTimeString()
}

func (mutable *MutableTempo) AddHours(hours int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddHours(hours))
}

func (mutable *MutableTempo) AddDays(days int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddDays(days))
}

func (mutable *MutableTempo) StartOf(unit Unit, options ...StartOfWeekOptions) *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOf(unit, options...))
}

func (mutable *MutableTempo) SetTimezone(name string) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTimezone(name)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) replace(next Tempo) *MutableTempo {
	mutable.value = next.value
	mutable.location = next.location
	return mutable
}

func (interval Interval) Inverted() bool {
	return interval.Start.After(interval.End)
}

func (interval Interval) Milliseconds() int {
	return interval.End.DiffInMilliseconds(interval.Start)
}

func (interval Interval) Seconds() int {
	return interval.End.DiffInSeconds(interval.Start)
}

func (interval Interval) Minutes() int {
	return interval.End.DiffInMinutes(interval.Start)
}

func (interval Interval) Hours() int {
	return interval.End.DiffInHours(interval.Start)
}

func (interval Interval) Days() int {
	return interval.End.DiffInDays(interval.Start)
}

func (interval Interval) Contains(input Tempo, inclusivity ...string) bool {
	mode := "[]"
	if len(inclusivity) > 0 {
		mode = inclusivity[0]
	}

	return input.Between(interval.Start, interval.End, mode)
}

func (interval Interval) Overlaps(other Interval) bool {
	return interval.Start.Before(other.End) && interval.End.After(other.Start)
}

func (interval Interval) Intersection(other Interval) (Interval, bool) {
	if !interval.Overlaps(other) {
		return Interval{}, false
	}

	return Interval{
		Start: Max(interval.Start, other.Start),
		End:   Min(interval.End, other.End),
	}, true
}

func (interval Interval) Union(other Interval) Interval {
	return Interval{
		Start: Min(interval.Start, other.Start),
		End:   Max(interval.End, other.End),
	}
}

func (period Period) Values() ([]Tempo, error) {
	values := make([]Tempo, 0)
	current := period.Start
	forward := period.End.SameOrAfter(period.Start)

	for {
		include := current.Before(period.End)
		if forward {
			include = current.Before(period.End)
			if period.IncludeEnd {
				include = current.SameOrBefore(period.End)
			}
		} else {
			include = current.After(period.End)
			if period.IncludeEnd {
				include = current.SameOrAfter(period.End)
			}
		}

		if !include {
			break
		}

		values = append(values, current)
		next := current.AddDuration(period.Step)
		if next.Same(current) {
			return nil, errors.New("tempo period step must advance the period")
		}
		current = next
	}

	return values, nil
}

func applyOptions(options ...Option) (config, error) {
	cfg := config{location: defaultLocation}
	for _, option := range options {
		if err := option(&cfg); err != nil {
			return config{}, err
		}
	}

	return cfg, nil
}

func loadLocation(name string) (*time.Location, error) {
	if name == "" || name == "UTC" {
		return time.UTC, nil
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("invalid tempo time zone %q: %w", name, err)
	}

	return location, nil
}

func parseInLocation(input string, location *time.Location) (time.Time, error) {
	if matches := dateOnlyPattern.FindStringSubmatch(input); matches != nil {
		year := mustInt(matches[1])
		month := mustInt(matches[2])
		day := mustInt(matches[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, location), nil
	}

	if !zonePattern.MatchString(input) {
		if matches := localPattern.FindStringSubmatch(input); matches != nil {
			year := mustInt(matches[1])
			month := mustInt(matches[2])
			day := mustInt(matches[3])
			hour := mustInt(defaultString(matches[4], "0"))
			minute := mustInt(defaultString(matches[5], "0"))
			second := mustInt(defaultString(matches[6], "0"))
			millisecond := mustInt(rightPad(defaultString(matches[7], "0"), 3))

			return time.Date(year, time.Month(month), day, hour, minute, second, millisecond*int(time.Millisecond), location), nil
		}
	}

	parsed, err := time.Parse(time.RFC3339Nano, input)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse tempo: %w", err)
	}

	return parsed, nil
}

func parseFromPattern(input string, pattern string, location *time.Location) (time.Time, error) {
	tokens := []string{"YYYY", "SSS", "YY", "ZZ", "MM", "DD", "HH", "hh", "mm", "ss", "Z", "M", "D", "H", "h", "m", "s", "A", "a"}
	groups := make([]string, 0)
	var expression strings.Builder
	expression.WriteString("^")

	for index := 0; index < len(pattern); {
		if pattern[index] == '[' {
			end := strings.IndexByte(pattern[index:], ']')
			if end >= 0 {
				expression.WriteString(regexp.QuoteMeta(pattern[index+1 : index+end]))
				index += end + 1
				continue
			}
		}

		matched := ""
		for _, token := range tokens {
			if strings.HasPrefix(pattern[index:], token) {
				matched = token
				break
			}
		}

		if matched == "" {
			expression.WriteString(regexp.QuoteMeta(pattern[index : index+1]))
			index++
			continue
		}

		groups = append(groups, matched)
		switch matched {
		case "A", "a":
			expression.WriteString(`(AM|PM|am|pm)`)
		case "Z":
			expression.WriteString(`(Z|[+-]\d{2}:\d{2})`)
		case "ZZ":
			expression.WriteString(`(Z|[+-]\d{4})`)
		case "YYYY":
			expression.WriteString(`(\d{4})`)
		case "YY", "MM", "DD", "HH", "hh", "mm", "ss":
			expression.WriteString(`(\d{2})`)
		case "SSS":
			expression.WriteString(`(\d{1,3})`)
		default:
			expression.WriteString(`(\d{1,2})`)
		}
		index += len(matched)
	}

	expression.WriteString("$")
	match := regexp.MustCompile(expression.String()).FindStringSubmatch(input)
	if match == nil {
		return time.Time{}, fmt.Errorf("input does not match tempo format: %s", input)
	}

	values := make(map[string]string, len(groups))
	for index, token := range groups {
		values[token] = match[index+1]
	}

	year := 1970
	if value := values["YYYY"]; value != "" {
		year = mustInt(value)
	} else if value := values["YY"]; value != "" {
		year = 2000 + mustInt(value)
	}

	hour := mustInt(firstPresent(values, "HH", "H", "hh", "h"))
	meridiem := firstPresent(values, "A", "a")
	if meridiem != "" {
		switch strings.ToLower(meridiem) {
		case "pm":
			if hour < 12 {
				hour += 12
			}
		case "am":
			if hour == 12 {
				hour = 0
			}
		}
	}

	components := Components{
		Year:        year,
		Month:       mustInt(defaultString(firstPresent(values, "MM", "M"), "1")),
		Day:         mustInt(defaultString(firstPresent(values, "DD", "D"), "1")),
		Hour:        hour,
		Minute:      mustInt(defaultString(firstPresent(values, "mm", "m"), "0")),
		Second:      mustInt(defaultString(firstPresent(values, "ss", "s"), "0")),
		Millisecond: mustInt(rightPad(defaultString(values["SSS"], "0"), 3)),
	}

	if offset := firstPresent(values, "Z", "ZZ"); offset != "" {
		offsetMinutes, err := parseOffsetMinutes(offset)
		if err != nil {
			return time.Time{}, err
		}
		utc := time.Date(
			components.Year,
			time.Month(components.Month),
			components.Day,
			components.Hour,
			components.Minute,
			components.Second,
			components.Millisecond*int(time.Millisecond),
			time.UTC,
		)

		return utc.Add(-time.Duration(offsetMinutes) * time.Minute), nil
	}

	return timeFromComponents(components, location), nil
}

func timeFromComponents(components Components, location *time.Location) time.Time {
	month := components.Month
	if month == 0 {
		month = 1
	}

	day := components.Day
	if day == 0 {
		day = 1
	}

	return time.Date(
		components.Year,
		time.Month(month),
		day,
		components.Hour,
		components.Minute,
		components.Second,
		components.Millisecond*int(time.Millisecond),
		location,
	)
}

func normalizeUnit(unit Unit) Unit {
	switch unit {
	case "milliseconds":
		return Millisecond
	case "seconds":
		return Second
	case "minutes":
		return Minute
	case "hours":
		return Hour
	case "days":
		return Day
	case "weeks":
		return Week
	case "months":
		return Month
	case "quarters":
		return Quarter
	case "years":
		return Year
	default:
		return unit
	}
}

func fixedUnitDuration(unit Unit) (time.Duration, bool) {
	switch normalizeUnit(unit) {
	case Millisecond:
		return time.Millisecond, true
	case Second:
		return time.Second, true
	case Minute:
		return time.Minute, true
	case Hour:
		return time.Hour, true
	case Day:
		return 24 * time.Hour, true
	case Week:
		return 7 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func unitDuration(unit Unit) time.Duration {
	if duration, ok := fixedUnitDuration(unit); ok {
		return duration
	}

	switch normalizeUnit(unit) {
	case Month:
		return 30 * 24 * time.Hour
	case Year:
		return 365 * 24 * time.Hour
	default:
		return time.Millisecond
	}
}

func bestRelativeUnit(milliseconds int64) Unit {
	absolute := milliseconds
	if absolute < 0 {
		absolute = -absolute
	}

	switch {
	case absolute < int64(time.Minute/time.Millisecond):
		return Second
	case absolute < int64(time.Hour/time.Millisecond):
		return Minute
	case absolute < int64((24*time.Hour)/time.Millisecond):
		return Hour
	case absolute < int64((7*24*time.Hour)/time.Millisecond):
		return Day
	case absolute < int64((30*24*time.Hour)/time.Millisecond):
		return Week
	case absolute < int64((365*24*time.Hour)/time.Millisecond):
		return Month
	default:
		return Year
	}
}

func monthDiff(left Tempo, right Tempo, unit Unit) float64 {
	sign := 1.0
	start := right
	end := left
	if left.Before(right) {
		sign = -1
		start = left
		end = right
	}

	startObject := start.ToObject()
	endObject := end.ToObject()
	months := (endObject.Year-startObject.Year)*12 + (endObject.Month - startObject.Month)
	if endObject.Day < startObject.Day {
		months--
	}

	value := float64(months)
	switch normalizeUnit(unit) {
	case Quarter:
		value /= 3
	case Year:
		value /= 12
	}

	return value * sign
}

func daysInMonth(year int, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func pad(value int, length int) string {
	result := strconv.Itoa(value)
	if value < 0 {
		result = strconv.Itoa(-value)
	}

	for len(result) < length {
		result = "0" + result
	}

	return result
}

func ordinal(value int) string {
	remainder := value % 100
	if remainder >= 11 && remainder <= 13 {
		return strconv.Itoa(value) + "th"
	}

	switch value % 10 {
	case 1:
		return strconv.Itoa(value) + "st"
	case 2:
		return strconv.Itoa(value) + "nd"
	case 3:
		return strconv.Itoa(value) + "rd"
	default:
		return strconv.Itoa(value) + "th"
	}
}

func formatOffset(offsetMinutes int, separator string) string {
	sign := "+"
	if offsetMinutes < 0 {
		sign = "-"
		offsetMinutes = -offsetMinutes
	}

	return fmt.Sprintf("%s%s%s%s", sign, pad(offsetMinutes/60, 2), separator, pad(offsetMinutes%60, 2))
}

func parseOffsetMinutes(input string) (int, error) {
	if input == "Z" {
		return 0, nil
	}

	clean := strings.ReplaceAll(input, ":", "")
	if len(clean) != 5 {
		return 0, fmt.Errorf("invalid tempo offset: %s", input)
	}

	sign := 1
	if clean[0] == '-' {
		sign = -1
	} else if clean[0] != '+' {
		return 0, fmt.Errorf("invalid tempo offset: %s", input)
	}

	hours, err := strconv.Atoi(clean[1:3])
	if err != nil {
		return 0, fmt.Errorf("invalid tempo offset: %s", input)
	}
	minutes, err := strconv.Atoi(clean[3:5])
	if err != nil {
		return 0, fmt.Errorf("invalid tempo offset: %s", input)
	}

	return sign * (hours*60 + minutes), nil
}

func firstPresent(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := values[key]; value != "" {
			return value
		}
	}

	return ""
}

func mustInt(input string) int {
	value, _ := strconv.Atoi(input)
	return value
}

func defaultString(input string, fallback string) string {
	if input == "" {
		return fallback
	}

	return input
}

func rightPad(input string, length int) string {
	for len(input) < length {
		input += "0"
	}

	if len(input) > length {
		return input[:length]
	}

	return input
}

func ternary(condition bool, left string, right string) string {
	if condition {
		return left
	}

	return right
}
