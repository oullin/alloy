package tempo

import (
	"time"

	"alloy.dev/backend/tempo/diff"
)

func (tempo Time) Diff(other Time, unit Unit, options ...DiffOptions) float64 {
	return diff.Between(tempo, other.State(), unit, diffOptions(options)...)
}

func (tempo Time) DiffAsDuration(other Time, options ...DiffOptions) Duration {
	value := Duration{Milliseconds: tempo.DiffInMilliseconds(other, options...)}

	return value.Normalize()
}

func (tempo Time) DiffAsDateInterval(other Time, options ...DiffOptions) Duration {
	return tempo.DiffAsDuration(other, options...)
}

func (tempo Time) DiffAsTempoInterval(other Time, options ...DiffOptions) Duration {
	return tempo.DiffAsDuration(other, options...)
}

func (tempo Time) DiffInMilliseconds(other Time, options ...DiffOptions) int {
	return diff.InMilliseconds(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInMicroseconds(other Time, options ...DiffOptions) int {
	return diff.InMicroseconds(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInSeconds(other Time, options ...DiffOptions) int {
	return diff.InSeconds(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInMinutes(other Time, options ...DiffOptions) int {
	return diff.InMinutes(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInHours(other Time, options ...DiffOptions) int {
	return diff.InHours(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInDays(other Time, options ...DiffOptions) int {
	return diff.InDays(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInWeeks(other Time, options ...DiffOptions) int {
	return diff.InWeeks(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInWeekdays(other Time, options ...DiffOptions) int {
	return diff.InWeekdays(tempo, other.State(), tempo.settingsSnapshot().WeekendDays, diffOptions(options)...)
}

func (tempo Time) DiffInWeekendDays(other Time, options ...DiffOptions) int {
	return diff.InWeekendDays(tempo, other.State(), tempo.settingsSnapshot().WeekendDays, diffOptions(options)...)
}

func (tempo Time) DiffInMonths(other Time, options ...DiffOptions) int {
	return diff.InMonths(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInQuarters(other Time, options ...DiffOptions) int {
	return diff.InQuarters(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInYears(other Time, options ...DiffOptions) int {
	return diff.InYears(tempo, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInUnit(unit Unit, other Time, options ...DiffOptions) int {
	return diff.InUnit(tempo, unit, other.State(), diffOptions(options)...)
}

func (tempo Time) DiffInDaysFiltered(other Time, predicate func(Time) bool, options ...DiffOptions) int {
	return diff.InDaysFiltered(tempo, other.State(), predicate, diffOptions(options)...)
}

func (tempo Time) DiffFiltered(other Time, predicate func(Time) bool, options ...DiffOptions) int {
	return tempo.DiffInDaysFiltered(other, predicate, options...)
}

func (tempo Time) DiffInHoursFiltered(other Time, predicate func(Time) bool, options ...DiffOptions) int {
	return diff.InHoursFiltered(tempo, other.State(), predicate, diffOptions(options)...)
}

func (tempo Time) SecondsSinceMidnight() int {
	return diff.SecondsSinceMidnight(tempo)
}

func (tempo Time) SecondsUntilEndOfDay() int {
	return diff.SecondsUntilEndOfDay(tempo)
}

func (tempo Time) Calendar(reference Time, formats ...map[string]string) string {
	value := tempo.StartOfDay().DiffInDays(reference.StartOfDay())
	key := "sameElse"

	switch {
	case value == 0:
		key = "sameDay"
	case value == 1:
		key = "nextDay"
	case value > 1 && value < 7:
		key = "nextWeek"
	case value == -1:
		key = "lastDay"
	case value < -1 && value > -7:
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

	return tempo.Format(pattern)
}

func (tempo Time) DiffForHumans(other Time, options ...HumanDiffOptions) string {
	opts := tempo.settingsSnapshot().HumanDiff

	if len(options) > 0 {
		opts = options[0]
	}

	return diff.ForHumans(tempo, other.State(), diff.HumanOptions{Absolute: opts.Absolute, Unit: opts.Unit})
}

func (tempo Time) From(other Time, options ...HumanDiffOptions) string {
	return tempo.DiffForHumans(other, options...)
}

func (tempo Time) Since(other Time, options ...HumanDiffOptions) string {
	return tempo.From(other, options...)
}

func (tempo Time) To(other Time, options ...HumanDiffOptions) string {
	return other.DiffForHumans(tempo, options...)
}

func (tempo Time) FromNow(options ...HumanDiffOptions) string {
	return tempo.DiffForHumans(Time{value: time.Now().UTC(), location: tempo.location}, options...)
}

func (tempo Time) ToNow(options ...HumanDiffOptions) string {
	return Time{value: time.Now().UTC(), location: tempo.location}.DiffForHumans(tempo, options...)
}

func (tempo Time) Ago(options ...HumanDiffOptions) string {
	return tempo.FromNow(options...)
}

func (tempo Time) Timespan(other Time, options ...HumanDiffOptions) string {
	opts := HumanDiffOptions{Absolute: true}

	if len(options) > 0 {
		opts = options[0]
		opts.Absolute = true
	}

	return tempo.DiffForHumans(other, opts)
}

func diffOptions(options []DiffOptions) []diff.Options {
	result := make([]diff.Options, 0, len(options))

	for _, option := range options {
		result = append(result, diff.Options{Absolute: option.Absolute, Float: option.Float})
	}

	return result
}
