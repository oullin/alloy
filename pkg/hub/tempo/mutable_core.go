package tempo

import (
	"time"

	"github.com/oullin/alloy/pkg/hub/tempo/setters"
)

func (mutable *MutableTime) Immutable() Time {
	return newTempoWithPolicy(mutable.value, mutable.location, mutable.runtime, mutable.settingsSnapshot(), mutable.serializer, mutable.toStringFormat)
}

func (mutable *MutableTime) Clone() *MutableTime {
	return NewMutable(mutable.Immutable())
}

func (mutable *MutableTime) AvoidMutation() *MutableTime {
	return mutable.Clone()
}

func (mutable *MutableTime) Cast() *MutableTime {
	return mutable.Clone()
}

func (mutable *MutableTime) Tempoize(input Time) *MutableTime {
	return mutable.replace(mutable.Immutable().Tempoize(input))
}

func (mutable *MutableTime) NowWithSameTz() *MutableTime {
	return mutable.replace(mutable.Immutable().NowWithSameTz())
}

func (mutable *MutableTime) Modify(modifier string) (*MutableTime, error) {
	next, err := mutable.Immutable().Modify(modifier)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) Change(modifier string) (*MutableTime, error) {
	return mutable.Modify(modifier)
}

func (mutable *MutableTime) Context() Context {
	return mutable.Immutable().Context()
}

func (mutable *MutableTime) WithRuntime(runtime Context) *MutableTime {
	return mutable.replace(mutable.Immutable().WithRuntime(runtime))
}

func (mutable *MutableTime) WithTranslator(translator Translator) *MutableTime {
	return mutable.replace(mutable.Immutable().WithTranslator(translator))
}

func (mutable *MutableTime) HasTranslator() bool {
	return mutable.Immutable().HasTranslator()
}

func (mutable *MutableTime) Timezone() string {
	return mutable.Immutable().Timezone()
}

func (mutable *MutableTime) Timestamp() int64 {
	return mutable.Immutable().Timestamp()
}

func (mutable *MutableTime) TimestampMs() int64 {
	return mutable.Immutable().TimestampMs()
}

func (mutable *MutableTime) PreciseTimestamp(precisions ...int) float64 {
	return mutable.Immutable().PreciseTimestamp(precisions...)
}

func (mutable *MutableTime) Unix() int64 {
	return mutable.Immutable().Unix()
}

func (mutable *MutableTime) Year() int {
	return mutable.Immutable().Year()
}

func (mutable *MutableTime) Month() int {
	return mutable.Immutable().Month()
}

func (mutable *MutableTime) Quarter() int {
	return mutable.Immutable().Quarter()
}

func (mutable *MutableTime) Day() int {
	return mutable.Immutable().Day()
}

func (mutable *MutableTime) DayOfWeek() int {
	return mutable.Immutable().DayOfWeek()
}

func (mutable *MutableTime) Weekday() int {
	return mutable.Immutable().Weekday()
}

func (mutable *MutableTime) SetWeekday(weekday time.Weekday) *MutableTime {
	return setters.SetWeekday(mutable, weekday)
}

func (mutable *MutableTime) ISOWeekday() int {
	return mutable.Immutable().ISOWeekday()
}

func (mutable *MutableTime) ISOWeek() (int, int) {
	return mutable.Immutable().ISOWeek()
}

func (mutable *MutableTime) ISOWeekYear() int {
	return mutable.Immutable().ISOWeekYear()
}

func (mutable *MutableTime) ISOWeekNumber() int {
	return mutable.Immutable().ISOWeekNumber()
}

func (mutable *MutableTime) WeeksInISOYear() int {
	return mutable.Immutable().WeeksInISOYear()
}

func (mutable *MutableTime) DayOfYear() int {
	return mutable.Immutable().DayOfYear()
}

func (mutable *MutableTime) SetDayOfYear(day int) *MutableTime {
	return setters.SetDayOfYear(mutable, day)
}

func (mutable *MutableTime) Hour() int {
	return mutable.Immutable().Hour()
}

func (mutable *MutableTime) Minute() int {
	return mutable.Immutable().Minute()
}

func (mutable *MutableTime) Second() int {
	return mutable.Immutable().Second()
}

func (mutable *MutableTime) Millisecond() int {
	return mutable.Immutable().Millisecond()
}

func (mutable *MutableTime) PaddedUnit(field string, length int) (string, bool) {
	return mutable.Immutable().PaddedUnit(field, length)
}

func (mutable *MutableTime) OffsetMinutes() int {
	return mutable.Immutable().OffsetMinutes()
}

func (mutable *MutableTime) OffsetString(separator string) string {
	return mutable.Immutable().OffsetString(separator)
}

func (mutable *MutableTime) UTCOffset() int {
	return mutable.Immutable().UTCOffset()
}

func (mutable *MutableTime) ZoneName() string {
	return mutable.Immutable().ZoneName()
}

func (mutable *MutableTime) MonthName() string {
	return mutable.Immutable().MonthName()
}

func (mutable *MutableTime) ShortMonthName() string {
	return mutable.Immutable().ShortMonthName()
}

func (mutable *MutableTime) DayName() string {
	return mutable.Immutable().DayName()
}

func (mutable *MutableTime) ShortDayName() string {
	return mutable.Immutable().ShortDayName()
}

func (mutable *MutableTime) MinDayName() string {
	return mutable.Immutable().MinDayName()
}

func (mutable *MutableTime) TranslateNumber(value int) string {
	return mutable.Immutable().TranslateNumber(value)
}

func (mutable *MutableTime) Translate(message string, replacements map[string]string) string {
	return mutable.Immutable().Translate(message, replacements)
}

func (mutable *MutableTime) TranslateWith(message string, replacements map[string]string) string {
	return mutable.Immutable().TranslateWith(message, replacements)
}

func (mutable *MutableTime) TranslationMessage(key string) (any, bool) {
	return mutable.Immutable().TranslationMessage(key)
}

func (mutable *MutableTime) IsUTC() bool {
	return mutable.Immutable().IsUTC()
}

func (mutable *MutableTime) IsLocal() bool {
	return mutable.Immutable().IsLocal()
}

func (mutable *MutableTime) IsDST() bool {
	return mutable.Immutable().IsDST()
}

func (mutable *MutableTime) IsLeapYear() bool {
	return mutable.Immutable().IsLeapYear()
}

func (mutable *MutableTime) DaysInYear() int {
	return mutable.Immutable().DaysInYear()
}

func (mutable *MutableTime) IsLongYear() bool {
	return mutable.Immutable().IsLongYear()
}

func (mutable *MutableTime) IsLongISOYear() bool {
	return mutable.Immutable().IsLongISOYear()
}

func (mutable *MutableTime) IsLastOfMonth() bool {
	return mutable.Immutable().IsLastOfMonth()
}

func (mutable *MutableTime) DaysInMonth() int {
	return mutable.Immutable().DaysInMonth()
}

func (mutable *MutableTime) IsWeekend() bool {
	return mutable.Immutable().IsWeekend()
}

func (mutable *MutableTime) IsSunday() bool {
	return mutable.Immutable().IsSunday()
}

func (mutable *MutableTime) IsMonday() bool {
	return mutable.Immutable().IsMonday()
}

func (mutable *MutableTime) IsTuesday() bool {
	return mutable.Immutable().IsTuesday()
}

func (mutable *MutableTime) IsWednesday() bool {
	return mutable.Immutable().IsWednesday()
}

func (mutable *MutableTime) IsThursday() bool {
	return mutable.Immutable().IsThursday()
}

func (mutable *MutableTime) IsFriday() bool {
	return mutable.Immutable().IsFriday()
}

func (mutable *MutableTime) IsSaturday() bool {
	return mutable.Immutable().IsSaturday()
}

func (mutable *MutableTime) IsDayOfWeek(weekday time.Weekday) bool {
	return mutable.Immutable().IsDayOfWeek(weekday)
}

func (mutable *MutableTime) IsWeekday() bool {
	return mutable.Immutable().IsWeekday()
}

func (mutable *MutableTime) IsPast(reference Time) bool {
	return mutable.Immutable().IsPast(reference)
}

func (mutable *MutableTime) IsFuture(reference Time) bool {
	return mutable.Immutable().IsFuture(reference)
}

func (mutable *MutableTime) IsNowOrPast() bool {
	return mutable.Immutable().IsNowOrPast()
}

func (mutable *MutableTime) IsNowOrFuture() bool {
	return mutable.Immutable().IsNowOrFuture()
}

func (mutable *MutableTime) IsToday(reference Time) bool {
	return mutable.Immutable().IsToday(reference)
}

func (mutable *MutableTime) IsTomorrow(reference Time) bool {
	return mutable.Immutable().IsTomorrow(reference)
}

func (mutable *MutableTime) IsYesterday(reference Time) bool {
	return mutable.Immutable().IsYesterday(reference)
}

func (mutable *MutableTime) IsMidnight() bool {
	return mutable.Immutable().IsMidnight()
}

func (mutable *MutableTime) IsMidday() bool {
	return mutable.Immutable().IsMidday()
}
