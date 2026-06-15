package tempo

func (interval Interval) Inverted() bool {
	return interval.Start.After(interval.End)
}

func (interval Interval) Milliseconds() int {
	return interval.End.DiffInMilliseconds(interval.Start)
}

func (interval Interval) Seconds() int {
	return interval.End.DiffInSeconds(interval.Start)
}

func (interval Interval) Minutes() int {
	return interval.End.DiffInMinutes(interval.Start)
}

func (interval Interval) Hours() int {
	return interval.End.DiffInHours(interval.Start)
}

func (interval Interval) Days() int {
	return interval.End.DiffInDays(interval.Start)
}

func (interval Interval) Weeks() int {
	return interval.End.DiffInWeeks(interval.Start)
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
	return Duration{Milliseconds: interval.Milliseconds()}.Normalize()
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

func (interval Interval) Contains(input Tempo, inclusivity ...string) bool {
	mode := "[]"
	if len(inclusivity) > 0 {
		mode = inclusivity[0]
	}

	return input.Between(interval.Start, interval.End, mode)
}

func (interval Interval) Overlaps(other Interval) bool {
	return interval.Start.Before(other.End) && interval.End.After(other.Start)
}

func (interval Interval) Intersection(other Interval) (Interval, bool) {
	if !interval.Overlaps(other) {
		return Interval{}, false
	}

	return Interval{
		Start: Max(interval.Start, other.Start),
		End:   Min(interval.End, other.End),
	}, true
}

func (interval Interval) Union(other Interval) Interval {
	return Interval{
		Start: Min(interval.Start, other.Start),
		End:   Max(interval.End, other.End),
	}
}
