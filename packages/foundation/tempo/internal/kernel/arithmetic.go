// Package kernel is the pure time math layer that all higher-level tempo
// packages compose on top of. It operates on (time.Time, *time.Location, Unit)
// only — no Time type, no Context, no Settings — so every consumer (immutable
// Time, mutable, feature packages) shares the same engine without
// circular dependencies.
package kernel

import (
	"time"

	"alloy.dev/foundation/tempo/calendar"
	"alloy.dev/foundation/tempo/duration"
)

func Add(value time.Time, location *time.Location, amount int, unit duration.Unit, monthsOverflow bool, yearsOverflow bool) time.Time {
	switch duration.NormalizeUnit(unit) {
	case duration.Millisecond:
		return value.Add(time.Duration(amount) * time.Millisecond).UTC()
	case duration.Second:
		return value.Add(time.Duration(amount) * time.Second).UTC()
	case duration.Minute:
		return value.Add(time.Duration(amount) * time.Minute).UTC()
	case duration.Hour:
		return value.Add(time.Duration(amount) * time.Hour).UTC()
	case duration.Day:
		return addDate(value, location, 0, 0, amount)
	case duration.Week:
		return addDate(value, location, 0, 0, amount*7)
	case duration.Month:
		if monthsOverflow {
			return addDate(value, location, 0, amount, 0)
		}

		return AddMonthsNoOverflow(value, location, amount)
	case duration.Quarter:
		if monthsOverflow {
			return addDate(value, location, 0, amount*3, 0)
		}

		return AddMonthsNoOverflow(value, location, amount*3)
	case duration.Year:
		if yearsOverflow {
			return addDate(value, location, amount, 0, 0)
		}

		return AddYearsNoOverflow(value, location, amount)
	default:
		return value.UTC()
	}
}

func AddMonthsNoOverflow(value time.Time, location *time.Location, months int) time.Time {
	local := value.In(location)
	target := time.Date(local.Year(), local.Month()+time.Month(months), 1, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), location)
	day := min(local.Day(), calendar.DaysInMonth(target.Year(), int(target.Month())))

	return time.Date(
		target.Year(),
		target.Month(),
		day,
		local.Hour(),
		local.Minute(),
		local.Second(),
		local.Nanosecond(),
		location,
	).UTC()
}

func AddYearsNoOverflow(value time.Time, location *time.Location, years int) time.Time {
	local := value.In(location)
	year := local.Year() + years
	day := min(local.Day(), calendar.DaysInMonth(year, int(local.Month())))

	return time.Date(
		year,
		local.Month(),
		day,
		local.Hour(),
		local.Minute(),
		local.Second(),
		local.Nanosecond(),
		location,
	).UTC()
}

func addDate(value time.Time, location *time.Location, years int, months int, days int) time.Time {
	return value.In(location).AddDate(years, months, days).UTC()
}
