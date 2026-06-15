package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/duration"
)

func ParseDuration(input string) (Duration, error) {
	return duration.Parse(input)
}

func Min(first Tempo, rest ...Tempo) Tempo {
	result := first

	for _, item := range rest {
		if item.Before(result) {
			result = item
		}
	}

	return result
}

func Max(first Tempo, rest ...Tempo) Tempo {
	result := first

	for _, item := range rest {
		if item.After(result) {
			result = item
		}
	}

	return result
}

func Minimum(first Tempo, rest ...Tempo) Tempo {
	return Min(first, rest...)
}

func Maximum(first Tempo, rest ...Tempo) Tempo {
	return Max(first, rest...)
}

func Average(start Tempo, end Tempo) Tempo {
	return Tempo{
		value:    time.UnixMilli((start.TimestampMs() + end.TimestampMs()) / 2).UTC(),
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
