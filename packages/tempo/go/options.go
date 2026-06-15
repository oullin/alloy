package tempo

import "time"

func WithTimezone(name string) Option {
	return func(cfg *config) error {
		location, err := loadLocation(name)

		if err != nil {
			return err
		}

		cfg.location = location
		cfg.settings.Timezone = location.String()

		return nil
	}
}

func WithRuntime(runtime Runtime) Option {
	return func(cfg *config) error {
		cfg.runtime = runtime

		return nil
	}
}

func WithConfig(appConfig *Config) Option {
	return func(cfg *config) error {
		if appConfig == nil {
			return nil
		}

		cfg.settings = appConfig.CloneSettings()
		cfg.serializer = appConfig.Serializer
		cfg.toStringFormat = appConfig.ToStringFormat
		cfg.runtime = NewRuntime(
			RuntimeLocale(appConfig.Settings.Locale),
			RuntimeFallbackLocale(appConfig.Settings.FallbackLocale),
		)

		if appConfig.Settings.Timezone != "" {
			location, err := loadLocation(appConfig.Settings.Timezone)

			if err != nil {
				return err
			}

			cfg.location = location
		}

		return nil
	}
}

func WithTranslator(translator Translator) Option {
	return func(cfg *config) error {
		cfg.runtime = cfg.runtime.With(RuntimeTranslator(translator))

		return nil
	}
}

func WithLocale(locale string) Option {
	return func(cfg *config) error {
		if locale == "" {
			return nil
		}

		cfg.settings.Locale = locale
		cfg.runtime = cfg.runtime.With(RuntimeLocale(locale))

		return nil
	}
}

func WithFallbackLocale(locale string) Option {
	return func(cfg *config) error {
		if locale == "" {
			return nil
		}

		cfg.settings.FallbackLocale = locale
		cfg.runtime = cfg.runtime.With(RuntimeFallbackLocale(locale))

		return nil
	}
}

func WithHumanDiffOptions(options HumanDiffOptions) Option {
	return func(cfg *config) error {
		cfg.settings.HumanDiff = options

		return nil
	}
}

func WithMidDayAt(hour int) Option {
	return func(cfg *config) error {
		cfg.settings.MidDayAt = hour

		return nil
	}
}

func WithMonthsOverflow(enabled bool) Option {
	return func(cfg *config) error {
		cfg.settings.MonthsOverflow = enabled

		return nil
	}
}

func WithStrictMode(enabled bool) Option {
	return func(cfg *config) error {
		cfg.settings.StrictMode = enabled

		return nil
	}
}

func WithTestNow(input Tempo) Option {
	return func(cfg *config) error {
		value := input
		cfg.settings.TestNow = &value

		return nil
	}
}

func WithWeekendDays(days []time.Weekday) Option {
	return func(cfg *config) error {
		cfg.settings.WeekendDays = append([]time.Weekday(nil), days...)

		return nil
	}
}

func WithYearsOverflow(enabled bool) Option {
	return func(cfg *config) error {
		cfg.settings.YearsOverflow = enabled

		return nil
	}
}

func WithSerializer(serializer Serializer) Option {
	return func(cfg *config) error {
		cfg.serializer = serializer

		return nil
	}
}

func WithToStringFormat(pattern string) Option {
	return func(cfg *config) error {
		cfg.toStringFormat = pattern

		return nil
	}
}

func ConfigLocale(locale string) ConfigOption {
	return func(cfg *Config) error {
		if locale != "" {
			cfg.Settings.Locale = locale
		}

		return nil
	}
}

func ConfigFallbackLocale(locale string) ConfigOption {
	return func(cfg *Config) error {
		if locale != "" {
			cfg.Settings.FallbackLocale = locale
		}

		return nil
	}
}

func ConfigHumanDiffOptions(options HumanDiffOptions) ConfigOption {
	return func(cfg *Config) error {
		cfg.Settings.HumanDiff = options

		return nil
	}
}

func ConfigMidDayAt(hour int) ConfigOption {
	return func(cfg *Config) error {
		cfg.Settings.MidDayAt = hour

		return nil
	}
}

func ConfigMonthsOverflow(enabled bool) ConfigOption {
	return func(cfg *Config) error {
		cfg.Settings.MonthsOverflow = enabled

		return nil
	}
}

func ConfigStrictMode(enabled bool) ConfigOption {
	return func(cfg *Config) error {
		cfg.Settings.StrictMode = enabled

		return nil
	}
}

func ConfigTestNow(input Tempo) ConfigOption {
	return func(cfg *Config) error {
		value := input
		cfg.Settings.TestNow = &value

		return nil
	}
}

func ConfigTimezone(name string) ConfigOption {
	return func(cfg *Config) error {
		location, err := loadLocation(name)

		if err != nil {
			return err
		}

		cfg.Settings.Timezone = location.String()

		return nil
	}
}

func ConfigWeekendDays(days []time.Weekday) ConfigOption {
	return func(cfg *Config) error {
		cfg.Settings.WeekendDays = append([]time.Weekday(nil), days...)

		return nil
	}
}

func ConfigYearsOverflow(enabled bool) ConfigOption {
	return func(cfg *Config) error {
		cfg.Settings.YearsOverflow = enabled

		return nil
	}
}

func ConfigSerializer(serializer Serializer) ConfigOption {
	return func(cfg *Config) error {
		cfg.Serializer = serializer

		return nil
	}
}

func ConfigToStringFormat(pattern string) ConfigOption {
	return func(cfg *Config) error {
		cfg.ToStringFormat = pattern

		return nil
	}
}
