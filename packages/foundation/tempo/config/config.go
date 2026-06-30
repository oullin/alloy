package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"alloy.dev/foundation/tempo/duration"
	"github.com/spf13/viper"
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
	return SettingsFromViper(NewViper())
}

func NewViper() *viper.Viper {
	reader := viper.New()
	reader.SetEnvPrefix("TEMPO")
	reader.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	reader.AutomaticEnv()

	reader.SetDefault("fallback_locale", "en-US")
	reader.SetDefault("human_diff.absolute", false)
	reader.SetDefault("human_diff.locale", "en-US")
	reader.SetDefault("human_diff.numeric", "always")
	reader.SetDefault("human_diff.style", "long")
	reader.SetDefault("human_diff.unit", "")
	reader.SetDefault("locale", "en-US")
	reader.SetDefault("mid_day_at", 12)
	reader.SetDefault("months_overflow", true)
	reader.SetDefault("strict_mode", true)
	reader.SetDefault("timezone", DefaultLocation().String())
	reader.SetDefault("weekend_days", []int{int(time.Sunday), int(time.Saturday)})
	reader.SetDefault("years_overflow", true)

	if configPath := strings.TrimSpace(os.Getenv("TEMPO_CONFIG")); configPath != "" {
		reader.SetConfigFile(configPath)
		_ = reader.ReadInConfig()
	}

	return reader
}

func SettingsFromViper(reader *viper.Viper) Settings {
	if reader == nil {
		reader = NewViper()
	}

	return Settings{
		FallbackLocale: reader.GetString("fallback_locale"),
		HumanDiff: HumanDiffOptions{
			Absolute: reader.GetBool("human_diff.absolute"),
			Locale:   reader.GetString("human_diff.locale"),
			Numeric:  reader.GetString("human_diff.numeric"),
			Style:    reader.GetString("human_diff.style"),
			Unit:     duration.Unit(reader.GetString("human_diff.unit")),
		},
		Locale:         reader.GetString("locale"),
		MidDayAt:       reader.GetInt("mid_day_at"),
		MonthsOverflow: reader.GetBool("months_overflow"),
		StrictMode:     reader.GetBool("strict_mode"),
		Timezone:       reader.GetString("timezone"),
		WeekendDays:    weekdaysFromInts(intSliceFromViper(reader, "weekend_days")),
		YearsOverflow:  reader.GetBool("years_overflow"),
	}
}

func intSliceFromViper(reader *viper.Viper, key string) []int {
	raw := reader.Get(key)

	switch values := raw.(type) {
	case []int:
		return values
	case []string:
		return intSliceFromStrings(values)
	case []any:
		result := make([]int, 0, len(values))

		for _, value := range values {
			switch typed := value.(type) {
			case int:
				result = append(result, typed)
			case string:
				result = append(result, intSliceFromStrings([]string{typed})...)
			default:
				if parsed, err := strconv.Atoi(fmt.Sprint(typed)); err == nil {
					result = append(result, parsed)
				}
			}
		}

		return result
	case string:
		return intSliceFromStrings([]string{values})
	default:
		return reader.GetIntSlice(key)
	}
}

func intSliceFromStrings(values []string) []int {
	result := make([]int, 0, len(values))

	for _, value := range values {
		for field := range strings.SplitSeq(value, ",") {
			trimmed := strings.TrimSpace(field)

			if trimmed == "" {
				continue
			}

			parsed, err := strconv.Atoi(trimmed)

			if err != nil {
				continue
			}

			result = append(result, parsed)
		}
	}

	return result
}

func weekdaysFromInts(values []int) []time.Weekday {
	if len(values) == 0 {
		return []time.Weekday{time.Sunday, time.Saturday}
	}

	days := make([]time.Weekday, 0, len(values))

	for _, value := range values {
		days = append(days, time.Weekday(value))
	}

	return days
}
