package tempo

import (
	"time"

	intervalpkg "alloy.dev/backend/tempo/interval"
)

func (interval Interval) span() intervalpkg.Span {
	return intervalpkg.New(interval.Start.value, interval.End.value)
}

func (interval Interval) Inverted() bool {
	return interval.span().Inverted()
}

func (interval Interval) Milliseconds() int {
	return interval.span().Milliseconds()
}

func (interval Interval) Seconds() int {
	return interval.span().Seconds()
}

func (interval Interval) Minutes() int {
	return interval.span().Minutes()
}

func (interval Interval) Hours() int {
	return interval.span().Hours()
}

func (interval Interval) Days() int {
	return interval.span().Days()
}

func (interval Interval) Weeks() int {
	return interval.span().Weeks()
}

func (interval Interval) Months() int {
	return interval.End.DiffInMonths(interval.Start)
}

func (interval Interval) Quarters() int {
	return interval.End.DiffInQuarters(interval.Start)
}

func (interval Interval) Years() int {
	return interval.End.DiffInYears(interval.Start)
}

func (interval Interval) ToDuration() Duration {
	value := Duration{Milliseconds: interval.Milliseconds()}

	return value.Normalize()
}

func (interval Interval) Invert() Interval {
	return Interval{Start: interval.End, End: interval.Start}
}

func (interval Interval) Abs() Interval {
	if interval.Inverted() {
		return interval.Invert()
	}

	return interval
}

func (interval Interval) Contains(input Time, inclusivity ...string) bool {
	mode := "[]"

	if len(inclusivity) > 0 {
		mode = inclusivity[0]
	}

	return interval.span().Contains(input.TimestampMs(), mode)
}

func (interval Interval) Overlaps(other Interval) bool {
	return interval.span().Overlaps(other.span())
}

func (interval Interval) Intersection(other Interval) (Interval, bool) {
	intersection, ok := interval.span().Intersection(other.span())

	if !ok {
		return Interval{}, false
	}

	return Interval{
		Start: newTempo(time.UnixMilli(intersection.StartMs), interval.Start.location, interval.Start.Context()),
		End:   newTempo(time.UnixMilli(intersection.EndMs), interval.End.location, interval.End.Context()),
	}, true
}

func (interval Interval) Union(other Interval) Interval {
	union := interval.span().Union(other.span())

	return Interval{
		Start: newTempo(time.UnixMilli(union.StartMs), interval.Start.location, interval.Start.Context()),
		End:   newTempo(time.UnixMilli(union.EndMs), interval.End.location, interval.End.Context()),
	}
}
