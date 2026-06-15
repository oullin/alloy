package tempo

import "time"

func (tempo Tempo) Add(value int, unit Unit) Tempo {
	switch normalizeUnit(unit) {
	case Millisecond:
		return tempo.addDuration(time.Duration(value) * time.Millisecond)
	case Second:
		return tempo.addDuration(time.Duration(value) * time.Second)
	case Minute:
		return tempo.addDuration(time.Duration(value) * time.Minute)
	case Hour:
		return tempo.addDuration(time.Duration(value) * time.Hour)
	case Day:
		return tempo.addDurationDate(0, 0, value)
	case Week:
		return tempo.addDurationDate(0, 0, value*7)
	case Month:
		return tempo.AddMonths(value)
	case Quarter:
		return tempo.AddMonths(value * 3)
	case Year:
		return tempo.AddYears(value)
	default:
		return tempo
	}
}

func (tempo Tempo) Sub(value int, unit Unit) Tempo {
	return tempo.Add(-value, unit)
}

func (tempo Tempo) AddUnit(unit Unit, value int) Tempo {
	return tempo.Add(value, unit)
}

func (tempo Tempo) AddUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) Tempo {
	next, err := tempo.Add(value, valueUnit).Clamp(tempo.StartOf(overflowUnit), tempo.EndOf(overflowUnit))
	if err != nil {
		return tempo
	}

	return next
}

func (tempo Tempo) AddRealUnit(unit Unit, value int) Tempo {
	return tempo.Add(value, unit)
}

func (tempo Tempo) AddUTCUnit(unit Unit, value int) Tempo {
	return tempo.Add(value, unit)
}

func (tempo Tempo) RawAdd(value int, unit Unit) Tempo {
	return tempo.Add(value, unit)
}

func (tempo Tempo) SubUnit(unit Unit, value int) Tempo {
	return tempo.Sub(value, unit)
}

func (tempo Tempo) SubUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) Tempo {
	return tempo.AddUnitNoOverflow(valueUnit, -value, overflowUnit)
}

func (tempo Tempo) SubRealUnit(unit Unit, value int) Tempo {
	return tempo.Sub(value, unit)
}

func (tempo Tempo) SubUTCUnit(unit Unit, value int) Tempo {
	return tempo.Sub(value, unit)
}

func (tempo Tempo) RawSub(value int, unit Unit) Tempo {
	return tempo.Sub(value, unit)
}

func (tempo Tempo) Subtract(value int, unit Unit) Tempo {
	return tempo.Sub(value, unit)
}

func (tempo Tempo) AddDuration(duration Duration) Tempo {
	return tempo.
		AddYears(duration.Years).
		AddMonths(duration.Quarters*3 + duration.Months).
		AddWeeks(duration.Weeks).
		AddDays(duration.Days).
		AddHours(duration.Hours).
		AddMinutes(duration.Minutes).
		AddSeconds(duration.Seconds).
		AddMilliseconds(duration.Milliseconds)
}

func (tempo Tempo) SubDuration(duration Duration) Tempo {
	return tempo.AddDuration(Duration{
		Years:        -duration.Years,
		Quarters:     -duration.Quarters,
		Months:       -duration.Months,
		Weeks:        -duration.Weeks,
		Days:         -duration.Days,
		Hours:        -duration.Hours,
		Minutes:      -duration.Minutes,
		Seconds:      -duration.Seconds,
		Milliseconds: -duration.Milliseconds,
	})
}

func (tempo Tempo) AddMilliseconds(milliseconds int) Tempo {
	return tempo.Add(milliseconds, Millisecond)
}

func (tempo Tempo) SubMilliseconds(milliseconds int) Tempo {
	return tempo.Sub(milliseconds, Millisecond)
}

func (tempo Tempo) AddSeconds(seconds int) Tempo {
	return tempo.Add(seconds, Second)
}

func (tempo Tempo) SubSeconds(seconds int) Tempo {
	return tempo.Sub(seconds, Second)
}

func (tempo Tempo) AddMinutes(minutes int) Tempo {
	return tempo.Add(minutes, Minute)
}

func (tempo Tempo) SubMinutes(minutes int) Tempo {
	return tempo.Sub(minutes, Minute)
}

func (tempo Tempo) AddHours(hours int) Tempo {
	return tempo.Add(hours, Hour)
}

func (tempo Tempo) SubHours(hours int) Tempo {
	return tempo.Sub(hours, Hour)
}

func (tempo Tempo) AddDays(days int) Tempo {
	return tempo.Add(days, Day)
}

func (tempo Tempo) SubDays(days int) Tempo {
	return tempo.Sub(days, Day)
}

func (tempo Tempo) AddWeekdays(days int) Tempo {
	if days == 0 {
		return tempo.Clone()
	}

	direction := 1
	if days < 0 {
		direction = -1
		days = -days
	}

	current := tempo.Clone()
	for days > 0 {
		current = current.AddDays(direction)
		if current.IsWeekday() {
			days--
		}
	}

	return current
}

func (tempo Tempo) SubWeekdays(days int) Tempo {
	return tempo.AddWeekdays(-days)
}

func (tempo Tempo) AddWeeks(weeks int) Tempo {
	return tempo.Add(weeks, Week)
}

func (tempo Tempo) SubWeeks(weeks int) Tempo {
	return tempo.Sub(weeks, Week)
}

func (tempo Tempo) AddMonths(months int) Tempo {
	if !tempoSettings.MonthsOverflow {
		return tempo.AddMonthsNoOverflow(months)
	}

	return tempo.addDurationDate(0, months, 0)
}

func (tempo Tempo) SubMonths(months int) Tempo {
	return tempo.AddMonths(-months)
}

func (tempo Tempo) AddMonthsNoOverflow(months int) Tempo {
	local := tempo.local()
	target := time.Date(local.Year(), local.Month()+time.Month(months), 1, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), tempo.location)
	day := min(local.Day(), daysInMonth(target.Year(), int(target.Month())))
	next := time.Date(
		target.Year(),
		target.Month(),
		day,
		local.Hour(),
		local.Minute(),
		local.Second(),
		local.Nanosecond(),
		tempo.location,
	)

	return Tempo{value: next.UTC(), location: tempo.location}
}

func (tempo Tempo) SubMonthsNoOverflow(months int) Tempo {
	return tempo.AddMonthsNoOverflow(-months)
}

func (tempo Tempo) AddQuarters(quarters int) Tempo {
	return tempo.AddMonths(quarters * 3)
}

func (tempo Tempo) SubQuarters(quarters int) Tempo {
	return tempo.AddQuarters(-quarters)
}

func (tempo Tempo) AddYears(years int) Tempo {
	if !tempoSettings.YearsOverflow {
		return tempo.AddYearsNoOverflow(years)
	}

	return tempo.addDurationDate(years, 0, 0)
}

func (tempo Tempo) SubYears(years int) Tempo {
	return tempo.AddYears(-years)
}

func (tempo Tempo) AddYearsNoOverflow(years int) Tempo {
	local := tempo.local()
	year := local.Year() + years
	day := min(local.Day(), daysInMonth(year, int(local.Month())))
	next := time.Date(
		year,
		local.Month(),
		day,
		local.Hour(),
		local.Minute(),
		local.Second(),
		local.Nanosecond(),
		tempo.location,
	)

	return Tempo{value: next.UTC(), location: tempo.location}
}

func (tempo Tempo) Age(reference Tempo) int {
	return reference.DiffInYears(tempo)
}

func (tempo Tempo) SubYearsNoOverflow(years int) Tempo {
	return tempo.AddYearsNoOverflow(-years)
}
