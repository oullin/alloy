package tempo

import (
	"github.com/oullin/alloy/tempo/comparison"
	"github.com/oullin/alloy/tempo/diff"
)

func (mutable *MutableTempo) IsStartOf(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Tempo().IsStartOf(unit, options...)
}

func (mutable *MutableTempo) IsStartOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Tempo().IsStartOfUnit(unit, options...)
}

func (mutable *MutableTempo) IsEndOf(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Tempo().IsEndOf(unit, options...)
}

func (mutable *MutableTempo) IsEndOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Tempo().IsEndOfUnit(unit, options...)
}

func (mutable *MutableTempo) IsCurrentUnit(unit Unit, reference Tempo) bool {
	return mutable.Tempo().IsCurrentUnit(unit, reference)
}

func (mutable *MutableTempo) IsStartOfMillisecond() bool {
	return mutable.Tempo().IsStartOfMillisecond()
}

func (mutable *MutableTempo) IsEndOfMillisecond() bool {
	return mutable.Tempo().IsEndOfMillisecond()
}

func (mutable *MutableTempo) IsStartOfSecond() bool {
	return mutable.Tempo().IsStartOfSecond()
}

func (mutable *MutableTempo) IsEndOfSecond() bool {
	return mutable.Tempo().IsEndOfSecond()
}

func (mutable *MutableTempo) IsStartOfMinute() bool {
	return mutable.Tempo().IsStartOfMinute()
}

func (mutable *MutableTempo) IsEndOfMinute() bool {
	return mutable.Tempo().IsEndOfMinute()
}

func (mutable *MutableTempo) IsStartOfHour() bool {
	return mutable.Tempo().IsStartOfHour()
}

func (mutable *MutableTempo) IsEndOfHour() bool {
	return mutable.Tempo().IsEndOfHour()
}

func (mutable *MutableTempo) IsStartOfDay() bool {
	return mutable.Tempo().IsStartOfDay()
}

func (mutable *MutableTempo) IsEndOfDay() bool {
	return mutable.Tempo().IsEndOfDay()
}

func (mutable *MutableTempo) IsStartOfWeek(options ...StartOfWeekOptions) bool {
	return mutable.Tempo().IsStartOfWeek(options...)
}

func (mutable *MutableTempo) IsEndOfWeek(options ...StartOfWeekOptions) bool {
	return mutable.Tempo().IsEndOfWeek(options...)
}

func (mutable *MutableTempo) IsStartOfMonth() bool {
	return mutable.Tempo().IsStartOfMonth()
}

func (mutable *MutableTempo) IsEndOfMonth() bool {
	return mutable.Tempo().IsEndOfMonth()
}

func (mutable *MutableTempo) IsStartOfQuarter() bool {
	return mutable.Tempo().IsStartOfQuarter()
}

func (mutable *MutableTempo) IsEndOfQuarter() bool {
	return mutable.Tempo().IsEndOfQuarter()
}

func (mutable *MutableTempo) IsStartOfYear() bool {
	return mutable.Tempo().IsStartOfYear()
}

func (mutable *MutableTempo) IsEndOfYear() bool {
	return mutable.Tempo().IsEndOfYear()
}

func (mutable *MutableTempo) IsStartOfDecade() bool {
	return mutable.Tempo().IsStartOfDecade()
}

func (mutable *MutableTempo) IsEndOfDecade() bool {
	return mutable.Tempo().IsEndOfDecade()
}

func (mutable *MutableTempo) IsStartOfCentury() bool {
	return mutable.Tempo().IsStartOfCentury()
}

func (mutable *MutableTempo) IsEndOfCentury() bool {
	return mutable.Tempo().IsEndOfCentury()
}

func (mutable *MutableTempo) IsStartOfMillennium() bool {
	return mutable.Tempo().IsStartOfMillennium()
}

func (mutable *MutableTempo) IsEndOfMillennium() bool {
	return mutable.Tempo().IsEndOfMillennium()
}

func (mutable *MutableTempo) IsStartOfTime() bool {
	return mutable.Tempo().IsStartOfTime()
}

func (mutable *MutableTempo) IsEndOfTime() bool {
	return mutable.Tempo().IsEndOfTime()
}

func (mutable *MutableTempo) Diff(other Tempo, unit Unit, options ...DiffOptions) float64 {
	return diff.Between(mutable, other.State(), unit, diffOptions(options)...)
}

func (mutable *MutableTempo) DiffAsDuration(other Tempo, options ...DiffOptions) Duration {
	value := Duration{Milliseconds: mutable.DiffInMilliseconds(other, options...)}

	return value.Normalize()
}

func (mutable *MutableTempo) DiffAsDateInterval(other Tempo, options ...DiffOptions) Duration {
	return mutable.DiffAsDuration(other, options...)
}

func (mutable *MutableTempo) DiffAsTempoInterval(other Tempo, options ...DiffOptions) Duration {
	return mutable.DiffAsDuration(other, options...)
}

func (mutable *MutableTempo) DiffInMilliseconds(other Tempo, options ...DiffOptions) int {
	return diff.InMilliseconds(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInMicroseconds(other Tempo, options ...DiffOptions) int {
	return diff.InMicroseconds(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInSeconds(other Tempo, options ...DiffOptions) int {
	return diff.InSeconds(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInMinutes(other Tempo, options ...DiffOptions) int {
	return diff.InMinutes(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInHours(other Tempo, options ...DiffOptions) int {
	return diff.InHours(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInDays(other Tempo, options ...DiffOptions) int {
	return diff.InDays(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInWeeks(other Tempo, options ...DiffOptions) int {
	return diff.InWeeks(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInWeekdays(other Tempo, options ...DiffOptions) int {
	return diff.InWeekdays(mutable, other.State(), mutable.settingsSnapshot().WeekendDays, diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInWeekendDays(other Tempo, options ...DiffOptions) int {
	return diff.InWeekendDays(mutable, other.State(), mutable.settingsSnapshot().WeekendDays, diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInMonths(other Tempo, options ...DiffOptions) int {
	return diff.InMonths(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInQuarters(other Tempo, options ...DiffOptions) int {
	return diff.InQuarters(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInYears(other Tempo, options ...DiffOptions) int {
	return diff.InYears(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInUnit(unit Unit, other Tempo, options ...DiffOptions) int {
	return diff.InUnit(mutable, unit, other.State(), diffOptions(options)...)
}

func (mutable *MutableTempo) DiffInDaysFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return mutable.Tempo().DiffInDaysFiltered(other, predicate, options...)
}

func (mutable *MutableTempo) DiffFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return mutable.DiffInDaysFiltered(other, predicate, options...)
}

func (mutable *MutableTempo) DiffInHoursFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return mutable.Tempo().DiffInHoursFiltered(other, predicate, options...)
}

func (mutable *MutableTempo) SecondsSinceMidnight() int {
	return diff.SecondsSinceMidnight(mutable)
}

func (mutable *MutableTempo) SecondsUntilEndOfDay() int {
	return diff.SecondsUntilEndOfDay(mutable)
}

func (mutable *MutableTempo) Calendar(reference Tempo, formats ...map[string]string) string {
	return mutable.Tempo().Calendar(reference, formats...)
}

func (mutable *MutableTempo) DiffForHumans(other Tempo, options ...HumanDiffOptions) string {
	opts := mutable.settingsSnapshot().HumanDiff

	if len(options) > 0 {
		opts = options[0]
	}

	return diff.ForHumans(mutable, other.State(), diff.HumanOptions{Absolute: opts.Absolute, Unit: opts.Unit})
}

func (mutable *MutableTempo) From(other Tempo, options ...HumanDiffOptions) string {
	return mutable.Tempo().From(other, options...)
}

func (mutable *MutableTempo) Since(other Tempo, options ...HumanDiffOptions) string {
	return mutable.Tempo().Since(other, options...)
}

func (mutable *MutableTempo) To(other Tempo, options ...HumanDiffOptions) string {
	return mutable.Tempo().To(other, options...)
}

func (mutable *MutableTempo) FromNow(options ...HumanDiffOptions) string {
	return mutable.Tempo().FromNow(options...)
}

func (mutable *MutableTempo) ToNow(options ...HumanDiffOptions) string {
	return mutable.Tempo().ToNow(options...)
}

func (mutable *MutableTempo) Ago(options ...HumanDiffOptions) string {
	return mutable.Tempo().Ago(options...)
}

func (mutable *MutableTempo) Timespan(other Tempo, options ...HumanDiffOptions) string {
	return mutable.Tempo().Timespan(other, options...)
}

func (mutable *MutableTempo) IsImmutable() bool {
	return false
}

func (mutable *MutableTempo) IsMutable() bool {
	return true
}

func (mutable *MutableTempo) Before(other Tempo, units ...Unit) bool {
	return comparison.Before(mutable, other.State(), units...)
}

func (mutable *MutableTempo) After(other Tempo, units ...Unit) bool {
	return comparison.After(mutable, other.State(), units...)
}

func (mutable *MutableTempo) Same(other Tempo, units ...Unit) bool {
	return comparison.Same(mutable, other.State(), units...)
}

func (mutable *MutableTempo) Is(other Tempo, units ...Unit) bool {
	return mutable.Same(other, units...)
}

func (mutable *MutableTempo) EqualTo(other Tempo, units ...Unit) bool {
	return mutable.Same(other, units...)
}

func (mutable *MutableTempo) Eq(other Tempo, units ...Unit) bool {
	return mutable.EqualTo(other, units...)
}

func (mutable *MutableTempo) NotEqualTo(other Tempo, units ...Unit) bool {
	return !mutable.Same(other, units...)
}

func (mutable *MutableTempo) Ne(other Tempo, units ...Unit) bool {
	return mutable.NotEqualTo(other, units...)
}

func (mutable *MutableTempo) GreaterThan(other Tempo, units ...Unit) bool {
	return mutable.After(other, units...)
}

func (mutable *MutableTempo) Gt(other Tempo, units ...Unit) bool {
	return mutable.GreaterThan(other, units...)
}

func (mutable *MutableTempo) GreaterThanOrEqualTo(other Tempo, units ...Unit) bool {
	return mutable.SameOrAfter(other, units...)
}

func (mutable *MutableTempo) Gte(other Tempo, units ...Unit) bool {
	return mutable.GreaterThanOrEqualTo(other, units...)
}

func (mutable *MutableTempo) LessThan(other Tempo, units ...Unit) bool {
	return mutable.Before(other, units...)
}

func (mutable *MutableTempo) Lt(other Tempo, units ...Unit) bool {
	return mutable.LessThan(other, units...)
}

func (mutable *MutableTempo) LessThanOrEqualTo(other Tempo, units ...Unit) bool {
	return mutable.SameOrBefore(other, units...)
}

func (mutable *MutableTempo) Lte(other Tempo, units ...Unit) bool {
	return mutable.LessThanOrEqualTo(other, units...)
}

func (mutable *MutableTempo) SameSecond(other Tempo) bool {
	return mutable.Same(other, Second)
}

func (mutable *MutableTempo) SameMinute(other Tempo) bool {
	return mutable.Same(other, Minute)
}

func (mutable *MutableTempo) SameHour(other Tempo) bool {
	return mutable.Same(other, Hour)
}

func (mutable *MutableTempo) SameDay(other Tempo) bool {
	return mutable.Same(other, Day)
}

func (mutable *MutableTempo) SameWeek(other Tempo) bool {
	return mutable.Same(other, Week)
}

func (mutable *MutableTempo) SameMonth(other Tempo) bool {
	return mutable.Same(other, Month)
}

func (mutable *MutableTempo) SameQuarter(other Tempo) bool {
	return mutable.Same(other, Quarter)
}

func (mutable *MutableTempo) SameYear(other Tempo) bool {
	return mutable.Same(other, Year)
}

func (mutable *MutableTempo) SameAs(pattern string, other Tempo) bool {
	return mutable.Tempo().SameAs(pattern, other)
}

func (mutable *MutableTempo) IsSameUnit(unit Unit, other Tempo) bool {
	return mutable.Same(other, unit)
}

func (mutable *MutableTempo) Birthday(other Tempo) bool {
	return mutable.Tempo().Birthday(other)
}

func (mutable *MutableTempo) Clamp(minimum Tempo, maximum Tempo) (*MutableTempo, error) {
	return comparison.Clamp(mutable, minimum.State(), maximum.State())
}

func (mutable *MutableTempo) Average(other Tempo) *MutableTempo {
	return comparison.Average(mutable, other.State())
}

func (mutable *MutableTempo) Closest(first Tempo, rest ...Tempo) *MutableTempo {
	return comparison.Closest(mutable, first.State(), states(rest)...)
}

func (mutable *MutableTempo) Farthest(first Tempo, rest ...Tempo) *MutableTempo {
	return comparison.Farthest(mutable, first.State(), states(rest)...)
}

func (mutable *MutableTempo) Min(other Tempo) *MutableTempo {
	return comparison.Min(mutable, other.State())
}

func (mutable *MutableTempo) Max(other Tempo) *MutableTempo {
	return comparison.Max(mutable, other.State())
}

func (mutable *MutableTempo) SameOrBefore(other Tempo, units ...Unit) bool {
	return comparison.SameOrBefore(mutable, other.State(), units...)
}

func (mutable *MutableTempo) SameOrAfter(other Tempo, units ...Unit) bool {
	return comparison.SameOrAfter(mutable, other.State(), units...)
}

func (mutable *MutableTempo) Between(start Tempo, end Tempo, inclusivity ...string) bool {
	return comparison.Between(mutable, start.State(), end.State(), inclusivity...)
}

func (mutable *MutableTempo) IsBetween(start Tempo, end Tempo, inclusivity ...string) bool {
	return mutable.Between(start, end, inclusivity...)
}

func (mutable *MutableTempo) BetweenIncluded(start Tempo, end Tempo) bool {
	return mutable.Between(start, end, "[]")
}

func (mutable *MutableTempo) BetweenExcluded(start Tempo, end Tempo) bool {
	return mutable.Between(start, end, "()")
}
