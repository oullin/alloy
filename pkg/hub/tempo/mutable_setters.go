package tempo

import (
	"time"

	"hara.sh/alloy/tempo/setters"
)

func (mutable *MutableTime) SetTimezone(name string) (*MutableTime, error) {
	location, err := loadLocation(name)

	if err != nil {
		return nil, err
	}

	mutable.location = location

	return mutable, nil
}

func (mutable *MutableTime) SetTimezoneKeepLocal(name string) (*MutableTime, error) {
	next, err := mutable.Immutable().SetTimezoneKeepLocal(name)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) ShiftTimezone(name string) (*MutableTime, error) {
	return mutable.SetTimezoneKeepLocal(name)
}

func (mutable *MutableTime) UTC() *MutableTime {
	mutable.location = time.UTC

	return mutable
}

func (mutable *MutableTime) Local() *MutableTime {
	mutable.location = time.Local

	return mutable
}

func (mutable *MutableTime) Set(components Components) (*MutableTime, error) {
	next, err := mutable.Immutable().Set(components)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetUnit(unit Unit, value int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetUnit(unit, value)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetYear(year int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetYear(year)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetMonth(month int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetMonth(month)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetDay(day int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetDay(day)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetDate(year int, month int, day int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetDate(year, month, day)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetDateFrom(source Time) (*MutableTime, error) {
	return mutable.SetDate(source.Year(), source.Month(), source.Day())
}

func (mutable *MutableTime) SetDateTime(year int, month int, day int, hour int, minute int, second int, millisecond int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetDateTime(year, month, day, hour, minute, second, millisecond)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetDateTimeFrom(source Time) (*MutableTime, error) {
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

func (mutable *MutableTime) SetHour(hour int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetHour(hour)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetMinute(minute int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetMinute(minute)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetSecond(second int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetSecond(second)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetMillisecond(millisecond int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetMillisecond(millisecond)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetTime(hour int, minute int, second int, millisecond int) (*MutableTime, error) {
	next, err := mutable.Immutable().SetTime(hour, minute, second, millisecond)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetTimeFrom(source Time) (*MutableTime, error) {
	return mutable.SetTime(source.Hour(), source.Minute(), source.Second(), source.Millisecond())
}

func (mutable *MutableTime) SetTimeFromTimeString(input string) (*MutableTime, error) {
	next, err := mutable.Immutable().SetTimeFromTimeString(input)

	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTime) SetTimestamp(timestamp int64) *MutableTime {
	return setters.SetTimestamp(mutable, timestamp)
}

func (mutable *MutableTime) SetISODate(year int, week int, day int) *MutableTime {
	return setters.SetISODate(mutable, year, week, day)
}

func (mutable *MutableTime) SetISOWeek(week int, days ...int) *MutableTime {
	return setters.SetISOWeek(mutable, week, days...)
}

func (mutable *MutableTime) SetISOWeekYear(year int, days ...int) *MutableTime {
	return setters.SetISOWeekYear(mutable, year, days...)
}

func (mutable *MutableTime) SetISOWeekday(day int) *MutableTime {
	return setters.SetISOWeekday(mutable, day)
}

func (mutable *MutableTime) ISOWeeksInYear() int {
	return mutable.Immutable().ISOWeeksInYear()
}

func (mutable *MutableTime) SetUnitNoOverflow(valueUnit Unit, value int, overflowUnit Unit) *MutableTime {
	return setters.SetUnitNoOverflow(mutable, valueUnit, value, overflowUnit)
}

func (mutable *MutableTime) Midday() *MutableTime {
	return setters.Midday(mutable, mutable.settingsSnapshot().MidDayAt)
}

func (mutable *MutableTime) MidDay() *MutableTime {
	return mutable.Midday()
}
