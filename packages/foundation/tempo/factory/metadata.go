package factory

import (
	"strings"
	"time"

	"github.com/oullin/alloy/packages/foundation/tempo/duration"
)

func CalendarFormats() map[string]string {
	return map[string]string{
		"lastDay":  "[Yesterday at] HH:mm",
		"lastWeek": "[Last] dddd [at] HH:mm",
		"nextDay":  "[Tomorrow at] HH:mm",
		"nextWeek": "dddd [at] HH:mm",
		"sameDay":  "[Today at] HH:mm",
		"sameElse": "YYYY-MM-DD",
	}
}

func ISOFormats() map[string]string {
	return map[string]string{
		"atom":     "YYYY-MM-DDTHH:mm:ssZZ",
		"cookie":   "ddd, DD-MMM-YYYY HH:mm:ss [GMT]",
		"date":     "YYYY-MM-DD",
		"dateTime": "YYYY-MM-DD HH:mm:ss",
		"iso8601":  "YYYY-MM-DDTHH:mm:ssZZ",
		"time":     "HH:mm:ss",
	}
}

func ISOUnits() []duration.Unit {
	return []duration.Unit{
		duration.Millisecond,
		duration.Second,
		duration.Minute,
		duration.Hour,
		duration.Day,
		duration.Week,
		duration.Month,
		duration.Quarter,
		duration.Year,
	}
}

func TimeFormatByPrecision(precision string) string {
	if precision == "millisecond" {
		return "HH:mm:ss.SSS"
	}

	return "HH:mm:ss"
}

func WeekStartsAt() time.Weekday {
	return time.Monday
}

func WeekEndsAt() time.Weekday {
	return time.Sunday
}

func LocaleHasSyntax(locale string) bool {
	return strings.TrimSpace(locale) != ""
}

func HasRelativeKeywords(input string) bool {
	value := strings.TrimSpace(strings.ToLower(input))

	return strings.HasPrefix(value, "+") ||
		strings.HasPrefix(value, "-") ||
		strings.Contains(value, "now") ||
		strings.Contains(value, "today") ||
		strings.Contains(value, "tomorrow") ||
		strings.Contains(value, "yesterday") ||
		strings.Contains(value, "next") ||
		strings.Contains(value, "last") ||
		strings.Contains(value, "ago")
}

func IsModifiableUnit(unit duration.Unit) bool {
	switch duration.NormalizeUnit(unit) {
	case duration.Millisecond, duration.Second, duration.Minute, duration.Hour, duration.Day, duration.Week, duration.Month, duration.Quarter, duration.Year, duration.Decade, duration.Century, duration.Millennium:
		return true
	default:
		return false
	}
}

func SingularUnit(unit duration.Unit) duration.Unit {
	return duration.NormalizeUnit(unit)
}

func PluralUnit(unit duration.Unit) string {
	return string(duration.NormalizeUnit(unit)) + "s"
}
