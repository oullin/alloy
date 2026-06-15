package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/boundaries"
	"github.com/oullin/alloy/tempo/internal/kernel"
)

func (tempo Tempo) StartOf(unit Unit, options ...StartOfWeekOptions) Tempo {
	return boundaries.StartOf(tempo, unit, weekOptions(options)...)
}

func (tempo Tempo) EndOf(unit Unit, options ...StartOfWeekOptions) Tempo {
	return boundaries.EndOf(tempo, unit, weekOptions(options)...)
}

func (tempo Tempo) IsStartOf(unit Unit, options ...StartOfWeekOptions) bool {
	return boundaries.IsStartOf(tempo, unit, weekOptions(options)...)
}

func (tempo Tempo) IsStartOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.IsStartOf(unit, options...)
}

func (tempo Tempo) IsEndOf(unit Unit, options ...StartOfWeekOptions) bool {
	return boundaries.IsEndOf(tempo, unit, weekOptions(options)...)
}

func (tempo Tempo) IsEndOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.IsEndOf(unit, options...)
}

func (tempo Tempo) IsStartOfTime() bool {
	return tempo.value.UnixMilli() <= -8640000000000000
}

func (tempo Tempo) IsEndOfTime() bool {
	return tempo.value.UnixMilli() >= 8640000000000000
}

func (tempo Tempo) IsCurrentUnit(unit Unit, reference Tempo) bool {
	return tempo.Same(reference, unit)
}

func (tempo Tempo) IsStartOfMillisecond() bool {
	return tempo.IsStartOf(Millisecond)
}

func (tempo Tempo) IsEndOfMillisecond() bool {
	return tempo.IsEndOf(Millisecond)
}

func (tempo Tempo) IsStartOfSecond() bool {
	return tempo.IsStartOf(Second)
}

func (tempo Tempo) IsEndOfSecond() bool {
	return tempo.IsEndOf(Second)
}

func (tempo Tempo) IsStartOfMinute() bool {
	return tempo.IsStartOf(Minute)
}

func (tempo Tempo) IsEndOfMinute() bool {
	return tempo.IsEndOf(Minute)
}

func (tempo Tempo) IsStartOfHour() bool {
	return tempo.IsStartOf(Hour)
}

func (tempo Tempo) IsEndOfHour() bool {
	return tempo.IsEndOf(Hour)
}

func (tempo Tempo) StartOfMillisecond() Tempo {
	return tempo.StartOf(Millisecond)
}

func (tempo Tempo) EndOfMillisecond() Tempo {
	return tempo.EndOf(Millisecond)
}

func (tempo Tempo) StartOfSecond() Tempo {
	return tempo.StartOf(Second)
}

func (tempo Tempo) EndOfSecond() Tempo {
	return tempo.EndOf(Second)
}

func (tempo Tempo) StartOfMinute() Tempo {
	return tempo.StartOf(Minute)
}

func (tempo Tempo) EndOfMinute() Tempo {
	return tempo.EndOf(Minute)
}

func (tempo Tempo) StartOfHour() Tempo {
	return tempo.StartOf(Hour)
}

func (tempo Tempo) EndOfHour() Tempo {
	return tempo.EndOf(Hour)
}

func (tempo Tempo) IsStartOfDay() bool {
	return tempo.IsStartOf(Day)
}

func (tempo Tempo) IsEndOfDay() bool {
	return tempo.IsEndOf(Day)
}

func (tempo Tempo) IsStartOfWeek(options ...StartOfWeekOptions) bool {
	return tempo.IsStartOf(Week, options...)
}

func (tempo Tempo) IsEndOfWeek(options ...StartOfWeekOptions) bool {
	return tempo.IsEndOf(Week, options...)
}

func (tempo Tempo) IsStartOfMonth() bool {
	return tempo.IsStartOf(Month)
}

func (tempo Tempo) IsEndOfMonth() bool {
	return tempo.IsEndOf(Month)
}

func (tempo Tempo) IsStartOfQuarter() bool {
	return tempo.IsStartOf(Quarter)
}

func (tempo Tempo) IsEndOfQuarter() bool {
	return tempo.IsEndOf(Quarter)
}

func (tempo Tempo) IsStartOfYear() bool {
	return tempo.IsStartOf(Year)
}

func (tempo Tempo) IsEndOfYear() bool {
	return tempo.IsEndOf(Year)
}

func (tempo Tempo) IsStartOfDecade() bool {
	return tempo.IsStartOf(Decade)
}

func (tempo Tempo) IsEndOfDecade() bool {
	return tempo.IsEndOf(Decade)
}

func (tempo Tempo) IsStartOfCentury() bool {
	return tempo.IsStartOf(Century)
}

func (tempo Tempo) IsEndOfCentury() bool {
	return tempo.IsEndOf(Century)
}

func (tempo Tempo) IsStartOfMillennium() bool {
	return tempo.IsStartOf(Millennium)
}

func (tempo Tempo) IsEndOfMillennium() bool {
	return tempo.IsEndOf(Millennium)
}

func (tempo Tempo) StartOfDay() Tempo {
	return tempo.StartOf(Day)
}

func (tempo Tempo) EndOfDay() Tempo {
	return tempo.EndOf(Day)
}

func (tempo Tempo) StartOfWeek(options ...StartOfWeekOptions) Tempo {
	return tempo.StartOf(Week, options...)
}

func (tempo Tempo) EndOfWeek(options ...StartOfWeekOptions) Tempo {
	return tempo.EndOf(Week, options...)
}

func (tempo Tempo) StartOfMonth() Tempo {
	return tempo.StartOf(Month)
}

func (tempo Tempo) EndOfMonth() Tempo {
	return tempo.EndOf(Month)
}

func (tempo Tempo) StartOfQuarter() Tempo {
	return tempo.StartOf(Quarter)
}

func (tempo Tempo) EndOfQuarter() Tempo {
	return tempo.EndOf(Quarter)
}

func (tempo Tempo) FirstOfMonth(weekdays ...time.Weekday) Tempo {
	return boundaries.FirstOfMonth(tempo, weekdays...)
}

func (tempo Tempo) LastOfMonth(weekdays ...time.Weekday) Tempo {
	return boundaries.LastOfMonth(tempo, weekdays...)
}

func (tempo Tempo) NthOfMonth(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return boundaries.NthOfMonth(tempo, occurrence, weekday)
}

func (tempo Tempo) FirstOfQuarter(weekdays ...time.Weekday) Tempo {
	return boundaries.FirstOfQuarter(tempo, weekdays...)
}

func (tempo Tempo) LastOfQuarter(weekdays ...time.Weekday) Tempo {
	return boundaries.LastOfQuarter(tempo, weekdays...)
}

func (tempo Tempo) NthOfQuarter(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return boundaries.NthOfQuarter(tempo, occurrence, weekday)
}

func (tempo Tempo) StartOfYear() Tempo {
	return tempo.StartOf(Year)
}

func (tempo Tempo) EndOfYear() Tempo {
	return tempo.EndOf(Year)
}

func (tempo Tempo) StartOfDecade() Tempo {
	return tempo.StartOf(Decade)
}

func (tempo Tempo) EndOfDecade() Tempo {
	return tempo.EndOf(Decade)
}

func (tempo Tempo) StartOfCentury() Tempo {
	return tempo.StartOf(Century)
}

func (tempo Tempo) EndOfCentury() Tempo {
	return tempo.EndOf(Century)
}

func (tempo Tempo) StartOfMillennium() Tempo {
	return tempo.StartOf(Millennium)
}

func (tempo Tempo) EndOfMillennium() Tempo {
	return tempo.EndOf(Millennium)
}

func (tempo Tempo) FirstOfYear(weekdays ...time.Weekday) Tempo {
	return boundaries.FirstOfYear(tempo, weekdays...)
}

func (tempo Tempo) LastOfYear(weekdays ...time.Weekday) Tempo {
	return boundaries.LastOfYear(tempo, weekdays...)
}

func (tempo Tempo) NthOfYear(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return boundaries.NthOfYear(tempo, occurrence, weekday)
}

func (tempo Tempo) Floor(unit Unit) Tempo {
	return boundaries.Floor(tempo, unit)
}

func (tempo Tempo) FloorUnit(unit Unit) Tempo {
	return tempo.Floor(unit)
}

func (tempo Tempo) FloorWeek(options ...StartOfWeekOptions) Tempo {
	return boundaries.FloorWeek(tempo, weekOptions(options)...)
}

func (tempo Tempo) Ceil(unit Unit) Tempo {
	return boundaries.Ceil(tempo, unit)
}

func (tempo Tempo) CeilUnit(unit Unit) Tempo {
	return tempo.Ceil(unit)
}

func (tempo Tempo) CeilWeek(options ...StartOfWeekOptions) Tempo {
	return boundaries.CeilWeek(tempo, weekOptions(options)...)
}

func (tempo Tempo) Round(unit Unit) Tempo {
	return boundaries.Round(tempo, unit)
}

func (tempo Tempo) RoundUnit(unit Unit) Tempo {
	return tempo.Round(unit)
}

func (tempo Tempo) RoundWeek(options ...StartOfWeekOptions) Tempo {
	return boundaries.RoundWeek(tempo, weekOptions(options)...)
}

func (tempo Tempo) Next(weekday time.Weekday) Tempo {
	return boundaries.Next(tempo, weekday)
}

func (tempo Tempo) Previous(weekday time.Weekday) Tempo {
	return boundaries.Previous(tempo, weekday)
}

func (tempo Tempo) NextOrSame(weekday time.Weekday) Tempo {
	return boundaries.NextOrSame(tempo, weekday)
}

func (tempo Tempo) PreviousOrSame(weekday time.Weekday) Tempo {
	return boundaries.PreviousOrSame(tempo, weekday)
}

func (tempo Tempo) NextWeekday() Tempo {
	return boundaries.NextWeekday(tempo, defaultConfig.Settings.WeekendDays)
}

func (tempo Tempo) PreviousWeekday() Tempo {
	return boundaries.PreviousWeekday(tempo, defaultConfig.Settings.WeekendDays)
}

func (tempo Tempo) NextWeekendDay() Tempo {
	return boundaries.NextWeekendDay(tempo, defaultConfig.Settings.WeekendDays)
}

func (tempo Tempo) PreviousWeekendDay() Tempo {
	return boundaries.PreviousWeekendDay(tempo, defaultConfig.Settings.WeekendDays)
}

func weekOptions(options []StartOfWeekOptions) []kernel.WeekOptions {
	result := make([]kernel.WeekOptions, 0, len(options))

	for _, option := range options {
		result = append(result, kernel.WeekOptions{WeekStartsOn: option.WeekStartsOn})
	}

	return result
}
