package tempo

import (
	"time"

	factorypkg "alloy.dev/backend/tempo/factory"
)

func Days() []string {
	return append([]string(nil), calendarDays()...)
}

func CalendarFormats() map[string]string {
	return factorypkg.CalendarFormats()
}

func ISOFormats() map[string]string {
	return factorypkg.ISOFormats()
}

func ISOUnits() []Unit {
	return factorypkg.ISOUnits()
}

func TimeFormatByPrecision(precision TimeStringPrecision) string {
	return factorypkg.TimeFormatByPrecision(string(precision))
}

func WeekStartsAt() time.Weekday { return factorypkg.WeekStartsAt() }

func WeekEndsAt() time.Weekday { return factorypkg.WeekEndsAt() }

func LocaleHasDiffSyntax(locale string) bool { return factorypkg.LocaleHasSyntax(locale) }

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
	return factorypkg.HasRelativeKeywords(input)
}

func IsModifiableUnit(unit Unit) bool {
	return factorypkg.IsModifiableUnit(unit)
}

func SingularUnit(unit Unit) Unit {
	return factorypkg.SingularUnit(unit)
}

func PluralUnit(unit Unit) string {
	return factorypkg.PluralUnit(unit)
}
