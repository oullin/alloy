package ranges

import tempopkg "github.com/oullin/alloy/tempo/tempo"

type Tempo struct {
	value tempopkg.Tempo
}

func From(value tempopkg.Tempo) Tempo {
	return Tempo{value: value}
}

func (tempo Tempo) Tempo() tempopkg.Tempo {
	return tempo.value
}

func (tempo Tempo) IntervalUntil(end tempopkg.Tempo) tempopkg.Interval {
	return tempo.value.IntervalUntil(end)
}

func (tempo Tempo) PeriodUntil(end tempopkg.Tempo, options ...tempopkg.PeriodOptions) tempopkg.Period {
	return tempo.value.PeriodUntil(end, options...)
}

func (tempo Tempo) Range(end tempopkg.Tempo, options ...tempopkg.PeriodOptions) tempopkg.Period {
	return tempo.value.Range(end, options...)
}
