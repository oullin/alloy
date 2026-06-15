package tempo

import (
	"strconv"
	"strings"
	"time"
)

func (tempo Tempo) Format(pattern string) string {
	local := tempo.local()
	offset := tempo.OffsetMinutes()
	hour12 := local.Hour() % 12
	if hour12 == 0 {
		hour12 = 12
	}

	values := map[string]string{
		"A":    ternary(local.Hour() < 12, "AM", "PM"),
		"a":    ternary(local.Hour() < 12, "am", "pm"),
		"D":    strconv.Itoa(local.Day()),
		"DD":   pad(local.Day(), 2),
		"Do":   ordinal(local.Day()),
		"d":    strconv.Itoa(int(local.Weekday())),
		"ddd":  local.Weekday().String()[:3],
		"dddd": local.Weekday().String(),
		"H":    strconv.Itoa(local.Hour()),
		"HH":   pad(local.Hour(), 2),
		"h":    strconv.Itoa(hour12),
		"hh":   pad(hour12, 2),
		"M":    strconv.Itoa(int(local.Month())),
		"MM":   pad(int(local.Month()), 2),
		"MMM":  local.Month().String()[:3],
		"MMMM": local.Month().String(),
		"m":    strconv.Itoa(local.Minute()),
		"mm":   pad(local.Minute(), 2),
		"S":    strconv.Itoa(tempo.Millisecond() / 100),
		"SSS":  pad(tempo.Millisecond(), 3),
		"s":    strconv.Itoa(local.Second()),
		"ss":   pad(local.Second(), 2),
		"X":    strconv.FormatInt(tempo.Timestamp(), 10),
		"x":    strconv.FormatInt(tempo.TimestampMs(), 10),
		"Y":    strconv.Itoa(local.Year()),
		"YY":   pad(local.Year()%100, 2),
		"YYYY": pad(local.Year(), 4),
		"Z":    formatOffset(offset, ":"),
		"ZZ":   formatOffset(offset, ""),
	}

	tokens := []string{"YYYY", "MMMM", "dddd", "MMM", "ddd", "SSS", "Do", "YY", "ZZ", "MM", "DD", "HH", "hh", "mm", "ss", "Z", "X", "x", "Y", "M", "D", "H", "h", "m", "s", "A", "a", "d"}
	var builder strings.Builder

	for index := 0; index < len(pattern); {
		if pattern[index] == '[' {
			end := strings.IndexByte(pattern[index:], ']')
			if end >= 0 {
				builder.WriteString(pattern[index+1 : index+end])
				index += end + 1
				continue
			}
		}

		matched := false
		for _, token := range tokens {
			if strings.HasPrefix(pattern[index:], token) {
				builder.WriteString(values[token])
				index += len(token)
				matched = true
				break
			}
		}

		if !matched {
			builder.WriteByte(pattern[index])
			index++
		}
	}

	return builder.String()
}

func (tempo Tempo) RawFormat(pattern string) string {
	return tempo.Format(pattern)
}

func (tempo Tempo) ISOFormat(pattern string) string {
	return tempo.Format(pattern)
}

func (tempo Tempo) TranslatedFormat(pattern string) string {
	return tempo.Format(pattern)
}

func (tempo Tempo) Ordinal(unit Unit) string {
	switch normalizeUnit(unit) {
	case Year:
		return ordinal(tempo.Year())
	case Quarter:
		return ordinal(tempo.Quarter())
	case Month:
		return ordinal(tempo.Month())
	default:
		return ordinal(tempo.Day())
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

func (tempo Tempo) GetDaysFromStartOfWeek(weekStartsOn time.Weekday) int {
	return (int(tempo.local().Weekday()) - int(weekStartsOn) + 7) % 7
}

func (tempo Tempo) SetDaysFromStartOfWeek(days int, weekStartsOn time.Weekday) Tempo {
	return tempo.StartOfWeek(StartOfWeekOptions{WeekStartsOn: weekStartsOn}).AddDays(days)
}

func (tempo Tempo) DateString() string {
	return tempo.Format("YYYY-MM-DD")
}

func (tempo Tempo) TimeString(precision ...TimeStringPrecision) string {
	base := tempo.Format("HH:mm:ss")
	if selectedPrecision(precision) == MillisecondPrecision {
		return base + "." + pad(tempo.Millisecond(), 3)
	}

	return base
}

func (tempo Tempo) DateTimeString() string {
	return tempo.Format("YYYY-MM-DD HH:mm:ss")
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
	return tempo.value.UTC().Format("2006-01-02T15:04:05.000Z")
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
	if defaultConfig.Serializer != nil {
		return defaultConfig.Serializer(tempo)
	}

	return tempo.ISOString()
}

func (tempo Tempo) Serialize() string {
	return tempo.JSONSerialize()
}

func (tempo Tempo) String() string {
	if defaultConfig.ToStringFormat != "" {
		return tempo.Format(defaultConfig.ToStringFormat)
	}

	return tempo.ISOString()
}

func (tempo Tempo) Time() time.Time {
	return tempo.value
}

func (tempo Tempo) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(tempo.JSONSerialize())), nil
}

func (tempo *Tempo) UnmarshalJSON(data []byte) error {
	input, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	location := tempo.location
	if location == nil {
		location = defaultLocation()
	}

	parsed, err := parseInLocation(input, location)
	if err != nil {
		return err
	}

	tempo.value = parsed.UTC()
	tempo.location = location
	return nil
}

func (tempo Tempo) ToObject() Object {
	local := tempo.local()

	return Object{
		Year:          local.Year(),
		Month:         int(local.Month()),
		Day:           local.Day(),
		Hour:          local.Hour(),
		Minute:        local.Minute(),
		Second:        local.Second(),
		Millisecond:   local.Nanosecond() / int(time.Millisecond),
		Timezone:      tempo.Timezone(),
		OffsetMinutes: tempo.OffsetMinutes(),
		Weekday:       int(local.Weekday()),
	}
}

func (tempo Tempo) ToMap() map[string]interface{} {
	object := tempo.ToObject()

	return map[string]interface{}{
		"year":          object.Year,
		"month":         object.Month,
		"day":           object.Day,
		"hour":          object.Hour,
		"minute":        object.Minute,
		"second":        object.Second,
		"millisecond":   object.Millisecond,
		"timeZone":      object.Timezone,
		"offsetMinutes": object.OffsetMinutes,
		"weekday":       object.Weekday,
	}
}

func (tempo Tempo) ToArray() [7]int {
	object := tempo.ToObject()
	return [7]int{
		object.Year,
		object.Month,
		object.Day,
		object.Hour,
		object.Minute,
		object.Second,
		object.Millisecond,
	}
}
