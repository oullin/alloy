package tempo

import (
	"github.com/oullin/alloy/packages/foundation/tempo/comparison"
	"github.com/oullin/alloy/packages/foundation/tempo/core"
)

func states(values []Time) []core.State {
	result := make([]core.State, 0, len(values))

	for _, value := range values {
		result = append(result, value.State())
	}

	return result
}

func (tempo Time) IsImmutable() bool {
	return true
}

func (tempo Time) IsMutable() bool {
	return false
}

func (tempo Time) Before(other Time, units ...Unit) bool {
	return comparison.Before(tempo, other.State(), units...)
}

func (tempo Time) After(other Time, units ...Unit) bool {
	return comparison.After(tempo, other.State(), units...)
}

func (tempo Time) Same(other Time, units ...Unit) bool {
	return comparison.Same(tempo, other.State(), units...)
}

func (tempo Time) Is(other Time, units ...Unit) bool {
	return tempo.Same(other, units...)
}

func (tempo Time) EqualTo(other Time, units ...Unit) bool {
	return tempo.Same(other, units...)
}

func (tempo Time) Eq(other Time, units ...Unit) bool {
	return tempo.EqualTo(other, units...)
}

func (tempo Time) NotEqualTo(other Time, units ...Unit) bool {
	return !tempo.Same(other, units...)
}

func (tempo Time) Ne(other Time, units ...Unit) bool {
	return tempo.NotEqualTo(other, units...)
}

func (tempo Time) GreaterThan(other Time, units ...Unit) bool {
	return tempo.After(other, units...)
}

func (tempo Time) Gt(other Time, units ...Unit) bool {
	return tempo.GreaterThan(other, units...)
}

func (tempo Time) GreaterThanOrEqualTo(other Time, units ...Unit) bool {
	return tempo.SameOrAfter(other, units...)
}

func (tempo Time) Gte(other Time, units ...Unit) bool {
	return tempo.GreaterThanOrEqualTo(other, units...)
}

func (tempo Time) LessThan(other Time, units ...Unit) bool {
	return tempo.Before(other, units...)
}

func (tempo Time) Lt(other Time, units ...Unit) bool {
	return tempo.LessThan(other, units...)
}

func (tempo Time) LessThanOrEqualTo(other Time, units ...Unit) bool {
	return tempo.SameOrBefore(other, units...)
}

func (tempo Time) Lte(other Time, units ...Unit) bool {
	return tempo.LessThanOrEqualTo(other, units...)
}

func (tempo Time) SameSecond(other Time) bool {
	return tempo.Same(other, Second)
}

func (tempo Time) SameMinute(other Time) bool {
	return tempo.Same(other, Minute)
}

func (tempo Time) SameHour(other Time) bool {
	return tempo.Same(other, Hour)
}

func (tempo Time) SameDay(other Time) bool {
	return tempo.Same(other, Day)
}

func (tempo Time) SameWeek(other Time) bool {
	return tempo.Same(other, Week)
}

func (tempo Time) SameMonth(other Time) bool {
	return tempo.Same(other, Month)
}

func (tempo Time) SameQuarter(other Time) bool {
	return tempo.Same(other, Quarter)
}

func (tempo Time) SameYear(other Time) bool {
	return tempo.Same(other, Year)
}

func (tempo Time) SameAs(pattern string, other Time) bool {
	return tempo.Format(pattern) == other.Format(pattern)
}

func (tempo Time) IsSameUnit(unit Unit, other Time) bool {
	return tempo.Same(other, unit)
}

func (tempo Time) Birthday(other Time) bool {
	return tempo.Month() == other.Month() && tempo.Day() == other.Day()
}

func (tempo Time) Clamp(minimum Time, maximum Time) (Time, error) {
	return comparison.Clamp(tempo, minimum.State(), maximum.State())
}

func (tempo Time) Average(other Time) Time {
	return comparison.Average(tempo, other.State())
}

func (tempo Time) Closest(first Time, rest ...Time) Time {
	return comparison.Closest(tempo, first.State(), states(rest)...)
}

func (tempo Time) Farthest(first Time, rest ...Time) Time {
	return comparison.Farthest(tempo, first.State(), states(rest)...)
}

func (tempo Time) Min(other Time) Time {
	return comparison.Min(tempo, other.State())
}

func (tempo Time) Max(other Time) Time {
	return comparison.Max(tempo, other.State())
}

func (tempo Time) SameOrBefore(other Time, units ...Unit) bool {
	return comparison.SameOrBefore(tempo, other.State(), units...)
}

func (tempo Time) SameOrAfter(other Time, units ...Unit) bool {
	return comparison.SameOrAfter(tempo, other.State(), units...)
}

func (tempo Time) Between(start Time, end Time, inclusivity ...string) bool {
	return comparison.Between(tempo, start.State(), end.State(), inclusivity...)
}

func (tempo Time) IsBetween(start Time, end Time, inclusivity ...string) bool {
	return tempo.Between(start, end, inclusivity...)
}

func (tempo Time) BetweenIncluded(start Time, end Time) bool {
	return tempo.Between(start, end, "[]")
}

func (tempo Time) BetweenExcluded(start Time, end Time) bool {
	return tempo.Between(start, end, "()")
}
