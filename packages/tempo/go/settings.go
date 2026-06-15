package tempo

import "time"

func SettingsState(settings ...Settings) Settings {
	if len(settings) > 0 {
		next := settings[0]
		if next.Locale != "" {
			defaultConfig.Settings.Locale = next.Locale
		}
		if next.FallbackLocale != "" {
			defaultConfig.Settings.FallbackLocale = next.FallbackLocale
		}
		if next.HumanDiff != (HumanDiffOptions{}) {
			defaultConfig.Settings.HumanDiff = next.HumanDiff
		}
		if next.MidDayAt != 0 {
			defaultConfig.Settings.MidDayAt = next.MidDayAt
		}
		defaultConfig.Settings.MonthsOverflow = next.MonthsOverflow
		defaultConfig.Settings.StrictMode = next.StrictMode
		defaultConfig.Settings.TestNow = next.TestNow
		if next.Timezone != "" {
			defaultConfig.Settings.Timezone = next.Timezone
		}
		if next.WeekendDays != nil {
			defaultConfig.Settings.WeekendDays = append([]time.Weekday(nil), next.WeekendDays...)
		}
		defaultConfig.Settings.YearsOverflow = next.YearsOverflow
	}

	return Settings{
		FallbackLocale: defaultConfig.Settings.FallbackLocale,
		HumanDiff:      defaultConfig.Settings.HumanDiff,
		Locale:         defaultConfig.Settings.Locale,
		MidDayAt:       defaultConfig.Settings.MidDayAt,
		MonthsOverflow: defaultConfig.Settings.MonthsOverflow,
		StrictMode:     defaultConfig.Settings.StrictMode,
		TestNow:        defaultConfig.Settings.TestNow,
		Timezone:       defaultConfig.Settings.Timezone,
		WeekendDays:    append([]time.Weekday(nil), defaultConfig.Settings.WeekendDays...),
		YearsOverflow:  defaultConfig.Settings.YearsOverflow,
	}
}

func GetLocale() string { return defaultConfig.Settings.Locale }

func SetLocale(locale string) { defaultConfig.Settings.Locale = locale }

func GetFallbackLocale() string { return defaultConfig.Settings.FallbackLocale }

func SetFallbackLocale(locale string) { defaultConfig.Settings.FallbackLocale = locale }

func GetHumanDiffOptions() HumanDiffOptions { return defaultConfig.Settings.HumanDiff }

func SetHumanDiffOptions(options HumanDiffOptions) { defaultConfig.Settings.HumanDiff = options }

func GetMidDayAt() int { return defaultConfig.Settings.MidDayAt }

func SetMidDayAt(hour int) { defaultConfig.Settings.MidDayAt = hour }

func GetWeekendDays() []time.Weekday {
	return append([]time.Weekday(nil), defaultConfig.Settings.WeekendDays...)
}

func SetWeekendDays(days []time.Weekday) {
	defaultConfig.Settings.WeekendDays = append([]time.Weekday(nil), days...)
}

func ShouldOverflowMonths() bool { return defaultConfig.Settings.MonthsOverflow }

func UseMonthsOverflow(enabled bool) { defaultConfig.Settings.MonthsOverflow = enabled }

func ResetMonthsOverflow() { defaultConfig.Settings.MonthsOverflow = true }

func ShouldOverflowYears() bool { return defaultConfig.Settings.YearsOverflow }

func UseYearsOverflow(enabled bool) { defaultConfig.Settings.YearsOverflow = enabled }

func ResetYearsOverflow() { defaultConfig.Settings.YearsOverflow = true }

func IsStrictModeEnabled() bool { return defaultConfig.Settings.StrictMode }

func UseStrictMode(enabled bool) { defaultConfig.Settings.StrictMode = enabled }

func GetTestNow() *Tempo { return defaultConfig.Settings.TestNow }

func SetTestNow(input *Tempo) { defaultConfig.Settings.TestNow = input }

func SetTestNowAndTimezone(input *Tempo, timezone string) {
	defaultConfig.Settings.TestNow = input
	defaultConfig.Settings.Timezone = timezone
}

func HasTestNow() bool { return defaultConfig.Settings.TestNow != nil }
