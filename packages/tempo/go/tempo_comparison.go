package tempo

import (
	"errors"
	"strings"
)

func (tempo Tempo) IsImmutable() bool {
	return true
}

func (tempo Tempo) IsMutable() bool {
	return false
}

func (tempo Tempo) Before(other Tempo, units ...Unit) bool {
	return tempo.compareValue(units...) < other.compareValue(units...)
}

func (tempo Tempo) After(other Tempo, units ...Unit) bool {
	return tempo.compareValue(units...) > other.compareValue(units...)
}

func (tempo Tempo) Same(other Tempo, units ...Unit) bool {
	return tempo.compareValue(units...) == other.compareValue(units...)
}

func (tempo Tempo) Is(other Tempo, units ...Unit) bool {
	return tempo.Same(other, units...)
}

func (tempo Tempo) EqualTo(other Tempo, units ...Unit) bool {
	return tempo.Same(other, units...)
}

func (tempo Tempo) Eq(other Tempo, units ...Unit) bool {
	return tempo.EqualTo(other, units...)
}

func (tempo Tempo) NotEqualTo(other Tempo, units ...Unit) bool {
	return !tempo.Same(other, units...)
}

func (tempo Tempo) Ne(other Tempo, units ...Unit) bool {
	return tempo.NotEqualTo(other, units...)
}

func (tempo Tempo) GreaterThan(other Tempo, units ...Unit) bool {
	return tempo.After(other, units...)
}

func (tempo Tempo) Gt(other Tempo, units ...Unit) bool {
	return tempo.GreaterThan(other, units...)
}

func (tempo Tempo) GreaterThanOrEqualTo(other Tempo, units ...Unit) bool {
	return tempo.SameOrAfter(other, units...)
}

func (tempo Tempo) Gte(other Tempo, units ...Unit) bool {
	return tempo.GreaterThanOrEqualTo(other, units...)
}

func (tempo Tempo) LessThan(other Tempo, units ...Unit) bool {
	return tempo.Before(other, units...)
}

func (tempo Tempo) Lt(other Tempo, units ...Unit) bool {
	return tempo.LessThan(other, units...)
}

func (tempo Tempo) LessThanOrEqualTo(other Tempo, units ...Unit) bool {
	return tempo.SameOrBefore(other, units...)
}

func (tempo Tempo) Lte(other Tempo, units ...Unit) bool {
	return tempo.LessThanOrEqualTo(other, units...)
}

func (tempo Tempo) SameSecond(other Tempo) bool {
	return tempo.Same(other, Second)
}

func (tempo Tempo) SameMinute(other Tempo) bool {
	return tempo.Same(other, Minute)
}

func (tempo Tempo) SameHour(other Tempo) bool {
	return tempo.Same(other, Hour)
}

func (tempo Tempo) SameDay(other Tempo) bool {
	return tempo.Same(other, Day)
}

func (tempo Tempo) SameWeek(other Tempo) bool {
	return tempo.Same(other, Week)
}

func (tempo Tempo) SameMonth(other Tempo) bool {
	return tempo.Same(other, Month)
}

func (tempo Tempo) SameQuarter(other Tempo) bool {
	return tempo.Same(other, Quarter)
}

func (tempo Tempo) SameYear(other Tempo) bool {
	return tempo.Same(other, Year)
}

func (tempo Tempo) SameAs(pattern string, other Tempo) bool {
	return tempo.Format(pattern) == other.Format(pattern)
}

func (tempo Tempo) IsSameUnit(unit Unit, other Tempo) bool {
	return tempo.Same(other, unit)
}

func (tempo Tempo) Birthday(other Tempo) bool {
	return tempo.Month() == other.Month() && tempo.Day() == other.Day()
}

func (tempo Tempo) Clamp(minimum Tempo, maximum Tempo) (Tempo, error) {
	if minimum.After(maximum) {
		return Tempo{}, errors.New("tempo clamp minimum must be before maximum")
	}

	if tempo.Before(minimum) {
		return minimum, nil
	}

	if tempo.After(maximum) {
		return maximum, nil
	}

	return tempo.Clone(), nil
}

func (tempo Tempo) Average(other Tempo) Tempo {
	return Average(tempo, other)
}

func (tempo Tempo) Closest(first Tempo, rest ...Tempo) Tempo {
	result := first
	bestDistance := absInt64(first.TimestampMs() - tempo.TimestampMs())

	for _, item := range rest {
		distance := absInt64(item.TimestampMs() - tempo.TimestampMs())

		if distance < bestDistance {
			result = item
			bestDistance = distance
		}
	}

	return result
}

func (tempo Tempo) Farthest(first Tempo, rest ...Tempo) Tempo {
	result := first
	bestDistance := absInt64(first.TimestampMs() - tempo.TimestampMs())

	for _, item := range rest {
		distance := absInt64(item.TimestampMs() - tempo.TimestampMs())

		if distance > bestDistance {
			result = item
			bestDistance = distance
		}
	}

	return result
}

func (tempo Tempo) Min(other Tempo) Tempo {
	return Min(tempo, other)
}

func (tempo Tempo) Max(other Tempo) Tempo {
	return Max(tempo, other)
}

func (tempo Tempo) Minimum(other Tempo) Tempo {
	return tempo.Min(other)
}

func (tempo Tempo) Maximum(other Tempo) Tempo {
	return tempo.Max(other)
}

func (tempo Tempo) SameOrBefore(other Tempo, units ...Unit) bool {
	return tempo.Same(other, units...) || tempo.Before(other, units...)
}

func (tempo Tempo) SameOrAfter(other Tempo, units ...Unit) bool {
	return tempo.Same(other, units...) || tempo.After(other, units...)
}

func (tempo Tempo) Between(start Tempo, end Tempo, inclusivity ...string) bool {
	if start.After(end) {
		start, end = end, start
	}

	mode := "[]"

	if len(inclusivity) > 0 {
		mode = inclusivity[0]
	}

	afterStart := tempo.After(start)

	if strings.HasPrefix(mode, "[") {
		afterStart = tempo.SameOrAfter(start)
	}

	beforeEnd := tempo.Before(end)

	if strings.HasSuffix(mode, "]") {
		beforeEnd = tempo.SameOrBefore(end)
	}

	return afterStart && beforeEnd
}

func (tempo Tempo) IsBetween(start Tempo, end Tempo, inclusivity ...string) bool {
	return tempo.Between(start, end, inclusivity...)
}

func (tempo Tempo) BetweenIncluded(start Tempo, end Tempo) bool {
	return tempo.Between(start, end, "[]")
}

func (tempo Tempo) BetweenExcluded(start Tempo, end Tempo) bool {
	return tempo.Between(start, end, "()")
}
