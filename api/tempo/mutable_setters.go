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
	next, err := mutable.Tempo().SetUnit(unit, value)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetYear(year int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetYear(year)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetMonth(month int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetMonth(month)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetDay(day int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetDay(day)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetDate(year int, month int, day int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetDate(year, month, day)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetDateFrom(source Tempo) (*MutableTempo, error) {
	return mutable.SetDate(source.Year(), source.Month(), source.Day())
}

func (mutable *MutableTempo) SetDateTime(year int, month int, day int, hour int, minute int, second int, millisecond int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetDateTime(year, month, day, hour, minute, second, millisecond)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetDateTimeFrom(source Tempo) (*MutableTempo, error) {
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

func (mutable *MutableTempo) SetHour(hour int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetHour(hour)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetMinute(minute int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetMinute(minute)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetSecond(second int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetSecond(second)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetMillisecond(millisecond int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetMillisecond(millisecond)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetTime(hour int, minute int, second int, millisecond int) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTime(hour, minute, second, millisecond)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetTimeFrom(source Tempo) (*MutableTempo, error) {
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

func (mutable *MutableTempo) SetISODate(year int, week int, day int) *MutableTempo {
	return setters.SetISODate(mutable, year, week, day)
}

func (mutable *MutableTempo) SetISOWeek(week int, days ...int) *MutableTempo {
	return setters.SetISOWeek(mutable, week, days...)
}

func (mutable *MutableTempo) SetISOWeekYear(year int, days ...int) *MutableTempo {
	return setters.SetISOWeekYear(mutable, year, days...)
}

func (mutable *MutableTempo) SetISOWeekday(day int) *MutableTempo {
	return setters.SetISOWeekday(mutable, day)
}

func (mutable *MutableTempo) ISOWeeksInYear() int {
	return mutable.Tempo().ISOWeeksInYear()
}

func (mutable *MutableTempo) SetUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) *MutableTempo {
	return setters.SetUnitNoOverflow(mutable, valueUnit, value, overflowUnit)
}

func (mutable *MutableTempo) Midday() *MutableTempo {
	return setters.Midday(mutable, mutable.settingsSnapshot().MidDayAt)
}

func (mutable *MutableTempo) MidDay() *MutableTempo {
	return mutable.Midday()
}
