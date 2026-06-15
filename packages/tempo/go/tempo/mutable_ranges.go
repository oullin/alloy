package tempo

import mutablepkg "github.com/oullin/alloy/tempo/mutable"

func (mutable *MutableTempo) IntervalUntil(end Tempo) Interval {
	return mutable.Tempo().IntervalUntil(end)
}

func (mutable *MutableTempo) PeriodUntil(end Tempo, options ...PeriodOptions) Period {
	return mutable.Tempo().PeriodUntil(end, options...)
}

func (mutable *MutableTempo) ToPeriod(end Tempo, options ...PeriodOptions) Period {
	return mutable.Tempo().ToPeriod(end, options...)
}

func (mutable *MutableTempo) Until(end Tempo, options ...PeriodOptions) Period {
	return mutable.Tempo().Until(end, options...)
}

func (mutable *MutableTempo) Range(end Tempo, options ...PeriodOptions) Period {
	return mutable.Tempo().Range(end, options...)
}

func (mutable *MutableTempo) replace(next Tempo) *MutableTempo {
	state := mutablepkg.Replace(
		mutablepkg.State[Runtime]{Value: mutable.value, Location: mutable.location, Runtime: mutable.Runtime()},
		mutablepkg.State[Runtime]{Value: next.value, Location: next.location, Runtime: next.Runtime()},
	)
	mutable.value = state.Value
	mutable.location = state.Location
	mutable.runtime = state.Runtime

	return mutable
}
