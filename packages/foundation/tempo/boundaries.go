package tempo

import (
	"time"

	"github.com/oullin/alloy/packages/foundation/tempo/boundaries"
	"github.com/oullin/alloy/packages/foundation/tempo/internal/kernel"
)

func (tempo Time) StartOf(unit Unit, options ...StartOfWeekOptions) Time {
	return boundaries.StartOf(tempo, unit, weekOptions(options)...)
}

func (tempo Time) EndOf(unit Unit, options ...StartOfWeekOptions) Time {
	return boundaries.EndOf(tempo, unit, weekOptions(options)...)
}

func (tempo Time) IsStartOf(unit Unit, options ...StartOfWeekOptions) bool {
	return boundaries.IsStartOf(tempo, unit, weekOptions(options)...)
}

func (tempo Time) IsStartOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.IsStartOf(unit, options...)
}

func (tempo Time) IsEndOf(unit Unit, options ...StartOfWeekOptions) bool {
	return boundaries.IsEndOf(tempo, unit, weekOptions(options)...)
}

func (tempo Time) IsEndOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.IsEndOf(unit, options...)
}

func (tempo Time) IsStartOfTime() bool {
	return tempo.value.UnixMilli() <= -8640000000000000
}

func (tempo Time) IsEndOfTime() bool {
	return tempo.value.UnixMilli() >= 8640000000000000
}

func (tempo Time) IsCurrentUnit(unit Unit, reference Time) bool {
	return tempo.Same(reference, unit)
}

func (tempo Time) IsStartOfMillisecond() bool {
	return tempo.IsStartOf(Millisecond)
}

func (tempo Time) IsEndOfMillisecond() bool {
	return tempo.IsEndOf(Millisecond)
}

func (tempo Time) IsStartOfSecond() bool {
	return tempo.IsStartOf(Second)
}

func (tempo Time) IsEndOfSecond() bool {
	return tempo.IsEndOf(Second)
}

func (tempo Time) IsStartOfMinute() bool {
	return tempo.IsStartOf(Minute)
}

func (tempo Time) IsEndOfMinute() bool {
	return tempo.IsEndOf(Minute)
}

func (tempo Time) IsStartOfHour() bool {
	return tempo.IsStartOf(Hour)
}

func (tempo Time) IsEndOfHour() bool {
	return tempo.IsEndOf(Hour)
}

func (tempo Time) StartOfMillisecond() Time {
	return tempo.StartOf(Millisecond)
}

func (tempo Time) EndOfMillisecond() Time {
	return tempo.EndOf(Millisecond)
}

func (tempo Time) StartOfSecond() Time {
	return tempo.StartOf(Second)
}

func (tempo Time) EndOfSecond() Time {
	return tempo.EndOf(Second)
}

func (tempo Time) StartOfMinute() Time {
	return tempo.StartOf(Minute)
}

func (tempo Time) EndOfMinute() Time {
	return tempo.EndOf(Minute)
}

func (tempo Time) StartOfHour() Time {
	return tempo.StartOf(Hour)
}

func (tempo Time) EndOfHour() Time {
	return tempo.EndOf(Hour)
}

func (tempo Time) IsStartOfDay() bool {
	return tempo.IsStartOf(Day)
}

func (tempo Time) IsEndOfDay() bool {
	return tempo.IsEndOf(Day)
}

func (tempo Time) IsStartOfWeek(options ...StartOfWeekOptions) bool {
	return tempo.IsStartOf(Week, options...)
}

func (tempo Time) IsEndOfWeek(options ...StartOfWeekOptions) bool {
	return tempo.IsEndOf(Week, options...)
}

func (tempo Time) IsStartOfMonth() bool {
	return tempo.IsStartOf(Month)
}

func (tempo Time) IsEndOfMonth() bool {
	return tempo.IsEndOf(Month)
}

func (tempo Time) IsStartOfQuarter() bool {
	return tempo.IsStartOf(Quarter)
}

func (tempo Time) IsEndOfQuarter() bool {
	return tempo.IsEndOf(Quarter)
}

func (tempo Time) IsStartOfYear() bool {
	return tempo.IsStartOf(Year)
}

func (tempo Time) IsEndOfYear() bool {
	return tempo.IsEndOf(Year)
}

func (tempo Time) IsStartOfDecade() bool {
	return tempo.IsStartOf(Decade)
}

func (tempo Time) IsEndOfDecade() bool {
	return tempo.IsEndOf(Decade)
}

func (tempo Time) IsStartOfCentury() bool {
	return tempo.IsStartOf(Century)
}

func (tempo Time) IsEndOfCentury() bool {
	return tempo.IsEndOf(Century)
}

func (tempo Time) IsStartOfMillennium() bool {
	return tempo.IsStartOf(Millennium)
}

func (tempo Time) IsEndOfMillennium() bool {
	return tempo.IsEndOf(Millennium)
}

func (tempo Time) StartOfDay() Time {
	return tempo.StartOf(Day)
}

func (tempo Time) EndOfDay() Time {
	return tempo.EndOf(Day)
}

func (tempo Time) StartOfWeek(options ...StartOfWeekOptions) Time {
	return tempo.StartOf(Week, options...)
}

func (tempo Time) EndOfWeek(options ...StartOfWeekOptions) Time {
	return tempo.EndOf(Week, options...)
}

func (tempo Time) StartOfMonth() Time {
	return tempo.StartOf(Month)
}

func (tempo Time) EndOfMonth() Time {
	return tempo.EndOf(Month)
}

func (tempo Time) StartOfQuarter() Time {
	return tempo.StartOf(Quarter)
}

func (tempo Time) EndOfQuarter() Time {
	return tempo.EndOf(Quarter)
}

func (tempo Time) FirstOfMonth(weekdays ...time.Weekday) Time {
	return boundaries.FirstOfMonth(tempo, weekdays...)
}

func (tempo Time) LastOfMonth(weekdays ...time.Weekday) Time {
	return boundaries.LastOfMonth(tempo, weekdays...)
}

func (tempo Time) NthOfMonth(occurrence int, weekday time.Weekday) (Time, bool) {
	return boundaries.NthOfMonth(tempo, occurrence, weekday)
}

func (tempo Time) FirstOfQuarter(weekdays ...time.Weekday) Time {
	return boundaries.FirstOfQuarter(tempo, weekdays...)
}

func (tempo Time) LastOfQuarter(weekdays ...time.Weekday) Time {
	return boundaries.LastOfQuarter(tempo, weekdays...)
}

func (tempo Time) NthOfQuarter(occurrence int, weekday time.Weekday) (Time, bool) {
	return boundaries.NthOfQuarter(tempo, occurrence, weekday)
}

func (tempo Time) StartOfYear() Time {
	return tempo.StartOf(Year)
}

func (tempo Time) EndOfYear() Time {
	return tempo.EndOf(Year)
}

func (tempo Time) StartOfDecade() Time {
	return tempo.StartOf(Decade)
}

func (tempo Time) EndOfDecade() Time {
	return tempo.EndOf(Decade)
}

func (tempo Time) StartOfCentury() Time {
	return tempo.StartOf(Century)
}

func (tempo Time) EndOfCentury() Time {
	return tempo.EndOf(Century)
}

func (tempo Time) StartOfMillennium() Time {
	return tempo.StartOf(Millennium)
}

func (tempo Time) EndOfMillennium() Time {
	return tempo.EndOf(Millennium)
}

func (tempo Time) FirstOfYear(weekdays ...time.Weekday) Time {
	return boundaries.FirstOfYear(tempo, weekdays...)
}

func (tempo Time) LastOfYear(weekdays ...time.Weekday) Time {
	return boundaries.LastOfYear(tempo, weekdays...)
}

func (tempo Time) NthOfYear(occurrence int, weekday time.Weekday) (Time, bool) {
	return boundaries.NthOfYear(tempo, occurrence, weekday)
}

func (tempo Time) Floor(unit Unit) Time {
	return boundaries.Floor(tempo, unit)
}

func (tempo Time) FloorUnit(unit Unit) Time {
	return tempo.Floor(unit)
}

func (tempo Time) FloorWeek(options ...StartOfWeekOptions) Time {
	return boundaries.FloorWeek(tempo, weekOptions(options)...)
}

func (tempo Time) Ceil(unit Unit) Time {
	return boundaries.Ceil(tempo, unit)
}

func (tempo Time) CeilUnit(unit Unit) Time {
	return tempo.Ceil(unit)
}

func (tempo Time) CeilWeek(options ...StartOfWeekOptions) Time {
	return boundaries.CeilWeek(tempo, weekOptions(options)...)
}

func (tempo Time) Round(unit Unit) Time {
	return boundaries.Round(tempo, unit)
}

func (tempo Time) RoundUnit(unit Unit) Time {
	return tempo.Round(unit)
}

func (tempo Time) RoundWeek(options ...StartOfWeekOptions) Time {
	return boundaries.RoundWeek(tempo, weekOptions(options)...)
}

func (tempo Time) Next(weekday time.Weekday) Time {
	return boundaries.Next(tempo, weekday)
}

func (tempo Time) Previous(weekday time.Weekday) Time {
	return boundaries.Previous(tempo, weekday)
}

func (tempo Time) NextOrSame(weekday time.Weekday) Time {
	return boundaries.NextOrSame(tempo, weekday)
}

func (tempo Time) PreviousOrSame(weekday time.Weekday) Time {
	return boundaries.PreviousOrSame(tempo, weekday)
}

func (tempo Time) NextWeekday() Time {
	return boundaries.NextWeekday(tempo, tempo.settingsSnapshot().WeekendDays)
}

func (tempo Time) PreviousWeekday() Time {
	return boundaries.PreviousWeekday(tempo, tempo.settingsSnapshot().WeekendDays)
}

func (tempo Time) NextWeekendDay() Time {
	return boundaries.NextWeekendDay(tempo, tempo.settingsSnapshot().WeekendDays)
}

func (tempo Time) PreviousWeekendDay() Time {
	return boundaries.PreviousWeekendDay(tempo, tempo.settingsSnapshot().WeekendDays)
}

func weekOptions(options []StartOfWeekOptions) []kernel.WeekOptions {
	result := make([]kernel.WeekOptions, 0, len(options))

	for _, option := range options {
		result = append(result, kernel.WeekOptions{WeekStartsOn: option.WeekStartsOn})
	}

	return result
}
