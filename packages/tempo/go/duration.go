package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/duration"
)

func ParseDuration(input string) (Duration, error) {
	return duration.Parse(input)
}

func Min(first Tempo, rest ...Tempo) Tempo {
	values := make([]int64, 0, len(rest)+1)
	items := append([]Tempo{first}, rest...)

	for _, item := range items {
		values = append(values, item.TimestampMs())
	}

	return items[EarlierIndex(values)]
}

func Max(first Tempo, rest ...Tempo) Tempo {
	values := make([]int64, 0, len(rest)+1)
	items := append([]Tempo{first}, rest...)

	for _, item := range items {
		values = append(values, item.TimestampMs())
	}

	return items[LaterIndex(values)]
}

func Average(start Tempo, end Tempo) Tempo {
	return Tempo{
		value:    time.UnixMilli(AverageMilliseconds(start.TimestampMs(), end.TimestampMs())).UTC(),
		location: start.location,
	}
}

func NewMutable(input Tempo) *MutableTempo {
	return &MutableTempo{
		value:          input.value,
		location:       input.location,
		runtime:        input.Runtime(),
		settings:       input.settingsSnapshot(),
		serializer:     input.serializer,
		toStringFormat: input.toStringFormat,
	}
}

func ParseMutable(input string, options ...Option) (*MutableTempo, error) {
	parsed, err := Parse(input, options...)

	if err != nil {
		return nil, err
	}

	return NewMutable(parsed), nil
}
