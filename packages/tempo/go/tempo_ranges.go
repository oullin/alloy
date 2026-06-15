package tempo

import "time"

func (tempo Tempo) IntervalUntil(end Tempo) Interval {
	return Interval{Start: tempo, End: end}
}

func (tempo Tempo) PeriodUntil(end Tempo, options ...PeriodOptions) Period {
	step := Duration{Days: 1}
	includeEnd := true
	if len(options) > 0 {
		if options[0].Step != (Duration{}) {
			step = options[0].Step
		}
		if options[0].ExcludeEnd {
			includeEnd = false
		} else if options[0].IncludeEnd {
			includeEnd = true
		}
	}

	return Period{Start: tempo, End: end, Step: step, IncludeEnd: includeEnd}
}

func (tempo Tempo) ToPeriod(end Tempo, options ...PeriodOptions) Period {
	return tempo.PeriodUntil(end, options...)
}

func (tempo Tempo) Until(end Tempo, options ...PeriodOptions) Period {
	return tempo.PeriodUntil(end, options...)
}

func (tempo Tempo) Range(end Tempo, options ...PeriodOptions) Period {
	return tempo.PeriodUntil(end, options...)
}

func (tempo Tempo) local() time.Time {
	return tempo.value.In(tempo.location)
}

func (tempo Tempo) addDuration(duration time.Duration) Tempo {
	return newTempo(tempo.value.Add(duration), tempo.location, tempo.Runtime())
}

func (tempo Tempo) addDurationDate(years int, months int, days int) Tempo {
	return newTempo(tempo.local().AddDate(years, months, days), tempo.location, tempo.Runtime())
}

func (tempo Tempo) fromLocal(local time.Time) Tempo {
	return newTempo(local, tempo.location, tempo.Runtime())
}

func (tempo Tempo) compareValue(units ...Unit) int64 {
	unit := Millisecond
	if len(units) > 0 {
		unit = units[0]
	}

	if normalizeUnit(unit) == Millisecond {
		return tempo.TimestampMs()
	}

	return tempo.StartOf(unit).TimestampMs()
}

func (tempo Tempo) diffFilteredDays(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	opts := DiffOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	sign := 1
	start := other.StartOf(Day)
	end := tempo.StartOf(Day)
	if tempo.Before(other, Day) {
		sign = -1
		start = tempo.StartOf(Day)
		end = other.StartOf(Day)
	}

	current := start
	count := 0
	for current.Before(end, Day) {
		current = current.AddDays(1)
		if current.SameOrBefore(end, Day) && predicate(current) {
			count++
		}
	}

	if opts.Absolute {
		return count
	}

	return count * sign
}
