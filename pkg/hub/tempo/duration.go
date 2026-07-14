package tempo

import (
	"time"

	"github.com/oullin/alloy/pkg/hub/tempo/duration"
)

func ParseDuration(input string) (Duration, error) {
	return duration.Parse(input)
}

func Min(first Time, rest ...Time) Time {
	values := make([]int64, 0, len(rest)+1)
	items := append([]Time{first}, rest...)

	for _, item := range items {
		values = append(values, item.TimestampMs())
	}

	return items[EarlierIndex(values)]
}

func Max(first Time, rest ...Time) Time {
	values := make([]int64, 0, len(rest)+1)
	items := append([]Time{first}, rest...)

	for _, item := range items {
		values = append(values, item.TimestampMs())
	}

	return items[LaterIndex(values)]
}

func Average(start Time, end Time) Time {
	return Time{
		value:    time.UnixMilli(AverageMilliseconds(start.TimestampMs(), end.TimestampMs())).UTC(),
		location: start.location,
	}
}

func NewMutable(input Time) *MutableTime {
	return &MutableTime{
		value:          input.value,
		location:       input.location,
		runtime:        input.Context(),
		settings:       input.settingsSnapshot(),
		serializer:     input.serializer,
		toStringFormat: input.toStringFormat,
	}
}

func ParseMutable(input string, options ...Option) (*MutableTime, error) {
	parsed, err := Parse(input, options...)

	if err != nil {
		return nil, err
	}

	return NewMutable(parsed), nil
}
