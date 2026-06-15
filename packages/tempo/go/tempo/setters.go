package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/setters"
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
	return setters.SetUnit(tempo, unit, value)
}

func (tempo Tempo) SetYear(year int) Tempo {
	return setters.SetYear(tempo, year)
}

func (tempo Tempo) SetMonth(month int) Tempo {
	return setters.SetMonth(tempo, month)
}

func (tempo Tempo) SetDay(day int) Tempo {
	return setters.SetDay(tempo, day)
}

func (tempo Tempo) SetDate(year int, month int, day int) Tempo {
	return setters.SetDate(tempo, year, month, day)
}

func (tempo Tempo) SetDateFrom(source Tempo) Tempo {
	return tempo.SetDate(source.Year(), source.Month(), source.Day())
}

func (tempo Tempo) SetDateTime(year int, month int, day int, hour int, minute int, second int, millisecond int) Tempo {
	return setters.SetDateTime(tempo, year, month, day, hour, minute, second, millisecond)
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
	return setters.SetHour(tempo, hour)
}

func (tempo Tempo) SetMinute(minute int) Tempo {
	return setters.SetMinute(tempo, minute)
}

func (tempo Tempo) SetSecond(second int) Tempo {
	return setters.SetSecond(tempo, second)
}

func (tempo Tempo) SetMillisecond(millisecond int) Tempo {
	return setters.SetMillisecond(tempo, millisecond)
}

func (tempo Tempo) SetTime(hour int, minute int, second int, millisecond int) Tempo {
	return setters.SetTime(tempo, hour, minute, second, millisecond)
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

func (tempo Tempo) SetTimestampFrom(timestamp int64) Tempo {
	return tempo.SetTimestamp(timestamp)
}

func (tempo Tempo) SetUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) Tempo {
	return setters.SetUnitNoOverflow(tempo, valueUnit, value, overflowUnit)
}

func (tempo Tempo) Midday() Tempo {
	return setters.Midday(tempo, defaultConfig.Settings.MidDayAt)
}

func (tempo Tempo) MidDay() Tempo {
	return tempo.Midday()
}
