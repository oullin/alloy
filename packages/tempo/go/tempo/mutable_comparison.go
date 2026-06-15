package tempo

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
	return mutable.Tempo().Diff(other, unit, options...)
}

func (mutable *MutableTempo) DiffAsDuration(other Tempo, options ...DiffOptions) Duration {
	return mutable.Tempo().DiffAsDuration(other, options...)
}

func (mutable *MutableTempo) DiffAsDateInterval(other Tempo, options ...DiffOptions) Duration {
	return mutable.Tempo().DiffAsDateInterval(other, options...)
}

func (mutable *MutableTempo) DiffAsTempoInterval(other Tempo, options ...DiffOptions) Duration {
	return mutable.Tempo().DiffAsTempoInterval(other, options...)
}

func (mutable *MutableTempo) DiffInMilliseconds(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInMilliseconds(other, options...)
}

func (mutable *MutableTempo) DiffInMicroseconds(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInMicroseconds(other, options...)
}

func (mutable *MutableTempo) DiffInSeconds(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInSeconds(other, options...)
}

func (mutable *MutableTempo) DiffInMinutes(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInMinutes(other, options...)
}

func (mutable *MutableTempo) DiffInHours(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInHours(other, options...)
}

func (mutable *MutableTempo) DiffInDays(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInDays(other, options...)
}

func (mutable *MutableTempo) DiffInWeeks(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInWeeks(other, options...)
}

func (mutable *MutableTempo) DiffInWeekdays(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInWeekdays(other, options...)
}

func (mutable *MutableTempo) DiffInWeekendDays(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInWeekendDays(other, options...)
}

func (mutable *MutableTempo) DiffInMonths(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInMonths(other, options...)
}

func (mutable *MutableTempo) DiffInQuarters(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInQuarters(other, options...)
}

func (mutable *MutableTempo) DiffInYears(other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInYears(other, options...)
}

func (mutable *MutableTempo) DiffInUnit(unit Unit, other Tempo, options ...DiffOptions) int {
	return mutable.Tempo().DiffInUnit(unit, other, options...)
}

func (mutable *MutableTempo) DiffInDaysFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return mutable.Tempo().DiffInDaysFiltered(other, predicate, options...)
}

func (mutable *MutableTempo) DiffFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return mutable.Tempo().DiffFiltered(other, predicate, options...)
}

func (mutable *MutableTempo) DiffInHoursFiltered(other Tempo, predicate func(Tempo) bool, options ...DiffOptions) int {
	return mutable.Tempo().DiffInHoursFiltered(other, predicate, options...)
}

func (mutable *MutableTempo) SecondsSinceMidnight() int {
	return mutable.Tempo().SecondsSinceMidnight()
}

func (mutable *MutableTempo) SecondsUntilEndOfDay() int {
	return mutable.Tempo().SecondsUntilEndOfDay()
}

func (mutable *MutableTempo) Calendar(reference Tempo, formats ...map[string]string) string {
	return mutable.Tempo().Calendar(reference, formats...)
}

func (mutable *MutableTempo) DiffForHumans(other Tempo, options ...HumanDiffOptions) string {
	return mutable.Tempo().DiffForHumans(other, options...)
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
	return mutable.Tempo().Before(other, units...)
}

func (mutable *MutableTempo) After(other Tempo, units ...Unit) bool {
	return mutable.Tempo().After(other, units...)
}

func (mutable *MutableTempo) Same(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Same(other, units...)
}

func (mutable *MutableTempo) Is(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Is(other, units...)
}

func (mutable *MutableTempo) EqualTo(other Tempo, units ...Unit) bool {
	return mutable.Tempo().EqualTo(other, units...)
}

func (mutable *MutableTempo) Eq(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Eq(other, units...)
}

func (mutable *MutableTempo) NotEqualTo(other Tempo, units ...Unit) bool {
	return mutable.Tempo().NotEqualTo(other, units...)
}

func (mutable *MutableTempo) Ne(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Ne(other, units...)
}

func (mutable *MutableTempo) GreaterThan(other Tempo, units ...Unit) bool {
	return mutable.Tempo().GreaterThan(other, units...)
}

func (mutable *MutableTempo) Gt(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Gt(other, units...)
}

func (mutable *MutableTempo) GreaterThanOrEqualTo(other Tempo, units ...Unit) bool {
	return mutable.Tempo().GreaterThanOrEqualTo(other, units...)
}

func (mutable *MutableTempo) Gte(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Gte(other, units...)
}

func (mutable *MutableTempo) LessThan(other Tempo, units ...Unit) bool {
	return mutable.Tempo().LessThan(other, units...)
}

func (mutable *MutableTempo) Lt(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Lt(other, units...)
}

func (mutable *MutableTempo) LessThanOrEqualTo(other Tempo, units ...Unit) bool {
	return mutable.Tempo().LessThanOrEqualTo(other, units...)
}

func (mutable *MutableTempo) Lte(other Tempo, units ...Unit) bool {
	return mutable.Tempo().Lte(other, units...)
}

func (mutable *MutableTempo) SameSecond(other Tempo) bool {
	return mutable.Tempo().SameSecond(other)
}

func (mutable *MutableTempo) SameMinute(other Tempo) bool {
	return mutable.Tempo().SameMinute(other)
}

func (mutable *MutableTempo) SameHour(other Tempo) bool {
	return mutable.Tempo().SameHour(other)
}

func (mutable *MutableTempo) SameDay(other Tempo) bool {
	return mutable.Tempo().SameDay(other)
}

func (mutable *MutableTempo) SameWeek(other Tempo) bool {
	return mutable.Tempo().SameWeek(other)
}

func (mutable *MutableTempo) SameMonth(other Tempo) bool {
	return mutable.Tempo().SameMonth(other)
}

func (mutable *MutableTempo) SameQuarter(other Tempo) bool {
	return mutable.Tempo().SameQuarter(other)
}

func (mutable *MutableTempo) SameYear(other Tempo) bool {
	return mutable.Tempo().SameYear(other)
}

func (mutable *MutableTempo) SameAs(pattern string, other Tempo) bool {
	return mutable.Tempo().SameAs(pattern, other)
}

func (mutable *MutableTempo) IsSameUnit(unit Unit, other Tempo) bool {
	return mutable.Tempo().IsSameUnit(unit, other)
}

func (mutable *MutableTempo) Birthday(other Tempo) bool {
	return mutable.Tempo().Birthday(other)
}

func (mutable *MutableTempo) Clamp(minimum Tempo, maximum Tempo) (*MutableTempo, error) {
	next, err := mutable.Tempo().Clamp(minimum, maximum)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) Average(other Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Average(other))
}

func (mutable *MutableTempo) Closest(first Tempo, rest ...Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Closest(first, rest...))
}

func (mutable *MutableTempo) Farthest(first Tempo, rest ...Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Farthest(first, rest...))
}

func (mutable *MutableTempo) Min(other Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Min(other))
}

func (mutable *MutableTempo) Max(other Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Max(other))
}

func (mutable *MutableTempo) Minimum(other Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Minimum(other))
}

func (mutable *MutableTempo) Maximum(other Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Maximum(other))
}

func (mutable *MutableTempo) SameOrBefore(other Tempo, units ...Unit) bool {
	return mutable.Tempo().SameOrBefore(other, units...)
}

func (mutable *MutableTempo) SameOrAfter(other Tempo, units ...Unit) bool {
	return mutable.Tempo().SameOrAfter(other, units...)
}

func (mutable *MutableTempo) Between(start Tempo, end Tempo, inclusivity ...string) bool {
	return mutable.Tempo().Between(start, end, inclusivity...)
}

func (mutable *MutableTempo) IsBetween(start Tempo, end Tempo, inclusivity ...string) bool {
	return mutable.Tempo().IsBetween(start, end, inclusivity...)
}

func (mutable *MutableTempo) BetweenIncluded(start Tempo, end Tempo) bool {
	return mutable.Tempo().BetweenIncluded(start, end)
}

func (mutable *MutableTempo) BetweenExcluded(start Tempo, end Tempo) bool {
	return mutable.Tempo().BetweenExcluded(start, end)
}
