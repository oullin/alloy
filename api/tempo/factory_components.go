package tempo

import "errors"

func Create(components Components) (Time, error) {
	factory, err := NewFactory(componentTimezoneOption(components))

	if err != nil {
		return Time{}, err
	}

	return factory.Create(components)
}

func CreateNormalized(components Components) (Time, error) {
	factory, err := NewFactory(componentTimezoneOption(components))

	if err != nil {
		return Time{}, err
	}

	return factory.CreateNormalized(components)
}

func Instance(input Time) Time {
	return input.Clone()
}

func CreateFromDate(year int, month int, day int, options ...Option) (Time, error) {
	factory, err := NewFactory(options...)

	if err != nil {
		return Time{}, err
	}

	return factory.CreateFromDate(year, month, day)
}

func CreateFromTime(hour int, minute int, second int, millisecond int, options ...Option) (Time, error) {
	factory, err := NewFactory(options...)

	if err != nil {
		return Time{}, err
	}

	return factory.CreateFromTime(hour, minute, second, millisecond)
}

func CreateFromTimeString(input string, options ...Option) (Time, error) {
	today, err := Today(options...)

	if err != nil {
		return Time{}, err
	}

	return today.SetTimeFromTimeString(input)
}

func FromObject(components Components) (Time, error) {
	return Create(components)
}

func componentTimezoneOption(components Components) Option {
	if components.Timezone == "" {
		return func(_ *config) error { return nil }
	}

	return WithTimezone(components.Timezone)
}

var errInvalidComponents = errors.New("invalid Time local date/time components")
