package tempo

import "time"

func Now(options ...Option) (Tempo, error) {
	factory, err := NewFactory(options...)

	if err != nil {
		return Tempo{}, err
	}

	return factory.Now(), nil
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
	factory, err := NewFactory(options...)

	if err != nil {
		return Tempo{}, err
	}

	return factory.FromTime(value), nil
}
