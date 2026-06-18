package tempo

import "errors"

func Create(components Components) (Tempo, error) {
	factory, err := NewFactory(componentTimezoneOption(components))

	if err != nil {
		return Tempo{}, err
	}

	return factory.Create(components)
}

func CreateNormalized(components Components) (Tempo, error) {
	factory, err := NewFactory(componentTimezoneOption(components))

	if err != nil {
		return Tempo{}, err
	}

	return factory.CreateNormalized(components)
}

func Instance(input Tempo) Tempo {
	return input.Clone()
}

func CreateFromDate(year int, month int, day int, options ...Option) (Tempo, error) {
	factory, err := NewFactory(options...)

	if err != nil {
		return Tempo{}, err
	}

	return factory.CreateFromDate(year, month, day)
}

func CreateFromTime(hour int, minute int, second int, millisecond int, options ...Option) (Tempo, error) {
	factory, err := NewFactory(options...)

	if err != nil {
		return Tempo{}, err
	}

	return factory.CreateFromTime(hour, minute, second, millisecond)
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

func componentTimezoneOption(components Components) Option {
	if components.Timezone == "" {
		return func(_ *config) error { return nil }
	}

	return WithTimezone(components.Timezone)
}

var errInvalidComponents = errors.New("invalid Tempo local date/time components")
