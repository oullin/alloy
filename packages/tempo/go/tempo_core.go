package tempo

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func (tempo Tempo) Clone() Tempo {
	return tempo
}

func (tempo Tempo) Runtime() Runtime {
	if tempo.runtime.Locale() == "" {
		return NewRuntime(
			RuntimeLocale(defaultConfig.Settings.Locale),
			RuntimeFallbackLocale(defaultConfig.Settings.FallbackLocale),
		)
	}

	return tempo.runtime
}

func (tempo Tempo) WithRuntime(runtime Runtime) Tempo {
	tempo.runtime = runtime
	return tempo
}

func (tempo Tempo) GetLocalTranslator() Translator {
	return tempo.Runtime().Translator()
}

func (tempo Tempo) SetLocalTranslator(translator Translator) Tempo {
	return tempo.WithRuntime(tempo.Runtime().With(RuntimeTranslator(translator)))
}

func (tempo Tempo) WithTranslator(translator Translator) Tempo {
	return tempo.SetLocalTranslator(translator)
}

func (tempo Tempo) HasLocalTranslator() bool {
	return tempo.Runtime().HasTranslator()
}

func (tempo Tempo) AvoidMutation() Tempo {
	return tempo.Clone()
}

func (tempo Tempo) Cast() Tempo {
	return tempo.Clone()
}

func (tempo Tempo) Tempoize(input Tempo) Tempo {
	return newTempo(input.value, input.location, tempo.Runtime())
}

func (tempo Tempo) NowWithSameTz() Tempo {
	return newTempo(time.Now(), tempo.location, tempo.Runtime())
}

func (tempo Tempo) Modify(modifier string) (Tempo, error) {
	value := strings.TrimSpace(modifier)
	if value == "" {
		return Tempo{}, errors.New("Tempo modifier cannot be empty")
	}

	if parsed, err := Parse(value, WithTimezone(tempo.Timezone())); err == nil {
		return parsed, nil
	}

	lower := strings.ToLower(value)
	if match := modifierPattern.FindStringSubmatch(lower); match != nil {
		amount, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return Tempo{}, err
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
		return Tempo{}, fmt.Errorf("invalid Tempo modifier: %s", modifier)
	}
}

func (tempo Tempo) Change(modifier string) (Tempo, error) {
	return tempo.Modify(modifier)
}

func (tempo Tempo) Immutable() Tempo {
	return tempo.Clone()
}

func (tempo Tempo) Mutable() *MutableTempo {
	return NewMutable(tempo)
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

func (tempo Tempo) GetTimestampMs() int64 {
	return tempo.TimestampMs()
}

func (tempo Tempo) GetPreciseTimestamp(precisions ...int) float64 {
	precision := 6
	if len(precisions) > 0 {
		precision = precisions[0]
	}

	return math.Round(float64(tempo.value.UnixNano()) / math.Pow10(9-precision))
}

func (tempo Tempo) Unix() int64 {
	return tempo.Timestamp()
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

func (tempo Tempo) Weekday() int {
	return tempo.DayOfWeek()
}

func (tempo Tempo) SetWeekday(weekday time.Weekday) Tempo {
	return tempo.AddDays(int(weekday) - tempo.DayOfWeek())
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

func (tempo Tempo) SetDayOfYear(day int) Tempo {
	return tempo.StartOfYear().
		AddDays(day-1).
		SetTime(tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
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

func (tempo Tempo) Get(field string) (any, bool) {
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

func (tempo Tempo) GetPaddedUnit(field string, length int) (string, bool) {
	value, ok := tempo.Get(field)
	if !ok {
		return "", false
	}

	number, ok := value.(int)
	if !ok {
		return "", false
	}

	return fmt.Sprintf("%0*d", length, number), true
}

func (tempo Tempo) OffsetMinutes() int {
	_, offset := tempo.local().Zone()
	return offset / 60
}

func (tempo Tempo) OffsetString(separator string) string {
	return formatOffset(tempo.OffsetMinutes(), separator)
}

func (tempo Tempo) GetOffsetString(separator string) string {
	return tempo.OffsetString(separator)
}

func (tempo Tempo) UTCOffset() int {
	return tempo.OffsetMinutes()
}

func (tempo Tempo) ZoneName() string {
	name, _ := tempo.local().Zone()
	return name
}

func (tempo Tempo) MonthName() string {
	return monthNames[tempo.Month()-1]
}

func (tempo Tempo) ShortMonthName() string {
	return shortMonthNames[tempo.Month()-1]
}

func (tempo Tempo) DayName() string {
	return dayNames[int(tempo.local().Weekday())]
}

func (tempo Tempo) ShortDayName() string {
	return shortDayNames[int(tempo.local().Weekday())]
}

func (tempo Tempo) MinDayName() string {
	name := tempo.ShortDayName()
	if len(name) < 2 {
		return name
	}

	return name[:2]
}

func (tempo Tempo) GetTranslatedMonthName() string { return tempo.MonthName() }

func (tempo Tempo) GetTranslatedShortMonthName() string { return tempo.ShortMonthName() }

func (tempo Tempo) GetTranslatedDayName() string { return tempo.DayName() }

func (tempo Tempo) GetTranslatedShortDayName() string { return tempo.ShortDayName() }

func (tempo Tempo) GetTranslatedMinDayName() string { return tempo.MinDayName() }

func (tempo Tempo) TranslateNumber(value int) string {
	return strconv.Itoa(value)
}

func (tempo Tempo) GetAltNumber(value int) string {
	return tempo.TranslateNumber(value)
}

func (tempo Tempo) Translate(message string, replacements map[string]string) string {
	if translated, ok := tempo.Runtime().Translate(message, replacements); ok {
		return translated
	}

	return tempo.TranslateWith(message, replacements)
}

func (tempo Tempo) TranslateWith(message string, replacements map[string]string) string {
	return replaceTranslationTokens(message, replacements)
}

func (tempo Tempo) GetTranslationMessage(key string) (any, bool) {
	return tempo.Runtime().Message(key)
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

func (tempo Tempo) DaysInYear() int {
	if tempo.IsLeapYear() {
		return 366
	}

	return 365
}

func (tempo Tempo) IsLongYear() bool {
	return tempo.WeeksInISOYear() == 53
}

func (tempo Tempo) IsLongISOYear() bool {
	return tempo.IsLongYear()
}

func (tempo Tempo) IsLastOfMonth() bool {
	return tempo.Day() == tempo.DaysInMonth()
}

func (tempo Tempo) DaysInMonth() int {
	return daysInMonth(tempo.Year(), tempo.Month())
}

func (tempo Tempo) IsWeekend() bool {
	weekday := tempo.local().Weekday()
	for _, weekendDay := range defaultConfig.Settings.WeekendDays {
		if weekday == weekendDay {
			return true
		}
	}

	return false
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

func (tempo Tempo) IsDayOfWeek(weekday time.Weekday) bool {
	return tempo.local().Weekday() == weekday
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

func (tempo Tempo) IsNowOrPast() bool {
	return tempo.SameOrBefore(Tempo{value: time.Now().UTC(), location: tempo.location})
}

func (tempo Tempo) IsNowOrFuture() bool {
	return tempo.SameOrAfter(Tempo{value: time.Now().UTC(), location: tempo.location})
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

func (tempo Tempo) IsMidnight() bool {
	return tempo.Hour() == 0 &&
		tempo.Minute() == 0 &&
		tempo.Second() == 0 &&
		tempo.Millisecond() == 0
}

func (tempo Tempo) IsMidday() bool {
	return tempo.Hour() == defaultConfig.Settings.MidDayAt &&
		tempo.Minute() == 0 &&
		tempo.Second() == 0 &&
		tempo.Millisecond() == 0
}
