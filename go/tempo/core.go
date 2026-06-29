package tempo

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	setterspkg "alloy.dev/api/tempo/setters"
)

func (tempo Time) Clone() Time {
	return tempo
}

func (tempo Time) Context() Context {
	if tempo.runtime.Locale() == "" {
		settings := tempo.settingsSnapshot()

		return NewRuntime(
			RuntimeLocale(settings.Locale),
			RuntimeFallbackLocale(settings.FallbackLocale),
		)
	}

	return tempo.runtime
}

func (tempo Time) WithRuntime(runtime Context) Time {
	tempo.runtime = runtime

	return tempo
}

func (tempo Time) WithTranslator(translator Translator) Time {
	return tempo.WithRuntime(tempo.Context().With(RuntimeTranslator(translator)))
}

func (tempo Time) HasTranslator() bool {
	return tempo.Context().HasTranslator()
}

func (tempo Time) AvoidMutation() Time {
	return tempo.Clone()
}

func (tempo Time) Cast() Time {
	return tempo.Clone()
}

func (tempo Time) Tempoize(input Time) Time {
	return newTempo(input.value, input.location, tempo.Context())
}

func (tempo Time) NowWithSameTz() Time {
	return newTempo(time.Now(), tempo.location, tempo.Context())
}

func (tempo Time) Modify(modifier string) (Time, error) {
	value := strings.TrimSpace(modifier)

	if value == "" {
		return Time{}, errors.New("Time modifier cannot be empty")
	}

	if parsed, err := Parse(value, WithTimezone(tempo.Timezone())); err == nil {
		return parsed, nil
	}

	lower := strings.ToLower(value)

	if match := modifierPattern.FindStringSubmatch(lower); match != nil {
		amount, err := strconv.ParseFloat(match[1], 64)

		if err != nil {
			return Time{}, err
		}

		return tempo.Add(int(math.Round(amount)), Unit(match[2])), nil
	}

	if match := movePattern.FindStringSubmatch(lower); match != nil {
		amount := 1

		if match[1] != "next" {
			amount = -1
		}

		return tempo.Add(amount, Unit(match[2])), nil
	}

	switch lower {
	case "now":
		return tempo.NowWithSameTz(), nil
	case "today":
		return tempo.NowWithSameTz().StartOfDay(), nil
	case "tomorrow":
		return tempo.NowWithSameTz().StartOfDay().AddDays(1), nil
	case "yesterday":
		return tempo.NowWithSameTz().StartOfDay().SubDays(1), nil
	default:
		return Time{}, fmt.Errorf("invalid Time modifier: %s", modifier)
	}
}

func (tempo Time) Change(modifier string) (Time, error) {
	return tempo.Modify(modifier)
}

func (tempo Time) Immutable() Time {
	return tempo.Clone()
}

func (tempo Time) Mutable() *MutableTime {
	return NewMutable(tempo)
}

func (tempo Time) Timezone() string {
	return tempo.location.String()
}

func (tempo Time) Timestamp() int64 {
	return tempo.value.Unix()
}

func (tempo Time) TimestampMs() int64 {
	return tempo.value.UnixMilli()
}

func (tempo Time) PreciseTimestamp(precisions ...int) float64 {
	precision := 6

	if len(precisions) > 0 {
		precision = precisions[0]
	}

	return math.Round(float64(tempo.value.UnixNano()) / math.Pow10(9-precision))
}

func (tempo Time) Unix() int64 {
	return tempo.Timestamp()
}

func (tempo Time) Year() int {
	return tempo.local().Year()
}

func (tempo Time) Month() int {
	return int(tempo.local().Month())
}

func (tempo Time) Quarter() int {
	return (tempo.Month()-1)/3 + 1
}

func (tempo Time) Day() int {
	return tempo.local().Day()
}

func (tempo Time) DayOfWeek() int {
	return int(tempo.local().Weekday())
}

func (tempo Time) Weekday() int {
	return tempo.DayOfWeek()
}

func (tempo Time) SetWeekday(weekday time.Weekday) Time {
	return tempo.AddDays(int(weekday) - tempo.DayOfWeek())
}

func (tempo Time) ISOWeekday() int {
	weekday := tempo.local().Weekday()

	if weekday == time.Sunday {
		return 7
	}

	return int(weekday)
}

func (tempo Time) ISOWeek() (int, int) {
	year, week := tempo.local().ISOWeek()

	return year, week
}

func (tempo Time) ISOWeekYear() int {
	year, _ := tempo.ISOWeek()

	return year
}

func (tempo Time) ISOWeekNumber() int {
	_, week := tempo.ISOWeek()

	return week
}

func (tempo Time) WeeksInISOYear() int {
	_, week := time.Date(tempo.ISOWeekYear(), time.December, 28, 0, 0, 0, 0, tempo.location).ISOWeek()

	return week
}

func (tempo Time) DayOfYear() int {
	return tempo.local().YearDay()
}

func (tempo Time) SetDayOfYear(day int) Time {
	return setterspkg.SetTime(
		tempo.StartOfYear().AddDays(day-1),
		tempo.Hour(),
		tempo.Minute(),
		tempo.Second(),
		tempo.Millisecond(),
	)
}

func (tempo Time) Hour() int {
	return tempo.local().Hour()
}

func (tempo Time) Minute() int {
	return tempo.local().Minute()
}

func (tempo Time) Second() int {
	return tempo.local().Second()
}

func (tempo Time) Millisecond() int {
	return tempo.local().Nanosecond() / int(time.Millisecond)
}

func (tempo Time) fieldValue(field string) (any, bool) {
	values := tempo.ToMap()
	value, ok := values[field]

	if ok {
		return value, true
	}

	switch field {
	case "timestamp":
		return tempo.Timestamp(), true
	case "timestampMs":
		return tempo.TimestampMs(), true
	default:
		return nil, false
	}
}

func (tempo Time) PaddedUnit(field string, length int) (string, bool) {
	value, ok := tempo.fieldValue(field)

	if !ok {
		return "", false
	}

	number, ok := value.(int)

	if !ok {
		return "", false
	}

	return fmt.Sprintf("%0*d", length, number), true
}

func (tempo Time) OffsetMinutes() int {
	_, offset := tempo.local().Zone()

	return offset / 60
}

func (tempo Time) OffsetString(separator string) string {
	return formatOffset(tempo.OffsetMinutes(), separator)
}

func (tempo Time) UTCOffset() int {
	return tempo.OffsetMinutes()
}

func (tempo Time) ZoneName() string {
	name, _ := tempo.local().Zone()

	return name
}

func (tempo Time) MonthName() string {
	return calendarMonthName(tempo.Month())
}

func (tempo Time) ShortMonthName() string {
	return calendarShortMonthName(tempo.Month())
}

func (tempo Time) DayName() string {
	return calendarDayName(int(tempo.local().Weekday()))
}

func (tempo Time) ShortDayName() string {
	return calendarShortDayName(int(tempo.local().Weekday()))
}

func (tempo Time) MinDayName() string {
	name := tempo.ShortDayName()

	if len(name) < 2 {
		return name
	}

	return name[:2]
}

func (tempo Time) TranslateNumber(value int) string {
	return strconv.Itoa(value)
}

func (tempo Time) Translate(message string, replacements map[string]string) string {
	if translated, ok := tempo.Context().Translate(message, replacements); ok {
		return translated
	}

	return tempo.TranslateWith(message, replacements)
}

func (tempo Time) TranslateWith(message string, replacements map[string]string) string {
	return replaceTranslationTokens(message, replacements)
}

func (tempo Time) TranslationMessage(key string) (any, bool) {
	return tempo.Context().Message(key)
}

func (tempo Time) IsUTC() bool {
	return tempo.location == time.UTC && tempo.OffsetMinutes() == 0
}

func (tempo Time) IsLocal() bool {
	return tempo.location.String() == time.Local.String()
}

func (tempo Time) IsDST() bool {
	local := tempo.local()
	january := time.Date(local.Year(), time.January, 1, 12, 0, 0, 0, tempo.location)
	july := time.Date(local.Year(), time.July, 1, 12, 0, 0, 0, tempo.location)
	_, januaryOffset := january.Zone()
	_, julyOffset := july.Zone()
	standardOffset := min(januaryOffset, julyOffset)
	_, currentOffset := local.Zone()

	return currentOffset > standardOffset
}

func (tempo Time) IsLeapYear() bool {
	year := tempo.Year()

	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func (tempo Time) DaysInYear() int {
	if tempo.IsLeapYear() {
		return 366
	}

	return 365
}

func (tempo Time) IsLongYear() bool {
	return tempo.WeeksInISOYear() == 53
}

func (tempo Time) IsLongISOYear() bool {
	return tempo.IsLongYear()
}

func (tempo Time) IsLastOfMonth() bool {
	return tempo.Day() == tempo.DaysInMonth()
}

func (tempo Time) DaysInMonth() int {
	return daysInMonth(tempo.Year(), tempo.Month())
}

func (tempo Time) IsWeekend() bool {
	weekday := tempo.local().Weekday()

	return slices.Contains(tempo.settingsSnapshot().WeekendDays, weekday)
}

func (tempo Time) IsSunday() bool {
	return tempo.local().Weekday() == time.Sunday
}

func (tempo Time) IsMonday() bool {
	return tempo.local().Weekday() == time.Monday
}

func (tempo Time) IsTuesday() bool {
	return tempo.local().Weekday() == time.Tuesday
}

func (tempo Time) IsWednesday() bool {
	return tempo.local().Weekday() == time.Wednesday
}

func (tempo Time) IsThursday() bool {
	return tempo.local().Weekday() == time.Thursday
}

func (tempo Time) IsFriday() bool {
	return tempo.local().Weekday() == time.Friday
}

func (tempo Time) IsSaturday() bool {
	return tempo.local().Weekday() == time.Saturday
}

func (tempo Time) IsDayOfWeek(weekday time.Weekday) bool {
	return tempo.local().Weekday() == weekday
}

func (tempo Time) IsWeekday() bool {
	return !tempo.IsWeekend()
}

func (tempo Time) IsPast(reference Time) bool {
	return tempo.Before(reference)
}

func (tempo Time) IsFuture(reference Time) bool {
	return tempo.After(reference)
}

func (tempo Time) IsNowOrPast() bool {
	return tempo.SameOrBefore(Time{value: time.Now().UTC(), location: tempo.location})
}

func (tempo Time) IsNowOrFuture() bool {
	return tempo.SameOrAfter(Time{value: time.Now().UTC(), location: tempo.location})
}

func (tempo Time) IsToday(reference Time) bool {
	return tempo.Same(reference, Day)
}

func (tempo Time) IsTomorrow(reference Time) bool {
	return tempo.Same(reference.AddDays(1), Day)
}

func (tempo Time) IsYesterday(reference Time) bool {
	return tempo.Same(reference.SubDays(1), Day)
}

func (tempo Time) IsMidnight() bool {
	return tempo.Hour() == 0 &&
		tempo.Minute() == 0 &&
		tempo.Second() == 0 &&
		tempo.Millisecond() == 0
}

func (tempo Time) IsMidday() bool {
	return tempo.Hour() == tempo.settingsSnapshot().MidDayAt &&
		tempo.Minute() == 0 &&
		tempo.Second() == 0 &&
		tempo.Millisecond() == 0
}
