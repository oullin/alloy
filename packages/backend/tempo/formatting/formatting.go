// Package formatting exposes generic time-string rendering helpers that
// work on any core.Bearer — both the immutable Time and the mutable
// *MutableTime — so Format, ToObject and the projection helpers share
// a single implementation.
//
// The package depends only on core, duration and calendar — never on
// the higher-level tempo package — keeping it safe to import anywhere
// in the module.
package formatting

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"alloy.dev/backend/tempo/calendar"
	"alloy.dev/backend/tempo/core"
)

type Object struct {
	Year          int
	Month         int
	Day           int
	Hour          int
	Minute        int
	Second        int
	Millisecond   int
	Timezone      string
	OffsetMinutes int
	Weekday       int
}

func Format[T core.Bearer[T]](bearer T, pattern string) string {
	state := bearer.State()
	local := state.Value.In(state.Location)
	offset := offsetMinutes(local)
	hour12 := local.Hour() % 12

	if hour12 == 0 {
		hour12 = 12
	}

	milli := local.Nanosecond() / int(time.Millisecond)
	values := map[string]string{
		"A":    ternary(local.Hour() < 12, "AM", "PM"),
		"a":    ternary(local.Hour() < 12, "am", "pm"),
		"D":    strconv.Itoa(local.Day()),
		"DD":   pad(local.Day(), 2),
		"Do":   Ordinal(local.Day()),
		"d":    strconv.Itoa(int(local.Weekday())),
		"ddd":  local.Weekday().String()[:3],
		"dddd": local.Weekday().String(),
		"H":    strconv.Itoa(local.Hour()),
		"HH":   pad(local.Hour(), 2),
		"h":    strconv.Itoa(hour12),
		"hh":   pad(hour12, 2),
		"M":    strconv.Itoa(int(local.Month())),
		"MM":   pad(int(local.Month()), 2),
		"MMM":  local.Month().String()[:3],
		"MMMM": local.Month().String(),
		"m":    strconv.Itoa(local.Minute()),
		"mm":   pad(local.Minute(), 2),
		"S":    strconv.Itoa(milli / 100),
		"SSS":  pad(milli, 3),
		"s":    strconv.Itoa(local.Second()),
		"ss":   pad(local.Second(), 2),
		"X":    strconv.FormatInt(state.Value.Unix(), 10),
		"x":    strconv.FormatInt(state.Value.UnixMilli(), 10),
		"Y":    strconv.Itoa(local.Year()),
		"YY":   pad(local.Year()%100, 2),
		"YYYY": pad(local.Year(), 4),
		"Z":    FormatOffset(offset, ":"),
		"ZZ":   FormatOffset(offset, ""),
	}

	tokens := []string{"YYYY", "MMMM", "dddd", "MMM", "ddd", "SSS", "Do", "YY", "ZZ", "MM", "DD", "HH", "hh", "mm", "ss", "Z", "X", "x", "Y", "M", "D", "H", "h", "m", "s", "A", "a", "d"}

	var builder strings.Builder

	for index := 0; index < len(pattern); {
		if pattern[index] == '[' {
			end := strings.IndexByte(pattern[index:], ']')

			if end >= 0 {
				builder.WriteString(pattern[index+1 : index+end])
				index += end + 1

				continue
			}
		}

		matched := false

		for _, token := range tokens {
			if strings.HasPrefix(pattern[index:], token) {
				builder.WriteString(values[token])
				index += len(token)
				matched = true

				break
			}
		}

		if !matched {
			builder.WriteByte(pattern[index])
			index++
		}
	}

	return builder.String()
}

func ToObject[T core.Bearer[T]](bearer T) Object {
	state := bearer.State()
	local := state.Value.In(state.Location)

	return Object{
		Year:          local.Year(),
		Month:         int(local.Month()),
		Day:           local.Day(),
		Hour:          local.Hour(),
		Minute:        local.Minute(),
		Second:        local.Second(),
		Millisecond:   local.Nanosecond() / int(time.Millisecond),
		Timezone:      state.Location.String(),
		OffsetMinutes: offsetMinutes(local),
		Weekday:       int(local.Weekday()),
	}
}

func ToMap[T core.Bearer[T]](bearer T) map[string]any {
	object := ToObject(bearer)

	return map[string]any{
		"year":          object.Year,
		"month":         object.Month,
		"day":           object.Day,
		"hour":          object.Hour,
		"minute":        object.Minute,
		"second":        object.Second,
		"millisecond":   object.Millisecond,
		"timeZone":      object.Timezone,
		"offsetMinutes": object.OffsetMinutes,
		"weekday":       object.Weekday,
	}
}

func ToArray[T core.Bearer[T]](bearer T) [7]int {
	object := ToObject(bearer)

	return [7]int{
		object.Year,
		object.Month,
		object.Day,
		object.Hour,
		object.Minute,
		object.Second,
		object.Millisecond,
	}
}

func ISOString[T core.Bearer[T]](bearer T) string {
	return bearer.State().Value.UTC().Format("2006-01-02T15:04:05.000Z")
}

func DateString[T core.Bearer[T]](bearer T) string {
	return Format(bearer, "YYYY-MM-DD")
}

func DateTimeString[T core.Bearer[T]](bearer T) string {
	return Format(bearer, "YYYY-MM-DD HH:mm:ss")
}

func TimeString[T core.Bearer[T]](bearer T, includeMilliseconds bool) string {
	base := Format(bearer, "HH:mm:ss")

	if includeMilliseconds {
		object := ToObject(bearer)

		return base + "." + pad(object.Millisecond, 3)
	}

	return base
}

func MonthName(month int) string {
	return calendar.MonthName(month)
}

func ShortMonthName(month int) string {
	return calendar.ShortMonthName(month)
}

func DayName(weekday int) string {
	return calendar.DayName(weekday)
}

func ShortDayName(weekday int) string {
	return calendar.ShortDayName(weekday)
}

func FormatOffset(offsetMinutes int, separator string) string {
	sign := "+"

	if offsetMinutes < 0 {
		sign = "-"
		offsetMinutes = -offsetMinutes
	}

	return fmt.Sprintf("%s%s%s%s", sign, pad(offsetMinutes/60, 2), separator, pad(offsetMinutes%60, 2))
}

func offsetMinutes(local time.Time) int {
	_, offset := local.Zone()

	return offset / 60
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

func Ordinal(value int) string {
	remainder := value % 100
	suffix := "th"

	if remainder >= 11 && remainder <= 13 {
		return strconv.Itoa(value) + suffix
	}

	switch value % 10 {
	case 1:
		suffix = "st"
	case 2:
		suffix = "nd"
	case 3:
		suffix = "rd"
	}

	return strconv.Itoa(value) + suffix
}

func ternary(condition bool, left string, right string) string {
	if condition {
		return left
	}

	return right
}
