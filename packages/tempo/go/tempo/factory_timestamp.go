package tempo

import factorypkg "github.com/oullin/alloy/tempo/factory"

func FromTimestamp(timestamp int64, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	return newTempo(factorypkg.FromTimestamp(timestamp), cfg.location, cfg.runtime), nil
}

func CreateFromTimestamp(timestamp int64, options ...Option) (Tempo, error) {
	return FromTimestamp(timestamp, options...)
}

func FromTimestampMs(timestamp int64, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	return newTempo(factorypkg.FromTimestampMs(timestamp), cfg.location, cfg.runtime), nil
}

func CreateFromTimestampMs(timestamp int64, options ...Option) (Tempo, error) {
	return FromTimestampMs(timestamp, options...)
}

func FromTimestampUTC(timestamp int64) (Tempo, error) {
	return FromTimestamp(timestamp, WithTimezone(defaultLocation().String()))
}

func FromTimestampMsUTC(timestamp int64) (Tempo, error) {
	return FromTimestampMs(timestamp, WithTimezone(defaultLocation().String()))
}

func CreateFromTimestampUTC(timestamp int64) (Tempo, error) {
	return FromTimestampUTC(timestamp)
}

func CreateFromTimestampMsUTC(timestamp int64) (Tempo, error) {
	return FromTimestampMsUTC(timestamp)
}
