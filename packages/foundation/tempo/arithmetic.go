package tempo

import "alloy.dev/foundation/tempo/arithmetic"

func (tempo Time) Add(value int, unit Unit) Time {
	return arithmetic.Add(tempo, value, unit, tempo.settingsSnapshot().MonthsOverflow, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Time) Sub(value int, unit Unit) Time {
	return tempo.Add(-value, unit)
}

func (tempo Time) AddNoOverflow(value int, valueUnit Unit, overflowUnit Unit) Time {
	next, err := tempo.Add(value, valueUnit).Clamp(tempo.StartOf(overflowUnit), tempo.EndOf(overflowUnit))

	if err != nil {
		return tempo
	}

	return next
}

func (tempo Time) SubNoOverflow(value int, valueUnit Unit, overflowUnit Unit) Time {
	return tempo.AddNoOverflow(-value, valueUnit, overflowUnit)
}

func (tempo Time) AddDuration(dur Duration) Time {
	return arithmetic.AddDuration(tempo, dur, tempo.settingsSnapshot().MonthsOverflow, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Time) SubDuration(dur Duration) Time {
	return arithmetic.SubDuration(tempo, dur, tempo.settingsSnapshot().MonthsOverflow, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Time) AddMilliseconds(milliseconds int) Time {
	return arithmetic.AddMilliseconds(tempo, milliseconds)
}

func (tempo Time) SubMilliseconds(milliseconds int) Time {
	return arithmetic.SubMilliseconds(tempo, milliseconds)
}

func (tempo Time) AddSeconds(seconds int) Time {
	return arithmetic.AddSeconds(tempo, seconds)
}

func (tempo Time) SubSeconds(seconds int) Time {
	return arithmetic.SubSeconds(tempo, seconds)
}

func (tempo Time) AddMinutes(minutes int) Time {
	return arithmetic.AddMinutes(tempo, minutes)
}

func (tempo Time) SubMinutes(minutes int) Time {
	return arithmetic.SubMinutes(tempo, minutes)
}

func (tempo Time) AddHours(hours int) Time {
	return arithmetic.AddHours(tempo, hours)
}

func (tempo Time) SubHours(hours int) Time {
	return arithmetic.SubHours(tempo, hours)
}

func (tempo Time) AddDays(days int) Time {
	return arithmetic.AddDays(tempo, days)
}

func (tempo Time) SubDays(days int) Time {
	return arithmetic.SubDays(tempo, days)
}

func (tempo Time) AddWeekdays(days int) Time {
	return arithmetic.AddWeekdays(tempo, days, tempo.settingsSnapshot().WeekendDays)
}

func (tempo Time) SubWeekdays(days int) Time {
	return arithmetic.SubWeekdays(tempo, days, tempo.settingsSnapshot().WeekendDays)
}

func (tempo Time) AddWeeks(weeks int) Time {
	return arithmetic.AddWeeks(tempo, weeks)
}

func (tempo Time) SubWeeks(weeks int) Time {
	return arithmetic.SubWeeks(tempo, weeks)
}

func (tempo Time) AddMonths(months int) Time {
	return arithmetic.AddMonths(tempo, months, tempo.settingsSnapshot().MonthsOverflow)
}

func (tempo Time) SubMonths(months int) Time {
	return arithmetic.SubMonths(tempo, months, tempo.settingsSnapshot().MonthsOverflow)
}

func (tempo Time) AddMonthsNoOverflow(months int) Time {
	return arithmetic.AddMonthsNoOverflow(tempo, months)
}

func (tempo Time) SubMonthsNoOverflow(months int) Time {
	return arithmetic.SubMonthsNoOverflow(tempo, months)
}

func (tempo Time) AddQuarters(quarters int) Time {
	return arithmetic.AddQuarters(tempo, quarters, tempo.settingsSnapshot().MonthsOverflow)
}

func (tempo Time) SubQuarters(quarters int) Time {
	return arithmetic.SubQuarters(tempo, quarters, tempo.settingsSnapshot().MonthsOverflow)
}

func (tempo Time) AddYears(years int) Time {
	return arithmetic.AddYears(tempo, years, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Time) SubYears(years int) Time {
	return arithmetic.SubYears(tempo, years, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Time) AddYearsNoOverflow(years int) Time {
	return arithmetic.AddYearsNoOverflow(tempo, years)
}

func (tempo Time) Age(reference Time) int {
	return reference.DiffInYears(tempo)
}

func (tempo Time) SubYearsNoOverflow(years int) Time {
	return arithmetic.SubYearsNoOverflow(tempo, years)
}
