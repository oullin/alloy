package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/boundaries"
)

func (mutable *MutableTempo) StartOf(unit Unit, options ...StartOfWeekOptions) *MutableTempo {
	return boundaries.StartOf(mutable, unit, weekOptions(options)...)
}

func (mutable *MutableTempo) EndOf(unit Unit, options ...StartOfWeekOptions) *MutableTempo {
	return boundaries.EndOf(mutable, unit, weekOptions(options)...)
}

func (mutable *MutableTempo) StartOfMillisecond() *MutableTempo {
	return mutable.StartOf(Millisecond)
}

func (mutable *MutableTempo) EndOfMillisecond() *MutableTempo {
	return mutable.EndOf(Millisecond)
}

func (mutable *MutableTempo) StartOfSecond() *MutableTempo {
	return mutable.StartOf(Second)
}

func (mutable *MutableTempo) EndOfSecond() *MutableTempo {
	return mutable.EndOf(Second)
}

func (mutable *MutableTempo) StartOfMinute() *MutableTempo {
	return mutable.StartOf(Minute)
}

func (mutable *MutableTempo) EndOfMinute() *MutableTempo {
	return mutable.EndOf(Minute)
}

func (mutable *MutableTempo) StartOfHour() *MutableTempo {
	return mutable.StartOf(Hour)
}

func (mutable *MutableTempo) EndOfHour() *MutableTempo {
	return mutable.EndOf(Hour)
}

func (mutable *MutableTempo) StartOfDay() *MutableTempo {
	return mutable.StartOf(Day)
}

func (mutable *MutableTempo) EndOfDay() *MutableTempo {
	return mutable.EndOf(Day)
}

func (mutable *MutableTempo) StartOfWeek(options ...StartOfWeekOptions) *MutableTempo {
	return mutable.StartOf(Week, options...)
}

func (mutable *MutableTempo) EndOfWeek(options ...StartOfWeekOptions) *MutableTempo {
	return mutable.EndOf(Week, options...)
}

func (mutable *MutableTempo) StartOfMonth() *MutableTempo {
	return mutable.StartOf(Month)
}

func (mutable *MutableTempo) EndOfMonth() *MutableTempo {
	return mutable.EndOf(Month)
}

func (mutable *MutableTempo) StartOfQuarter() *MutableTempo {
	return mutable.StartOf(Quarter)
}

func (mutable *MutableTempo) EndOfQuarter() *MutableTempo {
	return mutable.EndOf(Quarter)
}

func (mutable *MutableTempo) FirstOfMonth(weekdays ...time.Weekday) *MutableTempo {
	return boundaries.FirstOfMonth(mutable, weekdays...)
}

func (mutable *MutableTempo) LastOfMonth(weekdays ...time.Weekday) *MutableTempo {
	return boundaries.LastOfMonth(mutable, weekdays...)
}

func (mutable *MutableTempo) NthOfMonth(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return mutable.Tempo().NthOfMonth(occurrence, weekday)
}

func (mutable *MutableTempo) FirstOfQuarter(weekdays ...time.Weekday) *MutableTempo {
	return boundaries.FirstOfQuarter(mutable, weekdays...)
}

func (mutable *MutableTempo) LastOfQuarter(weekdays ...time.Weekday) *MutableTempo {
	return boundaries.LastOfQuarter(mutable, weekdays...)
}

func (mutable *MutableTempo) NthOfQuarter(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return mutable.Tempo().NthOfQuarter(occurrence, weekday)
}

func (mutable *MutableTempo) StartOfYear() *MutableTempo {
	return mutable.StartOf(Year)
}

func (mutable *MutableTempo) EndOfYear() *MutableTempo {
	return mutable.EndOf(Year)
}

func (mutable *MutableTempo) StartOfDecade() *MutableTempo {
	return mutable.StartOf(Decade)
}

func (mutable *MutableTempo) EndOfDecade() *MutableTempo {
	return mutable.EndOf(Decade)
}

func (mutable *MutableTempo) StartOfCentury() *MutableTempo {
	return mutable.StartOf(Century)
}

func (mutable *MutableTempo) EndOfCentury() *MutableTempo {
	return mutable.EndOf(Century)
}

func (mutable *MutableTempo) StartOfMillennium() *MutableTempo {
	return mutable.StartOf(Millennium)
}

func (mutable *MutableTempo) EndOfMillennium() *MutableTempo {
	return mutable.EndOf(Millennium)
}

func (mutable *MutableTempo) FirstOfYear(weekdays ...time.Weekday) *MutableTempo {
	return boundaries.FirstOfYear(mutable, weekdays...)
}

func (mutable *MutableTempo) LastOfYear(weekdays ...time.Weekday) *MutableTempo {
	return boundaries.LastOfYear(mutable, weekdays...)
}

func (mutable *MutableTempo) NthOfYear(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return mutable.Tempo().NthOfYear(occurrence, weekday)
}

func (mutable *MutableTempo) Floor(unit Unit) *MutableTempo {
	return boundaries.Floor(mutable, unit)
}

func (mutable *MutableTempo) FloorUnit(unit Unit) *MutableTempo {
	return mutable.Floor(unit)
}

func (mutable *MutableTempo) FloorWeek(options ...StartOfWeekOptions) *MutableTempo {
	return boundaries.FloorWeek(mutable, weekOptions(options)...)
}

func (mutable *MutableTempo) Ceil(unit Unit) *MutableTempo {
	return boundaries.Ceil(mutable, unit)
}

func (mutable *MutableTempo) CeilUnit(unit Unit) *MutableTempo {
	return mutable.Ceil(unit)
}

func (mutable *MutableTempo) CeilWeek(options ...StartOfWeekOptions) *MutableTempo {
	return boundaries.CeilWeek(mutable, weekOptions(options)...)
}

func (mutable *MutableTempo) Round(unit Unit) *MutableTempo {
	return boundaries.Round(mutable, unit)
}

func (mutable *MutableTempo) RoundUnit(unit Unit) *MutableTempo {
	return mutable.Round(unit)
}

func (mutable *MutableTempo) RoundWeek(options ...StartOfWeekOptions) *MutableTempo {
	return boundaries.RoundWeek(mutable, weekOptions(options)...)
}

func (mutable *MutableTempo) Next(weekday time.Weekday) *MutableTempo {
	return boundaries.Next(mutable, weekday)
}

func (mutable *MutableTempo) Previous(weekday time.Weekday) *MutableTempo {
	return boundaries.Previous(mutable, weekday)
}

func (mutable *MutableTempo) NextOrSame(weekday time.Weekday) *MutableTempo {
	return boundaries.NextOrSame(mutable, weekday)
}

func (mutable *MutableTempo) PreviousOrSame(weekday time.Weekday) *MutableTempo {
	return boundaries.PreviousOrSame(mutable, weekday)
}

func (mutable *MutableTempo) NextWeekday() *MutableTempo {
	return boundaries.NextWeekday(mutable, mutable.settingsSnapshot().WeekendDays)
}

func (mutable *MutableTempo) PreviousWeekday() *MutableTempo {
	return boundaries.PreviousWeekday(mutable, mutable.settingsSnapshot().WeekendDays)
}

func (mutable *MutableTempo) NextWeekendDay() *MutableTempo {
	return boundaries.NextWeekendDay(mutable, mutable.settingsSnapshot().WeekendDays)
}

func (mutable *MutableTempo) PreviousWeekendDay() *MutableTempo {
	return boundaries.PreviousWeekendDay(mutable, mutable.settingsSnapshot().WeekendDays)
}
