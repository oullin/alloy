package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/temporal"
)

func (tempo Tempo) StartOf(unit Unit, options ...StartOfWeekOptions) Tempo {
	weekOptions := make([]temporal.WeekOptions, 0, len(options))
	for _, option := range options {
		weekOptions = append(weekOptions, temporal.WeekOptions{WeekStartsOn: option.WeekStartsOn})
	}

	return tempo.with(temporal.StartOf(tempo.value, tempo.location, unit, weekOptions...), tempo.location)
}

func (tempo Tempo) EndOf(unit Unit, options ...StartOfWeekOptions) Tempo {
	weekOptions := make([]temporal.WeekOptions, 0, len(options))
	for _, option := range options {
		weekOptions = append(weekOptions, temporal.WeekOptions{WeekStartsOn: option.WeekStartsOn})
	}

	return tempo.with(temporal.EndOf(tempo.value, tempo.location, unit, weekOptions...), tempo.location)
}

func (tempo Tempo) IsStartOf(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.Same(tempo.StartOf(unit, options...))
}

func (tempo Tempo) IsStartOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.IsStartOf(unit, options...)
}

func (tempo Tempo) IsEndOf(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.Same(tempo.EndOf(unit, options...))
}

func (tempo Tempo) IsEndOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return tempo.IsEndOf(unit, options...)
}

func (tempo Tempo) IsStartOfTime() bool {
	return tempo.value.UnixMilli() <= -8640000000000000
}

func (tempo Tempo) IsEndOfTime() bool {
	return tempo.value.UnixMilli() >= 8640000000000000
}

func (tempo Tempo) IsCurrentUnit(unit Unit, reference Tempo) bool {
	return tempo.Same(reference, unit)
}

func (tempo Tempo) IsStartOfMillisecond() bool {
	return tempo.IsStartOf(Millisecond)
}

func (tempo Tempo) IsEndOfMillisecond() bool {
	return tempo.IsEndOf(Millisecond)
}

func (tempo Tempo) IsStartOfSecond() bool {
	return tempo.IsStartOf(Second)
}

func (tempo Tempo) IsEndOfSecond() bool {
	return tempo.IsEndOf(Second)
}

func (tempo Tempo) IsStartOfMinute() bool {
	return tempo.IsStartOf(Minute)
}

func (tempo Tempo) IsEndOfMinute() bool {
	return tempo.IsEndOf(Minute)
}

func (tempo Tempo) IsStartOfHour() bool {
	return tempo.IsStartOf(Hour)
}

func (tempo Tempo) IsEndOfHour() bool {
	return tempo.IsEndOf(Hour)
}

func (tempo Tempo) StartOfMillisecond() Tempo {
	return tempo.StartOf(Millisecond)
}

func (tempo Tempo) EndOfMillisecond() Tempo {
	return tempo.EndOf(Millisecond)
}

func (tempo Tempo) StartOfSecond() Tempo {
	return tempo.StartOf(Second)
}

func (tempo Tempo) EndOfSecond() Tempo {
	return tempo.EndOf(Second)
}

func (tempo Tempo) StartOfMinute() Tempo {
	return tempo.StartOf(Minute)
}

func (tempo Tempo) EndOfMinute() Tempo {
	return tempo.EndOf(Minute)
}

func (tempo Tempo) StartOfHour() Tempo {
	return tempo.StartOf(Hour)
}

func (tempo Tempo) EndOfHour() Tempo {
	return tempo.EndOf(Hour)
}

func (tempo Tempo) IsStartOfDay() bool {
	return tempo.IsStartOf(Day)
}

func (tempo Tempo) IsEndOfDay() bool {
	return tempo.IsEndOf(Day)
}

func (tempo Tempo) IsStartOfWeek(options ...StartOfWeekOptions) bool {
	return tempo.IsStartOf(Week, options...)
}

func (tempo Tempo) IsEndOfWeek(options ...StartOfWeekOptions) bool {
	return tempo.IsEndOf(Week, options...)
}

func (tempo Tempo) IsStartOfMonth() bool {
	return tempo.IsStartOf(Month)
}

func (tempo Tempo) IsEndOfMonth() bool {
	return tempo.IsEndOf(Month)
}

func (tempo Tempo) IsStartOfQuarter() bool {
	return tempo.IsStartOf(Quarter)
}

func (tempo Tempo) IsEndOfQuarter() bool {
	return tempo.IsEndOf(Quarter)
}

func (tempo Tempo) IsStartOfYear() bool {
	return tempo.IsStartOf(Year)
}

func (tempo Tempo) IsEndOfYear() bool {
	return tempo.IsEndOf(Year)
}

func (tempo Tempo) IsStartOfDecade() bool {
	return tempo.IsStartOf(Decade)
}

func (tempo Tempo) IsEndOfDecade() bool {
	return tempo.IsEndOf(Decade)
}

func (tempo Tempo) IsStartOfCentury() bool {
	return tempo.IsStartOf(Century)
}

func (tempo Tempo) IsEndOfCentury() bool {
	return tempo.IsEndOf(Century)
}

func (tempo Tempo) IsStartOfMillennium() bool {
	return tempo.IsStartOf(Millennium)
}

func (tempo Tempo) IsEndOfMillennium() bool {
	return tempo.IsEndOf(Millennium)
}

func (tempo Tempo) StartOfDay() Tempo {
	return tempo.StartOf(Day)
}

func (tempo Tempo) EndOfDay() Tempo {
	return tempo.EndOf(Day)
}

func (tempo Tempo) StartOfWeek(options ...StartOfWeekOptions) Tempo {
	return tempo.StartOf(Week, options...)
}

func (tempo Tempo) EndOfWeek(options ...StartOfWeekOptions) Tempo {
	return tempo.EndOf(Week, options...)
}

func (tempo Tempo) StartOfMonth() Tempo {
	return tempo.StartOf(Month)
}

func (tempo Tempo) EndOfMonth() Tempo {
	return tempo.EndOf(Month)
}

func (tempo Tempo) StartOfQuarter() Tempo {
	return tempo.StartOf(Quarter)
}

func (tempo Tempo) EndOfQuarter() Tempo {
	return tempo.EndOf(Quarter)
}

func (tempo Tempo) FirstOfMonth(weekdays ...time.Weekday) Tempo {
	first := tempo.StartOf(Month)
	if len(weekdays) == 0 {
		return first
	}

	target := weekdays[0]
	delta := (int(target) - int(first.local().Weekday()) + 7) % 7
	return first.AddDays(delta)
}

func (tempo Tempo) LastOfMonth(weekdays ...time.Weekday) Tempo {
	last := tempo.EndOf(Month).StartOf(Day)
	if len(weekdays) == 0 {
		return last
	}

	target := weekdays[0]
	delta := (int(last.local().Weekday()) - int(target) + 7) % 7
	return last.SubDays(delta)
}

func (tempo Tempo) NthOfMonth(occurrence int, weekday time.Weekday) (Tempo, bool) {
	if occurrence == 0 {
		return Tempo{}, false
	}

	month := tempo.Month()
	var candidate Tempo
	if occurrence > 0 {
		candidate = tempo.FirstOfMonth(weekday).AddWeeks(occurrence - 1)
	} else {
		candidate = tempo.LastOfMonth(weekday).SubWeeks(absInt(occurrence) - 1)
	}

	return candidate, candidate.Month() == month
}

func (tempo Tempo) FirstOfQuarter(weekdays ...time.Weekday) Tempo {
	first := tempo.StartOf(Quarter)
	if len(weekdays) == 0 {
		return first
	}

	target := weekdays[0]
	delta := (int(target) - int(first.local().Weekday()) + 7) % 7
	return first.AddDays(delta)
}

func (tempo Tempo) LastOfQuarter(weekdays ...time.Weekday) Tempo {
	last := tempo.EndOf(Quarter).StartOf(Day)
	if len(weekdays) == 0 {
		return last
	}

	target := weekdays[0]
	delta := (int(last.local().Weekday()) - int(target) + 7) % 7
	return last.SubDays(delta)
}

func (tempo Tempo) NthOfQuarter(occurrence int, weekday time.Weekday) (Tempo, bool) {
	if occurrence == 0 {
		return Tempo{}, false
	}

	quarter := tempo.Quarter()
	year := tempo.Year()
	candidate := tempo.FirstOfQuarter(weekday).AddWeeks(occurrence - 1)
	if occurrence < 0 {
		candidate = tempo.LastOfQuarter(weekday).SubWeeks(absInt(occurrence) - 1)
	}

	return candidate, candidate.Quarter() == quarter && candidate.Year() == year
}

func (tempo Tempo) StartOfYear() Tempo {
	return tempo.StartOf(Year)
}

func (tempo Tempo) EndOfYear() Tempo {
	return tempo.EndOf(Year)
}

func (tempo Tempo) StartOfDecade() Tempo {
	return tempo.StartOf(Decade)
}

func (tempo Tempo) EndOfDecade() Tempo {
	return tempo.EndOf(Decade)
}

func (tempo Tempo) StartOfCentury() Tempo {
	return tempo.StartOf(Century)
}

func (tempo Tempo) EndOfCentury() Tempo {
	return tempo.EndOf(Century)
}

func (tempo Tempo) StartOfMillennium() Tempo {
	return tempo.StartOf(Millennium)
}

func (tempo Tempo) EndOfMillennium() Tempo {
	return tempo.EndOf(Millennium)
}

func (tempo Tempo) FirstOfYear(weekdays ...time.Weekday) Tempo {
	first := tempo.StartOf(Year)
	if len(weekdays) == 0 {
		return first
	}

	target := weekdays[0]
	delta := (int(target) - int(first.local().Weekday()) + 7) % 7
	return first.AddDays(delta)
}

func (tempo Tempo) LastOfYear(weekdays ...time.Weekday) Tempo {
	last := tempo.EndOf(Year).StartOf(Day)
	if len(weekdays) == 0 {
		return last
	}

	target := weekdays[0]
	delta := (int(last.local().Weekday()) - int(target) + 7) % 7
	return last.SubDays(delta)
}

func (tempo Tempo) NthOfYear(occurrence int, weekday time.Weekday) (Tempo, bool) {
	if occurrence == 0 {
		return Tempo{}, false
	}

	year := tempo.Year()
	candidate := tempo.FirstOfYear(weekday).AddWeeks(occurrence - 1)
	if occurrence < 0 {
		candidate = tempo.LastOfYear(weekday).SubWeeks(absInt(occurrence) - 1)
	}

	return candidate, candidate.Year() == year
}

func (tempo Tempo) Floor(unit Unit) Tempo {
	fixed, ok := fixedUnitDuration(unit)
	if !ok {
		return tempo.StartOf(unit)
	}

	unixNano := tempo.value.UnixNano()
	fixedNano := int64(fixed)

	return Tempo{value: time.Unix(0, unixNano/fixedNano*fixedNano).UTC(), location: tempo.location}
}

func (tempo Tempo) FloorUnit(unit Unit) Tempo {
	return tempo.Floor(unit)
}

func (tempo Tempo) FloorWeek(options ...StartOfWeekOptions) Tempo {
	return tempo.StartOfWeek(options...)
}

func (tempo Tempo) Ceil(unit Unit) Tempo {
	floored := tempo.Floor(unit)
	if floored.Same(tempo) {
		return floored
	}

	return floored.Add(1, unit)
}

func (tempo Tempo) CeilUnit(unit Unit) Tempo {
	return tempo.Ceil(unit)
}

func (tempo Tempo) CeilWeek(options ...StartOfWeekOptions) Tempo {
	floored := tempo.FloorWeek(options...)
	if floored.Same(tempo) {
		return floored
	}

	return floored.AddWeeks(1)
}

func (tempo Tempo) Round(unit Unit) Tempo {
	fixed, ok := fixedUnitDuration(unit)
	if !ok {
		start := tempo.StartOf(unit)
		end := tempo.EndOf(unit)
		midpoint := start.TimestampMs() + (end.TimestampMs()-start.TimestampMs())/2
		if tempo.TimestampMs() >= midpoint {
			return tempo.Ceil(unit)
		}

		return start
	}

	return Tempo{value: tempo.value.Round(fixed).UTC(), location: tempo.location}
}

func (tempo Tempo) RoundUnit(unit Unit) Tempo {
	return tempo.Round(unit)
}

func (tempo Tempo) RoundWeek(options ...StartOfWeekOptions) Tempo {
	start := tempo.StartOfWeek(options...)
	end := tempo.EndOfWeek(options...)
	midpoint := start.TimestampMs() + (end.TimestampMs()-start.TimestampMs())/2
	if tempo.TimestampMs() >= midpoint {
		return tempo.CeilWeek(options...)
	}

	return start
}

func (tempo Tempo) Next(weekday time.Weekday) Tempo {
	delta := (int(weekday) - int(tempo.local().Weekday()) + 7) % 7
	if delta == 0 {
		delta = 7
	}

	return tempo.AddDays(delta)
}

func (tempo Tempo) Previous(weekday time.Weekday) Tempo {
	delta := (int(tempo.local().Weekday()) - int(weekday) + 7) % 7
	if delta == 0 {
		delta = 7
	}

	return tempo.SubDays(delta)
}

func (tempo Tempo) NextOrSame(weekday time.Weekday) Tempo {
	if tempo.local().Weekday() == weekday {
		return tempo.Clone()
	}

	return tempo.Next(weekday)
}

func (tempo Tempo) PreviousOrSame(weekday time.Weekday) Tempo {
	if tempo.local().Weekday() == weekday {
		return tempo.Clone()
	}

	return tempo.Previous(weekday)
}

func (tempo Tempo) NextWeekday() Tempo {
	next := tempo.AddDays(1)
	for next.IsWeekend() {
		next = next.AddDays(1)
	}

	return next
}

func (tempo Tempo) PreviousWeekday() Tempo {
	previous := tempo.SubDays(1)
	for previous.IsWeekend() {
		previous = previous.SubDays(1)
	}

	return previous
}

func (tempo Tempo) NextWeekendDay() Tempo {
	next := tempo.AddDays(1)
	for next.IsWeekday() {
		next = next.AddDays(1)
	}

	return next
}

func (tempo Tempo) PreviousWeekendDay() Tempo {
	previous := tempo.SubDays(1)
	for previous.IsWeekday() {
		previous = previous.SubDays(1)
	}

	return previous
}
