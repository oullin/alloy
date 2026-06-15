package tempo

import (
	"errors"
	"time"

	defaults "github.com/oullin/alloy/tempo/config"
)

type Config struct {
	Settings       Settings
	LastError      error
	Serializer     Serializer
	ToStringFormat string
}

func NewConfig(settings ...Settings) *Config {
	value := defaultSettings()
	if len(settings) > 0 {
		value = mergeSettings(value, settings[0])
	}

	return &Config{Settings: value}
}

func (config *Config) CloneSettings() Settings {
	return Settings{
		FallbackLocale: config.Settings.FallbackLocale,
		HumanDiff:      config.Settings.HumanDiff,
		Locale:         config.Settings.Locale,
		MidDayAt:       config.Settings.MidDayAt,
		MonthsOverflow: config.Settings.MonthsOverflow,
		StrictMode:     config.Settings.StrictMode,
		TestNow:        config.Settings.TestNow,
		Timezone:       config.Settings.Timezone,
		WeekendDays:    append([]time.Weekday(nil), config.Settings.WeekendDays...),
		YearsOverflow:  config.Settings.YearsOverflow,
	}
}

func (config *Config) Apply(settings Settings) {
	config.Settings = mergeSettings(config.Settings, settings)
}

func (config *Config) SetLastError(err error) {
	config.LastError = err
}

func (config *Config) GetLastError() error {
	if config == nil {
		return errors.New("tempo config is nil")
	}

	return config.LastError
}

func defaultSettings() Settings {
	settings := defaults.DefaultSettings()
	return Settings{
		FallbackLocale: settings.FallbackLocale,
		HumanDiff: HumanDiffOptions{
			Absolute: settings.HumanDiff.Absolute,
			Locale:   settings.HumanDiff.Locale,
			Numeric:  settings.HumanDiff.Numeric,
			Style:    settings.HumanDiff.Style,
			Unit:     settings.HumanDiff.Unit,
		},
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

func mergeSettings(current Settings, next Settings) Settings {
	if next.Locale != "" {
		current.Locale = next.Locale
	}
	if next.FallbackLocale != "" {
		current.FallbackLocale = next.FallbackLocale
	}
	if next.HumanDiff != (HumanDiffOptions{}) {
		current.HumanDiff = next.HumanDiff
	}
	if next.MidDayAt != 0 {
		current.MidDayAt = next.MidDayAt
	}
	current.MonthsOverflow = next.MonthsOverflow
	current.StrictMode = next.StrictMode
	current.TestNow = next.TestNow
	if next.Timezone != "" {
		current.Timezone = next.Timezone
	}
	if next.WeekendDays != nil {
		current.WeekendDays = append([]time.Weekday(nil), next.WeekendDays...)
	}
	current.YearsOverflow = next.YearsOverflow

	return current
}

var defaultConfig = NewConfig()
