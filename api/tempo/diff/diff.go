// Package diff exposes generic temporal difference helpers that work on
// any core.Bearer — both the immutable Time and the mutable
// *MutableTime — so Between/InDays/ForHumans share a single
// implementation.
//
// Read-only queries take a core.State for the right-hand side instead
// of another generic Bearer, so callers can diff any concrete bearer
// against any snapshot without paying for a second type parameter.
// Policy bits (HumanDiff defaults, weekend days) flow in as explicit
// arguments — diff itself never consults global settings.
//
// The package depends only on core, duration, calendar and
// internal/kernel — never on the higher-level tempo package — so the
// dependency graph stays acyclic.
package diff

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/oullin/alloy/tempo/arithmetic"
	"github.com/oullin/alloy/tempo/boundaries"
	"github.com/oullin/alloy/tempo/calendar"
	"github.com/oullin/alloy/tempo/comparison"
	"github.com/oullin/alloy/tempo/core"
	"github.com/oullin/alloy/tempo/duration"
)

type Options struct {
	Absolute bool
	Float    bool
}

type HumanOptions struct {
	Absolute bool
	Unit     duration.Unit
}

func Between[T core.Bearer[T]](bearer T, other core.State, unit duration.Unit, options ...Options) float64 {
	opts := Options{}

	if len(options) > 0 {
		opts = options[0]
	}

	state := bearer.State()
	dur := state.Value.Sub(other.Value)
	value := 0.0

	switch duration.NormalizeUnit(unit) {
	case duration.Millisecond:
		value = float64(dur.Milliseconds())
	case duration.Second:
		value = dur.Seconds()
	case duration.Minute:
		value = dur.Minutes()
	case duration.Hour:
		value = dur.Hours()
	case duration.Day:
		value = dur.Hours() / 24
	case duration.Week:
		value = dur.Hours() / (24 * 7)
	case duration.Month, duration.Quarter, duration.Year:
		value = monthDiff(state, other, unit)
	}

	if opts.Absolute {
		value = math.Abs(value)
	}

	if opts.Float {
		return value
	}

	return math.Trunc(value)
}

func InMilliseconds[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Millisecond, options...))
}

func InMicroseconds[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return InMilliseconds(bearer, other, options...) * 1000
}

func InSeconds[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Second, options...))
}

func InMinutes[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Minute, options...))
}

func InHours[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Hour, options...))
}

func InDays[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Day, options...))
}

func InWeeks[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Week, options...))
}

func InMonths[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Month, options...))
}

func InQuarters[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Quarter, options...))
}

func InYears[T core.Bearer[T]](bearer T, other core.State, options ...Options) int {
	return int(Between(bearer, other, duration.Year, options...))
}

func InUnit[T core.Bearer[T]](bearer T, unit duration.Unit, other core.State, options ...Options) int {
	return int(Between(bearer, other, unit, options...))
}

func InDaysFiltered[T core.Bearer[T]](bearer T, other core.State, predicate func(T) bool, options ...Options) int {
	opts := Options{}

	if len(options) > 0 {
		opts = options[0]
	}

	sign := 1
	startBearer := bearer.With(other.Value)
	startBearer = boundaries.StartOf(startBearer, duration.Day)
	endBearer := boundaries.StartOf(bearer, duration.Day)

	if comparison.Before(bearer, other, duration.Day) {
		sign = -1
		startBearer = boundaries.StartOf(bearer, duration.Day)
		endBearer = boundaries.StartOf(bearer.With(other.Value), duration.Day)
	}

	current := startBearer
	count := 0
	endState := endBearer.State()

	for comparison.Before(current, endState, duration.Day) {
		current = arithmetic.AddDays(current, 1)

		if comparison.SameOrBefore(current, endState, duration.Day) && predicate(current) {
			count++
		}
	}

	if opts.Absolute {
		return count
	}

	return count * sign
}

func InHoursFiltered[T core.Bearer[T]](bearer T, other core.State, predicate func(T) bool, options ...Options) int {
	opts := Options{}

	if len(options) > 0 {
		opts = options[0]
	}

	sign := 1
	startBearer := boundaries.StartOf(bearer.With(other.Value), duration.Hour)
	endBearer := boundaries.StartOf(bearer, duration.Hour)

	if comparison.Before(bearer, other, duration.Hour) {
		sign = -1
		startBearer = boundaries.StartOf(bearer, duration.Hour)
		endBearer = boundaries.StartOf(bearer.With(other.Value), duration.Hour)
	}

	current := startBearer
	endState := endBearer.State()
	count := 0

	for comparison.Before(current, endState, duration.Hour) {
		current = arithmetic.AddHours(current, 1)

		if comparison.SameOrBefore(current, endState, duration.Hour) && predicate(current) {
			count++
		}
	}

	if opts.Absolute || sign > 0 {
		return count
	}

	return -count
}

func InWeekdays[T core.Bearer[T]](bearer T, other core.State, weekendDays []time.Weekday, options ...Options) int {
	return InDaysFiltered(bearer, other, func(item T) bool {
		state := item.State()
		weekday := state.Value.In(state.Location).Weekday()

		return !slices.Contains(weekendDays, weekday)
	}, options...)
}

func InWeekendDays[T core.Bearer[T]](bearer T, other core.State, weekendDays []time.Weekday, options ...Options) int {
	return InDaysFiltered(bearer, other, func(item T) bool {
		state := item.State()
		weekday := state.Value.In(state.Location).Weekday()

		return slices.Contains(weekendDays, weekday)
	}, options...)
}

func SecondsSinceMidnight[T core.Bearer[T]](bearer T) int {
	startOfDay := boundaries.StartOf(bearer, duration.Day)

	return InSeconds(bearer, startOfDay.State(), Options{Absolute: true})
}

func SecondsUntilEndOfDay[T core.Bearer[T]](bearer T) int {
	endOfDay := boundaries.EndOf(bearer, duration.Day)

	return InSeconds(bearer, endOfDay.State(), Options{Absolute: true})
}

func ForHumans[T core.Bearer[T]](bearer T, other core.State, options HumanOptions) string {
	state := bearer.State()
	milliseconds := state.Value.UnixMilli() - other.Value.UnixMilli()
	unit := options.Unit

	if unit == "" {
		unit = duration.BestRelativeUnit(milliseconds)
	}

	value := int(math.Round(float64(milliseconds) / float64(duration.UnitDuration(unit).Milliseconds())))

	if options.Absolute && value < 0 {
		value = -value
	}

	unitName := string(duration.NormalizeUnit(unit))

	if value == 1 || value == -1 {
		unitName = strings.TrimSuffix(unitName, "s")
	} else {
		unitName += "s"
	}

	if value < 0 {
		return fmt.Sprintf("%d %s ago", -value, unitName)
	}

	return fmt.Sprintf("in %d %s", value, unitName)
}

func monthDiff(left core.State, right core.State, unit duration.Unit) float64 {
	sign := 1.0
	start := right
	end := left

	if left.Value.Before(right.Value) {
		sign = -1
		start = left
		end = right
	}

	startLocal := start.Value.In(start.Location)
	endLocal := end.Value.In(end.Location)
	value := float64(calendar.MonthDiff(
		startLocal.Year(),
		int(startLocal.Month()),
		startLocal.Day(),
		endLocal.Year(),
		int(endLocal.Month()),
		endLocal.Day(),
	))

	switch duration.NormalizeUnit(unit) {
	case duration.Quarter:
		value /= 3
	case duration.Year:
		value /= 12
	}

	return value * sign
}
