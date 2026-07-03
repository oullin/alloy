package tempo

import (
	"github.com/oullin/alloy/packages/foundation/tempo/arithmetic"
)

func (mutable *MutableTime) Add(value int, unit Unit) *MutableTime {
	return arithmetic.Add(mutable, value, unit, mutable.settingsSnapshot().MonthsOverflow, mutable.settingsSnapshot().YearsOverflow)
}

func (mutable *MutableTime) Sub(value int, unit Unit) *MutableTime {
	return mutable.Add(-value, unit)
}

func (mutable *MutableTime) AddNoOverflow(value int, valueUnit Unit, overflowUnit Unit) *MutableTime {
	return mutable.replace(mutable.Immutable().AddNoOverflow(value, valueUnit, overflowUnit))
}

func (mutable *MutableTime) SubNoOverflow(value int, valueUnit Unit, overflowUnit Unit) *MutableTime {
	return mutable.replace(mutable.Immutable().SubNoOverflow(value, valueUnit, overflowUnit))
}

func (mutable *MutableTime) AddDuration(dur Duration) *MutableTime {
	return arithmetic.AddDuration(mutable, dur, mutable.settingsSnapshot().MonthsOverflow, mutable.settingsSnapshot().YearsOverflow)
}

func (mutable *MutableTime) SubDuration(dur Duration) *MutableTime {
	return arithmetic.SubDuration(mutable, dur, mutable.settingsSnapshot().MonthsOverflow, mutable.settingsSnapshot().YearsOverflow)
}

func (mutable *MutableTime) AddMilliseconds(milliseconds int) *MutableTime {
	return arithmetic.AddMilliseconds(mutable, milliseconds)
}

func (mutable *MutableTime) SubMilliseconds(milliseconds int) *MutableTime {
	return arithmetic.SubMilliseconds(mutable, milliseconds)
}

func (mutable *MutableTime) AddSeconds(seconds int) *MutableTime {
	return arithmetic.AddSeconds(mutable, seconds)
}

func (mutable *MutableTime) SubSeconds(seconds int) *MutableTime {
	return arithmetic.SubSeconds(mutable, seconds)
}

func (mutable *MutableTime) AddMinutes(minutes int) *MutableTime {
	return arithmetic.AddMinutes(mutable, minutes)
}

func (mutable *MutableTime) SubMinutes(minutes int) *MutableTime {
	return arithmetic.SubMinutes(mutable, minutes)
}

func (mutable *MutableTime) AddHours(hours int) *MutableTime {
	return arithmetic.AddHours(mutable, hours)
}

func (mutable *MutableTime) SubHours(hours int) *MutableTime {
	return arithmetic.SubHours(mutable, hours)
}

func (mutable *MutableTime) AddDays(days int) *MutableTime {
	return arithmetic.AddDays(mutable, days)
}

func (mutable *MutableTime) SubDays(days int) *MutableTime {
	return arithmetic.SubDays(mutable, days)
}

func (mutable *MutableTime) AddWeekdays(days int) *MutableTime {
	return arithmetic.AddWeekdays(mutable, days, mutable.settingsSnapshot().WeekendDays)
}

func (mutable *MutableTime) SubWeekdays(days int) *MutableTime {
	return arithmetic.SubWeekdays(mutable, days, mutable.settingsSnapshot().WeekendDays)
}

func (mutable *MutableTime) AddWeeks(weeks int) *MutableTime {
	return arithmetic.AddWeeks(mutable, weeks)
}

func (mutable *MutableTime) SubWeeks(weeks int) *MutableTime {
	return arithmetic.SubWeeks(mutable, weeks)
}

func (mutable *MutableTime) AddMonths(months int) *MutableTime {
	return arithmetic.AddMonths(mutable, months, mutable.settingsSnapshot().MonthsOverflow)
}

func (mutable *MutableTime) SubMonths(months int) *MutableTime {
	return arithmetic.SubMonths(mutable, months, mutable.settingsSnapshot().MonthsOverflow)
}

func (mutable *MutableTime) AddMonthsNoOverflow(months int) *MutableTime {
	return arithmetic.AddMonthsNoOverflow(mutable, months)
}

func (mutable *MutableTime) SubMonthsNoOverflow(months int) *MutableTime {
	return arithmetic.SubMonthsNoOverflow(mutable, months)
}

func (mutable *MutableTime) AddQuarters(quarters int) *MutableTime {
	return arithmetic.AddQuarters(mutable, quarters, mutable.settingsSnapshot().MonthsOverflow)
}

func (mutable *MutableTime) SubQuarters(quarters int) *MutableTime {
	return arithmetic.SubQuarters(mutable, quarters, mutable.settingsSnapshot().MonthsOverflow)
}

func (mutable *MutableTime) AddYears(years int) *MutableTime {
	return arithmetic.AddYears(mutable, years, mutable.settingsSnapshot().YearsOverflow)
}

func (mutable *MutableTime) SubYears(years int) *MutableTime {
	return arithmetic.SubYears(mutable, years, mutable.settingsSnapshot().YearsOverflow)
}

func (mutable *MutableTime) AddYearsNoOverflow(years int) *MutableTime {
	return arithmetic.AddYearsNoOverflow(mutable, years)
}

func (mutable *MutableTime) SubYearsNoOverflow(years int) *MutableTime {
	return arithmetic.SubYearsNoOverflow(mutable, years)
}

func (mutable *MutableTime) Age(reference Time) int {
	return mutable.Immutable().Age(reference)
}
