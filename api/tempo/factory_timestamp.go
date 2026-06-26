package tempo

import factorypkg "github.com/oullin/alloy/api/tempo/factory"

func FromTimestamp(timestamp int64, options ...Option) (Time, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Time{}, err
	}

	return newTempo(factorypkg.FromTimestamp(timestamp), cfg.location, cfg.runtime), nil
}

func FromTimestampMs(timestamp int64, options ...Option) (Time, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Time{}, err
	}

	return newTempo(factorypkg.FromTimestampMs(timestamp), cfg.location, cfg.runtime), nil
}

func FromTimestampUTC(timestamp int64) (Time, error) {
	return FromTimestamp(timestamp, WithTimezone(defaultLocation().String()))
}

func FromTimestampMsUTC(timestamp int64) (Time, error) {
	return FromTimestampMs(timestamp, WithTimezone(defaultLocation().String()))
}
