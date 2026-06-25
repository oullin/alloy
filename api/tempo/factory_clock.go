package tempo

import "time"

func Now(options ...Option) (Time, error) {
	factory, err := NewFactory(options...)

	if err != nil {
		return Time{}, err
	}

	return factory.Now(), nil
}

func Today(options ...Option) (Time, error) {
	now, err := Now(options...)

	if err != nil {
		return Time{}, err
	}

	return now.StartOfDay(), nil
}

func Tomorrow(options ...Option) (Time, error) {
	today, err := Today(options...)

	if err != nil {
		return Time{}, err
	}

	return today.AddDays(1), nil
}

func Yesterday(options ...Option) (Time, error) {
	today, err := Today(options...)

	if err != nil {
		return Time{}, err
	}

	return today.SubDays(1), nil
}

func FromTime(value time.Time, options ...Option) (Time, error) {
	factory, err := NewFactory(options...)

	if err != nil {
		return Time{}, err
	}

	return factory.FromTime(value), nil
}
