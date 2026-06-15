package tempo

import "time"

func (mutable *MutableTempo) Add(value int, unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Add(value, unit))
}

func (mutable *MutableTempo) Sub(value int, unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Sub(value, unit))
}

func (mutable *MutableTempo) AddUnit(unit Unit, value int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddUnit(unit, value))
}

func (mutable *MutableTempo) AddUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddUnitNoOverflow(valueUnit, value, overflowUnit))
}

func (mutable *MutableTempo) AddRealUnit(unit Unit, value int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddRealUnit(unit, value))
}

func (mutable *MutableTempo) AddUTCUnit(unit Unit, value int) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddUTCUnit(unit, value))
}

func (mutable *MutableTempo) RawAdd(value int, unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().RawAdd(value, unit))
}

func (mutable *MutableTempo) SubUnit(unit Unit, value int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubUnit(unit, value))
}

func (mutable *MutableTempo) SubUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubUnitNoOverflow(valueUnit, value, overflowUnit))
}

func (mutable *MutableTempo) SubRealUnit(unit Unit, value int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubRealUnit(unit, value))
}

func (mutable *MutableTempo) SubUTCUnit(unit Unit, value int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SubUTCUnit(unit, value))
}

func (mutable *MutableTempo) RawSub(value int, unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().RawSub(value, unit))
}

func (mutable *MutableTempo) Subtract(value int, unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Subtract(value, unit))
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

func (mutable *MutableTempo) StartOfMillisecond() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfMillisecond())
}

func (mutable *MutableTempo) EndOfMillisecond() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfMillisecond())
}

func (mutable *MutableTempo) StartOfSecond() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfSecond())
}

func (mutable *MutableTempo) EndOfSecond() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfSecond())
}

func (mutable *MutableTempo) StartOfMinute() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfMinute())
}

func (mutable *MutableTempo) EndOfMinute() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfMinute())
}

func (mutable *MutableTempo) StartOfHour() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfHour())
}

func (mutable *MutableTempo) EndOfHour() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfHour())
}

func (mutable *MutableTempo) StartOfDay() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfDay())
}

func (mutable *MutableTempo) EndOfDay() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfDay())
}

func (mutable *MutableTempo) StartOfWeek(options ...StartOfWeekOptions) *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfWeek(options...))
}

func (mutable *MutableTempo) EndOfWeek(options ...StartOfWeekOptions) *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfWeek(options...))
}

func (mutable *MutableTempo) StartOfMonth() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfMonth())
}

func (mutable *MutableTempo) EndOfMonth() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfMonth())
}

func (mutable *MutableTempo) StartOfQuarter() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfQuarter())
}

func (mutable *MutableTempo) EndOfQuarter() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfQuarter())
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

func (mutable *MutableTempo) NthOfQuarter(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return mutable.Tempo().NthOfQuarter(occurrence, weekday)
}

func (mutable *MutableTempo) StartOfYear() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfYear())
}

func (mutable *MutableTempo) EndOfYear() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfYear())
}

func (mutable *MutableTempo) StartOfDecade() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfDecade())
}

func (mutable *MutableTempo) EndOfDecade() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfDecade())
}

func (mutable *MutableTempo) StartOfCentury() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfCentury())
}

func (mutable *MutableTempo) EndOfCentury() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfCentury())
}

func (mutable *MutableTempo) StartOfMillennium() *MutableTempo {
	return mutable.replace(mutable.Tempo().StartOfMillennium())
}

func (mutable *MutableTempo) EndOfMillennium() *MutableTempo {
	return mutable.replace(mutable.Tempo().EndOfMillennium())
}

func (mutable *MutableTempo) FirstOfYear(weekdays ...time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().FirstOfYear(weekdays...))
}

func (mutable *MutableTempo) LastOfYear(weekdays ...time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().LastOfYear(weekdays...))
}

func (mutable *MutableTempo) NthOfYear(occurrence int, weekday time.Weekday) (Tempo, bool) {
	return mutable.Tempo().NthOfYear(occurrence, weekday)
}

func (mutable *MutableTempo) Floor(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Floor(unit))
}

func (mutable *MutableTempo) FloorUnit(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().FloorUnit(unit))
}

func (mutable *MutableTempo) FloorWeek(options ...StartOfWeekOptions) *MutableTempo {
	return mutable.replace(mutable.Tempo().FloorWeek(options...))
}

func (mutable *MutableTempo) Ceil(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Ceil(unit))
}

func (mutable *MutableTempo) CeilUnit(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().CeilUnit(unit))
}

func (mutable *MutableTempo) CeilWeek(options ...StartOfWeekOptions) *MutableTempo {
	return mutable.replace(mutable.Tempo().CeilWeek(options...))
}

func (mutable *MutableTempo) Round(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().Round(unit))
}

func (mutable *MutableTempo) RoundUnit(unit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().RoundUnit(unit))
}

func (mutable *MutableTempo) RoundWeek(options ...StartOfWeekOptions) *MutableTempo {
	return mutable.replace(mutable.Tempo().RoundWeek(options...))
}

func (mutable *MutableTempo) Next(weekday time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().Next(weekday))
}

func (mutable *MutableTempo) Previous(weekday time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().Previous(weekday))
}

func (mutable *MutableTempo) NextOrSame(weekday time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().NextOrSame(weekday))
}

func (mutable *MutableTempo) PreviousOrSame(weekday time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().PreviousOrSame(weekday))
}

func (mutable *MutableTempo) NextWeekday() *MutableTempo {
	return mutable.replace(mutable.Tempo().NextWeekday())
}

func (mutable *MutableTempo) PreviousWeekday() *MutableTempo {
	return mutable.replace(mutable.Tempo().PreviousWeekday())
}

func (mutable *MutableTempo) NextWeekendDay() *MutableTempo {
	return mutable.replace(mutable.Tempo().NextWeekendDay())
}

func (mutable *MutableTempo) PreviousWeekendDay() *MutableTempo {
	return mutable.replace(mutable.Tempo().PreviousWeekendDay())
}
