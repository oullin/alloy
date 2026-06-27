// Package setters exposes generic field-replacement helpers that work on
// any core.Bearer — both the immutable Time and the mutable
// *MutableTime — so SetYear/SetMonth/SetDay/SetTime and friends share a
// single implementation.
//
// Field setters operate in the bearer's current location and emit a
// fresh time.Time via Bearer.With. Operations that change the bearer's
// location itself (SetTimezone, UTC, Local) cannot be expressed through
// the With contract and stay on the umbrella tempo type.
//
// The package depends only on core, duration and internal/kernel —
// never on the higher-level tempo package — keeping it safe to import
// anywhere in the module.
package setters

import (
	"fmt"
	"time"

	"alloy.dev/api/tempo/arithmetic"
	"alloy.dev/api/tempo/boundaries"
	"alloy.dev/api/tempo/comparison"
	"alloy.dev/api/tempo/core"
	"alloy.dev/api/tempo/duration"
	"alloy.dev/api/tempo/internal/kernel"
)

func SetYear[T core.Bearer[T]](bearer T, year int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(year, local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), state.Location).UTC())
}

func SetMonth[T core.Bearer[T]](bearer T, month int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(local.Year(), time.Month(month), local.Day(), local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), state.Location).UTC())
}

func SetDay[T core.Bearer[T]](bearer T, day int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(local.Year(), local.Month(), day, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), state.Location).UTC())
}

func SetHour[T core.Bearer[T]](bearer T, hour int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(local.Year(), local.Month(), local.Day(), hour, local.Minute(), local.Second(), local.Nanosecond(), state.Location).UTC())
}

func SetMinute[T core.Bearer[T]](bearer T, minute int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), minute, local.Second(), local.Nanosecond(), state.Location).UTC())
}

func SetSecond[T core.Bearer[T]](bearer T, second int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), second, local.Nanosecond(), state.Location).UTC())
}

func SetMillisecond[T core.Bearer[T]](bearer T, millisecond int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second(), millisecond*int(time.Millisecond), state.Location).UTC())
}

func SetDate[T core.Bearer[T]](bearer T, year int, month int, day int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(year, time.Month(month), day, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), state.Location).UTC())
}

func SetTime[T core.Bearer[T]](bearer T, hour int, minute int, second int, millisecond int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return bearer.With(time.Date(local.Year(), local.Month(), local.Day(), hour, minute, second, millisecond*int(time.Millisecond), state.Location).UTC())
}

func SetDateTime[T core.Bearer[T]](bearer T, year int, month int, day int, hour int, minute int, second int, millisecond int) T {
	state := bearer.State()

	return bearer.With(time.Date(year, time.Month(month), day, hour, minute, second, millisecond*int(time.Millisecond), state.Location).UTC())
}

func SetTimestamp[T core.Bearer[T]](bearer T, timestamp int64) T {
	return bearer.With(time.Unix(timestamp, 0).UTC())
}

func SetUnit[T core.Bearer[T]](bearer T, unit duration.Unit, value int) (T, error) {
	var zero T

	switch duration.NormalizeUnit(unit) {
	case duration.Year:
		return SetYear(bearer, value), nil
	case duration.Month:
		return SetMonth(bearer, value), nil
	case duration.Day:
		return SetDay(bearer, value), nil
	case duration.Hour:
		return SetHour(bearer, value), nil
	case duration.Minute:
		return SetMinute(bearer, value), nil
	case duration.Second:
		return SetSecond(bearer, value), nil
	case duration.Millisecond:
		return SetMillisecond(bearer, value), nil
	default:
		return zero, fmt.Errorf("tempo cannot set unit: %s", unit)
	}
}

func SetUnitNoOverflow[T core.Bearer[T]](bearer T, valueUnit duration.Unit, value int, overflowUnit duration.Unit) T {
	next, err := SetUnit(bearer, valueUnit, value)

	if err != nil {
		return bearer
	}

	clamped, err := comparison.Clamp(next, boundaries.StartOf(bearer, overflowUnit).State(), boundaries.EndOf(bearer, overflowUnit).State())

	if err != nil {
		return bearer
	}

	return clamped
}

func SetWeekday[T core.Bearer[T]](bearer T, weekday time.Weekday) T {
	state := bearer.State()
	current := int(state.Value.In(state.Location).Weekday())

	return arithmetic.AddDays(bearer, int(weekday)-current)
}

func SetDayOfYear[T core.Bearer[T]](bearer T, day int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)
	yearStart := boundaries.StartOf(bearer, duration.Year)
	withDay := arithmetic.AddDays(yearStart, day-1)

	return SetTime(withDay, local.Hour(), local.Minute(), local.Second(), local.Nanosecond()/int(time.Millisecond))
}

func SetISODate[T core.Bearer[T]](bearer T, year int, week int, day int) T {
	state := bearer.State()
	local := state.Value.In(state.Location)
	isoStart := bearer.With(kernel.StartOf(time.Date(year, time.January, 4, 0, 0, 0, 0, state.Location).UTC(), state.Location, duration.Week, kernel.WeekOptions{WeekStartsOn: time.Monday}))
	advanced := arithmetic.AddWeeks(isoStart, week-1)
	advanced = arithmetic.AddDays(advanced, day-1)

	return SetTime(advanced, local.Hour(), local.Minute(), local.Second(), local.Nanosecond()/int(time.Millisecond))
}

func Midday[T core.Bearer[T]](bearer T, hour int) T {
	return SetTime(bearer, hour, 0, 0, 0)
}

func SetISOWeek[T core.Bearer[T]](bearer T, week int, days ...int) T {
	state := bearer.State()
	year, _ := state.Value.In(state.Location).ISOWeek()
	day := isoWeekday(state)

	if len(days) > 0 {
		day = days[0]
	}

	return SetISODate(bearer, year, week, day)
}

func SetISOWeekYear[T core.Bearer[T]](bearer T, year int, days ...int) T {
	state := bearer.State()
	_, week := state.Value.In(state.Location).ISOWeek()
	day := isoWeekday(state)

	if len(days) > 0 {
		day = days[0]
	}

	return SetISODate(bearer, year, week, day)
}

func SetISOWeekday[T core.Bearer[T]](bearer T, day int) T {
	if day == 0 {
		day = 7
	}

	state := bearer.State()
	year, week := state.Value.In(state.Location).ISOWeek()

	return SetISODate(bearer, year, week, day)
}

func isoWeekday(state core.State) int {
	weekday := state.Value.In(state.Location).Weekday()

	if weekday == time.Sunday {
		return 7
	}

	return int(weekday)
}
