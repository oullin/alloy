package boundaries

import (
	"time"

	tempopkg "github.com/oullin/alloy/tempo/tempo"
)

type Tempo struct {
	value tempopkg.Tempo
}

func From(value tempopkg.Tempo) Tempo {
	return Tempo{value: value}
}

func (tempo Tempo) Tempo() tempopkg.Tempo {
	return tempo.value
}

func (tempo Tempo) StartOf(unit tempopkg.Unit, options ...tempopkg.StartOfWeekOptions) Tempo {
	return From(tempo.value.StartOf(unit, options...))
}

func (tempo Tempo) EndOf(unit tempopkg.Unit, options ...tempopkg.StartOfWeekOptions) Tempo {
	return From(tempo.value.EndOf(unit, options...))
}

func (tempo Tempo) IsStartOf(unit tempopkg.Unit, options ...tempopkg.StartOfWeekOptions) bool {
	return tempo.value.IsStartOf(unit, options...)
}

func (tempo Tempo) IsEndOf(unit tempopkg.Unit, options ...tempopkg.StartOfWeekOptions) bool {
	return tempo.value.IsEndOf(unit, options...)
}

func (tempo Tempo) StartOfDay() Tempo {
	return From(tempo.value.StartOfDay())
}

func (tempo Tempo) EndOfDay() Tempo {
	return From(tempo.value.EndOfDay())
}

func (tempo Tempo) StartOfWeek(options ...tempopkg.StartOfWeekOptions) Tempo {
	return From(tempo.value.StartOfWeek(options...))
}

func (tempo Tempo) EndOfWeek(options ...tempopkg.StartOfWeekOptions) Tempo {
	return From(tempo.value.EndOfWeek(options...))
}

func (tempo Tempo) StartOfMonth() Tempo {
	return From(tempo.value.StartOfMonth())
}

func (tempo Tempo) EndOfMonth() Tempo {
	return From(tempo.value.EndOfMonth())
}

func (tempo Tempo) StartOfQuarter() Tempo {
	return From(tempo.value.StartOfQuarter())
}

func (tempo Tempo) EndOfQuarter() Tempo {
	return From(tempo.value.EndOfQuarter())
}

func (tempo Tempo) StartOfYear() Tempo {
	return From(tempo.value.StartOfYear())
}

func (tempo Tempo) EndOfYear() Tempo {
	return From(tempo.value.EndOfYear())
}

func (tempo Tempo) Floor(unit tempopkg.Unit) Tempo {
	return From(tempo.value.Floor(unit))
}

func (tempo Tempo) Ceil(unit tempopkg.Unit) Tempo {
	return From(tempo.value.Ceil(unit))
}

func (tempo Tempo) Round(unit tempopkg.Unit) Tempo {
	return From(tempo.value.Round(unit))
}

func (tempo Tempo) Next(weekday time.Weekday) Tempo {
	return From(tempo.value.Next(weekday))
}

func (tempo Tempo) Previous(weekday time.Weekday) Tempo {
	return From(tempo.value.Previous(weekday))
}

func (tempo Tempo) NextWeekday() Tempo {
	return From(tempo.value.NextWeekday())
}

func (tempo Tempo) PreviousWeekday() Tempo {
	return From(tempo.value.PreviousWeekday())
}
