package tempo

import "github.com/oullin/alloy/tempo/arithmetic"

func (tempo Tempo) Add(value int, unit Unit) Tempo {
	return arithmetic.Add(tempo, value, unit, tempo.settingsSnapshot().MonthsOverflow, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Tempo) Sub(value int, unit Unit) Tempo {
	return tempo.Add(-value, unit)
}

func (tempo Tempo) AddNoOverflow(value int, valueUnit Unit, overflowUnit Unit) Tempo {
	next, err := tempo.Add(value, valueUnit).Clamp(tempo.StartOf(overflowUnit), tempo.EndOf(overflowUnit))

	if err != nil {
		return tempo
	}

	return next
}

func (tempo Tempo) SubNoOverflow(value int, valueUnit Unit, overflowUnit Unit) Tempo {
	return tempo.AddNoOverflow(-value, valueUnit, overflowUnit)
}

func (tempo Tempo) AddDuration(dur Duration) Tempo {
	return arithmetic.AddDuration(tempo, dur, tempo.settingsSnapshot().MonthsOverflow, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Tempo) SubDuration(dur Duration) Tempo {
	return arithmetic.SubDuration(tempo, dur, tempo.settingsSnapshot().MonthsOverflow, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Tempo) AddMilliseconds(milliseconds int) Tempo {
	return arithmetic.AddMilliseconds(tempo, milliseconds)
}

func (tempo Tempo) SubMilliseconds(milliseconds int) Tempo {
	return arithmetic.SubMilliseconds(tempo, milliseconds)
}

func (tempo Tempo) AddSeconds(seconds int) Tempo {
	return arithmetic.AddSeconds(tempo, seconds)
}

func (tempo Tempo) SubSeconds(seconds int) Tempo {
	return arithmetic.SubSeconds(tempo, seconds)
}

func (tempo Tempo) AddMinutes(minutes int) Tempo {
	return arithmetic.AddMinutes(tempo, minutes)
}

func (tempo Tempo) SubMinutes(minutes int) Tempo {
	return arithmetic.SubMinutes(tempo, minutes)
}

func (tempo Tempo) AddHours(hours int) Tempo {
	return arithmetic.AddHours(tempo, hours)
}

func (tempo Tempo) SubHours(hours int) Tempo {
	return arithmetic.SubHours(tempo, hours)
}

func (tempo Tempo) AddDays(days int) Tempo {
	return arithmetic.AddDays(tempo, days)
}

func (tempo Tempo) SubDays(days int) Tempo {
	return arithmetic.SubDays(tempo, days)
}

func (tempo Tempo) AddWeekdays(days int) Tempo {
	return arithmetic.AddWeekdays(tempo, days, tempo.settingsSnapshot().WeekendDays)
}

func (tempo Tempo) SubWeekdays(days int) Tempo {
	return arithmetic.SubWeekdays(tempo, days, tempo.settingsSnapshot().WeekendDays)
}

func (tempo Tempo) AddWeeks(weeks int) Tempo {
	return arithmetic.AddWeeks(tempo, weeks)
}

func (tempo Tempo) SubWeeks(weeks int) Tempo {
	return arithmetic.SubWeeks(tempo, weeks)
}

func (tempo Tempo) AddMonths(months int) Tempo {
	return arithmetic.AddMonths(tempo, months, tempo.settingsSnapshot().MonthsOverflow)
}

func (tempo Tempo) SubMonths(months int) Tempo {
	return arithmetic.SubMonths(tempo, months, tempo.settingsSnapshot().MonthsOverflow)
}

func (tempo Tempo) AddMonthsNoOverflow(months int) Tempo {
	return arithmetic.AddMonthsNoOverflow(tempo, months)
}

func (tempo Tempo) SubMonthsNoOverflow(months int) Tempo {
	return arithmetic.SubMonthsNoOverflow(tempo, months)
}

func (tempo Tempo) AddQuarters(quarters int) Tempo {
	return arithmetic.AddQuarters(tempo, quarters, tempo.settingsSnapshot().MonthsOverflow)
}

func (tempo Tempo) SubQuarters(quarters int) Tempo {
	return arithmetic.SubQuarters(tempo, quarters, tempo.settingsSnapshot().MonthsOverflow)
}

func (tempo Tempo) AddYears(years int) Tempo {
	return arithmetic.AddYears(tempo, years, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Tempo) SubYears(years int) Tempo {
	return arithmetic.SubYears(tempo, years, tempo.settingsSnapshot().YearsOverflow)
}

func (tempo Tempo) AddYearsNoOverflow(years int) Tempo {
	return arithmetic.AddYearsNoOverflow(tempo, years)
}

func (tempo Tempo) Age(reference Tempo) int {
	return reference.DiffInYears(tempo)
}

func (tempo Tempo) SubYearsNoOverflow(years int) Tempo {
	return arithmetic.SubYearsNoOverflow(tempo, years)
}
