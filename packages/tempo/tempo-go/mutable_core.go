package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/setters"
)

func (mutable *MutableTempo) Tempo() Tempo {
	return newTempoWithPolicy(mutable.value, mutable.location, mutable.runtime, mutable.settingsSnapshot(), mutable.serializer, mutable.toStringFormat)
}

func (mutable *MutableTempo) Clone() *MutableTempo {
	return NewMutable(mutable.Tempo())
}

func (mutable *MutableTempo) AvoidMutation() *MutableTempo {
	return mutable.Clone()
}

func (mutable *MutableTempo) Cast() *MutableTempo {
	return mutable.Clone()
}

func (mutable *MutableTempo) Tempoize(input Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().Tempoize(input))
}

func (mutable *MutableTempo) NowWithSameTz() *MutableTempo {
	return mutable.replace(mutable.Tempo().NowWithSameTz())
}

func (mutable *MutableTempo) Modify(modifier string) (*MutableTempo, error) {
	next, err := mutable.Tempo().Modify(modifier)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) Change(modifier string) (*MutableTempo, error) {
	return mutable.Modify(modifier)
}

func (mutable *MutableTempo) Immutable() Tempo {
	return mutable.Tempo()
}

func (mutable *MutableTempo) Runtime() Runtime {
	return mutable.Tempo().Runtime()
}

func (mutable *MutableTempo) WithRuntime(runtime Runtime) *MutableTempo {
	return mutable.replace(mutable.Tempo().WithRuntime(runtime))
}

func (mutable *MutableTempo) WithTranslator(translator Translator) *MutableTempo {
	return mutable.replace(mutable.Tempo().WithTranslator(translator))
}

func (mutable *MutableTempo) HasTranslator() bool {
	return mutable.Tempo().HasTranslator()
}

func (mutable *MutableTempo) Timezone() string {
	return mutable.Tempo().Timezone()
}

func (mutable *MutableTempo) Timestamp() int64 {
	return mutable.Tempo().Timestamp()
}

func (mutable *MutableTempo) TimestampMs() int64 {
	return mutable.Tempo().TimestampMs()
}

func (mutable *MutableTempo) PreciseTimestamp(precisions ...int) float64 {
	return mutable.Tempo().PreciseTimestamp(precisions...)
}

func (mutable *MutableTempo) Unix() int64 {
	return mutable.Tempo().Unix()
}

func (mutable *MutableTempo) Year() int {
	return mutable.Tempo().Year()
}

func (mutable *MutableTempo) Month() int {
	return mutable.Tempo().Month()
}

func (mutable *MutableTempo) Quarter() int {
	return mutable.Tempo().Quarter()
}

func (mutable *MutableTempo) Day() int {
	return mutable.Tempo().Day()
}

func (mutable *MutableTempo) DayOfWeek() int {
	return mutable.Tempo().DayOfWeek()
}

func (mutable *MutableTempo) Weekday() int {
	return mutable.Tempo().Weekday()
}

func (mutable *MutableTempo) SetWeekday(weekday time.Weekday) *MutableTempo {
	return setters.SetWeekday(mutable, weekday)
}

func (mutable *MutableTempo) ISOWeekday() int {
	return mutable.Tempo().ISOWeekday()
}

func (mutable *MutableTempo) ISOWeek() (int, int) {
	return mutable.Tempo().ISOWeek()
}

func (mutable *MutableTempo) ISOWeekYear() int {
	return mutable.Tempo().ISOWeekYear()
}

func (mutable *MutableTempo) ISOWeekNumber() int {
	return mutable.Tempo().ISOWeekNumber()
}

func (mutable *MutableTempo) WeeksInISOYear() int {
	return mutable.Tempo().WeeksInISOYear()
}

func (mutable *MutableTempo) DayOfYear() int {
	return mutable.Tempo().DayOfYear()
}

func (mutable *MutableTempo) SetDayOfYear(day int) *MutableTempo {
	return setters.SetDayOfYear(mutable, day)
}

func (mutable *MutableTempo) Hour() int {
	return mutable.Tempo().Hour()
}

func (mutable *MutableTempo) Minute() int {
	return mutable.Tempo().Minute()
}

func (mutable *MutableTempo) Second() int {
	return mutable.Tempo().Second()
}

func (mutable *MutableTempo) Millisecond() int {
	return mutable.Tempo().Millisecond()
}

func (mutable *MutableTempo) PaddedUnit(field string, length int) (string, bool) {
	return mutable.Tempo().PaddedUnit(field, length)
}

func (mutable *MutableTempo) OffsetMinutes() int {
	return mutable.Tempo().OffsetMinutes()
}

func (mutable *MutableTempo) OffsetString(separator string) string {
	return mutable.Tempo().OffsetString(separator)
}

func (mutable *MutableTempo) UTCOffset() int {
	return mutable.Tempo().UTCOffset()
}

func (mutable *MutableTempo) ZoneName() string {
	return mutable.Tempo().ZoneName()
}

func (mutable *MutableTempo) MonthName() string {
	return mutable.Tempo().MonthName()
}

func (mutable *MutableTempo) ShortMonthName() string {
	return mutable.Tempo().ShortMonthName()
}

func (mutable *MutableTempo) DayName() string {
	return mutable.Tempo().DayName()
}

func (mutable *MutableTempo) ShortDayName() string {
	return mutable.Tempo().ShortDayName()
}

func (mutable *MutableTempo) MinDayName() string {
	return mutable.Tempo().MinDayName()
}

func (mutable *MutableTempo) TranslateNumber(value int) string {
	return mutable.Tempo().TranslateNumber(value)
}

func (mutable *MutableTempo) Translate(message string, replacements map[string]string) string {
	return mutable.Tempo().Translate(message, replacements)
}

func (mutable *MutableTempo) TranslateWith(message string, replacements map[string]string) string {
	return mutable.Tempo().TranslateWith(message, replacements)
}

func (mutable *MutableTempo) TranslationMessage(key string) (any, bool) {
	return mutable.Tempo().TranslationMessage(key)
}

func (mutable *MutableTempo) IsUTC() bool {
	return mutable.Tempo().IsUTC()
}

func (mutable *MutableTempo) IsLocal() bool {
	return mutable.Tempo().IsLocal()
}

func (mutable *MutableTempo) IsDST() bool {
	return mutable.Tempo().IsDST()
}

func (mutable *MutableTempo) IsLeapYear() bool {
	return mutable.Tempo().IsLeapYear()
}

func (mutable *MutableTempo) DaysInYear() int {
	return mutable.Tempo().DaysInYear()
}

func (mutable *MutableTempo) IsLongYear() bool {
	return mutable.Tempo().IsLongYear()
}

func (mutable *MutableTempo) IsLongISOYear() bool {
	return mutable.Tempo().IsLongISOYear()
}

func (mutable *MutableTempo) IsLastOfMonth() bool {
	return mutable.Tempo().IsLastOfMonth()
}

func (mutable *MutableTempo) DaysInMonth() int {
	return mutable.Tempo().DaysInMonth()
}

func (mutable *MutableTempo) IsWeekend() bool {
	return mutable.Tempo().IsWeekend()
}

func (mutable *MutableTempo) IsSunday() bool {
	return mutable.Tempo().IsSunday()
}

func (mutable *MutableTempo) IsMonday() bool {
	return mutable.Tempo().IsMonday()
}

func (mutable *MutableTempo) IsTuesday() bool {
	return mutable.Tempo().IsTuesday()
}

func (mutable *MutableTempo) IsWednesday() bool {
	return mutable.Tempo().IsWednesday()
}

func (mutable *MutableTempo) IsThursday() bool {
	return mutable.Tempo().IsThursday()
}

func (mutable *MutableTempo) IsFriday() bool {
	return mutable.Tempo().IsFriday()
}

func (mutable *MutableTempo) IsSaturday() bool {
	return mutable.Tempo().IsSaturday()
}

func (mutable *MutableTempo) IsDayOfWeek(weekday time.Weekday) bool {
	return mutable.Tempo().IsDayOfWeek(weekday)
}

func (mutable *MutableTempo) IsWeekday() bool {
	return mutable.Tempo().IsWeekday()
}

func (mutable *MutableTempo) IsPast(reference Tempo) bool {
	return mutable.Tempo().IsPast(reference)
}

func (mutable *MutableTempo) IsFuture(reference Tempo) bool {
	return mutable.Tempo().IsFuture(reference)
}

func (mutable *MutableTempo) IsNowOrPast() bool {
	return mutable.Tempo().IsNowOrPast()
}

func (mutable *MutableTempo) IsNowOrFuture() bool {
	return mutable.Tempo().IsNowOrFuture()
}

func (mutable *MutableTempo) IsToday(reference Tempo) bool {
	return mutable.Tempo().IsToday(reference)
}

func (mutable *MutableTempo) IsTomorrow(reference Tempo) bool {
	return mutable.Tempo().IsTomorrow(reference)
}

func (mutable *MutableTempo) IsYesterday(reference Tempo) bool {
	return mutable.Tempo().IsYesterday(reference)
}

func (mutable *MutableTempo) IsMidnight() bool {
	return mutable.Tempo().IsMidnight()
}

func (mutable *MutableTempo) IsMidday() bool {
	return mutable.Tempo().IsMidday()
}
