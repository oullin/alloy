package duration

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Unit string

const (
	Millisecond Unit = "millisecond"
	Second      Unit = "second"
	Minute      Unit = "minute"
	Hour        Unit = "hour"
	Day         Unit = "day"
	Week        Unit = "week"
	Month       Unit = "month"
	Quarter     Unit = "quarter"
	Year        Unit = "year"
	Decade      Unit = "decade"
	Century     Unit = "century"
	Millennium  Unit = "millennium"
)

type Duration struct {
	Years        int
	Quarters     int
	Months       int
	Weeks        int
	Days         int
	Hours        int
	Minutes      int
	Seconds      int
	Milliseconds int
}

var durationPattern = regexp.MustCompile(`^(-)?P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)

func Parse(input string) (Duration, error) {
	matches := durationPattern.FindStringSubmatch(input)
	if matches == nil {
		return Duration{}, fmt.Errorf("invalid tempo duration: %s", input)
	}

	sign := 1
	if matches[1] == "-" {
		sign = -1
	}

	seconds := 0
	milliseconds := 0
	if matches[8] != "" {
		value, err := strconv.ParseFloat(matches[8], 64)
		if err != nil {
			return Duration{}, fmt.Errorf("invalid tempo duration seconds: %w", err)
		}
		seconds = int(math.Trunc(value))
		milliseconds = int(math.Round((value - math.Trunc(value)) * 1000))
	}

	return Duration{
		Years:        sign * mustInt(defaultString(matches[2], "0")),
		Months:       sign * mustInt(defaultString(matches[3], "0")),
		Weeks:        sign * mustInt(defaultString(matches[4], "0")),
		Days:         sign * mustInt(defaultString(matches[5], "0")),
		Hours:        sign * mustInt(defaultString(matches[6], "0")),
		Minutes:      sign * mustInt(defaultString(matches[7], "0")),
		Seconds:      sign * seconds,
		Milliseconds: sign * milliseconds,
	}.Normalize(), nil
}

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

func (duration Duration) ToMap() map[string]int {
	return map[string]int{
		"years":        duration.Years,
		"quarters":     duration.Quarters,
		"months":       duration.Months,
		"weeks":        duration.Weeks,
		"days":         duration.Days,
		"hours":        duration.Hours,
		"minutes":      duration.Minutes,
		"seconds":      duration.Seconds,
		"milliseconds": duration.Milliseconds,
	}
}

func (duration Duration) ToArray() [9]int {
	return [9]int{
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
}

func (duration Duration) ISOString() string {
	if duration.IsZero() {
		return "PT0S"
	}

	normalized := duration.Normalize()
	sign := ""
	if normalized.direction() < 0 {
		sign = "-"
	}
	value := normalized.Abs()
	dateParts := strings.Builder{}
	if value.Years != 0 {
		dateParts.WriteString(strconv.Itoa(value.Years) + "Y")
	}
	if value.Months != 0 {
		dateParts.WriteString(strconv.Itoa(value.Months) + "M")
	}
	if value.Days != 0 {
		dateParts.WriteString(strconv.Itoa(value.Days) + "D")
	}

	timeParts := strings.Builder{}
	if value.Hours != 0 {
		timeParts.WriteString(strconv.Itoa(value.Hours) + "H")
	}
	if value.Minutes != 0 {
		timeParts.WriteString(strconv.Itoa(value.Minutes) + "M")
	}
	if value.Seconds != 0 || value.Milliseconds != 0 {
		seconds := strconv.Itoa(value.Seconds)
		if value.Milliseconds != 0 {
			seconds += "." + pad(value.Milliseconds, 3)
		}
		timeParts.WriteString(seconds + "S")
	}

	result := sign + "P" + dateParts.String()
	if timeParts.Len() > 0 {
		result += "T" + timeParts.String()
	}

	return result
}

func (duration Duration) String() string {
	return duration.ISOString()
}

func (duration Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(duration.ISOString())), nil
}

func (duration *Duration) UnmarshalJSON(data []byte) error {
	input, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	parsed, err := Parse(input)
	if err != nil {
		return err
	}

	*duration = parsed
	return nil
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

func NormalizeUnit(unit Unit) Unit {
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

func FixedUnitDuration(unit Unit) (time.Duration, bool) {
	switch NormalizeUnit(unit) {
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

func UnitDuration(unit Unit) time.Duration {
	if duration, ok := FixedUnitDuration(unit); ok {
		return duration
	}

	switch NormalizeUnit(unit) {
	case Month:
		return 30 * 24 * time.Hour
	case Year:
		return 365 * 24 * time.Hour
	default:
		return time.Millisecond
	}
}

func BestRelativeUnit(milliseconds int64) Unit {
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

func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}

func pad(value int, length int) string {
	result := strconv.Itoa(value)
	if value < 0 {
		result = strconv.Itoa(-value)
	}

	for len(result) < length {
		result = "0" + result
	}

	return result
}

func mustInt(input string) int {
	value, _ := strconv.Atoi(input)
	return value
}

func defaultString(input string, fallback string) string {
	if input == "" {
		return fallback
	}

	return input
}
