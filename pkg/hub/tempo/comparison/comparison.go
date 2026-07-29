// Package comparison exposes generic ordering and selection helpers that
// work on any core.Bearer — both the immutable Time and the mutable
// *MutableTime — so before/after/same/clamp/min/max/closest/farthest
// live in one place.
//
// Comparison queries take a core.State for the right-hand side rather
// than another generic Bearer, so callers can compare against any
// concrete bearer's snapshot without paying for a second type parameter.
// The package depends only on core, duration and internal/kernel — never
// on the higher-level tempo package — keeping it safe to import anywhere.
package comparison

import (
	"errors"
	"strings"
	"time"

	"hara.sh/alloy/tempo/core"
	"hara.sh/alloy/tempo/duration"
	"hara.sh/alloy/tempo/internal/kernel"
)

func Before[T core.Bearer[T]](bearer T, other core.State, units ...duration.Unit) bool {
	return compareValue(bearer.State(), units...) < compareValue(other, units...)
}

func After[T core.Bearer[T]](bearer T, other core.State, units ...duration.Unit) bool {
	return compareValue(bearer.State(), units...) > compareValue(other, units...)
}

func Same[T core.Bearer[T]](bearer T, other core.State, units ...duration.Unit) bool {
	return compareValue(bearer.State(), units...) == compareValue(other, units...)
}

func SameOrBefore[T core.Bearer[T]](bearer T, other core.State, units ...duration.Unit) bool {
	value := compareValue(bearer.State(), units...)
	target := compareValue(other, units...)

	return value <= target
}

func SameOrAfter[T core.Bearer[T]](bearer T, other core.State, units ...duration.Unit) bool {
	value := compareValue(bearer.State(), units...)
	target := compareValue(other, units...)

	return value >= target
}

func Between[T core.Bearer[T]](bearer T, start core.State, end core.State, inclusivity ...string) bool {
	if compareValue(start) > compareValue(end) {
		start, end = end, start
	}

	mode := "[]"

	if len(inclusivity) > 0 {
		mode = inclusivity[0]
	}

	value := compareValue(bearer.State())
	startValue := compareValue(start)
	endValue := compareValue(end)
	afterStart := value > startValue

	if strings.HasPrefix(mode, "[") {
		afterStart = value >= startValue
	}

	beforeEnd := value < endValue

	if strings.HasSuffix(mode, "]") {
		beforeEnd = value <= endValue
	}

	return afterStart && beforeEnd
}

func Clamp[T core.Bearer[T]](bearer T, minimum core.State, maximum core.State) (T, error) {
	var zero T

	if compareValue(minimum) > compareValue(maximum) {
		return zero, errors.New("tempo clamp minimum must be before maximum")
	}

	value := compareValue(bearer.State())

	if value < compareValue(minimum) {
		return bearer.With(minimum.Value), nil
	}

	if value > compareValue(maximum) {
		return bearer.With(maximum.Value), nil
	}

	return bearer.With(bearer.State().Value), nil
}

func Average[T core.Bearer[T]](bearer T, other core.State) T {
	state := bearer.State()
	avg := kernel.AverageMilliseconds(state.Value.UnixMilli(), other.Value.UnixMilli())

	return bearer.With(time.UnixMilli(avg).UTC())
}

func Closest[T core.Bearer[T]](bearer T, first core.State, rest ...core.State) T {
	reference := bearer.State().Value.UnixMilli()
	bestState := first
	bestDistance := kernel.DistanceInt64(first.Value.UnixMilli(), reference)

	for _, candidate := range rest {
		distance := kernel.DistanceInt64(candidate.Value.UnixMilli(), reference)

		if distance < bestDistance {
			bestState = candidate
			bestDistance = distance
		}
	}

	return bearer.With(bestState.Value)
}

func Farthest[T core.Bearer[T]](bearer T, first core.State, rest ...core.State) T {
	reference := bearer.State().Value.UnixMilli()
	bestState := first
	bestDistance := kernel.DistanceInt64(first.Value.UnixMilli(), reference)

	for _, candidate := range rest {
		distance := kernel.DistanceInt64(candidate.Value.UnixMilli(), reference)

		if distance > bestDistance {
			bestState = candidate
			bestDistance = distance
		}
	}

	return bearer.With(bestState.Value)
}

func Min[T core.Bearer[T]](bearer T, other core.State) T {
	if compareValue(bearer.State()) <= compareValue(other) {
		return bearer.With(bearer.State().Value)
	}

	return bearer.With(other.Value)
}

func Max[T core.Bearer[T]](bearer T, other core.State) T {
	if compareValue(bearer.State()) >= compareValue(other) {
		return bearer.With(bearer.State().Value)
	}

	return bearer.With(other.Value)
}

func compareValue(state core.State, units ...duration.Unit) int64 {
	unit := duration.Millisecond

	if len(units) > 0 {
		unit = units[0]
	}

	if duration.NormalizeUnit(unit) == duration.Millisecond {
		return state.Value.UnixMilli()
	}

	return kernel.CompareValue(state.Value, state.Location, unit) / int64(time.Millisecond)
}
