package tempo

import (
	"github.com/oullin/alloy/tempo/arithmetic"
)

func (mutable *MutableTempo) Add(value int, unit Unit) *MutableTempo {
	return arithmetic.Add(mutable, value, unit, defaultConfig.Settings.MonthsOverflow, defaultConfig.Settings.YearsOverflow)
}

func (mutable *MutableTempo) Sub(value int, unit Unit) *MutableTempo {
	return mutable.Add(-value, unit)
}

func (mutable *MutableTempo) AddUnit(unit Unit, value int) *MutableTempo {
	return mutable.Add(value, unit)
}

func (mutable *MutableTempo) AddUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) *MutableTempo {
	return mutable.replace(mutable.Tempo().AddUnitNoOverflow(valueUnit, value, overflowUnit))
}

func (mutable *MutableTempo) AddRealUnit(unit Unit, value int) *MutableTempo {
	return mutable.Add(value, unit)
}

func (mutable *MutableTempo) AddUTCUnit(unit Unit, value int) *MutableTempo {
	return mutable.Add(value, unit)
}

func (mutable *MutableTempo) RawAdd(value int, unit Unit) *MutableTempo {
	return mutable.Add(value, unit)
}

func (mutable *MutableTempo) SubUnit(unit Unit, value int) *MutableTempo {
	return mutable.Sub(value, unit)
}

func (mutable *MutableTempo) SubUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) *MutableTempo {
	return mutable.AddUnitNoOverflow(valueUnit, -value, overflowUnit)
}

func (mutable *MutableTempo) SubRealUnit(unit Unit, value int) *MutableTempo {
	return mutable.Sub(value, unit)
}

func (mutable *MutableTempo) SubUTCUnit(unit Unit, value int) *MutableTempo {
	return mutable.Sub(value, unit)
}

func (mutable *MutableTempo) RawSub(value int, unit Unit) *MutableTempo {
	return mutable.Sub(value, unit)
}

func (mutable *MutableTempo) Subtract(value int, unit Unit) *MutableTempo {
	return mutable.Sub(value, unit)
}

func (mutable *MutableTempo) AddDuration(dur Duration) *MutableTempo {
	return arithmetic.AddDuration(mutable, dur, defaultConfig.Settings.MonthsOverflow, defaultConfig.Settings.YearsOverflow)
}

func (mutable *MutableTempo) SubDuration(dur Duration) *MutableTempo {
	return arithmetic.SubDuration(mutable, dur, defaultConfig.Settings.MonthsOverflow, defaultConfig.Settings.YearsOverflow)
}

func (mutable *MutableTempo) AddMilliseconds(milliseconds int) *MutableTempo {
	return arithmetic.AddMilliseconds(mutable, milliseconds)
}

func (mutable *MutableTempo) SubMilliseconds(milliseconds int) *MutableTempo {
	return arithmetic.SubMilliseconds(mutable, milliseconds)
}

func (mutable *MutableTempo) AddSeconds(seconds int) *MutableTempo {
	return arithmetic.AddSeconds(mutable, seconds)
}

func (mutable *MutableTempo) SubSeconds(seconds int) *MutableTempo {
	return arithmetic.SubSeconds(mutable, seconds)
}

func (mutable *MutableTempo) AddMinutes(minutes int) *MutableTempo {
	return arithmetic.AddMinutes(mutable, minutes)
}

func (mutable *MutableTempo) SubMinutes(minutes int) *MutableTempo {
	return arithmetic.SubMinutes(mutable, minutes)
}

func (mutable *MutableTempo) AddHours(hours int) *MutableTempo {
	return arithmetic.AddHours(mutable, hours)
}

func (mutable *MutableTempo) SubHours(hours int) *MutableTempo {
	return arithmetic.SubHours(mutable, hours)
}

func (mutable *MutableTempo) AddDays(days int) *MutableTempo {
	return arithmetic.AddDays(mutable, days)
}

func (mutable *MutableTempo) SubDays(days int) *MutableTempo {
	return arithmetic.SubDays(mutable, days)
}

func (mutable *MutableTempo) AddWeekdays(days int) *MutableTempo {
	return arithmetic.AddWeekdays(mutable, days, defaultConfig.Settings.WeekendDays)
}

func (mutable *MutableTempo) SubWeekdays(days int) *MutableTempo {
	return arithmetic.SubWeekdays(mutable, days, defaultConfig.Settings.WeekendDays)
}

func (mutable *MutableTempo) AddWeeks(weeks int) *MutableTempo {
	return arithmetic.AddWeeks(mutable, weeks)
}

func (mutable *MutableTempo) SubWeeks(weeks int) *MutableTempo {
	return arithmetic.SubWeeks(mutable, weeks)
}

func (mutable *MutableTempo) AddMonths(months int) *MutableTempo {
	return arithmetic.AddMonths(mutable, months, defaultConfig.Settings.MonthsOverflow)
}

func (mutable *MutableTempo) SubMonths(months int) *MutableTempo {
	return arithmetic.SubMonths(mutable, months, defaultConfig.Settings.MonthsOverflow)
}

func (mutable *MutableTempo) AddMonthsNoOverflow(months int) *MutableTempo {
	return arithmetic.AddMonthsNoOverflow(mutable, months)
}

func (mutable *MutableTempo) SubMonthsNoOverflow(months int) *MutableTempo {
	return arithmetic.SubMonthsNoOverflow(mutable, months)
}

func (mutable *MutableTempo) AddQuarters(quarters int) *MutableTempo {
	return arithmetic.AddQuarters(mutable, quarters, defaultConfig.Settings.MonthsOverflow)
}

func (mutable *MutableTempo) SubQuarters(quarters int) *MutableTempo {
	return arithmetic.SubQuarters(mutable, quarters, defaultConfig.Settings.MonthsOverflow)
}

func (mutable *MutableTempo) AddYears(years int) *MutableTempo {
	return arithmetic.AddYears(mutable, years, defaultConfig.Settings.YearsOverflow)
}

func (mutable *MutableTempo) SubYears(years int) *MutableTempo {
	return arithmetic.SubYears(mutable, years, defaultConfig.Settings.YearsOverflow)
}

func (mutable *MutableTempo) AddYearsNoOverflow(years int) *MutableTempo {
	return arithmetic.AddYearsNoOverflow(mutable, years)
}

func (mutable *MutableTempo) SubYearsNoOverflow(years int) *MutableTempo {
	return arithmetic.SubYearsNoOverflow(mutable, years)
}

func (mutable *MutableTempo) Age(reference Tempo) int {
	return mutable.Tempo().Age(reference)
}
