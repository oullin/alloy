package tempo

import (
	"strings"
	"time"
)

func GetDays() []string {
	return append([]string(nil), calendarDays()...)
}

func GetCalendarFormats() map[string]string {
	return map[string]string{
		"lastDay":  "[Yesterday at] HH:mm",
		"lastWeek": "[Last] dddd [at] HH:mm",
		"nextDay":  "[Tomorrow at] HH:mm",
		"nextWeek": "dddd [at] HH:mm",
		"sameDay":  "[Today at] HH:mm",
		"sameElse": "YYYY-MM-DD",
	}
}

func GetIsoFormats() map[string]string {
	return map[string]string{
		"atom":     "YYYY-MM-DDTHH:mm:ssZZ",
		"cookie":   "ddd, DD-MMM-YYYY HH:mm:ss [GMT]",
		"date":     "YYYY-MM-DD",
		"dateTime": "YYYY-MM-DD HH:mm:ss",
		"iso8601":  "YYYY-MM-DDTHH:mm:ssZZ",
		"time":     "HH:mm:ss",
	}
}

func GetIsoUnits() []Unit {
	return []Unit{Millisecond, Second, Minute, Hour, Day, Week, Month, Quarter, Year}
}

func GetTimeFormatByPrecision(precision TimeStringPrecision) string {
	if precision == MillisecondPrecision {
		return "HH:mm:ss.SSS"
	}

	return "HH:mm:ss"
}

func GetWeekStartsAt() time.Weekday { return time.Monday }

func GetWeekEndsAt() time.Weekday { return time.Sunday }

func LocaleHasDiffSyntax(locale string) bool { return strings.TrimSpace(locale) != "" }

func LocaleHasDiffOneDayWords(locale string) bool { return LocaleHasDiffSyntax(locale) }

func LocaleHasDiffTwoDayWords(locale string) bool { return LocaleHasDiffSyntax(locale) }

func LocaleHasPeriodSyntax(locale string) bool { return LocaleHasDiffSyntax(locale) }

func LocaleHasShortUnits(locale string) bool { return LocaleHasDiffSyntax(locale) }

func Sleep(seconds float64) {
	if seconds <= 0 {
		return
	}

	time.Sleep(time.Duration(seconds * float64(time.Second)))
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

func IsModifiableUnit(unit Unit) bool {
	switch normalizeUnit(unit) {
	case Millisecond, Second, Minute, Hour, Day, Week, Month, Quarter, Year, Decade, Century, Millennium:
		return true
	default:
		return false
	}
}

func SingularUnit(unit Unit) Unit {
	return normalizeUnit(unit)
}

func PluralUnit(unit Unit) string {
	return string(normalizeUnit(unit)) + "s"
}
