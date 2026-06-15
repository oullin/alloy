package config

import (
	"time"

	"github.com/oullin/alloy/tempo/duration"
)

type HumanDiffOptions struct {
	Absolute bool
	Locale   string
	Numeric  string
	Style    string
	Unit     duration.Unit
}

type Settings struct {
	FallbackLocale string
	HumanDiff      HumanDiffOptions
	Locale         string
	MidDayAt       int
	MonthsOverflow bool
	StrictMode     bool
	Timezone       string
	WeekendDays    []time.Weekday
	YearsOverflow  bool
}

func DefaultLocation() *time.Location {
	return time.UTC
}

func DefaultSettings() Settings {
	return Settings{
		FallbackLocale: "en-US",
		HumanDiff:      HumanDiffOptions{Locale: "en-US", Numeric: "always", Style: "long"},
		Locale:         "en-US",
		MidDayAt:       12,
		MonthsOverflow: true,
		StrictMode:     true,
		Timezone:       DefaultLocation().String(),
		WeekendDays:    []time.Weekday{time.Sunday, time.Saturday},
		YearsOverflow:  true,
	}
}
