package tempo

import "time"

func normalizeUnit(unit Unit) Unit {
	switch unit {
	case "milliseconds":
		return Millisecond
	case "seconds":
		return Second
	case "minutes":
		return Minute
	case "hours":
		return Hour
	case "days":
		return Day
	case "weeks":
		return Week
	case "months":
		return Month
	case "quarters":
		return Quarter
	case "years":
		return Year
	case "decades":
		return Decade
	case "centuries":
		return Century
	case "millenniums", "millennia":
		return Millennium
	default:
		return unit
	}
}

func fixedUnitDuration(unit Unit) (time.Duration, bool) {
	switch normalizeUnit(unit) {
	case Millisecond:
		return time.Millisecond, true
	case Second:
		return time.Second, true
	case Minute:
		return time.Minute, true
	case Hour:
		return time.Hour, true
	case Day:
		return 24 * time.Hour, true
	case Week:
		return 7 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func unitDuration(unit Unit) time.Duration {
	if duration, ok := fixedUnitDuration(unit); ok {
		return duration
	}

	switch normalizeUnit(unit) {
	case Month:
		return 30 * 24 * time.Hour
	case Year:
		return 365 * 24 * time.Hour
	default:
		return time.Millisecond
	}
}

func bestRelativeUnit(milliseconds int64) Unit {
	absolute := milliseconds
	if absolute < 0 {
		absolute = -absolute
	}

	switch {
	case absolute < int64(time.Minute/time.Millisecond):
		return Second
	case absolute < int64(time.Hour/time.Millisecond):
		return Minute
	case absolute < int64((24*time.Hour)/time.Millisecond):
		return Hour
	case absolute < int64((7*24*time.Hour)/time.Millisecond):
		return Day
	case absolute < int64((30*24*time.Hour)/time.Millisecond):
		return Week
	case absolute < int64((365*24*time.Hour)/time.Millisecond):
		return Month
	default:
		return Year
	}
}
