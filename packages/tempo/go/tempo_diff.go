package tempo

import (
	"fmt"
	"math"
	"strings"
	"time"
)

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

func (tempo Tempo) DiffAsDuration(other Tempo, options ...DiffOptions) Duration {
	return Duration{Milliseconds: tempo.DiffInMilliseconds(other, options...)}.Normalize()
}

func (tempo Tempo) DiffAsDateInterval(other Tempo, options ...DiffOptions) Duration {
	return tempo.DiffAsDuration(other, options...)
}

func (tempo Tempo) DiffAsTempoInterval(other Tempo, options ...DiffOptions) Duration {
	return tempo.DiffAsDuration(other, options...)
}

func (tempo Tempo) DiffInMilliseconds(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Millisecond, options...))
}

func (tempo Tempo) DiffInMicroseconds(other Tempo, options ...DiffOptions) int {
	return tempo.DiffInMilliseconds(other, options...) * 1000
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

func (tempo Tempo) DiffInQuarters(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Quarter, options...))
}

func (tempo Tempo) DiffInYears(other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, Year, options...))
}

func (tempo Tempo) DiffInUnit(unit Unit, other Tempo, options ...DiffOptions) int {
	return int(tempo.Diff(other, unit, options...))
}

func (tempo Tempo) DiffInDaysFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return tempo.diffFilteredDays(other, predicate, options...)
}

func (tempo Tempo) DiffFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return tempo.DiffInDaysFiltered(other, predicate, options...)
}

func (tempo Tempo) DiffInHoursFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	opts := DiffOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	sign := 1
	start := other.StartOfHour()
	end := tempo.StartOfHour()
	if tempo.Before(other, Hour) {
		sign = -1
		start = tempo.StartOfHour()
		end = other.StartOfHour()
	}

	count := 0
	current := start
	for current.Before(end, Hour) {
		current = current.AddHours(1)
		if current.SameOrBefore(end, Hour) && predicate(current) {
			count++
		}
	}

	if opts.Absolute || sign > 0 {
		return count
	}

	return -count
}

func (tempo Tempo) SecondsSinceMidnight() int {
	return tempo.DiffInSeconds(tempo.StartOfDay(), DiffOptions{Absolute: true})
}

func (tempo Tempo) SecondsUntilEndOfDay() int {
	return tempo.DiffInSeconds(tempo.EndOfDay(), DiffOptions{Absolute: true})
}

func (tempo Tempo) Calendar(reference Tempo, formats ...map[string]string) string {
	diff := tempo.StartOfDay().DiffInDays(reference.StartOfDay())
	key := "sameElse"
	switch {
	case diff == 0:
		key = "sameDay"
	case diff == 1:
		key = "nextDay"
	case diff > 1 && diff < 7:
		key = "nextWeek"
	case diff == -1:
		key = "lastDay"
	case diff < -1 && diff > -7:
		key = "lastWeek"
	}

	defaults := map[string]string{
		"lastDay":  "[Yesterday at] HH:mm",
		"lastWeek": "[Last] dddd [at] HH:mm",
		"nextDay":  "[Tomorrow at] HH:mm",
		"nextWeek": "dddd [at] HH:mm",
		"sameDay":  "[Today at] HH:mm",
		"sameElse": "YYYY-MM-DD",
	}
	pattern := defaults[key]
	if len(formats) > 0 {
		if custom, ok := formats[0][key]; ok {
			pattern = custom
		}
	}

	return tempo.ISOFormat(pattern)
}

func (tempo Tempo) DiffForHumans(other Tempo, options ...HumanDiffOptions) string {
	opts := tempoSettings.HumanDiff
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

func (tempo Tempo) From(other Tempo, options ...HumanDiffOptions) string {
	return tempo.DiffForHumans(other, options...)
}

func (tempo Tempo) Since(other Tempo, options ...HumanDiffOptions) string {
	return tempo.From(other, options...)
}

func (tempo Tempo) To(other Tempo, options ...HumanDiffOptions) string {
	return other.DiffForHumans(tempo, options...)
}

func (tempo Tempo) FromNow(options ...HumanDiffOptions) string {
	return tempo.DiffForHumans(Tempo{value: time.Now().UTC(), location: tempo.location}, options...)
}

func (tempo Tempo) ToNow(options ...HumanDiffOptions) string {
	return Tempo{value: time.Now().UTC(), location: tempo.location}.DiffForHumans(tempo, options...)
}

func (tempo Tempo) Ago(options ...HumanDiffOptions) string {
	return tempo.FromNow(options...)
}

func (tempo Tempo) Timespan(other Tempo, options ...HumanDiffOptions) string {
	opts := HumanDiffOptions{Absolute: true}
	if len(options) > 0 {
		opts = options[0]
		opts.Absolute = true
	}

	return tempo.DiffForHumans(other, opts)
}
