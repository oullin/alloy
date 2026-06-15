package tempo

import (
	"fmt"
	"time"
)

func (tempo Tempo) SetTimezone(name string) (Tempo, error) {
	location, err := loadLocation(name)
	if err != nil {
		return Tempo{}, err
	}

	return newTempo(tempo.value, location, tempo.Runtime()), nil
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

	return newTempo(next, location, tempo.Runtime()), nil
}

func (tempo Tempo) ShiftTimezone(name string) (Tempo, error) {
	return tempo.SetTimezoneKeepLocal(name)
}

func (tempo Tempo) UTC() Tempo {
	return newTempo(tempo.value, time.UTC, tempo.Runtime())
}

func (tempo Tempo) Local() Tempo {
	return newTempo(tempo.value, time.Local, tempo.Runtime())
}

func (tempo Tempo) fromObject(object Object, location *time.Location) Tempo {
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

	return newTempo(next, location, tempo.Runtime())
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

	return tempo.fromObject(object, location), nil
}

func (tempo Tempo) SetUnit(unit Unit, value int) (Tempo, error) {
	switch normalizeUnit(unit) {
	case Year:
		return tempo.SetYear(value), nil
	case Month:
		return tempo.SetMonth(value), nil
	case Day:
		return tempo.SetDay(value), nil
	case Hour:
		return tempo.SetHour(value), nil
	case Minute:
		return tempo.SetMinute(value), nil
	case Second:
		return tempo.SetSecond(value), nil
	case Millisecond:
		return tempo.SetMillisecond(value), nil
	default:
		return Tempo{}, fmt.Errorf("tempo cannot set unit: %s", unit)
	}
}

func (tempo Tempo) SetYear(year int) Tempo {
	object := tempo.ToObject()
	object.Year = year

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetMonth(month int) Tempo {
	object := tempo.ToObject()
	object.Month = month

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetDay(day int) Tempo {
	object := tempo.ToObject()
	object.Day = day

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetDate(year int, month int, day int) Tempo {
	object := tempo.ToObject()
	object.Year = year
	object.Month = month
	object.Day = day

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetDateFrom(source Tempo) Tempo {
	return tempo.SetDate(source.Year(), source.Month(), source.Day())
}

func (tempo Tempo) SetDateTime(year int, month int, day int, hour int, minute int, second int, millisecond int) Tempo {
	object := tempo.ToObject()
	object.Year = year
	object.Month = month
	object.Day = day
	object.Hour = hour
	object.Minute = minute
	object.Second = second
	object.Millisecond = millisecond

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetDateTimeFrom(source Tempo) Tempo {
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

func (tempo Tempo) SetHour(hour int) Tempo {
	object := tempo.ToObject()
	object.Hour = hour

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetMinute(minute int) Tempo {
	object := tempo.ToObject()
	object.Minute = minute

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetSecond(second int) Tempo {
	object := tempo.ToObject()
	object.Second = second

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetMillisecond(millisecond int) Tempo {
	object := tempo.ToObject()
	object.Millisecond = millisecond

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetTime(hour int, minute int, second int, millisecond int) Tempo {
	object := tempo.ToObject()
	object.Hour = hour
	object.Minute = minute
	object.Second = second
	object.Millisecond = millisecond

	return tempo.fromObject(object, tempo.location)
}

func (tempo Tempo) SetTimeFrom(source Tempo) Tempo {
	return tempo.SetTime(source.Hour(), source.Minute(), source.Second(), source.Millisecond())
}

func (tempo Tempo) SetTimeFromTimeString(input string) (Tempo, error) {
	parsed, err := Parse(tempo.DateString()+"T"+input, WithTimezone(tempo.Timezone()))
	if err != nil {
		return Tempo{}, err
	}

	return tempo.SetTimeFrom(parsed), nil
}

func (tempo Tempo) SetTimestamp(timestamp int64) Tempo {
	return newTempo(time.Unix(timestamp, 0), tempo.location, tempo.Runtime())
}

func (tempo Tempo) SetISODate(year int, week int, day int) Tempo {
	isoYearStart := Tempo{
		value:    time.Date(year, time.January, 4, 0, 0, 0, 0, tempo.location).UTC(),
		location: tempo.location,
		runtime:  tempo.Runtime(),
	}.StartOfWeek(StartOfWeekOptions{WeekStartsOn: time.Monday})

	return isoYearStart.
		AddWeeks(week-1).
		AddDays(day-1).
		SetTime(tempo.Hour(), tempo.Minute(), tempo.Second(), tempo.Millisecond())
}

func (tempo Tempo) SetISOWeek(week int, days ...int) Tempo {
	day := tempo.ISOWeekday()
	if len(days) > 0 {
		day = days[0]
	}

	return tempo.SetISODate(tempo.ISOWeekYear(), week, day)
}

func (tempo Tempo) SetISOWeekYear(year int, days ...int) Tempo {
	day := tempo.ISOWeekday()
	if len(days) > 0 {
		day = days[0]
	}

	return tempo.SetISODate(year, tempo.ISOWeekNumber(), day)
}

func (tempo Tempo) SetISOWeekday(day int) Tempo {
	if day == 0 {
		day = 7
	}

	return tempo.SetISODate(tempo.ISOWeekYear(), tempo.ISOWeekNumber(), day)
}

func (tempo Tempo) ISOWeeksInYear() int {
	return tempo.WeeksInISOYear()
}

func (tempo Tempo) SetTimestampFrom(timestamp int64) Tempo {
	return tempo.SetTimestamp(timestamp)
}

func (tempo Tempo) SetUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) Tempo {
	next, err := tempo.SetUnit(valueUnit, value)
	if err != nil {
		return tempo
	}

	clamped, err := next.Clamp(tempo.StartOf(overflowUnit), tempo.EndOf(overflowUnit))
	if err != nil {
		return tempo
	}

	return clamped
}

func (tempo Tempo) Midday() Tempo {
	return tempo.SetTime(defaultConfig.Settings.MidDayAt, 0, 0, 0)
}

func (tempo Tempo) MidDay() Tempo {
	return tempo.Midday()
}
