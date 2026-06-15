package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/diff"
)

func (tempo Tempo) Diff(other Tempo, unit Unit, options ...DiffOptions) float64 {
	return diff.Diff(tempo, other.State(), unit, diffOptions(options)...)
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
	return diff.DiffInMilliseconds(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInMicroseconds(other Tempo, options ...DiffOptions) int {
	return diff.DiffInMicroseconds(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInSeconds(other Tempo, options ...DiffOptions) int {
	return diff.DiffInSeconds(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInMinutes(other Tempo, options ...DiffOptions) int {
	return diff.DiffInMinutes(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInHours(other Tempo, options ...DiffOptions) int {
	return diff.DiffInHours(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInDays(other Tempo, options ...DiffOptions) int {
	return diff.DiffInDays(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInWeeks(other Tempo, options ...DiffOptions) int {
	return diff.DiffInWeeks(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInWeekdays(other Tempo, options ...DiffOptions) int {
	return diff.DiffInWeekdays(tempo, other.State(), defaultConfig.Settings.WeekendDays, diffOptions(options)...)
}

func (tempo Tempo) DiffInWeekendDays(other Tempo, options ...DiffOptions) int {
	return diff.DiffInWeekendDays(tempo, other.State(), defaultConfig.Settings.WeekendDays, diffOptions(options)...)
}

func (tempo Tempo) DiffInMonths(other Tempo, options ...DiffOptions) int {
	return diff.DiffInMonths(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInQuarters(other Tempo, options ...DiffOptions) int {
	return diff.DiffInQuarters(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInYears(other Tempo, options ...DiffOptions) int {
	return diff.DiffInYears(tempo, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInUnit(unit Unit, other Tempo, options ...DiffOptions) int {
	return diff.DiffInUnit(tempo, unit, other.State(), diffOptions(options)...)
}

func (tempo Tempo) DiffInDaysFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return diff.DiffInDaysFiltered(tempo, other.State(), predicate, diffOptions(options)...)
}

func (tempo Tempo) DiffFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return tempo.DiffInDaysFiltered(other, predicate, options...)
}

func (tempo Tempo) DiffInHoursFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return diff.DiffInHoursFiltered(tempo, other.State(), predicate, diffOptions(options)...)
}

func (tempo Tempo) SecondsSinceMidnight() int {
	return diff.SecondsSinceMidnight(tempo)
}

func (tempo Tempo) SecondsUntilEndOfDay() int {
	return diff.SecondsUntilEndOfDay(tempo)
}

func (tempo Tempo) Calendar(reference Tempo, formats ...map[string]string) string {
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

	return tempo.ISOFormat(pattern)
}

func (tempo Tempo) DiffForHumans(other Tempo, options ...HumanDiffOptions) string {
	opts := defaultConfig.Settings.HumanDiff

	if len(options) > 0 {
		opts = options[0]
	}

	return diff.ForHumans(tempo, other.State(), diff.HumanOptions{Absolute: opts.Absolute, Unit: opts.Unit})
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

func diffOptions(options []DiffOptions) []diff.Options {
	result := make([]diff.Options, 0, len(options))

	for _, option := range options {
		result = append(result, diff.Options{Absolute: option.Absolute, Float: option.Float})
	}

	return result
}
