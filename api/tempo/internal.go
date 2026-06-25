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

func newTempo(value time.Time, location *time.Location, runtime Context) Time {
	settings := defaultSettings()

	return newTempoWithPolicy(value, location, runtime, settings, nil, "")
}

func newTempoWithPolicy(value time.Time, location *time.Location, runtime Context, settings Settings, serializer Serializer, toStringFormat string) Time {
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

	return Time{
		value:          value.UTC(),
		location:       location,
		runtime:        runtime,
		settings:       cloneSettings(settings),
		serializer:     serializer,
		toStringFormat: toStringFormat,
	}
}

func (tempo Time) with(value time.Time, location *time.Location) Time {
	return newTempoWithPolicy(value, location, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat)
}

func (tempo Time) settingsSnapshot() Settings {
	return cloneSettings(normalizeSettings(tempo.settings))
}

func (mutable *MutableTime) settingsSnapshot() Settings {
	if mutable == nil {
		return defaultSettings()
	}

	return cloneSettings(normalizeSettings(mutable.settings))
}
