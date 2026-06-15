package diff

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

func (tempo Tempo) Diff(other tempopkg.Tempo, unit tempopkg.Unit, options ...tempopkg.DiffOptions) float64 {
	return tempo.value.Diff(other, unit, options...)
}

func (tempo Tempo) DiffAsDuration(other tempopkg.Tempo, options ...tempopkg.DiffOptions) tempopkg.Duration {
	return tempo.value.DiffAsDuration(other, options...)
}

func (tempo Tempo) DiffInDays(other tempopkg.Tempo, options ...tempopkg.DiffOptions) int {
	return tempo.value.DiffInDays(other, options...)
}

func (tempo Tempo) DiffInMonths(other tempopkg.Tempo, options ...tempopkg.DiffOptions) int {
	return tempo.value.DiffInMonths(other, options...)
}

func (tempo Tempo) DiffInYears(other tempopkg.Tempo, options ...tempopkg.DiffOptions) int {
	return tempo.value.DiffInYears(other, options...)
}

func (tempo Tempo) ForHumans(other tempopkg.Tempo, options ...tempopkg.HumanDiffOptions) string {
	return tempo.value.DiffForHumans(other, options...)
}

func (tempo Tempo) FromNow(options ...tempopkg.HumanDiffOptions) string {
	return tempo.value.FromNow(options...)
}

func (tempo Tempo) ToNow(options ...tempopkg.HumanDiffOptions) string {
	return tempo.value.ToNow(options...)
}
