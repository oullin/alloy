package kernel

import (
	"time"

	"github.com/oullin/alloy/pkg/hub/tempo/duration"
)

type WeekOptions struct {
	WeekStartsOn time.Weekday
}

func StartOf(value time.Time, location *time.Location, unit duration.Unit, options ...WeekOptions) time.Time {
	local := value.In(location)

	switch duration.NormalizeUnit(unit) {
	case duration.Millisecond:
		return value.UTC()
	case duration.Second:
		return time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second(), 0, location).UTC()
	case duration.Minute:
		return time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), 0, 0, location).UTC()
	case duration.Hour:
		return time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), 0, 0, 0, location).UTC()
	case duration.Day:
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location).UTC()
	case duration.Week:
		weekStartsOn := time.Monday

		if len(options) > 0 {
			weekStartsOn = options[0].WeekStartsOn
		}

		delta := (int(local.Weekday()) - int(weekStartsOn) + 7) % 7

		return Add(StartOf(value, location, duration.Day), location, -delta, duration.Day, true, true)
	case duration.Month:
		return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, location).UTC()
	case duration.Quarter:
		month := time.Month(((int(local.Month())-1)/3)*3 + 1)

		return time.Date(local.Year(), month, 1, 0, 0, 0, 0, location).UTC()
	case duration.Year:
		return time.Date(local.Year(), time.January, 1, 0, 0, 0, 0, location).UTC()
	case duration.Decade:
		return time.Date(local.Year()-local.Year()%10, time.January, 1, 0, 0, 0, 0, location).UTC()
	case duration.Century:
		return time.Date(local.Year()-(local.Year()-1)%100, time.January, 1, 0, 0, 0, 0, location).UTC()
	case duration.Millennium:
		return time.Date(local.Year()-(local.Year()-1)%1000, time.January, 1, 0, 0, 0, 0, location).UTC()
	default:
		return value.UTC()
	}
}

func EndOf(value time.Time, location *time.Location, unit duration.Unit, options ...WeekOptions) time.Time {
	switch duration.NormalizeUnit(unit) {
	case duration.Millisecond:
		return value.UTC()
	case duration.Second:
		return Add(StartOf(value, location, duration.Second), location, 1, duration.Second, true, true).Add(-time.Millisecond).UTC()
	case duration.Minute:
		return Add(StartOf(value, location, duration.Minute), location, 1, duration.Minute, true, true).Add(-time.Millisecond).UTC()
	case duration.Hour:
		return Add(StartOf(value, location, duration.Hour), location, 1, duration.Hour, true, true).Add(-time.Millisecond).UTC()
	case duration.Day:
		return Add(StartOf(value, location, duration.Day), location, 1, duration.Day, true, true).Add(-time.Millisecond).UTC()
	case duration.Week:
		return Add(StartOf(value, location, duration.Week, options...), location, 1, duration.Week, true, true).Add(-time.Millisecond).UTC()
	case duration.Month:
		return Add(StartOf(value, location, duration.Month), location, 1, duration.Month, true, true).Add(-time.Millisecond).UTC()
	case duration.Quarter:
		return Add(StartOf(value, location, duration.Quarter), location, 1, duration.Quarter, true, true).Add(-time.Millisecond).UTC()
	case duration.Year:
		return Add(StartOf(value, location, duration.Year), location, 1, duration.Year, true, true).Add(-time.Millisecond).UTC()
	case duration.Decade:
		return Add(StartOf(value, location, duration.Decade), location, 10, duration.Year, true, true).Add(-time.Millisecond).UTC()
	case duration.Century:
		return Add(StartOf(value, location, duration.Century), location, 100, duration.Year, true, true).Add(-time.Millisecond).UTC()
	case duration.Millennium:
		return Add(StartOf(value, location, duration.Millennium), location, 1000, duration.Year, true, true).Add(-time.Millisecond).UTC()
	default:
		return value.UTC()
	}
}

func CompareValue(value time.Time, location *time.Location, unit duration.Unit) int64 {
	if unit == "" {
		return value.UnixNano()
	}

	return StartOf(value, location, unit).UnixNano()
}
