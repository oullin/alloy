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
