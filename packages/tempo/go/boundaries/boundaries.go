// Package boundaries exposes generic snap-to-unit operations that work on any
// core.Bearer — both the immutable Tempo and the mutable *MutableTempo — so
// the StartOf/EndOf/Floor/Ceil/Round and Nth-weekday math live in one place.
//
// Functions here are pure mechanism: callers (typically Tempo/Mutable method
// shims) decide policy such as weekend days by passing the relevant slice.
// The package itself depends only on core, duration, calendar and
// internal/kernel — never on the higher-level tempo package — keeping it
// safe to import from anywhere in the module.
package boundaries

import (
	"time"

	"github.com/oullin/alloy/tempo/arithmetic"
	"github.com/oullin/alloy/tempo/core"
	"github.com/oullin/alloy/tempo/duration"
	"github.com/oullin/alloy/tempo/internal/kernel"
)

func StartOf[T core.Bearer[T]](bearer T, unit duration.Unit, options ...kernel.WeekOptions) T {
	state := bearer.State()

	return bearer.With(kernel.StartOf(state.Value, state.Location, unit, options...))
}

func EndOf[T core.Bearer[T]](bearer T, unit duration.Unit, options ...kernel.WeekOptions) T {
	state := bearer.State()

	return bearer.With(kernel.EndOf(state.Value, state.Location, unit, options...))
}

func IsStartOf[T core.Bearer[T]](bearer T, unit duration.Unit, options ...kernel.WeekOptions) bool {
	state := bearer.State()
	target := kernel.StartOf(state.Value, state.Location, unit, options...)

	return state.Value.Equal(target)
}

func IsEndOf[T core.Bearer[T]](bearer T, unit duration.Unit, options ...kernel.WeekOptions) bool {
	state := bearer.State()
	target := kernel.EndOf(state.Value, state.Location, unit, options...)

	return state.Value.Equal(target)
}

func FirstOfMonth[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	first := StartOf(bearer, duration.Month)

	if len(weekdays) == 0 {
		return first
	}

	state := first.State()
	target := weekdays[0]
	delta := (int(target) - int(state.Value.In(state.Location).Weekday()) + 7) % 7

	return arithmetic.AddDays(first, delta)
}

func LastOfMonth[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	last := StartOf(EndOf(bearer, duration.Month), duration.Day)

	if len(weekdays) == 0 {
		return last
	}

	state := last.State()
	target := weekdays[0]
	delta := (int(state.Value.In(state.Location).Weekday()) - int(target) + 7) % 7

	return arithmetic.AddDays(last, -delta)
}

func NthOfMonth[T core.Bearer[T]](bearer T, occurrence int, weekday time.Weekday) (T, bool) {
	var zero T

	if occurrence == 0 {
		return zero, false
	}

	state := bearer.State()
	originMonth := state.Value.In(state.Location).Month()

	var candidate T

	if occurrence > 0 {
		candidate = arithmetic.AddWeeks(FirstOfMonth(bearer, weekday), occurrence-1)
	} else {
		candidate = arithmetic.AddWeeks(LastOfMonth(bearer, weekday), -(abs(occurrence) - 1))
	}

	candidateState := candidate.State()

	return candidate, candidateState.Value.In(candidateState.Location).Month() == originMonth
}

func FirstOfQuarter[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	first := StartOf(bearer, duration.Quarter)

	if len(weekdays) == 0 {
		return first
	}

	state := first.State()
	target := weekdays[0]
	delta := (int(target) - int(state.Value.In(state.Location).Weekday()) + 7) % 7

	return arithmetic.AddDays(first, delta)
}

func LastOfQuarter[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	last := StartOf(EndOf(bearer, duration.Quarter), duration.Day)

	if len(weekdays) == 0 {
		return last
	}

	state := last.State()
	target := weekdays[0]
	delta := (int(state.Value.In(state.Location).Weekday()) - int(target) + 7) % 7

	return arithmetic.AddDays(last, -delta)
}

func NthOfQuarter[T core.Bearer[T]](bearer T, occurrence int, weekday time.Weekday) (T, bool) {
	var zero T

	if occurrence == 0 {
		return zero, false
	}

	state := bearer.State()
	local := state.Value.In(state.Location)
	originQuarter := (int(local.Month())-1)/3 + 1
	originYear := local.Year()

	var candidate T

	if occurrence > 0 {
		candidate = arithmetic.AddWeeks(FirstOfQuarter(bearer, weekday), occurrence-1)
	} else {
		candidate = arithmetic.AddWeeks(LastOfQuarter(bearer, weekday), -(abs(occurrence) - 1))
	}

	candidateState := candidate.State()
	candidateLocal := candidateState.Value.In(candidateState.Location)
	candidateQuarter := (int(candidateLocal.Month())-1)/3 + 1

	return candidate, candidateQuarter == originQuarter && candidateLocal.Year() == originYear
}

func FirstOfYear[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	first := StartOf(bearer, duration.Year)

	if len(weekdays) == 0 {
		return first
	}

	state := first.State()
	target := weekdays[0]
	delta := (int(target) - int(state.Value.In(state.Location).Weekday()) + 7) % 7

	return arithmetic.AddDays(first, delta)
}

func LastOfYear[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	last := StartOf(EndOf(bearer, duration.Year), duration.Day)

	if len(weekdays) == 0 {
		return last
	}

	state := last.State()
	target := weekdays[0]
	delta := (int(state.Value.In(state.Location).Weekday()) - int(target) + 7) % 7

	return arithmetic.AddDays(last, -delta)
}

func NthOfYear[T core.Bearer[T]](bearer T, occurrence int, weekday time.Weekday) (T, bool) {
	var zero T

	if occurrence == 0 {
		return zero, false
	}

	state := bearer.State()
	originYear := state.Value.In(state.Location).Year()

	var candidate T

	if occurrence > 0 {
		candidate = arithmetic.AddWeeks(FirstOfYear(bearer, weekday), occurrence-1)
	} else {
		candidate = arithmetic.AddWeeks(LastOfYear(bearer, weekday), -(abs(occurrence) - 1))
	}

	candidateState := candidate.State()

	return candidate, candidateState.Value.In(candidateState.Location).Year() == originYear
}

func Floor[T core.Bearer[T]](bearer T, unit duration.Unit) T {
	fixed, ok := duration.FixedUnitDuration(unit)

	if !ok {
		return StartOf(bearer, unit)
	}

	state := bearer.State()
	unixNano := state.Value.UnixNano()
	fixedNano := int64(fixed)

	return bearer.With(time.Unix(0, unixNano/fixedNano*fixedNano).UTC())
}

func FloorWeek[T core.Bearer[T]](bearer T, options ...kernel.WeekOptions) T {
	return StartOf(bearer, duration.Week, options...)
}

func Ceil[T core.Bearer[T]](bearer T, unit duration.Unit) T {
	floored := Floor(bearer, unit)
	state := bearer.State()
	flooredState := floored.State()

	if flooredState.Value.Equal(state.Value) {
		return floored
	}

	return bearer.With(kernel.Add(flooredState.Value, flooredState.Location, 1, unit, true, true))
}

func CeilWeek[T core.Bearer[T]](bearer T, options ...kernel.WeekOptions) T {
	floored := FloorWeek(bearer, options...)
	state := bearer.State()
	flooredState := floored.State()

	if flooredState.Value.Equal(state.Value) {
		return floored
	}

	return bearer.With(kernel.Add(flooredState.Value, flooredState.Location, 1, duration.Week, true, true))
}

func Round[T core.Bearer[T]](bearer T, unit duration.Unit) T {
	fixed, ok := duration.FixedUnitDuration(unit)

	if !ok {
		start := StartOf(bearer, unit)
		end := EndOf(bearer, unit)
		startState := start.State()
		endState := end.State()
		state := bearer.State()
		midpoint := startState.Value.UnixMilli() + (endState.Value.UnixMilli()-startState.Value.UnixMilli())/2

		if state.Value.UnixMilli() >= midpoint {
			return Ceil(bearer, unit)
		}

		return start
	}

	state := bearer.State()

	return bearer.With(state.Value.Round(fixed).UTC())
}

func RoundWeek[T core.Bearer[T]](bearer T, options ...kernel.WeekOptions) T {
	start := StartOf(bearer, duration.Week, options...)
	end := EndOf(bearer, duration.Week, options...)
	startState := start.State()
	endState := end.State()
	state := bearer.State()
	midpoint := startState.Value.UnixMilli() + (endState.Value.UnixMilli()-startState.Value.UnixMilli())/2

	if state.Value.UnixMilli() >= midpoint {
		return CeilWeek(bearer, options...)
	}

	return start
}

func Next[T core.Bearer[T]](bearer T, weekday time.Weekday) T {
	state := bearer.State()
	delta := (int(weekday) - int(state.Value.In(state.Location).Weekday()) + 7) % 7

	if delta == 0 {
		delta = 7
	}

	return arithmetic.AddDays(bearer, delta)
}

func Previous[T core.Bearer[T]](bearer T, weekday time.Weekday) T {
	state := bearer.State()
	delta := (int(state.Value.In(state.Location).Weekday()) - int(weekday) + 7) % 7

	if delta == 0 {
		delta = 7
	}

	return arithmetic.AddDays(bearer, -delta)
}

func NextOrSame[T core.Bearer[T]](bearer T, weekday time.Weekday) T {
	state := bearer.State()

	if state.Value.In(state.Location).Weekday() == weekday {
		return bearer.With(state.Value)
	}

	return Next(bearer, weekday)
}

func PreviousOrSame[T core.Bearer[T]](bearer T, weekday time.Weekday) T {
	state := bearer.State()

	if state.Value.In(state.Location).Weekday() == weekday {
		return bearer.With(state.Value)
	}

	return Previous(bearer, weekday)
}

func NextWeekday[T core.Bearer[T]](bearer T, weekendDays []time.Weekday) T {
	current := arithmetic.AddDays(bearer, 1)

	for isWeekend(current, weekendDays) {
		current = arithmetic.AddDays(current, 1)
	}

	return current
}

func PreviousWeekday[T core.Bearer[T]](bearer T, weekendDays []time.Weekday) T {
	current := arithmetic.AddDays(bearer, -1)

	for isWeekend(current, weekendDays) {
		current = arithmetic.AddDays(current, -1)
	}

	return current
}

func NextWeekendDay[T core.Bearer[T]](bearer T, weekendDays []time.Weekday) T {
	current := arithmetic.AddDays(bearer, 1)

	for !isWeekend(current, weekendDays) {
		current = arithmetic.AddDays(current, 1)
	}

	return current
}

func PreviousWeekendDay[T core.Bearer[T]](bearer T, weekendDays []time.Weekday) T {
	current := arithmetic.AddDays(bearer, -1)

	for !isWeekend(current, weekendDays) {
		current = arithmetic.AddDays(current, -1)
	}

	return current
}

func isWeekend[T core.Bearer[T]](bearer T, weekendDays []time.Weekday) bool {
	state := bearer.State()
	weekday := state.Value.In(state.Location).Weekday()

	for _, day := range weekendDays {
		if weekday == day {
			return true
		}
	}

	return false
}

func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
