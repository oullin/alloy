package tempo

import (
	"fmt"
	"time"
)

const isoLayout = time.RFC3339Nano

// Tempo is an immutable timestamp wrapper with Carbon-compatible fixture behavior.
type Tempo struct {
	value time.Time
}

// Parse creates a Tempo from an RFC3339/RFC3339Nano timestamp.
func Parse(input string) (Tempo, error) {
	parsed, err := time.Parse(isoLayout, input)
	if err != nil {
		return Tempo{}, fmt.Errorf("parse tempo: %w", err)
	}

	return Tempo{value: parsed.UTC()}, nil
}

// FromTimestamp creates a Tempo from a Unix timestamp in seconds.
func FromTimestamp(timestamp int64) Tempo {
	return Tempo{value: time.Unix(timestamp, 0).UTC()}
}

// AddDays returns a new Tempo offset by the given number of days.
func (tempo Tempo) AddDays(days int) Tempo {
	return Tempo{value: tempo.value.AddDate(0, 0, days)}
}

// AddMonths returns a new Tempo offset by the given number of months.
func (tempo Tempo) AddMonths(months int) Tempo {
	return Tempo{value: tempo.value.AddDate(0, months, 0)}
}

// DiffInDays returns the whole-day difference between tempo and other.
func (tempo Tempo) DiffInDays(other Tempo) int {
	return int(tempo.value.Sub(other.value).Hours() / 24)
}

func (tempo Tempo) Before(other Tempo) bool {
	return tempo.value.Before(other.value)
}

func (tempo Tempo) After(other Tempo) bool {
	return tempo.value.After(other.value)
}

func (tempo Tempo) DateString() string {
	return tempo.value.Format(time.DateOnly)
}

func (tempo Tempo) ISOString() string {
	return tempo.value.Format("2006-01-02T15:04:05.000Z")
}

func (tempo Tempo) Time() time.Time {
	return tempo.value
}
