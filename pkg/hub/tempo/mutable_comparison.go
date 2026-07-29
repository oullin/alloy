package tempo

import (
	"hara.sh/alloy/tempo/comparison"
	"hara.sh/alloy/tempo/diff"
)

func (mutable *MutableTime) IsStartOf(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Immutable().IsStartOf(unit, options...)
}

func (mutable *MutableTime) IsStartOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Immutable().IsStartOfUnit(unit, options...)
}

func (mutable *MutableTime) IsEndOf(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Immutable().IsEndOf(unit, options...)
}

func (mutable *MutableTime) IsEndOfUnit(unit Unit, options ...StartOfWeekOptions) bool {
	return mutable.Immutable().IsEndOfUnit(unit, options...)
}

func (mutable *MutableTime) IsCurrentUnit(unit Unit, reference Time) bool {
	return mutable.Immutable().IsCurrentUnit(unit, reference)
}

func (mutable *MutableTime) IsStartOfMillisecond() bool {
	return mutable.Immutable().IsStartOfMillisecond()
}

func (mutable *MutableTime) IsEndOfMillisecond() bool {
	return mutable.Immutable().IsEndOfMillisecond()
}

func (mutable *MutableTime) IsStartOfSecond() bool {
	return mutable.Immutable().IsStartOfSecond()
}

func (mutable *MutableTime) IsEndOfSecond() bool {
	return mutable.Immutable().IsEndOfSecond()
}

func (mutable *MutableTime) IsStartOfMinute() bool {
	return mutable.Immutable().IsStartOfMinute()
}

func (mutable *MutableTime) IsEndOfMinute() bool {
	return mutable.Immutable().IsEndOfMinute()
}

func (mutable *MutableTime) IsStartOfHour() bool {
	return mutable.Immutable().IsStartOfHour()
}

func (mutable *MutableTime) IsEndOfHour() bool {
	return mutable.Immutable().IsEndOfHour()
}

func (mutable *MutableTime) IsStartOfDay() bool {
	return mutable.Immutable().IsStartOfDay()
}

func (mutable *MutableTime) IsEndOfDay() bool {
	return mutable.Immutable().IsEndOfDay()
}

func (mutable *MutableTime) IsStartOfWeek(options ...StartOfWeekOptions) bool {
	return mutable.Immutable().IsStartOfWeek(options...)
}

func (mutable *MutableTime) IsEndOfWeek(options ...StartOfWeekOptions) bool {
	return mutable.Immutable().IsEndOfWeek(options...)
}

func (mutable *MutableTime) IsStartOfMonth() bool {
	return mutable.Immutable().IsStartOfMonth()
}

func (mutable *MutableTime) IsEndOfMonth() bool {
	return mutable.Immutable().IsEndOfMonth()
}

func (mutable *MutableTime) IsStartOfQuarter() bool {
	return mutable.Immutable().IsStartOfQuarter()
}

func (mutable *MutableTime) IsEndOfQuarter() bool {
	return mutable.Immutable().IsEndOfQuarter()
}

func (mutable *MutableTime) IsStartOfYear() bool {
	return mutable.Immutable().IsStartOfYear()
}

func (mutable *MutableTime) IsEndOfYear() bool {
	return mutable.Immutable().IsEndOfYear()
}

func (mutable *MutableTime) IsStartOfDecade() bool {
	return mutable.Immutable().IsStartOfDecade()
}

func (mutable *MutableTime) IsEndOfDecade() bool {
	return mutable.Immutable().IsEndOfDecade()
}

func (mutable *MutableTime) IsStartOfCentury() bool {
	return mutable.Immutable().IsStartOfCentury()
}

func (mutable *MutableTime) IsEndOfCentury() bool {
	return mutable.Immutable().IsEndOfCentury()
}

func (mutable *MutableTime) IsStartOfMillennium() bool {
	return mutable.Immutable().IsStartOfMillennium()
}

func (mutable *MutableTime) IsEndOfMillennium() bool {
	return mutable.Immutable().IsEndOfMillennium()
}

func (mutable *MutableTime) IsStartOfTime() bool {
	return mutable.Immutable().IsStartOfTime()
}

func (mutable *MutableTime) IsEndOfTime() bool {
	return mutable.Immutable().IsEndOfTime()
}

func (mutable *MutableTime) Diff(other Time, unit Unit, options ...DiffOptions) float64 {
	return diff.Between(mutable, other.State(), unit, diffOptions(options)...)
}

func (mutable *MutableTime) DiffAsDuration(other Time, options ...DiffOptions) Duration {
	value := Duration{Milliseconds: mutable.DiffInMilliseconds(other, options...)}

	return value.Normalize()
}

func (mutable *MutableTime) DiffAsDateInterval(other Time, options ...DiffOptions) Duration {
	return mutable.DiffAsDuration(other, options...)
}

func (mutable *MutableTime) DiffAsTempoInterval(other Time, options ...DiffOptions) Duration {
	return mutable.DiffAsDuration(other, options...)
}

func (mutable *MutableTime) DiffInMilliseconds(other Time, options ...DiffOptions) int {
	return diff.InMilliseconds(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInMicroseconds(other Time, options ...DiffOptions) int {
	return diff.InMicroseconds(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInSeconds(other Time, options ...DiffOptions) int {
	return diff.InSeconds(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInMinutes(other Time, options ...DiffOptions) int {
	return diff.InMinutes(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInHours(other Time, options ...DiffOptions) int {
	return diff.InHours(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInDays(other Time, options ...DiffOptions) int {
	return diff.InDays(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInWeeks(other Time, options ...DiffOptions) int {
	return diff.InWeeks(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInWeekdays(other Time, options ...DiffOptions) int {
	return diff.InWeekdays(mutable, other.State(), mutable.settingsSnapshot().WeekendDays, diffOptions(options)...)
}

func (mutable *MutableTime) DiffInWeekendDays(other Time, options ...DiffOptions) int {
	return diff.InWeekendDays(mutable, other.State(), mutable.settingsSnapshot().WeekendDays, diffOptions(options)...)
}

func (mutable *MutableTime) DiffInMonths(other Time, options ...DiffOptions) int {
	return diff.InMonths(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInQuarters(other Time, options ...DiffOptions) int {
	return diff.InQuarters(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInYears(other Time, options ...DiffOptions) int {
	return diff.InYears(mutable, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInUnit(unit Unit, other Time, options ...DiffOptions) int {
	return diff.InUnit(mutable, unit, other.State(), diffOptions(options)...)
}

func (mutable *MutableTime) DiffInDaysFiltered(other Time, predicate func(Time) bool, options ...DiffOptions) int {
	return mutable.Immutable().DiffInDaysFiltered(other, predicate, options...)
}

func (mutable *MutableTime) DiffFiltered(other Time, predicate func(Time) bool, options ...DiffOptions) int {
	return mutable.DiffInDaysFiltered(other, predicate, options...)
}

func (mutable *MutableTime) DiffInHoursFiltered(other Time, predicate func(Time) bool, options ...DiffOptions) int {
	return mutable.Immutable().DiffInHoursFiltered(other, predicate, options...)
}

func (mutable *MutableTime) SecondsSinceMidnight() int {
	return diff.SecondsSinceMidnight(mutable)
}

func (mutable *MutableTime) SecondsUntilEndOfDay() int {
	return diff.SecondsUntilEndOfDay(mutable)
}

func (mutable *MutableTime) Calendar(reference Time, formats ...map[string]string) string {
	return mutable.Immutable().Calendar(reference, formats...)
}

func (mutable *MutableTime) DiffForHumans(other Time, options ...HumanDiffOptions) string {
	opts := mutable.settingsSnapshot().HumanDiff

	if len(options) > 0 {
		opts = options[0]
	}

	return diff.ForHumans(mutable, other.State(), diff.HumanOptions{Absolute: opts.Absolute, Unit: opts.Unit})
}

func (mutable *MutableTime) From(other Time, options ...HumanDiffOptions) string {
	return mutable.Immutable().From(other, options...)
}

func (mutable *MutableTime) Since(other Time, options ...HumanDiffOptions) string {
	return mutable.Immutable().Since(other, options...)
}

func (mutable *MutableTime) To(other Time, options ...HumanDiffOptions) string {
	return mutable.Immutable().To(other, options...)
}

func (mutable *MutableTime) FromNow(options ...HumanDiffOptions) string {
	return mutable.Immutable().FromNow(options...)
}

func (mutable *MutableTime) ToNow(options ...HumanDiffOptions) string {
	return mutable.Immutable().ToNow(options...)
}

func (mutable *MutableTime) Ago(options ...HumanDiffOptions) string {
	return mutable.Immutable().Ago(options...)
}

func (mutable *MutableTime) Timespan(other Time, options ...HumanDiffOptions) string {
	return mutable.Immutable().Timespan(other, options...)
}

func (mutable *MutableTime) IsImmutable() bool {
	return false
}

func (mutable *MutableTime) IsMutable() bool {
	return true
}

func (mutable *MutableTime) Before(other Time, units ...Unit) bool {
	return comparison.Before(mutable, other.State(), units...)
}

func (mutable *MutableTime) After(other Time, units ...Unit) bool {
	return comparison.After(mutable, other.State(), units...)
}

func (mutable *MutableTime) Same(other Time, units ...Unit) bool {
	return comparison.Same(mutable, other.State(), units...)
}

func (mutable *MutableTime) Is(other Time, units ...Unit) bool {
	return mutable.Same(other, units...)
}

func (mutable *MutableTime) EqualTo(other Time, units ...Unit) bool {
	return mutable.Same(other, units...)
}

func (mutable *MutableTime) Eq(other Time, units ...Unit) bool {
	return mutable.EqualTo(other, units...)
}

func (mutable *MutableTime) NotEqualTo(other Time, units ...Unit) bool {
	return !mutable.Same(other, units...)
}

func (mutable *MutableTime) Ne(other Time, units ...Unit) bool {
	return mutable.NotEqualTo(other, units...)
}

func (mutable *MutableTime) GreaterThan(other Time, units ...Unit) bool {
	return mutable.After(other, units...)
}

func (mutable *MutableTime) Gt(other Time, units ...Unit) bool {
	return mutable.GreaterThan(other, units...)
}

func (mutable *MutableTime) GreaterThanOrEqualTo(other Time, units ...Unit) bool {
	return mutable.SameOrAfter(other, units...)
}

func (mutable *MutableTime) Gte(other Time, units ...Unit) bool {
	return mutable.GreaterThanOrEqualTo(other, units...)
}

func (mutable *MutableTime) LessThan(other Time, units ...Unit) bool {
	return mutable.Before(other, units...)
}

func (mutable *MutableTime) Lt(other Time, units ...Unit) bool {
	return mutable.LessThan(other, units...)
}

func (mutable *MutableTime) LessThanOrEqualTo(other Time, units ...Unit) bool {
	return mutable.SameOrBefore(other, units...)
}

func (mutable *MutableTime) Lte(other Time, units ...Unit) bool {
	return mutable.LessThanOrEqualTo(other, units...)
}

func (mutable *MutableTime) SameSecond(other Time) bool {
	return mutable.Same(other, Second)
}

func (mutable *MutableTime) SameMinute(other Time) bool {
	return mutable.Same(other, Minute)
}

func (mutable *MutableTime) SameHour(other Time) bool {
	return mutable.Same(other, Hour)
}

func (mutable *MutableTime) SameDay(other Time) bool {
	return mutable.Same(other, Day)
}

func (mutable *MutableTime) SameWeek(other Time) bool {
	return mutable.Same(other, Week)
}

func (mutable *MutableTime) SameMonth(other Time) bool {
	return mutable.Same(other, Month)
}

func (mutable *MutableTime) SameQuarter(other Time) bool {
	return mutable.Same(other, Quarter)
}

func (mutable *MutableTime) SameYear(other Time) bool {
	return mutable.Same(other, Year)
}

func (mutable *MutableTime) SameAs(pattern string, other Time) bool {
	return mutable.Immutable().SameAs(pattern, other)
}

func (mutable *MutableTime) IsSameUnit(unit Unit, other Time) bool {
	return mutable.Same(other, unit)
}

func (mutable *MutableTime) Birthday(other Time) bool {
	return mutable.Immutable().Birthday(other)
}

func (mutable *MutableTime) Clamp(minimum Time, maximum Time) (*MutableTime, error) {
	return comparison.Clamp(mutable, minimum.State(), maximum.State())
}

func (mutable *MutableTime) Average(other Time) *MutableTime {
	return comparison.Average(mutable, other.State())
}

func (mutable *MutableTime) Closest(first Time, rest ...Time) *MutableTime {
	return comparison.Closest(mutable, first.State(), states(rest)...)
}

func (mutable *MutableTime) Farthest(first Time, rest ...Time) *MutableTime {
	return comparison.Farthest(mutable, first.State(), states(rest)...)
}

func (mutable *MutableTime) Min(other Time) *MutableTime {
	return comparison.Min(mutable, other.State())
}

func (mutable *MutableTime) Max(other Time) *MutableTime {
	return comparison.Max(mutable, other.State())
}

func (mutable *MutableTime) SameOrBefore(other Time, units ...Unit) bool {
	return comparison.SameOrBefore(mutable, other.State(), units...)
}

func (mutable *MutableTime) SameOrAfter(other Time, units ...Unit) bool {
	return comparison.SameOrAfter(mutable, other.State(), units...)
}

func (mutable *MutableTime) Between(start Time, end Time, inclusivity ...string) bool {
	return comparison.Between(mutable, start.State(), end.State(), inclusivity...)
}

func (mutable *MutableTime) IsBetween(start Time, end Time, inclusivity ...string) bool {
	return mutable.Between(start, end, inclusivity...)
}

func (mutable *MutableTime) BetweenIncluded(start Time, end Time) bool {
	return mutable.Between(start, end, "[]")
}

func (mutable *MutableTime) BetweenExcluded(start Time, end Time) bool {
	return mutable.Between(start, end, "()")
}
