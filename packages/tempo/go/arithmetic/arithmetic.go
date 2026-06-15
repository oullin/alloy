package arithmetic

import tempopkg "github.com/oullin/alloy/tempo/tempo"

type Tempo struct {
	value tempopkg.Tempo
}

func From(value tempopkg.Tempo) Tempo {
	return Tempo{value: value}
}

func (tempo Tempo) Tempo() tempopkg.Tempo {
	return tempo.value
}

func (tempo Tempo) Add(value int, unit tempopkg.Unit) Tempo {
	return From(tempo.value.Add(value, unit))
}

func (tempo Tempo) Sub(value int, unit tempopkg.Unit) Tempo {
	return From(tempo.value.Sub(value, unit))
}

func (tempo Tempo) AddDuration(duration tempopkg.Duration) Tempo {
	return From(tempo.value.AddDuration(duration))
}

func (tempo Tempo) SubDuration(duration tempopkg.Duration) Tempo {
	return From(tempo.value.SubDuration(duration))
}

func (tempo Tempo) AddMilliseconds(milliseconds int) Tempo {
	return From(tempo.value.AddMilliseconds(milliseconds))
}

func (tempo Tempo) SubMilliseconds(milliseconds int) Tempo {
	return From(tempo.value.SubMilliseconds(milliseconds))
}

func (tempo Tempo) AddSeconds(seconds int) Tempo {
	return From(tempo.value.AddSeconds(seconds))
}

func (tempo Tempo) SubSeconds(seconds int) Tempo {
	return From(tempo.value.SubSeconds(seconds))
}

func (tempo Tempo) AddMinutes(minutes int) Tempo {
	return From(tempo.value.AddMinutes(minutes))
}

func (tempo Tempo) SubMinutes(minutes int) Tempo {
	return From(tempo.value.SubMinutes(minutes))
}

func (tempo Tempo) AddHours(hours int) Tempo {
	return From(tempo.value.AddHours(hours))
}

func (tempo Tempo) SubHours(hours int) Tempo {
	return From(tempo.value.SubHours(hours))
}

func (tempo Tempo) AddDays(days int) Tempo {
	return From(tempo.value.AddDays(days))
}

func (tempo Tempo) SubDays(days int) Tempo {
	return From(tempo.value.SubDays(days))
}

func (tempo Tempo) AddWeekdays(days int) Tempo {
	return From(tempo.value.AddWeekdays(days))
}

func (tempo Tempo) SubWeekdays(days int) Tempo {
	return From(tempo.value.SubWeekdays(days))
}

func (tempo Tempo) AddWeeks(weeks int) Tempo {
	return From(tempo.value.AddWeeks(weeks))
}

func (tempo Tempo) SubWeeks(weeks int) Tempo {
	return From(tempo.value.SubWeeks(weeks))
}

func (tempo Tempo) AddMonths(months int) Tempo {
	return From(tempo.value.AddMonths(months))
}

func (tempo Tempo) SubMonths(months int) Tempo {
	return From(tempo.value.SubMonths(months))
}

func (tempo Tempo) AddMonthsNoOverflow(months int) Tempo {
	return From(tempo.value.AddMonthsNoOverflow(months))
}

func (tempo Tempo) SubMonthsNoOverflow(months int) Tempo {
	return From(tempo.value.SubMonthsNoOverflow(months))
}

func (tempo Tempo) AddQuarters(quarters int) Tempo {
	return From(tempo.value.AddQuarters(quarters))
}

func (tempo Tempo) SubQuarters(quarters int) Tempo {
	return From(tempo.value.SubQuarters(quarters))
}

func (tempo Tempo) AddYears(years int) Tempo {
	return From(tempo.value.AddYears(years))
}

func (tempo Tempo) SubYears(years int) Tempo {
	return From(tempo.value.SubYears(years))
}

func (tempo Tempo) AddYearsNoOverflow(years int) Tempo {
	return From(tempo.value.AddYearsNoOverflow(years))
}

func (tempo Tempo) SubYearsNoOverflow(years int) Tempo {
	return From(tempo.value.SubYearsNoOverflow(years))
}

func (tempo Tempo) Age(reference tempopkg.Tempo) int {
	return tempo.value.Age(reference)
}
