package tempo

import "time"

func SettingsState(settings ...Settings) Settings {
	if len(settings) > 0 {
		next := settings[0]
		if next.Locale != "" {
			tempoSettings.Locale = next.Locale
		}
		if next.FallbackLocale != "" {
			tempoSettings.FallbackLocale = next.FallbackLocale
		}
		if next.HumanDiff != (HumanDiffOptions{}) {
			tempoSettings.HumanDiff = next.HumanDiff
		}
		if next.MidDayAt != 0 {
			tempoSettings.MidDayAt = next.MidDayAt
		}
		tempoSettings.MonthsOverflow = next.MonthsOverflow
		tempoSettings.StrictMode = next.StrictMode
		tempoSettings.TestNow = next.TestNow
		if next.Timezone != "" {
			tempoSettings.Timezone = next.Timezone
		}
		if next.WeekendDays != nil {
			tempoSettings.WeekendDays = append([]time.Weekday(nil), next.WeekendDays...)
		}
		tempoSettings.YearsOverflow = next.YearsOverflow
	}

	return Settings{
		FallbackLocale: tempoSettings.FallbackLocale,
		HumanDiff:      tempoSettings.HumanDiff,
		Locale:         tempoSettings.Locale,
		MidDayAt:       tempoSettings.MidDayAt,
		MonthsOverflow: tempoSettings.MonthsOverflow,
		StrictMode:     tempoSettings.StrictMode,
		TestNow:        tempoSettings.TestNow,
		Timezone:       tempoSettings.Timezone,
		WeekendDays:    append([]time.Weekday(nil), tempoSettings.WeekendDays...),
		YearsOverflow:  tempoSettings.YearsOverflow,
	}
}

func GetLocale() string { return tempoSettings.Locale }

func SetLocale(locale string) { tempoSettings.Locale = locale }

func GetFallbackLocale() string { return tempoSettings.FallbackLocale }

func SetFallbackLocale(locale string) { tempoSettings.FallbackLocale = locale }

func GetHumanDiffOptions() HumanDiffOptions { return tempoSettings.HumanDiff }

func SetHumanDiffOptions(options HumanDiffOptions) { tempoSettings.HumanDiff = options }

func GetMidDayAt() int { return tempoSettings.MidDayAt }

func SetMidDayAt(hour int) { tempoSettings.MidDayAt = hour }

func GetWeekendDays() []time.Weekday {
	return append([]time.Weekday(nil), tempoSettings.WeekendDays...)
}

func SetWeekendDays(days []time.Weekday) {
	tempoSettings.WeekendDays = append([]time.Weekday(nil), days...)
}

func ShouldOverflowMonths() bool { return tempoSettings.MonthsOverflow }

func UseMonthsOverflow(enabled bool) { tempoSettings.MonthsOverflow = enabled }

func ResetMonthsOverflow() { tempoSettings.MonthsOverflow = true }

func ShouldOverflowYears() bool { return tempoSettings.YearsOverflow }

func UseYearsOverflow(enabled bool) { tempoSettings.YearsOverflow = enabled }

func ResetYearsOverflow() { tempoSettings.YearsOverflow = true }

func IsStrictModeEnabled() bool { return tempoSettings.StrictMode }

func UseStrictMode(enabled bool) { tempoSettings.StrictMode = enabled }

func GetTestNow() *Tempo { return tempoSettings.TestNow }

func SetTestNow(input *Tempo) { tempoSettings.TestNow = input }

func SetTestNowAndTimezone(input *Tempo, timezone string) {
	tempoSettings.TestNow = input
	tempoSettings.Timezone = timezone
}

func HasTestNow() bool { return tempoSettings.TestNow != nil }
