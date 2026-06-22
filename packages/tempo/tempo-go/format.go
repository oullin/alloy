package tempo

import (
	"strconv"
	"strings"
	"time"

	"github.com/oullin/alloy/tempo/formatting"
)

func (tempo Tempo) Format(pattern string) string {
	return formatting.Format(tempo, pattern)
}

func (tempo Tempo) Ordinal(unit Unit) string {
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

func (tempo Tempo) Meridiem(lowercase bool) string {
	value := "AM"

	if tempo.Hour() >= 12 {
		value = "PM"
	}

	if lowercase {
		return strings.ToLower(value)
	}

	return value
}

func (tempo Tempo) Week() int {
	return tempo.ISOWeekNumber()
}

func (tempo Tempo) WeekYear() int {
	return tempo.ISOWeekYear()
}

func (tempo Tempo) WeeksInYear() int {
	return tempo.WeeksInISOYear()
}

func (tempo Tempo) DaysFromStartOfWeek(weekStartsOn time.Weekday) int {
	return (int(tempo.local().Weekday()) - int(weekStartsOn) + 7) % 7
}

func (tempo Tempo) SetDaysFromStartOfWeek(days int, weekStartsOn time.Weekday) Tempo {
	return tempo.StartOfWeek(StartOfWeekOptions{WeekStartsOn: weekStartsOn}).AddDays(days)
}

func (tempo Tempo) DateString() string {
	return formatting.DateString(tempo)
}

func (tempo Tempo) TimeString(precision ...TimeStringPrecision) string {
	return formatting.TimeString(tempo, selectedPrecision(precision) == MillisecondPrecision)
}

func (tempo Tempo) DateTimeString() string {
	return formatting.DateTimeString(tempo)
}

func (tempo Tempo) FormattedDateString() string {
	return tempo.Format("MMM D, YYYY")
}

func (tempo Tempo) FormattedDayDateString() string {
	return tempo.Format("ddd, MMM D, YYYY")
}

func (tempo Tempo) DayDateTimeString() string {
	return tempo.Format("ddd, MMM D, YYYY h:mm A")
}

func (tempo Tempo) DateTimeLocalString(precision ...TimeStringPrecision) string {
	return tempo.DateString() + "T" + tempo.TimeString(precision...)
}

func (tempo Tempo) ISOString() string {
	return formatting.ISOString(tempo)
}

func (tempo Tempo) ISO8601String() string {
	return tempo.Format("YYYY-MM-DDTHH:mm:ssZ")
}

func (tempo Tempo) ISO8601ZuluString(precision ...TimeStringPrecision) string {
	return tempo.UTC().DateTimeLocalString(precision...) + "Z"
}

func (tempo Tempo) RFC3339String(precision ...TimeStringPrecision) string {
	return tempo.DateTimeLocalString(precision...) + tempo.OffsetString(":")
}

func (tempo Tempo) RFC7231String() string {
	return tempo.UTC().Format("ddd, DD MMM YYYY HH:mm:ss [GMT]")
}

func (tempo Tempo) RFC822String() string {
	return tempo.Format("ddd, DD MMM YY HH:mm:ss ZZ")
}

func (tempo Tempo) RFC850String() string {
	return tempo.Format("dddd, DD-MMM-YY HH:mm:ss ZZ")
}

func (tempo Tempo) RFC1036String() string {
	return tempo.RFC822String()
}

func (tempo Tempo) RFC1123String() string {
	return tempo.RSSString()
}

func (tempo Tempo) RFC2822String() string {
	return tempo.RSSString()
}

func (tempo Tempo) W3CString() string {
	return tempo.RFC3339String()
}

func (tempo Tempo) CookieString() string {
	return tempo.UTC().Format("ddd, DD-MMM-YYYY HH:mm:ss [GMT]")
}

func (tempo Tempo) AtomString() string {
	return tempo.RFC3339String()
}

func (tempo Tempo) RSSString() string {
	return tempo.Format("ddd, DD MMM YYYY HH:mm:ss ZZ")
}

func (tempo Tempo) UnixString() string {
	return strconv.FormatInt(tempo.Timestamp(), 10)
}

func (tempo Tempo) JSONSerialize() string {
	if tempo.serializer != nil {
		return tempo.serializer(tempo)
	}

	return tempo.ISOString()
}

func (tempo Tempo) Serialize() string {
	return tempo.JSONSerialize()
}

func (tempo Tempo) String() string {
	if tempo.toStringFormat != "" {
		return tempo.Format(tempo.toStringFormat)
	}

	return tempo.ISOString()
}

func (tempo Tempo) Time() time.Time {
	return tempo.value
}

func (tempo Tempo) MarshalJSON() ([]byte, error) {
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

func (tempo Tempo) ToObject() Object {
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

func (tempo Tempo) ToMap() map[string]any {
	return formatting.ToMap(tempo)
}

func (tempo Tempo) ToArray() [7]int {
	return formatting.ToArray(tempo)
}
