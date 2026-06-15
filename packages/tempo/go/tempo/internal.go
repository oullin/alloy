package tempo

import "time"

func applyOptions(options ...Option) (config, error) {
	cfg := config{
		location: defaultLocation(),
		app:      defaultConfig,
		runtime: NewRuntime(
			RuntimeLocale(defaultConfig.Settings.Locale),
			RuntimeFallbackLocale(defaultConfig.Settings.FallbackLocale),
		),
	}

	for _, option := range options {
		if err := option(&cfg); err != nil {
			return config{}, err
		}
	}

	return cfg, nil
}

func newTempo(value time.Time, location *time.Location, runtime Runtime) Tempo {
	if location == nil {
		location = defaultLocation()
	}

	if runtime.Locale() == "" {
		runtime = NewRuntime(
			RuntimeLocale(defaultConfig.Settings.Locale),
			RuntimeFallbackLocale(defaultConfig.Settings.FallbackLocale),
		)
	}

	return Tempo{value: value.UTC(), location: location, runtime: runtime}
}

func (tempo Tempo) with(value time.Time, location *time.Location) Tempo {
	return newTempo(value, location, tempo.Runtime())
}
