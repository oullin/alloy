package tempo

func (mutable *MutableTempo) SetTimezone(name string) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTimezone(name)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetTimezoneKeepLocal(name string) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTimezoneKeepLocal(name)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) ShiftTimezone(name string) (*MutableTempo, error) {
	next, err := mutable.Tempo().ShiftTimezone(name)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) UTC() *MutableTempo {
	return mutable.replace(mutable.Tempo().UTC())
}

func (mutable *MutableTempo) Local() *MutableTempo {
	return mutable.replace(mutable.Tempo().Local())
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

func (mutable *MutableTempo) SetYear(year int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetYear(year))
}

func (mutable *MutableTempo) SetMonth(month int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetMonth(month))
}

func (mutable *MutableTempo) SetDay(day int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetDay(day))
}

func (mutable *MutableTempo) SetDate(year int, month int, day int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetDate(year, month, day))
}

func (mutable *MutableTempo) SetDateFrom(source Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetDateFrom(source))
}

func (mutable *MutableTempo) SetDateTime(year int, month int, day int, hour int, minute int, second int, millisecond int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetDateTime(year, month, day, hour, minute, second, millisecond))
}

func (mutable *MutableTempo) SetDateTimeFrom(source Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetDateTimeFrom(source))
}

func (mutable *MutableTempo) SetHour(hour int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetHour(hour))
}

func (mutable *MutableTempo) SetMinute(minute int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetMinute(minute))
}

func (mutable *MutableTempo) SetSecond(second int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetSecond(second))
}

func (mutable *MutableTempo) SetMillisecond(millisecond int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetMillisecond(millisecond))
}

func (mutable *MutableTempo) SetTime(hour int, minute int, second int, millisecond int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetTime(hour, minute, second, millisecond))
}

func (mutable *MutableTempo) SetTimeFrom(source Tempo) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetTimeFrom(source))
}

func (mutable *MutableTempo) SetTimeFromTimeString(input string) (*MutableTempo, error) {
	next, err := mutable.Tempo().SetTimeFromTimeString(input)
	if err != nil {
		return nil, err
	}

	return mutable.replace(next), nil
}

func (mutable *MutableTempo) SetTimestamp(timestamp int64) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetTimestamp(timestamp))
}

func (mutable *MutableTempo) SetTimestampFrom(timestamp int64) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetTimestampFrom(timestamp))
}

func (mutable *MutableTempo) SetISODate(year int, week int, day int) *MutableTempo {
	return mutable.replace(mutable.Tempo().SetISODate(year, week, day))
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
	return mutable.replace(mutable.Tempo().SetUnitNoOverflow(valueUnit, value, overflowUnit))
}

func (mutable *MutableTempo) Midday() *MutableTempo {
	return mutable.replace(mutable.Tempo().Midday())
}

func (mutable *MutableTempo) MidDay() *MutableTempo {
	return mutable.replace(mutable.Tempo().MidDay())
}
