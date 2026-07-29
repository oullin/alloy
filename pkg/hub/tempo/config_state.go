package tempo

import (
	"time"

	defaults "hara.sh/alloy/tempo/config"
)

type Config struct {
	Settings       Settings
	Serializer     Serializer
	ToStringFormat string
}

func NewConfig(options ...ConfigOption) (*Config, error) {
	cfg := &Config{Settings: defaultSettings()}

	for _, option := range options {
		if err := option(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (config *Config) CloneSettings() Settings {
	if config == nil {
		return defaultSettings()
	}

	return cloneSettings(config.Settings)
}

func (config *Config) Apply(options ...ConfigOption) error {
	if config == nil {
		return nil
	}

	for _, option := range options {
		if err := option(config); err != nil {
			return err
		}
	}

	return nil
}

func cloneSettings(settings Settings) Settings {
	return Settings{
		FallbackLocale: settings.FallbackLocale,
		HumanDiff:      settings.HumanDiff,
		Locale:         settings.Locale,
		MidDayAt:       settings.MidDayAt,
		MonthsOverflow: settings.MonthsOverflow,
		StrictMode:     settings.StrictMode,
		TestNow:        settings.TestNow,
		Timezone:       settings.Timezone,
		WeekendDays:    append([]time.Weekday(nil), settings.WeekendDays...),
		YearsOverflow:  settings.YearsOverflow,
	}
}

func defaultSettings() Settings {
	settings := defaults.DefaultSettings()

	return Settings{
		FallbackLocale: settings.FallbackLocale,
		HumanDiff:      settings.HumanDiff,
		Locale:         settings.Locale,
		MidDayAt:       settings.MidDayAt,
		MonthsOverflow: settings.MonthsOverflow,
		StrictMode:     settings.StrictMode,
		Timezone:       settings.Timezone,
		WeekendDays:    append([]time.Weekday(nil), settings.WeekendDays...),
		YearsOverflow:  settings.YearsOverflow,
	}
}

func defaultLocation() *time.Location {
	return defaults.DefaultLocation()
}

func normalizeSettings(settings Settings) Settings {
	if settings.Locale == "" && settings.FallbackLocale == "" && settings.WeekendDays == nil && settings.MidDayAt == 0 {
		return defaultSettings()
	}

	defaults := defaultSettings()

	if settings.Locale == "" {
		settings.Locale = defaults.Locale
	}

	if settings.FallbackLocale == "" {
		settings.FallbackLocale = defaults.FallbackLocale
	}

	if settings.HumanDiff == (HumanDiffOptions{}) {
		settings.HumanDiff = defaults.HumanDiff
	}

	if settings.MidDayAt == 0 {
		settings.MidDayAt = defaults.MidDayAt
	}

	if settings.Timezone == "" {
		settings.Timezone = defaults.Timezone
	}

	if settings.WeekendDays == nil {
		settings.WeekendDays = append([]time.Weekday(nil), defaults.WeekendDays...)
	}

	return settings
}
