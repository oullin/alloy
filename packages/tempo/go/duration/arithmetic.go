package duration

import "time"

func (duration Duration) Plus(other Duration) Duration {
	return Duration{
		Years:        duration.Years + other.Years,
		Quarters:     duration.Quarters + other.Quarters,
		Months:       duration.Months + other.Months,
		Weeks:        duration.Weeks + other.Weeks,
		Days:         duration.Days + other.Days,
		Hours:        duration.Hours + other.Hours,
		Minutes:      duration.Minutes + other.Minutes,
		Seconds:      duration.Seconds + other.Seconds,
		Milliseconds: duration.Milliseconds + other.Milliseconds,
	}
}

func (duration Duration) Minus(other Duration) Duration {
	return duration.Plus(other.Negated())
}

func (duration Duration) Negated() Duration {
	return Duration{
		Years:        -duration.Years,
		Quarters:     -duration.Quarters,
		Months:       -duration.Months,
		Weeks:        -duration.Weeks,
		Days:         -duration.Days,
		Hours:        -duration.Hours,
		Minutes:      -duration.Minutes,
		Seconds:      -duration.Seconds,
		Milliseconds: -duration.Milliseconds,
	}
}

func (duration Duration) Abs() Duration {
	return Duration{
		Years:        absInt(duration.Years),
		Quarters:     absInt(duration.Quarters),
		Months:       absInt(duration.Months),
		Weeks:        absInt(duration.Weeks),
		Days:         absInt(duration.Days),
		Hours:        absInt(duration.Hours),
		Minutes:      absInt(duration.Minutes),
		Seconds:      absInt(duration.Seconds),
		Milliseconds: absInt(duration.Milliseconds),
	}
}

func (duration Duration) Normalize() Duration {
	sign := duration.direction()
	value := duration.Abs()
	milliseconds := value.Milliseconds
	seconds := value.Seconds + milliseconds/1000
	milliseconds %= 1000
	minutes := value.Minutes + seconds/60
	seconds %= 60
	hours := value.Hours + minutes/60
	minutes %= 60
	days := value.Days + hours/24 + value.Weeks*7
	hours %= 24
	months := value.Months + value.Quarters*3
	years := value.Years + months/12
	months %= 12

	return Duration{
		Years:        sign * years,
		Months:       sign * months,
		Days:         sign * days,
		Hours:        sign * hours,
		Minutes:      sign * minutes,
		Seconds:      sign * seconds,
		Milliseconds: sign * milliseconds,
	}
}

func (duration Duration) Total(unit Unit) float64 {
	milliseconds := duration.totalMilliseconds()

	if fixed, ok := FixedUnitDuration(unit); ok {
		return float64(milliseconds) / float64(fixed.Milliseconds())
	}

	months := float64(duration.Years*12+duration.Quarters*3+duration.Months) + float64(milliseconds)/float64((30*24*time.Hour).Milliseconds())

	switch NormalizeUnit(unit) {
	case Month:
		return months
	case Quarter:
		return months / 3
	case Year:
		return months / 12
	default:
		return float64(milliseconds)
	}
}

func (duration Duration) IsZero() bool {
	return duration == (Duration{})
}

func (duration Duration) IsPositive() bool {
	return !duration.IsZero() && duration.direction() > 0
}

func (duration Duration) IsNegative() bool {
	return duration.direction() < 0
}

func (duration Duration) totalMilliseconds() int64 {
	return int64(duration.Weeks*7+duration.Days)*int64((24*time.Hour)/time.Millisecond) +
		int64(duration.Hours)*int64(time.Hour/time.Millisecond) +
		int64(duration.Minutes)*int64(time.Minute/time.Millisecond) +
		int64(duration.Seconds)*int64(time.Second/time.Millisecond) +
		int64(duration.Milliseconds)
}

func (duration Duration) direction() int {
	values := []int{
		duration.Years,
		duration.Quarters,
		duration.Months,
		duration.Weeks,
		duration.Days,
		duration.Hours,
		duration.Minutes,
		duration.Seconds,
		duration.Milliseconds,
	}

	for _, value := range values {
		if value < 0 {
			return -1
		}

		if value > 0 {
			return 1
		}
	}

	return 1
}
