package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/duration"
	"github.com/oullin/alloy/tempo/setters"
)

func (tempo Tempo) SetTimezone(name string) (Tempo, error) {
	location, err := loadLocation(name)

	if err != nil {
		return Tempo{}, err
	}

	return newTempoWithPolicy(tempo.value, location, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat), nil
}

func (tempo Tempo) SetTimezoneKeepLocal(name string) (Tempo, error) {
	location, err := loadLocation(name)

	if err != nil {
		return Tempo{}, err
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

	return newTempoWithPolicy(next, location, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat), nil
}

func (tempo Tempo) ShiftTimezone(name string) (Tempo, error) {
	return tempo.SetTimezoneKeepLocal(name)
}

func (tempo Tempo) UTC() Tempo {
	return newTempoWithPolicy(tempo.value, time.UTC, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat)
}

func (tempo Tempo) Local() Tempo {
	return newTempoWithPolicy(tempo.value, time.Local, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat)
}

func (tempo Tempo) fromObject(object Object, location *time.Location) (Tempo, error) {
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
		return Tempo{}, errInvalidComponents
	}

	return newTempoWithPolicy(next, location, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat), nil
}

func (tempo Tempo) Set(components Components) (Tempo, error) {
	object := tempo.ToObject()
	location := tempo.location

	if components.Timezone != "" {
		nextLocation, err := loadLocation(components.Timezone)

		if err != nil {
			return Tempo{}, err
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

func (tempo Tempo) SetUnit(unit Unit, value int) (Tempo, error) {
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
		return Tempo{}, errInvalidComponents
	}
}

func (tempo Tempo) SetYear(year int) (Tempo, error) {
	return tempo.SetDateTime(year, tempo.Month(), tempo.Day(), tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Tempo) SetMonth(month int) (Tempo, error) {
	return tempo.SetDateTime(tempo.Year(), month, tempo.Day(), tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Tempo) SetDay(day int) (Tempo, error) {
	return tempo.SetDateTime(tempo.Year(), tempo.Month(), day, tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Tempo) SetDate(year int, month int, day int) (Tempo, error) {
	return tempo.SetDateTime(year, month, day, tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Tempo) SetDateFrom(source Tempo) (Tempo, error) {
	return tempo.SetDate(source.Year(), source.Month(), source.Day())
}

func (tempo Tempo) SetDateTime(year int, month int, day int, hour int, minute int, second int, millisecond int) (Tempo, error) {
	components := Components{Year: year, Month: month, Day: day, Hour: hour, Minute: minute, Second: second, Millisecond: millisecond}
	value := timeFromComponents(components, tempo.location)

	if !componentsMatchTime(components, value, tempo.location) {
		return Tempo{}, errInvalidComponents
	}

	return newTempoWithPolicy(value, tempo.location, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat), nil
}

func (tempo Tempo) SetDateTimeFrom(source Tempo) (Tempo, error) {
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

func (tempo Tempo) SetHour(hour int) (Tempo, error) {
	return tempo.SetTime(hour, tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Tempo) SetMinute(minute int) (Tempo, error) {
	return tempo.SetTime(tempo.Hour(), minute, tempo.Second(), tempo.Millisecond())
}

func (tempo Tempo) SetSecond(second int) (Tempo, error) {
	return tempo.SetTime(tempo.Hour(), tempo.Minute(), second, tempo.Millisecond())
}

func (tempo Tempo) SetMillisecond(millisecond int) (Tempo, error) {
	return tempo.SetTime(tempo.Hour(), tempo.Minute(), tempo.Second(), millisecond)
}

func (tempo Tempo) SetTime(hour int, minute int, second int, millisecond int) (Tempo, error) {
	return tempo.SetDateTime(tempo.Year(), tempo.Month(), tempo.Day(), hour, minute, second, millisecond)
}

func (tempo Tempo) SetTimeFrom(source Tempo) (Tempo, error) {
	return tempo.SetTime(source.Hour(), source.Minute(), source.Second(), source.Millisecond())
}

func (tempo Tempo) SetTimeFromTimeString(input string) (Tempo, error) {
	parsed, err := newParserWithPolicy(tempo.location, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat).Parse(tempo.DateString() + "T" + input)

	if err != nil {
		return Tempo{}, err
	}

	return tempo.SetTimeFrom(parsed)
}

func (tempo Tempo) SetTimestamp(timestamp int64) Tempo {
	return setters.SetTimestamp(tempo, timestamp)
}

func (tempo Tempo) SetISODate(year int, week int, day int) Tempo {
	return setters.SetISODate(tempo, year, week, day)
}

func (tempo Tempo) SetISOWeek(week int, days ...int) Tempo {
	return setters.SetISOWeek(tempo, week, days...)
}

func (tempo Tempo) SetISOWeekYear(year int, days ...int) Tempo {
	return setters.SetISOWeekYear(tempo, year, days...)
}

func (tempo Tempo) SetISOWeekday(day int) Tempo {
	return setters.SetISOWeekday(tempo, day)
}

func (tempo Tempo) ISOWeeksInYear() int {
	return tempo.WeeksInISOYear()
}

func (tempo Tempo) SetUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) Tempo {
	return setters.SetUnitNoOverflow(tempo, valueUnit, value, overflowUnit)
}

func (tempo Tempo) Midday() Tempo {
	return setters.Midday(tempo, tempo.settingsSnapshot().MidDayAt)
}

func (tempo Tempo) MidDay() Tempo {
	return tempo.Midday()
}
