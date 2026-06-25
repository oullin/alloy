package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/formatting"
)

func (mutable *MutableTime) DateString() string {
	return formatting.DateString(mutable)
}

func (mutable *MutableTime) TimeString(precision ...TimeStringPrecision) string {
	return formatting.TimeString(mutable, selectedPrecision(precision) == MillisecondPrecision)
}

func (mutable *MutableTime) DateTimeString() string {
	return formatting.DateTimeString(mutable)
}

func (mutable *MutableTime) FormattedDateString() string {
	return mutable.Immutable().FormattedDateString()
}

func (mutable *MutableTime) FormattedDayDateString() string {
	return mutable.Immutable().FormattedDayDateString()
}

func (mutable *MutableTime) DayDateTimeString() string {
	return mutable.Immutable().DayDateTimeString()
}

func (mutable *MutableTime) DateTimeLocalString(precision ...TimeStringPrecision) string {
	return mutable.Immutable().DateTimeLocalString(precision...)
}

func (mutable *MutableTime) ISOString() string {
	return formatting.ISOString(mutable)
}

func (mutable *MutableTime) ISO8601String() string {
	return mutable.Immutable().ISO8601String()
}

func (mutable *MutableTime) ISO8601ZuluString(precision ...TimeStringPrecision) string {
	return mutable.Immutable().ISO8601ZuluString(precision...)
}

func (mutable *MutableTime) RFC3339String(precision ...TimeStringPrecision) string {
	return mutable.Immutable().RFC3339String(precision...)
}

func (mutable *MutableTime) RFC7231String() string {
	return mutable.Immutable().RFC7231String()
}

func (mutable *MutableTime) RFC822String() string {
	return mutable.Immutable().RFC822String()
}

func (mutable *MutableTime) RFC850String() string {
	return mutable.Immutable().RFC850String()
}

func (mutable *MutableTime) RFC1036String() string {
	return mutable.Immutable().RFC1036String()
}

func (mutable *MutableTime) RFC1123String() string {
	return mutable.Immutable().RFC1123String()
}

func (mutable *MutableTime) RFC2822String() string {
	return mutable.Immutable().RFC2822String()
}

func (mutable *MutableTime) W3CString() string {
	return mutable.Immutable().W3CString()
}

func (mutable *MutableTime) CookieString() string {
	return mutable.Immutable().CookieString()
}

func (mutable *MutableTime) AtomString() string {
	return mutable.Immutable().AtomString()
}

func (mutable *MutableTime) RSSString() string {
	return mutable.Immutable().RSSString()
}

func (mutable *MutableTime) UnixString() string {
	return mutable.Immutable().UnixString()
}

func (mutable *MutableTime) JSONSerialize() string {
	return mutable.Immutable().JSONSerialize()
}

func (mutable *MutableTime) Serialize() string {
	return mutable.Immutable().Serialize()
}

func (mutable *MutableTime) String() string {
	return mutable.Immutable().String()
}

func (mutable *MutableTime) Time() time.Time {
	return mutable.value
}

func (mutable *MutableTime) MarshalJSON() ([]byte, error) {
	return mutable.Immutable().MarshalJSON()
}

func (mutable *MutableTime) UnmarshalJSON(data []byte) error {
	value, location, err := parseSerializedJSON(data, mutable.location)

	if err != nil {
		return err
	}

	mutable.value = value
	mutable.location = location

	return nil
}

func (mutable *MutableTime) ToObject() Object {
	return mutable.Immutable().ToObject()
}

func (mutable *MutableTime) ToMap() map[string]any {
	return formatting.ToMap(mutable)
}

func (mutable *MutableTime) ToArray() [7]int {
	return formatting.ToArray(mutable)
}

func (mutable *MutableTime) Format(pattern string) string {
	return formatting.Format(mutable, pattern)
}

func (mutable *MutableTime) Ordinal(unit Unit) string {
	return mutable.Immutable().Ordinal(unit)
}

func (mutable *MutableTime) Meridiem(lowercase bool) string {
	return mutable.Immutable().Meridiem(lowercase)
}

func (mutable *MutableTime) Week() int {
	return mutable.Immutable().Week()
}

func (mutable *MutableTime) WeekYear() int {
	return mutable.Immutable().WeekYear()
}

func (mutable *MutableTime) WeeksInYear() int {
	return mutable.Immutable().WeeksInYear()
}

func (mutable *MutableTime) DaysFromStartOfWeek(weekStartsOn time.Weekday) int {
	return mutable.Immutable().DaysFromStartOfWeek(weekStartsOn)
}

func (mutable *MutableTime) SetDaysFromStartOfWeek(days int, weekStartsOn time.Weekday) *MutableTime {
	return mutable.replace(mutable.Immutable().SetDaysFromStartOfWeek(days, weekStartsOn))
}
