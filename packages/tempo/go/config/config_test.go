package config

import (
	"testing"
	"time"
)

func TestDefaultSettingsReadEnvironment(t *testing.T) {
	t.Setenv("TEMPO_LOCALE", "en-GB")
	t.Setenv("TEMPO_FALLBACK_LOCALE", "fr-FR")
	t.Setenv("TEMPO_MID_DAY_AT", "13")
	t.Setenv("TEMPO_MONTHS_OVERFLOW", "false")
	t.Setenv("TEMPO_HUMAN_DIFF_UNIT", "day")
	t.Setenv("TEMPO_WEEKEND_DAYS", "5, 6")

	settings := DefaultSettings()

	if settings.Locale != "en-GB" {
		t.Fatalf("expected locale from env, got %q", settings.Locale)
	}

	if settings.FallbackLocale != "fr-FR" {
		t.Fatalf("expected fallback locale from env, got %q", settings.FallbackLocale)
	}

	if settings.MidDayAt != 13 {
		t.Fatalf("expected mid-day value from env, got %d", settings.MidDayAt)
	}

	if settings.MonthsOverflow {
		t.Fatal("expected months overflow to be disabled from env")
	}

	if settings.HumanDiff.Unit != "day" {
		t.Fatalf("expected human diff unit from env, got %q", settings.HumanDiff.Unit)
	}

	expectedWeekend := []time.Weekday{time.Friday, time.Saturday}

	if len(settings.WeekendDays) != len(expectedWeekend) {
		t.Fatalf("expected %d weekend days, got %d", len(expectedWeekend), len(settings.WeekendDays))
	}

	for index, expected := range expectedWeekend {
		if settings.WeekendDays[index] != expected {
			t.Fatalf("expected weekend day %d to be %s, got %s", index, expected, settings.WeekendDays[index])
		}
	}
}
