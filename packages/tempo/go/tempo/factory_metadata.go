package tempo

import (
	"time"

	factorypkg "github.com/oullin/alloy/tempo/factory"
)

func GetDays() []string {
	return append([]string(nil), calendarDays()...)
}

func GetCalendarFormats() map[string]string {
	return factorypkg.CalendarFormats()
}

func GetIsoFormats() map[string]string {
	return factorypkg.ISOFormats()
}

func GetIsoUnits() []Unit {
	return factorypkg.ISOUnits()
}

func GetTimeFormatByPrecision(precision TimeStringPrecision) string {
	return factorypkg.TimeFormatByPrecision(string(precision))
}

func GetWeekStartsAt() time.Weekday { return factorypkg.WeekStartsAt() }

func GetWeekEndsAt() time.Weekday { return factorypkg.WeekEndsAt() }

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
