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

type TimeStringPrecision string

const (
	SecondPrecision      TimeStringPrecision = "second"
	MillisecondPrecision TimeStringPrecision = "millisecond"
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
	ExcludeEnd bool
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

type Factory struct {
	now      *time.Time
	location *time.Location
}

var (
	defaultLocation = time.UTC
	dateOnlyPattern = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
	durationPattern = regexp.MustCompile(`^(-)?P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)
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

func TryParse(input string, options ...Option) (Tempo, bool) {
	tempo, err := Parse(input, options...)
	return tempo, err == nil
}

func CanParse(input string, options ...Option) bool {
	_, ok := TryParse(input, options...)
	return ok
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

func TryFromFormat(input string, pattern string, options ...Option) (Tempo, bool) {
	tempo, err := FromFormat(input, pattern, options...)
	return tempo, err == nil
}

func HasFormat(input string, pattern string, options ...Option) bool {
	_, ok := TryFromFormat(input, pattern, options...)
	return ok
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

func NewFactory(options ...Option) (Factory, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Factory{}, err
	}

	return Factory{location: cfg.location}, nil
}

func NewFactoryWithTestNow(input Tempo, options ...Option) (Factory, error) {
	cfg := config{location: input.location}
	for _, option := range options {
		if err := option(&cfg); err != nil {
			return Factory{}, err
		}
	}

	now := input.value
	return Factory{now: &now, location: cfg.location}, nil
}

func (factory Factory) Now() Tempo {
	if factory.now != nil {
		return Tempo{value: factory.now.UTC(), location: factory.location}
	}

	return Tempo{value: time.Now().UTC(), location: factory.location}
}

func (factory Factory) ImmutableNow() Tempo {
	return factory.Now()
}

func (factory Factory) MutableNow() *MutableTempo {
	return NewMutable(factory.Now())
}

func (factory Factory) FromTime(value time.Time) Tempo {
	return Tempo{value: value.UTC(), location: factory.location}
}

func (factory Factory) Parse(input string) (Tempo, error) {
	parsed, err := parseInLocation(input, factory.location)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: parsed.UTC(), location: factory.location}, nil
}

func (factory Factory) TryParse(input string) (Tempo, bool) {
	tempo, err := factory.Parse(input)
	return tempo, err == nil
}

func (factory Factory) CanParse(input string) bool {
	_, ok := factory.TryParse(input)
	return ok
}

func (factory Factory) FromFormat(input string, pattern string) (Tempo, error) {
	parsed, err := parseFromPattern(input, pattern, factory.location)
	if err != nil {
		return Tempo{}, err
	}

	return Tempo{value: parsed.UTC(), location: factory.location}, nil
}

func (factory Factory) TryFromFormat(input string, pattern string) (Tempo, bool) {
	tempo, err := factory.FromFormat(input, pattern)
	return tempo, err == nil
}

func (factory Factory) HasFormat(input string, pattern string) bool {
	_, ok := factory.TryFromFormat(input, pattern)
	return ok
}

func (factory Factory) Create(components Components) (Tempo, error) {
	location := factory.location
	if components.Timezone != "" {
		nextLocation, err := loadLocation(components.Timezone)
		if err != nil {
			return Tempo{}, err
		}
		location = nextLocation
	}

	return Tempo{value: timeFromComponents(components, location).UTC(), location: location}, nil
}

func (factory Factory) FromObject(components Components) (Tempo, error) {
	return factory.Create(components)
}

func (factory Factory) FromTimestamp(timestamp int64) Tempo {
	return Tempo{value: time.Unix(timestamp, 0).UTC(), location: factory.location}
}

func (factory Factory) FromTimestampMs(timestamp int64) Tempo {
	return Tempo{value: time.UnixMilli(timestamp).UTC(), location: factory.location}
}

func ParseDuration(input string) (Duration, error) {
	matches := durationPattern.FindStringSubmatch(input)
	if matches == nil {
		return Duration{}, fmt.Errorf("invalid tempo duration: %s", input)
	}

	sign := 1
	if matches[1] == "-" {
		sign = -1
	}

	seconds := 0
	milliseconds := 0
	if matches[8] != "" {
		value, err := strconv.ParseFloat(matches[8], 64)
		if err != nil {
			return Duration{}, fmt.Errorf("invalid tempo duration seconds: %w", err)
		}
		seconds = int(math.Trunc(value))
		milliseconds = int(math.Round((value - math.Trunc(value)) * 1000))
	}

	return Duration{
		Years:        sign * mustInt(defaultString(matches[2], "0")),
		Months:       sign * mustInt(defaultString(matches[3], "0")),
		Weeks:        sign * mustInt(defaultString(matches[4], "0")),
		Days:         sign * mustInt(defaultString(matches[5], "0")),
		Hours:        sign * mustInt(defaultString(matches[6], "0")),
		Minutes:      sign * mustInt(defaultString(matches[7], "0")),
		Seconds:      sign * seconds,
		Milliseconds: sign * milliseconds,
	}.Normalize(), nil
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

func Average(start Tempo, end Tempo) Tempo {
	return Tempo{
		value:    time.UnixMilli((start.TimestampMs() + end.TimestampMs()) / 2).UTC(),
		location: start.location,
	}
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

func (duration Duration) Plus(other Duration) Duration {
	return Duration{
		Years:        duration.Years + other.Years,
		Quarters:     duration.Quarters + other.Quarters,
		Months:       duration.Months + other.Months,
		Weeks:        duration.Weeks + other.Weeks,
		Days:         duration.Days + other.Days,
		Hours:        duration.Hours + other.Hours,
		Minutes:      duration.Minutes + other.Minutes,
		Seconds:      duration.Seconds + other.Seconds,
		Milliseconds: duration.Milliseconds + other.Milliseconds,
	}
}

func (duration Duration) Minus(other Duration) Duration {
	return duration.Plus(other.Negated())
}

func (duration Duration) Negated() Duration {
	return Duration{
		Years:        -duration.Years,
		Quarters:     -duration.Quarters,
		Months:       -duration.Months,
		Weeks:        -duration.Weeks,
		Days:         -duration.Days,
		Hours:        -duration.Hours,
		Minutes:      -duration.Minutes,
		Seconds:      -duration.Seconds,
		Milliseconds: -duration.Milliseconds,
	}
}

func (duration Duration) Abs() Duration {
	return Duration{
		Years:        absInt(duration.Years),
		Quarters:     absInt(duration.Quarters),
		Months:       absInt(duration.Months),
		Weeks:        absInt(duration.Weeks),
		Days:         absInt(duration.Days),
		Hours:        absInt(duration.Hours),
		Minutes:      absInt(duration.Minutes),
		Seconds:      absInt(duration.Seconds),
		Milliseconds: absInt(duration.Milliseconds),
	}
}

func (duration Duration) Normalize() Duration {
	sign := duration.direction()
	value := duration.Abs()
	milliseconds := value.Milliseconds
	seconds := value.Seconds + milliseconds/1000
	milliseconds %= 1000
	minutes := value.Minutes + seconds/60
	seconds %= 60
	hours := value.Hours + minutes/60
	minutes %= 60
	days := value.Days + hours/24 + value.Weeks*7
	hours %= 24
	months := value.Months + value.Quarters*3
	years := value.Years + months/12
	months %= 12

	return Duration{
		Years:        sign * years,
		Months:       sign * months,
		Days:         sign * days,
		Hours:        sign * hours,
		Minutes:      sign * minutes,
		Seconds:      sign * seconds,
		Milliseconds: sign * milliseconds,
	}
}

func (duration Duration) Total(unit Unit) float64 {
	milliseconds := duration.totalMilliseconds()
	if fixed, ok := fixedUnitDuration(unit); ok {
		return float64(milliseconds) / float64(fixed.Milliseconds())
	}

	months := float64(duration.Years*12+duration.Quarters*3+duration.Months) + float64(milliseconds)/float64((30*24*time.Hour).Milliseconds())
	switch normalizeUnit(unit) {
	case Month:
		return months
	case Quarter:
		return months / 3
	case Year:
		return months / 12
	default:
		return float64(milliseconds)
	}
}

func (duration Duration) IsZero() bool {
	return duration == (Duration{})
}

func (duration Duration) ISOString() string {
	if duration.IsZero() {
		return "PT0S"
	}

	normalized := duration.Normalize()
	sign := ""
	if normalized.direction() < 0 {
		sign = "-"
	}
	value := normalized.Abs()
	dateParts := strings.Builder{}
	if value.Years != 0 {
		dateParts.WriteString(strconv.Itoa(value.Years) + "Y")
	}
	if value.Months != 0 {
		dateParts.WriteString(strconv.Itoa(value.Months) + "M")
	}
	if value.Days != 0 {
		dateParts.WriteString(strconv.Itoa(value.Days) + "D")
	}

	timeParts := strings.Builder{}
	if value.Hours != 0 {
		timeParts.WriteString(strconv.Itoa(value.Hours) + "H")
	}
	if value.Minutes != 0 {
		timeParts.WriteString(strconv.Itoa(value.Minutes) + "M")
	}
	if value.Seconds != 0 || value.Milliseconds != 0 {
		seconds := strconv.Itoa(value.Seconds)
		if value.Milliseconds != 0 {
			seconds += "." + pad(value.Milliseconds, 3)
		}
		timeParts.WriteString(seconds + "S")
	}

	result := sign + "P" + dateParts.String()
	if timeParts.Len() > 0 {
		result += "T" + timeParts.String()
	}

	return result
}

func (duration Duration) String() string {
	return duration.ISOString()
}

func (duration Duration) totalMilliseconds() int64 {
	return int64(duration.Weeks*7+duration.Days)*int64((24*time.Hour)/time.Millisecond) +
		int64(duration.Hours)*int64(time.Hour/time.Millisecond) +
		int64(duration.Minutes)*int64(time.Minute/time.Millisecond) +
		int64(duration.Seconds)*int64(time.Second/time.Millisecond) +
		int64(duration.Milliseconds)
}

func (duration Duration) direction() int {
	values := []int{
		duration.Years,
		duration.Quarters,
		duration.Months,
		duration.Weeks,
		duration.Days,
		duration.Hours,
		duration.Minutes,
		duration.Seconds,
		duration.Milliseconds,
	}
	for _, value := range values {
		if value < 0 {
			return -1
		}
		if value > 0 {
			return 1
		}
	}

	return 1
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

func (tempo Tempo) OffsetString(separator string) string {
	return formatOffset(tempo.OffsetMinutes(), separator)
}

func (tempo Tempo) ZoneName() string {
	name, _ := tempo.local().Zone()
	return name
}

func (tempo Tempo) IsUTC() bool {
	return tempo.location == time.UTC && tempo.OffsetMinutes() == 0
}

func (tempo Tempo) IsLocal() bool {
	return tempo.location.String() == time.Local.String()
}

func (tempo Tempo) IsDST() bool {
	local := tempo.local()
	january := time.Date(local.Year(), time.January, 1, 12, 0, 0, 0, tempo.location)
	july := time.Date(local.Year(), time.July, 1, 12, 0, 0, 0, tempo.location)
	_, januaryOffset := january.Zone()
	_, julyOffset := july.Zone()
	standardOffset := min(januaryOffset, julyOffset)
	_, currentOffset := local.Zone()

	return currentOffset > standardOffset
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

func (tempo Tempo) Local() Tempo {
	return Tempo{value: tempo.value, location: time.Local}
}

func (tempo Tempo) fromObject(object Object, location *time.Location) Tempo {
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

	return Tempo{value: next.UTC(), location: location}
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

	return tempo.fromObject(object, location), nil
}

func (tempo Tempo) SetYear(year int) Tempo {
	object := tempo.ToObject()
	object.Year = year

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetMonth(month int) Tempo {
	object := tempo.ToObject()
	object.Month = month

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetDay(day int) Tempo {
	object := tempo.ToObject()
	object.Day = day

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetDate(year int, month int, day int) Tempo {
	object := tempo.ToObject()
	object.Year = year
	object.Month = month
	object.Day = day

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetHour(hour int) Tempo {
	object := tempo.ToObject()
	object.Hour = hour

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetMinute(minute int) Tempo {
	object := tempo.ToObject()
	object.Minute = minute

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetSecond(second int) Tempo {
	object := tempo.ToObject()
	object.Second = second

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetMillisecond(millisecond int) Tempo {
	object := tempo.ToObject()
	object.Millisecond = millisecond

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetTime(hour int, minute int, second int, millisecond int) Tempo {
	object := tempo.ToObject()
	object.Hour = hour
	object.Minute = minute
	object.Second = second
	object.Millisecond = millisecond

	return tempo.fromObject(object, tempo.location)
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

func (tempo Tempo) AddWeekdays(days int) Tempo {
	if days == 0 {
		return tempo.Clone()
	}

	direction := 1
	if days < 0 {
		direction = -1
		days = -days
	}

	current := tempo.Clone()
	for days > 0 {
		current = current.AddDays(direction)
		if current.IsWeekday() {
			days--
		}
	}

	return current
}

func (tempo Tempo) SubWeekdays(days int) Tempo {
	return tempo.AddWeekdays(-days)
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

func (tempo Tempo) IsStartOf(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.Same(tempo.StartOf(unit, options...))
}

func (tempo Tempo) IsEndOf(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.Same(tempo.EndOf(unit, options...))
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

func (tempo Tempo) NthOfMonth(occurrence int, weekday time.Weekday) (Tempo, bool) {
	if occurrence == 0 {
		return Tempo{}, false
	}

	month := tempo.Month()
	var candidate Tempo
	if occurrence > 0 {
		candidate = tempo.FirstOfMonth(weekday).AddWeeks(occurrence - 1)
	} else {
		candidate = tempo.LastOfMonth(weekday).SubWeeks(absInt(occurrence) - 1)
	}

	return candidate, candidate.Month() == month
}

func (tempo Tempo) FirstOfQuarter(weekdays ...time.Weekday) Tempo {
	first := tempo.StartOf(Quarter)
	if len(weekdays) == 0 {
		return first
	}

	target := weekdays[0]
	delta := (int(target) - int(first.local().Weekday()) + 7) % 7
	return first.AddDays(delta)
}

func (tempo Tempo) LastOfQuarter(weekdays ...time.Weekday) Tempo {
	last := tempo.EndOf(Quarter).StartOf(Day)
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

func (tempo Tempo) FirstOfYear(weekdays ...time.Weekday) Tempo {
	first := tempo.StartOf(Year)
	if len(weekdays) == 0 {
		return first
	}

	target := weekdays[0]
	delta := (int(target) - int(first.local().Weekday()) + 7) % 7
	return first.AddDays(delta)
}

func (tempo Tempo) LastOfYear(weekdays ...time.Weekday) Tempo {
	last := tempo.EndOf(Year).StartOf(Day)
	if len(weekdays) == 0 {
		return last
	}

	target := weekdays[0]
	delta := (int(last.local().Weekday()) - int(target) + 7) % 7
	return last.SubDays(delta)
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

func (tempo Tempo) DiffInWeekdays(other Tempo, options ...DiffOptions) int {
	return tempo.diffFilteredDays(other, func(item Tempo) bool {
		return item.IsWeekday()
	}, options...)
}

func (tempo Tempo) DiffInWeekendDays(other Tempo, options ...DiffOptions) int {
	return tempo.diffFilteredDays(other, func(item Tempo) bool {
		return item.IsWeekend()
	}, options...)
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

func (tempo Tempo) SameSecond(other Tempo) bool {
	return tempo.Same(other, Second)
}

func (tempo Tempo) SameMinute(other Tempo) bool {
	return tempo.Same(other, Minute)
}

func (tempo Tempo) SameHour(other Tempo) bool {
	return tempo.Same(other, Hour)
}

func (tempo Tempo) SameDay(other Tempo) bool {
	return tempo.Same(other, Day)
}

func (tempo Tempo) SameWeek(other Tempo) bool {
	return tempo.Same(other, Week)
}

func (tempo Tempo) SameMonth(other Tempo) bool {
	return tempo.Same(other, Month)
}

func (tempo Tempo) SameQuarter(other Tempo) bool {
	return tempo.Same(other, Quarter)
}

func (tempo Tempo) SameYear(other Tempo) bool {
	return tempo.Same(other, Year)
}

func (tempo Tempo) Birthday(other Tempo) bool {
	return tempo.Month() == other.Month() && tempo.Day() == other.Day()
}

func (tempo Tempo) Clamp(minimum Tempo, maximum Tempo) (Tempo, error) {
	if minimum.After(maximum) {
		return Tempo{}, errors.New("tempo clamp minimum must be before maximum")
	}

	if tempo.Before(minimum) {
		return minimum, nil
	}

	if tempo.After(maximum) {
		return maximum, nil
	}

	return tempo.Clone(), nil
}

func (tempo Tempo) Average(other Tempo) Tempo {
	return Average(tempo, other)
}

func (tempo Tempo) Closest(first Tempo, rest ...Tempo) Tempo {
	result := first
	bestDistance := absInt64(first.TimestampMs() - tempo.TimestampMs())

	for _, item := range rest {
		distance := absInt64(item.TimestampMs() - tempo.TimestampMs())
		if distance < bestDistance {
			result = item
			bestDistance = distance
		}
	}

	return result
}

func (tempo Tempo) Farthest(first Tempo, rest ...Tempo) Tempo {
	result := first
	bestDistance := absInt64(first.TimestampMs() - tempo.TimestampMs())

	for _, item := range rest {
		distance := absInt64(item.TimestampMs() - tempo.TimestampMs())
		if distance > bestDistance {
			result = item
			bestDistance = distance
		}
	}

	return result
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

func (tempo Tempo) TimeString(precision ...TimeStringPrecision) string {
	base := tempo.Format("HH:mm:ss")
	if selectedPrecision(precision) == MillisecondPrecision {
		return base + "." + pad(tempo.Millisecond(), 3)
	}

	return base
}

func (tempo Tempo) DateTimeString() string {
	return tempo.Format("YYYY-MM-DD HH:mm:ss")
}

func (tempo Tempo) DateTimeLocalString(precision ...TimeStringPrecision) string {
	return tempo.DateString() + "T" + tempo.TimeString(precision...)
}

func (tempo Tempo) ISOString() string {
	return tempo.value.UTC().Format("2006-01-02T15:04:05.000Z")
}

func (tempo Tempo) ISO8601String() string {
	return tempo.Format("YYYY-MM-DDTHH:mm:ssZ")
}

func (tempo Tempo) RFC3339String(precision ...TimeStringPrecision) string {
	return tempo.DateTimeLocalString(precision...) + tempo.OffsetString(":")
}

func (tempo Tempo) RFC7231String() string {
	return tempo.UTC().Format("ddd, DD MMM YYYY HH:mm:ss [GMT]")
}

func (tempo Tempo) CookieString() string {
	return tempo.UTC().Format("ddd, DD-MMM-YYYY HH:mm:ss [GMT]")
}

func (tempo Tempo) AtomString() string {
	return tempo.RFC3339String()
}

func (tempo Tempo) RSSString() string {
	return tempo.Format("ddd, DD MMM YYYY HH:mm:ss ZZ")
}

func (tempo Tempo) UnixString() string {
	return strconv.FormatInt(tempo.Timestamp(), 10)
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

func (tempo Tempo) ToMap() map[string]interface{} {
	object := tempo.ToObject()

	return map[string]interface{}{
		"year":          object.Year,
		"month":         object.Month,
		"day":           object.Day,
		"hour":          object.Hour,
		"minute":        object.Minute,
		"second":        object.Second,
		"millisecond":   object.Millisecond,
		"timeZone":      object.Timezone,
		"offsetMinutes": object.OffsetMinutes,
		"weekday":       object.Weekday,
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
		if options[0].ExcludeEnd {
			includeEnd = false
		} else if options[0].IncludeEnd {
			includeEnd = true
		}
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

func (tempo Tempo) diffFilteredDays(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	opts := DiffOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	sign := 1
	start := other.StartOf(Day)
	end := tempo.StartOf(Day)
	if tempo.Before(other, Day) {
		sign = -1
		start = tempo.StartOf(Day)
		end = other.StartOf(Day)
	}

	current := start
	count := 0
	for current.Before(end, Day) {
		current = current.AddDays(1)
		if current.SameOrBefore(end, Day) && predicate(current) {
			count++
		}
	}

	if opts.Absolute {
		return count
	}

	return count * sign
}

func (mutable *MutableTempo) Tempo() Tempo {
	return Tempo{value: mutable.value, location: mutable.location}
}

func (mutable *MutableTempo) Clone() *MutableTempo {
	return NewMutable(mutable.Tempo())
}

func (mutable *MutableTempo) Timezone() string {
	return mutable.Tempo().Timezone()
}

func (mutable *MutableTempo) Timestamp() int64 {
	return mutable.Tempo().Timestamp()
}

func (mutable *MutableTempo) TimestampMs() int64 {
	return mutable.Tempo().TimestampMs()
}

func (mutable *MutableTempo) Year() int {
	return mutable.Tempo().Year()
}

func (mutable *MutableTempo) Month() int {
	return mutable.Tempo().Month()
}

func (mutable *MutableTempo) Quarter() int {
	return mutable.Tempo().Quarter()
}

func (mutable *MutableTempo) Day() int {
	return mutable.Tempo().Day()
}

func (mutable *MutableTempo) DayOfWeek() int {
	return mutable.Tempo().DayOfWeek()
}

func (mutable *MutableTempo) ISOWeekday() int {
	return mutable.Tempo().ISOWeekday()
}

func (mutable *MutableTempo) ISOWeek() (int, int) {
	return mutable.Tempo().ISOWeek()
}

func (mutable *MutableTempo) ISOWeekYear() int {
	return mutable.Tempo().ISOWeekYear()
}

func (mutable *MutableTempo) ISOWeekNumber() int {
	return mutable.Tempo().ISOWeekNumber()
}

func (mutable *MutableTempo) WeeksInISOYear() int {
	return mutable.Tempo().WeeksInISOYear()
}

func (mutable *MutableTempo) DayOfYear() int {
	return mutable.Tempo().DayOfYear()
}

func (mutable *MutableTempo) Hour() int {
	return mutable.Tempo().Hour()
}

func (mutable *MutableTempo) Minute() int {
	return mutable.Tempo().Minute()
}

func (mutable *MutableTempo) Second() int {
	return mutable.Tempo().Second()
}

func (mutable *MutableTempo) Millisecond() int {
	return mutable.Tempo().Millisecond()
}

func (mutable *MutableTempo) OffsetMinutes() int {
	return mutable.Tempo().OffsetMinutes()
}

func (mutable *MutableTempo) OffsetString(separator string) string {
	return mutable.Tempo().OffsetString(separator)
}

func (mutable *MutableTempo) ZoneName() string {
	return mutable.Tempo().ZoneName()
}

func (mutable *MutableTempo) IsUTC() bool {
	return mutable.Tempo().IsUTC()
}

func (mutable *MutableTempo) IsLocal() bool {
	return mutable.Tempo().IsLocal()
}

func (mutable *MutableTempo) IsDST() bool {
	return mutable.Tempo().IsDST()
}

func (mutable *MutableTempo) IsLeapYear() bool {
	return mutable.Tempo().IsLeapYear()
}

func (mutable *MutableTempo) DaysInMonth() int {
	return mutable.Tempo().DaysInMonth()
}

func (mutable *MutableTempo) IsWeekend() bool {
	return mutable.Tempo().IsWeekend()
}

func (mutable *MutableTempo) IsSunday() bool {
	return mutable.Tempo().IsSunday()
}

func (mutable *MutableTempo) IsMonday() bool {
	return mutable.Tempo().IsMonday()
}

func (mutable *MutableTempo) IsTuesday() bool {
	return mutable.Tempo().IsTuesday()
}

func (mutable *MutableTempo) IsWednesday() bool {
	return mutable.Tempo().IsWednesday()
}

func (mutable *MutableTempo) IsThursday() bool {
	return mutable.Tempo().IsThursday()
}

func (mutable *MutableTempo) IsFriday() bool {
	return mutable.Tempo().IsFriday()
}

func (mutable *MutableTempo) IsSaturday() bool {
	return mutable.Tempo().IsSaturday()
}

func (mutable *MutableTempo) IsWeekday() bool {
	return mutable.Tempo().IsWeekday()
}

func (mutable *MutableTempo) IsPast(reference Tempo) bool {
	return mutable.Tempo().IsPast(reference)
}

func (mutable *MutableTempo) IsFuture(reference Tempo) bool {
	return mutable.Tempo().IsFuture(reference)
}

func (mutable *MutableTempo) IsToday(reference Tempo) bool {
	return mutable.Tempo().IsToday(reference)
}

func (mutable *MutableTempo) IsTomorrow(reference Tempo) bool {
	return mutable.Tempo().IsTomorrow(reference)
}

func (mutable *MutableTempo) IsYesterday(reference Tempo) bool {
	return mutable.Tempo().IsYesterday(reference)
}

func (mutable *MutableTempo) DateString() string {
	return mutable.Tempo().DateString()
}

func (mutable *MutableTempo) TimeString(precision ...TimeStringPrecision) string {
	return mutable.Tempo().TimeString(precision...)
}

func (mutable *MutableTempo) DateTimeString() string {
	return mutable.Tempo().DateTimeString()
}

func (mutable *MutableTempo) DateTimeLocalString(precision ...TimeStringPrecision) string {
	return mutable.Tempo().DateTimeLocalString(precision...)
}

func (mutable *MutableTempo) ISOString() string {
	return mutable.Tempo().ISOString()
}

func (mutable *MutableTempo) ISO8601String() string {
	return mutable.Tempo().ISO8601String()
}

func (mutable *MutableTempo) RFC3339String(precision ...TimeStringPrecision) string {
	return mutable.Tempo().RFC3339String(precision...)
}

func (mutable *MutableTempo) RFC7231String() string {
	return mutable.Tempo().RFC7231String()
}

func (mutable *MutableTempo) CookieString() string {
	return mutable.Tempo().CookieString()
}

func (mutable *MutableTempo) AtomString() string {
	return mutable.Tempo().AtomString()
}

func (mutable *MutableTempo) RSSString() string {
	return mutable.Tempo().RSSString()
}

func (mutable *MutableTempo) UnixString() string {
	return mutable.Tempo().UnixString()
}

func (mutable *MutableTempo) Time() time.Time {
	return mutable.Tempo().Time()
}

func (mutable *MutableTempo) ToObject() Object {
	return mutable.Tempo().ToObject()
}

func (mutable *MutableTempo) ToMap() map[string]interface{} {
	return mutable.Tempo().ToMap()
}

func (mutable *MutableTempo) ToArray() [7]int {
	return mutable.Tempo().ToArray()
}

func (mutable *MutableTempo) Format(pattern string) string {
	return mutable.Tempo().Format(pattern)
}

func (mutable *MutableTempo) SetTimezone(name string) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTimezone(name)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetTimezoneKeepLocal(name string) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTimezoneKeepLocal(name)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) UTC() *MutableTempo {
	return mutable.replace(mutable.Tempo().UTC())
}

func (mutable *MutableTempo) Local() *MutableTempo {
	return mutable.replace(mutable.Tempo().Local())
}

func (mutable *MutableTempo) Set(components Components) (*MutableTempo, error) {
	next, err := mutable.Tempo().Set(components)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetYear(year int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetYear(year))
}

func (mutable *MutableTempo) SetMonth(month int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetMonth(month))
}

func (mutable *MutableTempo) SetDay(day int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetDay(day))
}

func (mutable *MutableTempo) SetDate(year int, month int, day int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetDate(year, month, day))
}

func (mutable *MutableTempo) SetHour(hour int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetHour(hour))
}

func (mutable *MutableTempo) SetMinute(minute int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetMinute(minute))
}

func (mutable *MutableTempo) SetSecond(second int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetSecond(second))
}

func (mutable *MutableTempo) SetMillisecond(millisecond int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetMillisecond(millisecond))
}

func (mutable *MutableTempo) SetTime(hour int, minute int, second int, millisecond int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetTime(hour, minute, second, millisecond))
}

func (mutable *MutableTempo) Add(value int, unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Add(value, unit))
}

func (mutable *MutableTempo) Sub(value int, unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Sub(value, unit))
}

func (mutable *MutableTempo) AddDuration(duration Duration) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddDuration(duration))
}

func (mutable *MutableTempo) SubDuration(duration Duration) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubDuration(duration))
}

func (mutable *MutableTempo) AddMilliseconds(milliseconds int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddMilliseconds(milliseconds))
}

func (mutable *MutableTempo) SubMilliseconds(milliseconds int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubMilliseconds(milliseconds))
}

func (mutable *MutableTempo) AddSeconds(seconds int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddSeconds(seconds))
}

func (mutable *MutableTempo) SubSeconds(seconds int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubSeconds(seconds))
}

func (mutable *MutableTempo) AddMinutes(minutes int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddMinutes(minutes))
}

func (mutable *MutableTempo) SubMinutes(minutes int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubMinutes(minutes))
}

func (mutable *MutableTempo) AddHours(hours int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddHours(hours))
}

func (mutable *MutableTempo) SubHours(hours int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubHours(hours))
}

func (mutable *MutableTempo) AddDays(days int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddDays(days))
}

func (mutable *MutableTempo) SubDays(days int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubDays(days))
}

func (mutable *MutableTempo) AddWeekdays(days int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddWeekdays(days))
}

func (mutable *MutableTempo) SubWeekdays(days int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubWeekdays(days))
}

func (mutable *MutableTempo) AddWeeks(weeks int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddWeeks(weeks))
}

func (mutable *MutableTempo) SubWeeks(weeks int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubWeeks(weeks))
}

func (mutable *MutableTempo) AddMonths(months int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddMonths(months))
}

func (mutable *MutableTempo) SubMonths(months int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubMonths(months))
}

func (mutable *MutableTempo) AddMonthsNoOverflow(months int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddMonthsNoOverflow(months))
}

func (mutable *MutableTempo) SubMonthsNoOverflow(months int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubMonthsNoOverflow(months))
}

func (mutable *MutableTempo) AddQuarters(quarters int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddQuarters(quarters))
}

func (mutable *MutableTempo) SubQuarters(quarters int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubQuarters(quarters))
}

func (mutable *MutableTempo) AddYears(years int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddYears(years))
}

func (mutable *MutableTempo) SubYears(years int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubYears(years))
}

func (mutable *MutableTempo) AddYearsNoOverflow(years int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddYearsNoOverflow(years))
}

func (mutable *MutableTempo) SubYearsNoOverflow(years int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubYearsNoOverflow(years))
}

func (mutable *MutableTempo) Age(reference Tempo) int {
	return mutable.Tempo().Age(reference)
}

func (mutable *MutableTempo) StartOf(unit Unit, options ...StartOfWeekOptions) *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOf(unit, options...))
}

func (mutable *MutableTempo) EndOf(unit Unit, options ...StartOfWeekOptions) *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOf(unit, options...))
}

func (mutable *MutableTempo) StartOfDay() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfDay())
}

func (mutable *MutableTempo) EndOfDay() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfDay())
}

func (mutable *MutableTempo) StartOfMonth() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfMonth())
}

func (mutable *MutableTempo) EndOfMonth() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfMonth())
}

func (mutable *MutableTempo) FirstOfMonth(weekdays ...time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().FirstOfMonth(weekdays...))
}

func (mutable *MutableTempo) LastOfMonth(weekdays ...time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().LastOfMonth(weekdays...))
}

func (mutable *MutableTempo) NthOfMonth(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return mutable.Tempo().NthOfMonth(occurrence, weekday)
}

func (mutable *MutableTempo) FirstOfQuarter(weekdays ...time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().FirstOfQuarter(weekdays...))
}

func (mutable *MutableTempo) LastOfQuarter(weekdays ...time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().LastOfQuarter(weekdays...))
}

func (mutable *MutableTempo) StartOfYear() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfYear())
}

func (mutable *MutableTempo) EndOfYear() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfYear())
}

func (mutable *MutableTempo) FirstOfYear(weekdays ...time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().FirstOfYear(weekdays...))
}

func (mutable *MutableTempo) LastOfYear(weekdays ...time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().LastOfYear(weekdays...))
}

func (mutable *MutableTempo) Floor(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Floor(unit))
}

func (mutable *MutableTempo) Ceil(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Ceil(unit))
}

func (mutable *MutableTempo) Round(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Round(unit))
}

func (mutable *MutableTempo) Next(weekday time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().Next(weekday))
}

func (mutable *MutableTempo) Previous(weekday time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().Previous(weekday))
}

func (mutable *MutableTempo) NextWeekday() *MutableTempo {
	return mutable.replace(mutable.Tempo().NextWeekday())
}

func (mutable *MutableTempo) PreviousWeekday() *MutableTempo {
	return mutable.replace(mutable.Tempo().PreviousWeekday())
}

func (mutable *MutableTempo) IsStartOf(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Tempo().IsStartOf(unit, options...)
}

func (mutable *MutableTempo) IsEndOf(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Tempo().IsEndOf(unit, options...)
}

func (mutable *MutableTempo) Diff(other Tempo, unit Unit, options ...DiffOptions) float64 {
	return mutable.Tempo().Diff(other, unit, options...)
}

func (mutable *MutableTempo) DiffInMilliseconds(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInMilliseconds(other, options...)
}

func (mutable *MutableTempo) DiffInSeconds(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInSeconds(other, options...)
}

func (mutable *MutableTempo) DiffInMinutes(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInMinutes(other, options...)
}

func (mutable *MutableTempo) DiffInHours(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInHours(other, options...)
}

func (mutable *MutableTempo) DiffInDays(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInDays(other, options...)
}

func (mutable *MutableTempo) DiffInWeeks(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInWeeks(other, options...)
}

func (mutable *MutableTempo) DiffInWeekdays(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInWeekdays(other, options...)
}

func (mutable *MutableTempo) DiffInWeekendDays(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInWeekendDays(other, options...)
}

func (mutable *MutableTempo) DiffInMonths(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInMonths(other, options...)
}

func (mutable *MutableTempo) DiffInYears(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInYears(other, options...)
}

func (mutable *MutableTempo) DiffForHumans(other Tempo, options ...HumanDiffOptions) string {
	return mutable.Tempo().DiffForHumans(other, options...)
}

func (mutable *MutableTempo) Before(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Before(other, units...)
}

func (mutable *MutableTempo) After(other Tempo, units ...Unit) bool {
	return mutable.Tempo().After(other, units...)
}

func (mutable *MutableTempo) Same(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Same(other, units...)
}

func (mutable *MutableTempo) SameSecond(other Tempo) bool {
	return mutable.Tempo().SameSecond(other)
}

func (mutable *MutableTempo) SameMinute(other Tempo) bool {
	return mutable.Tempo().SameMinute(other)
}

func (mutable *MutableTempo) SameHour(other Tempo) bool {
	return mutable.Tempo().SameHour(other)
}

func (mutable *MutableTempo) SameDay(other Tempo) bool {
	return mutable.Tempo().SameDay(other)
}

func (mutable *MutableTempo) SameWeek(other Tempo) bool {
	return mutable.Tempo().SameWeek(other)
}

func (mutable *MutableTempo) SameMonth(other Tempo) bool {
	return mutable.Tempo().SameMonth(other)
}

func (mutable *MutableTempo) SameQuarter(other Tempo) bool {
	return mutable.Tempo().SameQuarter(other)
}

func (mutable *MutableTempo) SameYear(other Tempo) bool {
	return mutable.Tempo().SameYear(other)
}

func (mutable *MutableTempo) Birthday(other Tempo) bool {
	return mutable.Tempo().Birthday(other)
}

func (mutable *MutableTempo) Clamp(minimum Tempo, maximum Tempo) (*MutableTempo, error) {
	next, err := mutable.Tempo().Clamp(minimum, maximum)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) Average(other Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Average(other))
}

func (mutable *MutableTempo) Closest(first Tempo, rest ...Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Closest(first, rest...))
}

func (mutable *MutableTempo) Farthest(first Tempo, rest ...Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Farthest(first, rest...))
}

func (mutable *MutableTempo) SameOrBefore(other Tempo, units ...Unit) bool {
	return mutable.Tempo().SameOrBefore(other, units...)
}

func (mutable *MutableTempo) SameOrAfter(other Tempo, units ...Unit) bool {
	return mutable.Tempo().SameOrAfter(other, units...)
}

func (mutable *MutableTempo) Between(start Tempo, end Tempo, inclusivity ...string) bool {
	return mutable.Tempo().Between(start, end, inclusivity...)
}

func (mutable *MutableTempo) IntervalUntil(end Tempo) Interval {
	return mutable.Tempo().IntervalUntil(end)
}

func (mutable *MutableTempo) PeriodUntil(end Tempo, options ...PeriodOptions) Period {
	return mutable.Tempo().PeriodUntil(end, options...)
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

func (interval Interval) ToDuration() Duration {
	return Duration{Milliseconds: interval.Milliseconds()}.Normalize()
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
		if (forward && next.Before(current)) || (!forward && next.After(current)) {
			return nil, errors.New("tempo period step must advance toward the end")
		}
		current = next
	}

	return values, nil
}

func (period Period) First() (Tempo, bool, error) {
	values, err := period.Values()
	if err != nil {
		return Tempo{}, false, err
	}
	if len(values) == 0 {
		return Tempo{}, false, nil
	}

	return values[0], true, nil
}

func (period Period) Last() (Tempo, bool, error) {
	values, err := period.Values()
	if err != nil {
		return Tempo{}, false, err
	}
	if len(values) == 0 {
		return Tempo{}, false, nil
	}

	return values[len(values)-1], true, nil
}

func (period Period) Count() (int, error) {
	values, err := period.Values()
	if err != nil {
		return 0, err
	}

	return len(values), nil
}

func (period Period) IsEmpty() (bool, error) {
	count, err := period.Count()
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (period Period) Contains(input Tempo) bool {
	forward := period.End.SameOrAfter(period.Start)
	afterStart := input.SameOrAfter(period.Start)
	if !forward {
		afterStart = input.SameOrBefore(period.Start)
	}

	beforeEnd := input.Before(period.End)
	if forward {
		if period.IncludeEnd {
			beforeEnd = input.SameOrBefore(period.End)
		}
	} else {
		beforeEnd = input.After(period.End)
		if period.IncludeEnd {
			beforeEnd = input.SameOrAfter(period.End)
		}
	}

	return afterStart && beforeEnd
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
	tokens := []string{"YYYY", "MMMM", "dddd", "MMM", "ddd", "SSS", "Do", "YY", "ZZ", "MM", "DD", "HH", "hh", "mm", "ss", "Z", "M", "D", "H", "h", "m", "s", "A", "a"}
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
		case "MMM", "MMMM", "ddd", "dddd":
			expression.WriteString(`([\p{L}.]+)`)
		case "Do":
			expression.WriteString(`(\d{1,2})(?:st|nd|rd|th)`)
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
		Day:         mustInt(defaultString(firstPresent(values, "DD", "Do", "D"), "1")),
		Hour:        hour,
		Minute:      mustInt(defaultString(firstPresent(values, "mm", "m"), "0")),
		Second:      mustInt(defaultString(firstPresent(values, "ss", "s"), "0")),
		Millisecond: mustInt(rightPad(defaultString(values["SSS"], "0"), 3)),
	}
	if month, ok := monthNumberFromName(firstPresent(values, "MMMM", "MMM")); ok {
		components.Month = month
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

func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}

	return value
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

func selectedPrecision(precision []TimeStringPrecision) TimeStringPrecision {
	if len(precision) > 0 && precision[0] == MillisecondPrecision {
		return MillisecondPrecision
	}

	return SecondPrecision
}

func monthNumberFromName(input string) (int, bool) {
	normalized := strings.TrimSuffix(strings.ToLower(input), ".")
	names := map[string]int{
		"jan":       1,
		"january":   1,
		"feb":       2,
		"february":  2,
		"mar":       3,
		"march":     3,
		"apr":       4,
		"april":     4,
		"may":       5,
		"jun":       6,
		"june":      6,
		"jul":       7,
		"july":      7,
		"aug":       8,
		"august":    8,
		"sep":       9,
		"sept":      9,
		"september": 9,
		"oct":       10,
		"october":   10,
		"nov":       11,
		"november":  11,
		"dec":       12,
		"december":  12,
	}
	month, ok := names[normalized]

	return month, ok
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
