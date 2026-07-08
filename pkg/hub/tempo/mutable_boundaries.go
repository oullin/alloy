package tempo

import (
	"time"

	"github.com/oullin/alloy/pkg/hub/tempo/boundaries"
)

func (mutable *MutableTime) StartOf(unit Unit, options ...StartOfWeekOptions) *MutableTime {
	return boundaries.StartOf(mutable, unit, weekOptions(options)...)
}

func (mutable *MutableTime) EndOf(unit Unit, options ...StartOfWeekOptions) *MutableTime {
	return boundaries.EndOf(mutable, unit, weekOptions(options)...)
}

func (mutable *MutableTime) StartOfMillisecond() *MutableTime {
	return mutable.StartOf(Millisecond)
}

func (mutable *MutableTime) EndOfMillisecond() *MutableTime {
	return mutable.EndOf(Millisecond)
}

func (mutable *MutableTime) StartOfSecond() *MutableTime {
	return mutable.StartOf(Second)
}

func (mutable *MutableTime) EndOfSecond() *MutableTime {
	return mutable.EndOf(Second)
}

func (mutable *MutableTime) StartOfMinute() *MutableTime {
	return mutable.StartOf(Minute)
}

func (mutable *MutableTime) EndOfMinute() *MutableTime {
	return mutable.EndOf(Minute)
}

func (mutable *MutableTime) StartOfHour() *MutableTime {
	return mutable.StartOf(Hour)
}

func (mutable *MutableTime) EndOfHour() *MutableTime {
	return mutable.EndOf(Hour)
}

func (mutable *MutableTime) StartOfDay() *MutableTime {
	return mutable.StartOf(Day)
}

func (mutable *MutableTime) EndOfDay() *MutableTime {
	return mutable.EndOf(Day)
}

func (mutable *MutableTime) StartOfWeek(options ...StartOfWeekOptions) *MutableTime {
	return mutable.StartOf(Week, options...)
}

func (mutable *MutableTime) EndOfWeek(options ...StartOfWeekOptions) *MutableTime {
	return mutable.EndOf(Week, options...)
}

func (mutable *MutableTime) StartOfMonth() *MutableTime {
	return mutable.StartOf(Month)
}

func (mutable *MutableTime) EndOfMonth() *MutableTime {
	return mutable.EndOf(Month)
}

func (mutable *MutableTime) StartOfQuarter() *MutableTime {
	return mutable.StartOf(Quarter)
}

func (mutable *MutableTime) EndOfQuarter() *MutableTime {
	return mutable.EndOf(Quarter)
}

func (mutable *MutableTime) FirstOfMonth(weekdays ...time.Weekday) *MutableTime {
	return boundaries.FirstOfMonth(mutable, weekdays...)
}

func (mutable *MutableTime) LastOfMonth(weekdays ...time.Weekday) *MutableTime {
	return boundaries.LastOfMonth(mutable, weekdays...)
}

func (mutable *MutableTime) NthOfMonth(occurrence int, weekday time.Weekday) (Time, bool) {
	return mutable.Immutable().NthOfMonth(occurrence, weekday)
}

func (mutable *MutableTime) FirstOfQuarter(weekdays ...time.Weekday) *MutableTime {
	return boundaries.FirstOfQuarter(mutable, weekdays...)
}

func (mutable *MutableTime) LastOfQuarter(weekdays ...time.Weekday) *MutableTime {
	return boundaries.LastOfQuarter(mutable, weekdays...)
}

func (mutable *MutableTime) NthOfQuarter(occurrence int, weekday time.Weekday) (Time, bool) {
	return mutable.Immutable().NthOfQuarter(occurrence, weekday)
}

func (mutable *MutableTime) StartOfYear() *MutableTime {
	return mutable.StartOf(Year)
}

func (mutable *MutableTime) EndOfYear() *MutableTime {
	return mutable.EndOf(Year)
}

func (mutable *MutableTime) StartOfDecade() *MutableTime {
	return mutable.StartOf(Decade)
}

func (mutable *MutableTime) EndOfDecade() *MutableTime {
	return mutable.EndOf(Decade)
}

func (mutable *MutableTime) StartOfCentury() *MutableTime {
	return mutable.StartOf(Century)
}

func (mutable *MutableTime) EndOfCentury() *MutableTime {
	return mutable.EndOf(Century)
}

func (mutable *MutableTime) StartOfMillennium() *MutableTime {
	return mutable.StartOf(Millennium)
}

func (mutable *MutableTime) EndOfMillennium() *MutableTime {
	return mutable.EndOf(Millennium)
}

func (mutable *MutableTime) FirstOfYear(weekdays ...time.Weekday) *MutableTime {
	return boundaries.FirstOfYear(mutable, weekdays...)
}

func (mutable *MutableTime) LastOfYear(weekdays ...time.Weekday) *MutableTime {
	return boundaries.LastOfYear(mutable, weekdays...)
}

func (mutable *MutableTime) NthOfYear(occurrence int, weekday time.Weekday) (Time, bool) {
	return mutable.Immutable().NthOfYear(occurrence, weekday)
}

func (mutable *MutableTime) Floor(unit Unit) *MutableTime {
	return boundaries.Floor(mutable, unit)
}

func (mutable *MutableTime) FloorUnit(unit Unit) *MutableTime {
	return mutable.Floor(unit)
}

func (mutable *MutableTime) FloorWeek(options ...StartOfWeekOptions) *MutableTime {
	return boundaries.FloorWeek(mutable, weekOptions(options)...)
}

func (mutable *MutableTime) Ceil(unit Unit) *MutableTime {
	return boundaries.Ceil(mutable, unit)
}

func (mutable *MutableTime) CeilUnit(unit Unit) *MutableTime {
	return mutable.Ceil(unit)
}

func (mutable *MutableTime) CeilWeek(options ...StartOfWeekOptions) *MutableTime {
	return boundaries.CeilWeek(mutable, weekOptions(options)...)
}

func (mutable *MutableTime) Round(unit Unit) *MutableTime {
	return boundaries.Round(mutable, unit)
}

func (mutable *MutableTime) RoundUnit(unit Unit) *MutableTime {
	return mutable.Round(unit)
}

func (mutable *MutableTime) RoundWeek(options ...StartOfWeekOptions) *MutableTime {
	return boundaries.RoundWeek(mutable, weekOptions(options)...)
}

func (mutable *MutableTime) Next(weekday time.Weekday) *MutableTime {
	return boundaries.Next(mutable, weekday)
}

func (mutable *MutableTime) Previous(weekday time.Weekday) *MutableTime {
	return boundaries.Previous(mutable, weekday)
}

func (mutable *MutableTime) NextOrSame(weekday time.Weekday) *MutableTime {
	return boundaries.NextOrSame(mutable, weekday)
}

func (mutable *MutableTime) PreviousOrSame(weekday time.Weekday) *MutableTime {
	return boundaries.PreviousOrSame(mutable, weekday)
}

func (mutable *MutableTime) NextWeekday() *MutableTime {
	return boundaries.NextWeekday(mutable, mutable.settingsSnapshot().WeekendDays)
}

func (mutable *MutableTime) PreviousWeekday() *MutableTime {
	return boundaries.PreviousWeekday(mutable, mutable.settingsSnapshot().WeekendDays)
}

func (mutable *MutableTime) NextWeekendDay() *MutableTime {
	return boundaries.NextWeekendDay(mutable, mutable.settingsSnapshot().WeekendDays)
}

func (mutable *MutableTime) PreviousWeekendDay() *MutableTime {
	return boundaries.PreviousWeekendDay(mutable, mutable.settingsSnapshot().WeekendDays)
}
