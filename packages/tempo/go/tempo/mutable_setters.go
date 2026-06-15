package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/setters"
)

func (mutable *MutableTempo) SetTimezone(name string) (*MutableTempo, error) {
	location, err := loadLocation(name)

	if err != nil {
		return nil, err
	}

	mutable.location = location

	return mutable, nil
}

func (mutable *MutableTempo) SetTimezoneKeepLocal(name string) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTimezoneKeepLocal(name)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) ShiftTimezone(name string) (*MutableTempo, error) {
	return mutable.SetTimezoneKeepLocal(name)
}

func (mutable *MutableTempo) UTC() *MutableTempo {
	mutable.location = time.UTC

	return mutable
}

func (mutable *MutableTempo) Local() *MutableTempo {
	mutable.location = time.Local

	return mutable
}

func (mutable *MutableTempo) Set(components Components) (*MutableTempo, error) {
	next, err := mutable.Tempo().Set(components)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetUnit(unit Unit, value int) (*MutableTempo, error) {
	return setters.SetUnit(mutable, unit, value)
}

func (mutable *MutableTempo) SetYear(year int) *MutableTempo {
	return setters.SetYear(mutable, year)
}

func (mutable *MutableTempo) SetMonth(month int) *MutableTempo {
	return setters.SetMonth(mutable, month)
}

func (mutable *MutableTempo) SetDay(day int) *MutableTempo {
	return setters.SetDay(mutable, day)
}

func (mutable *MutableTempo) SetDate(year int, month int, day int) *MutableTempo {
	return setters.SetDate(mutable, year, month, day)
}

func (mutable *MutableTempo) SetDateFrom(source Tempo) *MutableTempo {
	return mutable.SetDate(source.Year(), source.Month(), source.Day())
}

func (mutable *MutableTempo) SetDateTime(year int, month int, day int, hour int, minute int, second int, millisecond int) *MutableTempo {
	return setters.SetDateTime(mutable, year, month, day, hour, minute, second, millisecond)
}

func (mutable *MutableTempo) SetDateTimeFrom(source Tempo) *MutableTempo {
	return mutable.SetDateTime(
		source.Year(),
		source.Month(),
		source.Day(),
		source.Hour(),
		source.Minute(),
		source.Second(),
		source.Millisecond(),
	)
}

func (mutable *MutableTempo) SetHour(hour int) *MutableTempo {
	return setters.SetHour(mutable, hour)
}

func (mutable *MutableTempo) SetMinute(minute int) *MutableTempo {
	return setters.SetMinute(mutable, minute)
}

func (mutable *MutableTempo) SetSecond(second int) *MutableTempo {
	return setters.SetSecond(mutable, second)
}

func (mutable *MutableTempo) SetMillisecond(millisecond int) *MutableTempo {
	return setters.SetMillisecond(mutable, millisecond)
}

func (mutable *MutableTempo) SetTime(hour int, minute int, second int, millisecond int) *MutableTempo {
	return setters.SetTime(mutable, hour, minute, second, millisecond)
}

func (mutable *MutableTempo) SetTimeFrom(source Tempo) *MutableTempo {
	return mutable.SetTime(source.Hour(), source.Minute(), source.Second(), source.Millisecond())
}

func (mutable *MutableTempo) SetTimeFromTimeString(input string) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTimeFromTimeString(input)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetTimestamp(timestamp int64) *MutableTempo {
	return setters.SetTimestamp(mutable, timestamp)
}

func (mutable *MutableTempo) SetTimestampFrom(timestamp int64) *MutableTempo {
	return mutable.SetTimestamp(timestamp)
}

func (mutable *MutableTempo) SetISODate(year int, week int, day int) *MutableTempo {
	return setters.SetISODate(mutable, year, week, day)
}

func (mutable *MutableTempo) SetISOWeek(week int, days ...int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetISOWeek(week, days...))
}

func (mutable *MutableTempo) SetISOWeekYear(year int, days ...int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetISOWeekYear(year, days...))
}

func (mutable *MutableTempo) SetISOWeekday(day int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetISOWeekday(day))
}

func (mutable *MutableTempo) ISOWeeksInYear() int {
	return mutable.Tempo().ISOWeeksInYear()
}

func (mutable *MutableTempo) SetUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) *MutableTempo {
	return setters.SetUnitNoOverflow(mutable, valueUnit, value, overflowUnit)
}

func (mutable *MutableTempo) Midday() *MutableTempo {
	return setters.Midday(mutable, defaultConfig.Settings.MidDayAt)
}

func (mutable *MutableTempo) MidDay() *MutableTempo {
	return mutable.Midday()
}
