package tempo

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
	mutable.value = next.value
	mutable.location = next.location
	mutable.runtime = next.Runtime()
	mutable.settings = next.settingsSnapshot()
	mutable.serializer = next.serializer
	mutable.toStringFormat = next.toStringFormat

	return mutable
}
