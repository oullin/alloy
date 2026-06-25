package tempo

import "time"

func (tempo Time) IntervalUntil(end Time) Interval {
	return Interval{Start: tempo, End: end}
}

func (tempo Time) PeriodUntil(end Time, options ...PeriodOptions) Period {
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

func (tempo Time) ToPeriod(end Time, options ...PeriodOptions) Period {
	return tempo.PeriodUntil(end, options...)
}

func (tempo Time) Until(end Time, options ...PeriodOptions) Period {
	return tempo.PeriodUntil(end, options...)
}

func (tempo Time) Range(end Time, options ...PeriodOptions) Period {
	return tempo.PeriodUntil(end, options...)
}

func (tempo Time) local() time.Time {
	return tempo.value.In(tempo.location)
}
