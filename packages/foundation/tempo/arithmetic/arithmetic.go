// Package arithmetic exposes generic time-shift operations that work on any
// core.Bearer — both the immutable Time and the mutable *MutableTime — so
// the underlying math is written once and instantiated per concrete type.
//
// Functions here are pure mechanism: callers (typically Time/Mutable method
// shims) decide policy such as month/year overflow by passing the relevant
// flags. The package itself depends only on core, duration, calendar and
// internal/kernel — never on the higher-level tempo package — keeping it
// safe to import from anywhere in the module.
package arithmetic

import (
	"slices"
	"time"

	"alloy.dev/foundation/tempo/core"
	"alloy.dev/foundation/tempo/duration"
	"alloy.dev/foundation/tempo/internal/kernel"
)

func Add[T core.Bearer[T]](bearer T, amount int, unit duration.Unit, monthsOverflow bool, yearsOverflow bool) T {
	state := bearer.State()

	return bearer.With(kernel.Add(state.Value, state.Location, amount, unit, monthsOverflow, yearsOverflow))
}

func Sub[T core.Bearer[T]](bearer T, amount int, unit duration.Unit, monthsOverflow bool, yearsOverflow bool) T {
	return Add(bearer, -amount, unit, monthsOverflow, yearsOverflow)
}

func AddMilliseconds[T core.Bearer[T]](bearer T, milliseconds int) T {
	return Add(bearer, milliseconds, duration.Millisecond, false, false)
}

func SubMilliseconds[T core.Bearer[T]](bearer T, milliseconds int) T {
	return AddMilliseconds(bearer, -milliseconds)
}

func AddSeconds[T core.Bearer[T]](bearer T, seconds int) T {
	return Add(bearer, seconds, duration.Second, false, false)
}

func SubSeconds[T core.Bearer[T]](bearer T, seconds int) T {
	return AddSeconds(bearer, -seconds)
}

func AddMinutes[T core.Bearer[T]](bearer T, minutes int) T {
	return Add(bearer, minutes, duration.Minute, false, false)
}

func SubMinutes[T core.Bearer[T]](bearer T, minutes int) T {
	return AddMinutes(bearer, -minutes)
}

func AddHours[T core.Bearer[T]](bearer T, hours int) T {
	return Add(bearer, hours, duration.Hour, false, false)
}

func SubHours[T core.Bearer[T]](bearer T, hours int) T {
	return AddHours(bearer, -hours)
}

func AddDays[T core.Bearer[T]](bearer T, days int) T {
	return Add(bearer, days, duration.Day, false, false)
}

func SubDays[T core.Bearer[T]](bearer T, days int) T {
	return AddDays(bearer, -days)
}

func AddWeeks[T core.Bearer[T]](bearer T, weeks int) T {
	return Add(bearer, weeks, duration.Week, false, false)
}

func SubWeeks[T core.Bearer[T]](bearer T, weeks int) T {
	return AddWeeks(bearer, -weeks)
}

func AddMonths[T core.Bearer[T]](bearer T, months int, monthsOverflow bool) T {
	if !monthsOverflow {
		return AddMonthsNoOverflow(bearer, months)
	}

	state := bearer.State()

	return bearer.With(state.Value.In(state.Location).AddDate(0, months, 0).UTC())
}

func SubMonths[T core.Bearer[T]](bearer T, months int, monthsOverflow bool) T {
	return AddMonths(bearer, -months, monthsOverflow)
}

func AddMonthsNoOverflow[T core.Bearer[T]](bearer T, months int) T {
	state := bearer.State()

	return bearer.With(kernel.AddMonthsNoOverflow(state.Value, state.Location, months))
}

func SubMonthsNoOverflow[T core.Bearer[T]](bearer T, months int) T {
	return AddMonthsNoOverflow(bearer, -months)
}

func AddQuarters[T core.Bearer[T]](bearer T, quarters int, monthsOverflow bool) T {
	return AddMonths(bearer, quarters*3, monthsOverflow)
}

func SubQuarters[T core.Bearer[T]](bearer T, quarters int, monthsOverflow bool) T {
	return AddQuarters(bearer, -quarters, monthsOverflow)
}

func AddYears[T core.Bearer[T]](bearer T, years int, yearsOverflow bool) T {
	if !yearsOverflow {
		return AddYearsNoOverflow(bearer, years)
	}

	state := bearer.State()

	return bearer.With(state.Value.In(state.Location).AddDate(years, 0, 0).UTC())
}

func SubYears[T core.Bearer[T]](bearer T, years int, yearsOverflow bool) T {
	return AddYears(bearer, -years, yearsOverflow)
}

func AddYearsNoOverflow[T core.Bearer[T]](bearer T, years int) T {
	state := bearer.State()

	return bearer.With(kernel.AddYearsNoOverflow(state.Value, state.Location, years))
}

func SubYearsNoOverflow[T core.Bearer[T]](bearer T, years int) T {
	return AddYearsNoOverflow(bearer, -years)
}

func AddWeekdays[T core.Bearer[T]](bearer T, days int, weekendDays []time.Weekday) T {
	if days == 0 {
		return bearer
	}

	direction := 1

	if days < 0 {
		direction = -1
		days = -days
	}

	current := bearer

	for days > 0 {
		current = AddDays(current, direction)
		state := current.State()
		weekday := state.Value.In(state.Location).Weekday()

		if !isWeekend(weekday, weekendDays) {
			days--
		}
	}

	return current
}

func SubWeekdays[T core.Bearer[T]](bearer T, days int, weekendDays []time.Weekday) T {
	return AddWeekdays(bearer, -days, weekendDays)
}

// AddDuration applies every non-zero field of the multi-unit Duration in a
// fixed order — years → months (rolling quarters into months) → weeks → days
// → hours → minutes → seconds → milliseconds — so calendar-aware shifts land
// before fixed-length ones. The overflow flags propagate to the month/year
// stages exactly as if each AddX had been invoked individually.
func AddDuration[T core.Bearer[T]](bearer T, dur duration.Span, monthsOverflow bool, yearsOverflow bool) T {
	current := bearer
	current = AddYears(current, dur.Years, yearsOverflow)
	current = AddMonths(current, dur.Quarters*3+dur.Months, monthsOverflow)
	current = AddWeeks(current, dur.Weeks)
	current = AddDays(current, dur.Days)
	current = AddHours(current, dur.Hours)
	current = AddMinutes(current, dur.Minutes)
	current = AddSeconds(current, dur.Seconds)
	current = AddMilliseconds(current, dur.Milliseconds)

	return current
}

func SubDuration[T core.Bearer[T]](bearer T, dur duration.Span, monthsOverflow bool, yearsOverflow bool) T {
	return AddDuration(bearer, duration.Span{
		Years:        -dur.Years,
		Quarters:     -dur.Quarters,
		Months:       -dur.Months,
		Weeks:        -dur.Weeks,
		Days:         -dur.Days,
		Hours:        -dur.Hours,
		Minutes:      -dur.Minutes,
		Seconds:      -dur.Seconds,
		Milliseconds: -dur.Milliseconds,
	}, monthsOverflow, yearsOverflow)
}

func isWeekend(weekday time.Weekday, weekendDays []time.Weekday) bool {
	return slices.Contains(weekendDays, weekday)
}
