package tempo

import (
	"strconv"
	"strings"
	"time"

	"alloy.dev/api/tempo/formatting"
)

func (tempo Time) Format(pattern string) string {
	return formatting.Format(tempo, pattern)
}

func (tempo Time) Ordinal(unit Unit) string {
	switch normalizeUnit(unit) {
	case Year:
		return formatting.Ordinal(tempo.Year())
	case Quarter:
		return formatting.Ordinal(tempo.Quarter())
	case Month:
		return formatting.Ordinal(tempo.Month())
	default:
		return formatting.Ordinal(tempo.Day())
	}
}

func (tempo Time) Meridiem(lowercase bool) string {
	value := "AM"

	if tempo.Hour() >= 12 {
		value = "PM"
	}

	if lowercase {
		return strings.ToLower(value)
	}

	return value
}

func (tempo Time) Week() int {
	return tempo.ISOWeekNumber()
}

func (tempo Time) WeekYear() int {
	return tempo.ISOWeekYear()
}

func (tempo Time) WeeksInYear() int {
	return tempo.WeeksInISOYear()
}

func (tempo Time) DaysFromStartOfWeek(weekStartsOn time.Weekday) int {
	return (int(tempo.local().Weekday()) - int(weekStartsOn) + 7) % 7
}

func (tempo Time) SetDaysFromStartOfWeek(days int, weekStartsOn time.Weekday) Time {
	return tempo.StartOfWeek(StartOfWeekOptions{WeekStartsOn: weekStartsOn}).AddDays(days)
}

func (tempo Time) DateString() string {
	return formatting.DateString(tempo)
}

func (tempo Time) TimeString(precision ...TimeStringPrecision) string {
	return formatting.TimeString(tempo, selectedPrecision(precision) == MillisecondPrecision)
}

func (tempo Time) DateTimeString() string {
	return formatting.DateTimeString(tempo)
}

func (tempo Time) FormattedDateString() string {
	return tempo.Format("MMM D, YYYY")
}

func (tempo Time) FormattedDayDateString() string {
	return tempo.Format("ddd, MMM D, YYYY")
}

func (tempo Time) DayDateTimeString() string {
	return tempo.Format("ddd, MMM D, YYYY h:mm A")
}

func (tempo Time) DateTimeLocalString(precision ...TimeStringPrecision) string {
	return tempo.DateString() + "T" + tempo.TimeString(precision...)
}

func (tempo Time) ISOString() string {
	return formatting.ISOString(tempo)
}

func (tempo Time) ISO8601String() string {
	return tempo.Format("YYYY-MM-DDTHH:mm:ssZ")
}

func (tempo Time) ISO8601ZuluString(precision ...TimeStringPrecision) string {
	return tempo.UTC().DateTimeLocalString(precision...) + "Z"
}

func (tempo Time) RFC3339String(precision ...TimeStringPrecision) string {
	return tempo.DateTimeLocalString(precision...) + tempo.OffsetString(":")
}

func (tempo Time) RFC7231String() string {
	return tempo.UTC().Format("ddd, DD MMM YYYY HH:mm:ss [GMT]")
}

func (tempo Time) RFC822String() string {
	return tempo.Format("ddd, DD MMM YY HH:mm:ss ZZ")
}

func (tempo Time) RFC850String() string {
	return tempo.Format("dddd, DD-MMM-YY HH:mm:ss ZZ")
}

func (tempo Time) RFC1036String() string {
	return tempo.RFC822String()
}

func (tempo Time) RFC1123String() string {
	return tempo.RSSString()
}

func (tempo Time) RFC2822String() string {
	return tempo.RSSString()
}

func (tempo Time) W3CString() string {
	return tempo.RFC3339String()
}

func (tempo Time) CookieString() string {
	return tempo.UTC().Format("ddd, DD-MMM-YYYY HH:mm:ss [GMT]")
}

func (tempo Time) AtomString() string {
	return tempo.RFC3339String()
}

func (tempo Time) RSSString() string {
	return tempo.Format("ddd, DD MMM YYYY HH:mm:ss ZZ")
}

func (tempo Time) UnixString() string {
	return strconv.FormatInt(tempo.Timestamp(), 10)
}

func (tempo Time) JSONSerialize() string {
	if tempo.serializer != nil {
		return tempo.serializer(tempo)
	}

	return tempo.ISOString()
}

func (tempo Time) Serialize() string {
	return tempo.JSONSerialize()
}

func (tempo Time) String() string {
	if tempo.toStringFormat != "" {
		return tempo.Format(tempo.toStringFormat)
	}

	return tempo.ISOString()
}

func (tempo Time) Time() time.Time {
	return tempo.value
}

func (tempo Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(tempo.JSONSerialize())), nil
}

func parseSerializedJSON(data []byte, location *time.Location) (time.Time, *time.Location, error) {
	input, err := strconv.Unquote(string(data))

	if err != nil {
		return time.Time{}, nil, err
	}

	if location == nil {
		location = defaultLocation()
	}

	parsed, err := parseInLocation(input, location)

	if err != nil {
		return time.Time{}, nil, err
	}

	return parsed.UTC(), location, nil
}

func (tempo Time) ToObject() Object {
	object := formatting.ToObject(tempo)

	return Object{
		Year:          object.Year,
		Month:         object.Month,
		Day:           object.Day,
		Hour:          object.Hour,
		Minute:        object.Minute,
		Second:        object.Second,
		Millisecond:   object.Millisecond,
		Timezone:      tempo.Timezone(),
		OffsetMinutes: object.OffsetMinutes,
		Weekday:       object.Weekday,
	}
}

func (tempo Time) ToMap() map[string]any {
	return formatting.ToMap(tempo)
}

func (tempo Time) ToArray() [7]int {
	return formatting.ToArray(tempo)
}
