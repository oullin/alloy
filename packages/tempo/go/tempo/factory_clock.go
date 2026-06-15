package tempo

import "time"

func Now(options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	appConfig := cfg.app

	if appConfig == nil {
		appConfig = defaultConfig
	}

	if appConfig.Settings.TestNow != nil {
		location := cfg.location

		if len(options) == 0 && appConfig.Settings.Timezone != "" {
			configured, err := loadLocation(appConfig.Settings.Timezone)

			if err != nil {
				return Tempo{}, err
			}

			location = configured
		}

		return newTempo(appConfig.Settings.TestNow.value, location, cfg.runtime), nil
	}

	return newTempo(time.Now(), cfg.location, cfg.runtime), nil
}

func Today(options ...Option) (Tempo, error) {
	now, err := Now(options...)

	if err != nil {
		return Tempo{}, err
	}

	return now.StartOfDay(), nil
}

func Tomorrow(options ...Option) (Tempo, error) {
	today, err := Today(options...)

	if err != nil {
		return Tempo{}, err
	}

	return today.AddDays(1), nil
}

func Yesterday(options ...Option) (Tempo, error) {
	today, err := Today(options...)

	if err != nil {
		return Tempo{}, err
	}

	return today.SubDays(1), nil
}

func FromTime(value time.Time, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	return newTempo(value, cfg.location, cfg.runtime), nil
}
