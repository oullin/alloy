package tempo

import (
	"time"

	"github.com/oullin/alloy/api/tempo/duration"
	"github.com/oullin/alloy/api/tempo/setters"
)

func (tempo Time) SetTimezone(name string) (Time, error) {
	location, err := loadLocation(name)

	if err != nil {
		return Time{}, err
	}

	return newTempoWithPolicy(tempo.value, location, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat), nil
}

func (tempo Time) SetTimezoneKeepLocal(name string) (Time, error) {
	location, err := loadLocation(name)

	if err != nil {
		return Time{}, err
	}

	object := tempo.ToObject()
	next := time.Date(
		object.Year,
		time.Month(object.Month),
		object.Day,
		object.Hour,
		object.Minute,
		object.Second,
		object.Millisecond*int(time.Millisecond),
		location,
	)

	return newTempoWithPolicy(next, location, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat), nil
}

func (tempo Time) ShiftTimezone(name string) (Time, error) {
	return tempo.SetTimezoneKeepLocal(name)
}

func (tempo Time) UTC() Time {
	return newTempoWithPolicy(tempo.value, time.UTC, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat)
}

func (tempo Time) Local() Time {
	return newTempoWithPolicy(tempo.value, time.Local, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat)
}

func (tempo Time) fromObject(object Object, location *time.Location) (Time, error) {
	components := Components{
		Year:        object.Year,
		Month:       object.Month,
		Day:         object.Day,
		Hour:        object.Hour,
		Minute:      object.Minute,
		Second:      object.Second,
		Millisecond: object.Millisecond,
	}
	next := time.Date(
		object.Year,
		time.Month(object.Month),
		object.Day,
		object.Hour,
		object.Minute,
		object.Second,
		object.Millisecond*int(time.Millisecond),
		location,
	)

	if !componentsMatchTime(components, next, location) {
		return Time{}, errInvalidComponents
	}

	return newTempoWithPolicy(next, location, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat), nil
}

func (tempo Time) Set(components Components) (Time, error) {
	object := tempo.ToObject()
	location := tempo.location

	if components.Timezone != "" {
		nextLocation, err := loadLocation(components.Timezone)

		if err != nil {
			return Time{}, err
		}

		location = nextLocation
	}

	if components.Year != 0 {
		object.Year = components.Year
	}

	if components.Month != 0 {
		object.Month = components.Month
	}

	if components.Day != 0 {
		object.Day = components.Day
	}

	if components.Hour != 0 {
		object.Hour = components.Hour
	}

	if components.Minute != 0 {
		object.Minute = components.Minute
	}

	if components.Second != 0 {
		object.Second = components.Second
	}

	if components.Millisecond != 0 {
		object.Millisecond = components.Millisecond
	}

	return tempo.fromObject(object, location)
}

func (tempo Time) SetUnit(unit Unit, value int) (Time, error) {
	switch duration.NormalizeUnit(unit) {
	case duration.Year:
		return tempo.SetYear(value)
	case duration.Month:
		return tempo.SetMonth(value)
	case duration.Day:
		return tempo.SetDay(value)
	case duration.Hour:
		return tempo.SetHour(value)
	case duration.Minute:
		return tempo.SetMinute(value)
	case duration.Second:
		return tempo.SetSecond(value)
	case duration.Millisecond:
		return tempo.SetMillisecond(value)
	default:
		return Time{}, errInvalidComponents
	}
}

func (tempo Time) SetYear(year int) (Time, error) {
	return tempo.SetDateTime(year, tempo.Month(), tempo.Day(), tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Time) SetMonth(month int) (Time, error) {
	return tempo.SetDateTime(tempo.Year(), month, tempo.Day(), tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Time) SetDay(day int) (Time, error) {
	return tempo.SetDateTime(tempo.Year(), tempo.Month(), day, tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Time) SetDate(year int, month int, day int) (Time, error) {
	return tempo.SetDateTime(year, month, day, tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Time) SetDateFrom(source Time) (Time, error) {
	return tempo.SetDate(source.Year(), source.Month(), source.Day())
}

func (tempo Time) SetDateTime(year int, month int, day int, hour int, minute int, second int, millisecond int) (Time, error) {
	components := Components{Year: year, Month: month, Day: day, Hour: hour, Minute: minute, Second: second, Millisecond: millisecond}
	value := timeFromComponents(components, tempo.location)

	if !componentsMatchTime(components, value, tempo.location) {
		return Time{}, errInvalidComponents
	}

	return newTempoWithPolicy(value, tempo.location, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat), nil
}

func (tempo Time) SetDateTimeFrom(source Time) (Time, error) {
	return tempo.SetDateTime(
		source.Year(),
		source.Month(),
		source.Day(),
		source.Hour(),
		source.Minute(),
		source.Second(),
		source.Millisecond(),
	)
}

func (tempo Time) SetHour(hour int) (Time, error) {
	return tempo.SetTime(hour, tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Time) SetMinute(minute int) (Time, error) {
	return tempo.SetTime(tempo.Hour(), minute, tempo.Second(), tempo.Millisecond())
}

func (tempo Time) SetSecond(second int) (Time, error) {
	return tempo.SetTime(tempo.Hour(), tempo.Minute(), second, tempo.Millisecond())
}

func (tempo Time) SetMillisecond(millisecond int) (Time, error) {
	return tempo.SetTime(tempo.Hour(), tempo.Minute(), tempo.Second(), millisecond)
}

func (tempo Time) SetTime(hour int, minute int, second int, millisecond int) (Time, error) {
	return tempo.SetDateTime(tempo.Year(), tempo.Month(), tempo.Day(), hour, minute, second, millisecond)
}

func (tempo Time) SetTimeFrom(source Time) (Time, error) {
	return tempo.SetTime(source.Hour(), source.Minute(), source.Second(), source.Millisecond())
}

func (tempo Time) SetTimeFromTimeString(input string) (Time, error) {
	parsed, err := newParserWithPolicy(tempo.location, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat).Parse(tempo.DateString() + "T" + input)

	if err != nil {
		return Time{}, err
	}

	return tempo.SetTimeFrom(parsed)
}

func (tempo Time) SetTimestamp(timestamp int64) Time {
	return setters.SetTimestamp(tempo, timestamp)
}

func (tempo Time) SetISODate(year int, week int, day int) Time {
	return setters.SetISODate(tempo, year, week, day)
}

func (tempo Time) SetISOWeek(week int, days ...int) Time {
	return setters.SetISOWeek(tempo, week, days...)
}

func (tempo Time) SetISOWeekYear(year int, days ...int) Time {
	return setters.SetISOWeekYear(tempo, year, days...)
}

func (tempo Time) SetISOWeekday(day int) Time {
	return setters.SetISOWeekday(tempo, day)
}

func (tempo Time) ISOWeeksInYear() int {
	return tempo.WeeksInISOYear()
}

func (tempo Time) SetUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) Time {
	return setters.SetUnitNoOverflow(tempo, valueUnit, value, overflowUnit)
}

func (tempo Time) Midday() Time {
	return setters.Midday(tempo, tempo.settingsSnapshot().MidDayAt)
}

func (tempo Time) MidDay() Time {
	return tempo.Midday()
}
