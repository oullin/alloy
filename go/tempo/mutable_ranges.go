package tempo

func (mutable *MutableTime) IntervalUntil(end Time) Interval {
	return mutable.Immutable().IntervalUntil(end)
}

func (mutable *MutableTime) PeriodUntil(end Time, options ...PeriodOptions) Period {
	return mutable.Immutable().PeriodUntil(end, options...)
}

func (mutable *MutableTime) ToPeriod(end Time, options ...PeriodOptions) Period {
	return mutable.Immutable().ToPeriod(end, options...)
}

func (mutable *MutableTime) Until(end Time, options ...PeriodOptions) Period {
	return mutable.Immutable().Until(end, options...)
}

func (mutable *MutableTime) Range(end Time, options ...PeriodOptions) Period {
	return mutable.Immutable().Range(end, options...)
}

func (mutable *MutableTime) replace(next Time) *MutableTime {
	mutable.value = next.value
	mutable.location = next.location
	mutable.runtime = next.Context()
	mutable.settings = next.settingsSnapshot()
	mutable.serializer = next.serializer
	mutable.toStringFormat = next.toStringFormat

	return mutable
}
