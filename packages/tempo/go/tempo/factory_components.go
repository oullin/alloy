package tempo

import "errors"

func Create(components Components) (Tempo, error) {
	location, err := loadLocation(components.Timezone)

	if err != nil {
		return Tempo{}, err
	}

	return newTempo(
		timeFromComponents(components, location),
		location,
		NewRuntime(
			RuntimeLocale(defaultConfig.Settings.Locale),
			RuntimeFallbackLocale(defaultConfig.Settings.FallbackLocale),
		),
	), nil
}

func CreateSafe(components Components) (Tempo, error) {
	location, err := loadLocation(components.Timezone)

	if err != nil {
		return Tempo{}, err
	}

	value := timeFromComponents(components, location)

	if !componentsMatchTime(components, value, location) {
		return Tempo{}, errors.New("invalid Tempo local date/time components")
	}

	return newTempo(
		value,
		location,
		NewRuntime(
			RuntimeLocale(defaultConfig.Settings.Locale),
			RuntimeFallbackLocale(defaultConfig.Settings.FallbackLocale),
		),
	), nil
}

func CreateStrict(components Components) (Tempo, error) {
	return CreateSafe(components)
}

func Instance(input Tempo) Tempo {
	return input.Clone()
}

func CreateFromDate(year int, month int, day int, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	return newTempo(
		timeFromComponents(Components{Year: year, Month: month, Day: day}, cfg.location),
		cfg.location,
		cfg.runtime,
	), nil
}

func CreateMidnightDate(year int, month int, day int, options ...Option) (Tempo, error) {
	return CreateFromDate(year, month, day, options...)
}

func CreateFromTime(hour int, minute int, second int, millisecond int, options ...Option) (Tempo, error) {
	today, err := Today(options...)

	if err != nil {
		return Tempo{}, err
	}

	return today.SetTime(hour, minute, second, millisecond), nil
}

func CreateFromTimeString(input string, options ...Option) (Tempo, error) {
	today, err := Today(options...)

	if err != nil {
		return Tempo{}, err
	}

	return today.SetTimeFromTimeString(input)
}

func FromObject(components Components) (Tempo, error) {
	return Create(components)
}
