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
	"slices"
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
	return firstOfUnit(bearer, duration.Month, weekdays...)
}

func LastOfMonth[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	return lastOfUnit(bearer, duration.Month, weekdays...)
}

func NthOfMonth[T core.Bearer[T]](bearer T, occurrence int, weekday time.Weekday) (T, bool) {
	return nthOfUnit(bearer, occurrence, weekday, duration.Month, sameMonth)
}

func FirstOfQuarter[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	return firstOfUnit(bearer, duration.Quarter, weekdays...)
}

func LastOfQuarter[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	return lastOfUnit(bearer, duration.Quarter, weekdays...)
}

func NthOfQuarter[T core.Bearer[T]](bearer T, occurrence int, weekday time.Weekday) (T, bool) {
	return nthOfUnit(bearer, occurrence, weekday, duration.Quarter, sameQuarter)
}

func FirstOfYear[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	return firstOfUnit(bearer, duration.Year, weekdays...)
}

func LastOfYear[T core.Bearer[T]](bearer T, weekdays ...time.Weekday) T {
	return lastOfUnit(bearer, duration.Year, weekdays...)
}

func NthOfYear[T core.Bearer[T]](bearer T, occurrence int, weekday time.Weekday) (T, bool) {
	return nthOfUnit(bearer, occurrence, weekday, duration.Year, sameYear)
}

func firstOfUnit[T core.Bearer[T]](bearer T, unit duration.Unit, weekdays ...time.Weekday) T {
	first := StartOf(bearer, unit)

	if len(weekdays) == 0 {
		return first
	}

	state := first.State()
	target := weekdays[0]
	delta := (int(target) - int(state.Value.In(state.Location).Weekday()) + 7) % 7

	return arithmetic.AddDays(first, delta)
}

func lastOfUnit[T core.Bearer[T]](bearer T, unit duration.Unit, weekdays ...time.Weekday) T {
	last := StartOf(EndOf(bearer, unit), duration.Day)

	if len(weekdays) == 0 {
		return last
	}

	state := last.State()
	target := weekdays[0]
	delta := (int(state.Value.In(state.Location).Weekday()) - int(target) + 7) % 7

	return arithmetic.AddDays(last, -delta)
}

func nthOfUnit[T core.Bearer[T]](
	bearer T,
	occurrence int,
	weekday time.Weekday,
	unit duration.Unit,
	sameUnit func(core.State, core.State) bool,
) (T, bool) {
	var zero T

	if occurrence == 0 {
		return zero, false
	}

	var candidate T

	if occurrence > 0 {
		candidate = arithmetic.AddWeeks(firstOfUnit(bearer, unit, weekday), occurrence-1)
	} else {
		candidate = arithmetic.AddWeeks(lastOfUnit(bearer, unit, weekday), -(abs(occurrence) - 1))
	}

	return candidate, sameUnit(bearer.State(), candidate.State())
}

func sameMonth(origin core.State, candidate core.State) bool {
	originMonth := origin.Value.In(origin.Location).Month()
	candidateMonth := candidate.Value.In(candidate.Location).Month()

	return candidateMonth == originMonth
}

func sameQuarter(origin core.State, candidate core.State) bool {
	originLocal := origin.Value.In(origin.Location)
	candidateLocal := candidate.Value.In(candidate.Location)

	return quarterOf(candidateLocal) == quarterOf(originLocal) && candidateLocal.Year() == originLocal.Year()
}

func sameYear(origin core.State, candidate core.State) bool {
	return candidate.Value.In(candidate.Location).Year() == origin.Value.In(origin.Location).Year()
}

func quarterOf(value time.Time) int {
	return (int(value.Month())-1)/3 + 1
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

	return slices.Contains(weekendDays, weekday)
}

func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
