package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/duration"
	domain "github.com/oullin/alloy/tempo/tempo"
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

	return items[domain.EarlierIndex(values)]
}

func Max(first Tempo, rest ...Tempo) Tempo {
	values := make([]int64, 0, len(rest)+1)
	items := append([]Tempo{first}, rest...)

	for _, item := range items {
		values = append(values, item.TimestampMs())
	}

	return items[domain.LaterIndex(values)]
}

func Minimum(first Tempo, rest ...Tempo) Tempo {
	return Min(first, rest...)
}

func Maximum(first Tempo, rest ...Tempo) Tempo {
	return Max(first, rest...)
}

func Average(start Tempo, end Tempo) Tempo {
	return Tempo{
		value:    time.UnixMilli(domain.AverageMilliseconds(start.TimestampMs(), end.TimestampMs())).UTC(),
		location: start.location,
	}
}

func NewMutable(input Tempo) *MutableTempo {
	return &MutableTempo{value: input.value, location: input.location, runtime: input.Runtime()}
}

func ParseMutable(input string, options ...Option) (*MutableTempo, error) {
	parsed, err := Parse(input, options...)

	if err != nil {
		return nil, err
	}

	return NewMutable(parsed), nil
}
