package tempo

import "time"

func (mutable *MutableTempo) DateString() string {
	return mutable.Tempo().DateString()
}

func (mutable *MutableTempo) TimeString(precision ...TimeStringPrecision) string {
	return mutable.Tempo().TimeString(precision...)
}

func (mutable *MutableTempo) DateTimeString() string {
	return mutable.Tempo().DateTimeString()
}

func (mutable *MutableTempo) FormattedDateString() string {
	return mutable.Tempo().FormattedDateString()
}

func (mutable *MutableTempo) FormattedDayDateString() string {
	return mutable.Tempo().FormattedDayDateString()
}

func (mutable *MutableTempo) DayDateTimeString() string {
	return mutable.Tempo().DayDateTimeString()
}

func (mutable *MutableTempo) DateTimeLocalString(precision ...TimeStringPrecision) string {
	return mutable.Tempo().DateTimeLocalString(precision...)
}

func (mutable *MutableTempo) ISOString() string {
	return mutable.Tempo().ISOString()
}

func (mutable *MutableTempo) ISO8601String() string {
	return mutable.Tempo().ISO8601String()
}

func (mutable *MutableTempo) ISO8601ZuluString(precision ...TimeStringPrecision) string {
	return mutable.Tempo().ISO8601ZuluString(precision...)
}

func (mutable *MutableTempo) RFC3339String(precision ...TimeStringPrecision) string {
	return mutable.Tempo().RFC3339String(precision...)
}

func (mutable *MutableTempo) RFC7231String() string {
	return mutable.Tempo().RFC7231String()
}

func (mutable *MutableTempo) RFC822String() string {
	return mutable.Tempo().RFC822String()
}

func (mutable *MutableTempo) RFC850String() string {
	return mutable.Tempo().RFC850String()
}

func (mutable *MutableTempo) RFC1036String() string {
	return mutable.Tempo().RFC1036String()
}

func (mutable *MutableTempo) RFC1123String() string {
	return mutable.Tempo().RFC1123String()
}

func (mutable *MutableTempo) RFC2822String() string {
	return mutable.Tempo().RFC2822String()
}

func (mutable *MutableTempo) W3CString() string {
	return mutable.Tempo().W3CString()
}

func (mutable *MutableTempo) CookieString() string {
	return mutable.Tempo().CookieString()
}

func (mutable *MutableTempo) AtomString() string {
	return mutable.Tempo().AtomString()
}

func (mutable *MutableTempo) RSSString() string {
	return mutable.Tempo().RSSString()
}

func (mutable *MutableTempo) UnixString() string {
	return mutable.Tempo().UnixString()
}

func (mutable *MutableTempo) JSONSerialize() string {
	return mutable.Tempo().JSONSerialize()
}

func (mutable *MutableTempo) Serialize() string {
	return mutable.Tempo().Serialize()
}

func (mutable *MutableTempo) String() string {
	return mutable.Tempo().String()
}

func (mutable *MutableTempo) Time() time.Time {
	return mutable.Tempo().Time()
}

func (mutable *MutableTempo) MarshalJSON() ([]byte, error) {
	return mutable.Tempo().MarshalJSON()
}

func (mutable *MutableTempo) UnmarshalJSON(data []byte) error {
	tempo := mutable.Tempo()
	if err := tempo.UnmarshalJSON(data); err != nil {
		return err
	}

	mutable.value = tempo.value
	mutable.location = tempo.location
	return nil
}

func (mutable *MutableTempo) ToObject() Object {
	return mutable.Tempo().ToObject()
}

func (mutable *MutableTempo) ToMap() map[string]interface{} {
	return mutable.Tempo().ToMap()
}

func (mutable *MutableTempo) ToArray() [7]int {
	return mutable.Tempo().ToArray()
}

func (mutable *MutableTempo) Format(pattern string) string {
	return mutable.Tempo().Format(pattern)
}

func (mutable *MutableTempo) RawFormat(pattern string) string {
	return mutable.Tempo().RawFormat(pattern)
}

func (mutable *MutableTempo) ISOFormat(pattern string) string {
	return mutable.Tempo().ISOFormat(pattern)
}

func (mutable *MutableTempo) TranslatedFormat(pattern string) string {
	return mutable.Tempo().TranslatedFormat(pattern)
}

func (mutable *MutableTempo) Ordinal(unit Unit) string {
	return mutable.Tempo().Ordinal(unit)
}

func (mutable *MutableTempo) Meridiem(lowercase bool) string {
	return mutable.Tempo().Meridiem(lowercase)
}

func (mutable *MutableTempo) Week() int {
	return mutable.Tempo().Week()
}

func (mutable *MutableTempo) WeekYear() int {
	return mutable.Tempo().WeekYear()
}

func (mutable *MutableTempo) WeeksInYear() int {
	return mutable.Tempo().WeeksInYear()
}

func (mutable *MutableTempo) GetDaysFromStartOfWeek(weekStartsOn time.Weekday) int {
	return mutable.Tempo().GetDaysFromStartOfWeek(weekStartsOn)
}

func (mutable *MutableTempo) SetDaysFromStartOfWeek(days int, weekStartsOn time.Weekday) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetDaysFromStartOfWeek(days, weekStartsOn))
}
