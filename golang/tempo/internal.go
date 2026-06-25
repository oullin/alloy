package tempo

import "time"

func applyOptions(options ...Option) (config, error) {
	settings := defaultSettings()
	cfg := config{
		location: defaultLocation(),
		settings: settings,
		runtime: NewRuntime(
			RuntimeLocale(settings.Locale),
			RuntimeFallbackLocale(settings.FallbackLocale),
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
	settings := defaultSettings()

	return newTempoWithPolicy(value, location, runtime, settings, nil, "")
}

func newTempoWithPolicy(value time.Time, location *time.Location, runtime Runtime, settings Settings, serializer Serializer, toStringFormat string) Tempo {
	if location == nil {
		location = defaultLocation()
	}

	settings = normalizeSettings(settings)

	if runtime.Locale() == "" {
		runtime = NewRuntime(
			RuntimeLocale(settings.Locale),
			RuntimeFallbackLocale(settings.FallbackLocale),
		)
	}

	return Tempo{
		value:          value.UTC(),
		location:       location,
		runtime:        runtime,
		settings:       cloneSettings(settings),
		serializer:     serializer,
		toStringFormat: toStringFormat,
	}
}

func (tempo Tempo) with(value time.Time, location *time.Location) Tempo {
	return newTempoWithPolicy(value, location, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat)
}

func (tempo Tempo) settingsSnapshot() Settings {
	return cloneSettings(normalizeSettings(tempo.settings))
}

func (mutable *MutableTempo) settingsSnapshot() Settings {
	if mutable == nil {
		return defaultSettings()
	}

	return cloneSettings(normalizeSettings(mutable.settings))
}
