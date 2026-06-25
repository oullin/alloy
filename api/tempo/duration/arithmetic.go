package duration

import "time"

func (value *Span) Plus(other Span) Span {
	return Span{
		Years:        value.Years + other.Years,
		Quarters:     value.Quarters + other.Quarters,
		Months:       value.Months + other.Months,
		Weeks:        value.Weeks + other.Weeks,
		Days:         value.Days + other.Days,
		Hours:        value.Hours + other.Hours,
		Minutes:      value.Minutes + other.Minutes,
		Seconds:      value.Seconds + other.Seconds,
		Milliseconds: value.Milliseconds + other.Milliseconds,
	}
}

func (value *Span) Minus(other Span) Span {
	return value.Plus(other.Negated())
}

func (value *Span) Negated() Span {
	return Span{
		Years:        -value.Years,
		Quarters:     -value.Quarters,
		Months:       -value.Months,
		Weeks:        -value.Weeks,
		Days:         -value.Days,
		Hours:        -value.Hours,
		Minutes:      -value.Minutes,
		Seconds:      -value.Seconds,
		Milliseconds: -value.Milliseconds,
	}
}

func (value *Span) Abs() Span {
	return Span{
		Years:        absInt(value.Years),
		Quarters:     absInt(value.Quarters),
		Months:       absInt(value.Months),
		Weeks:        absInt(value.Weeks),
		Days:         absInt(value.Days),
		Hours:        absInt(value.Hours),
		Minutes:      absInt(value.Minutes),
		Seconds:      absInt(value.Seconds),
		Milliseconds: absInt(value.Milliseconds),
	}
}

func (value *Span) Normalize() Span {
	sign := value.direction()
	absolute := value.Abs()
	milliseconds := absolute.Milliseconds
	seconds := absolute.Seconds + milliseconds/1000
	milliseconds %= 1000
	minutes := absolute.Minutes + seconds/60
	seconds %= 60
	hours := absolute.Hours + minutes/60
	minutes %= 60
	days := absolute.Days + hours/24 + absolute.Weeks*7
	hours %= 24
	months := absolute.Months + absolute.Quarters*3
	years := absolute.Years + months/12
	months %= 12

	return Span{
		Years:        sign * years,
		Months:       sign * months,
		Days:         sign * days,
		Hours:        sign * hours,
		Minutes:      sign * minutes,
		Seconds:      sign * seconds,
		Milliseconds: sign * milliseconds,
	}
}

func (value *Span) Total(unit Unit) float64 {
	milliseconds := value.totalMilliseconds()

	if fixed, ok := FixedUnitDuration(unit); ok {
		return float64(milliseconds) / float64(fixed.Milliseconds())
	}

	months := float64(value.Years*12+value.Quarters*3+value.Months) + float64(milliseconds)/float64((30*24*time.Hour).Milliseconds())

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

func (value *Span) IsZero() bool {
	return *value == (Span{})
}

func (value *Span) IsPositive() bool {
	return !value.IsZero() && value.direction() > 0
}

func (value *Span) IsNegative() bool {
	return value.direction() < 0
}

func (value *Span) totalMilliseconds() int64 {
	return int64(value.Weeks*7+value.Days)*int64((24*time.Hour)/time.Millisecond) +
		int64(value.Hours)*int64(time.Hour/time.Millisecond) +
		int64(value.Minutes)*int64(time.Minute/time.Millisecond) +
		int64(value.Seconds)*int64(time.Second/time.Millisecond) +
		int64(value.Milliseconds)
}

func (value *Span) direction() int {
	values := []int{
		value.Years,
		value.Quarters,
		value.Months,
		value.Weeks,
		value.Days,
		value.Hours,
		value.Minutes,
		value.Seconds,
		value.Milliseconds,
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
