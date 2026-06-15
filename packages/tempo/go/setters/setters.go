package setters

import tempopkg "github.com/oullin/alloy/tempo/tempo"

type Tempo struct {
	value tempopkg.Tempo
}

func From(value tempopkg.Tempo) Tempo {
	return Tempo{value: value}
}

func (tempo Tempo) Tempo() tempopkg.Tempo {
	return tempo.value
}

func (tempo Tempo) SetTimezone(name string) (Tempo, error) {
	next, err := tempo.value.SetTimezone(name)

	if err != nil {
		return Tempo{}, err
	}

	return From(next), nil
}

func (tempo Tempo) SetTimezoneKeepLocal(name string) (Tempo, error) {
	next, err := tempo.value.SetTimezoneKeepLocal(name)

	if err != nil {
		return Tempo{}, err
	}

	return From(next), nil
}

func (tempo Tempo) Set(components tempopkg.Components) (Tempo, error) {
	next, err := tempo.value.Set(components)

	if err != nil {
		return Tempo{}, err
	}

	return From(next), nil
}

func (tempo Tempo) SetYear(year int) Tempo {
	return From(tempo.value.SetYear(year))
}

func (tempo Tempo) SetMonth(month int) Tempo {
	return From(tempo.value.SetMonth(month))
}

func (tempo Tempo) SetDay(day int) Tempo {
	return From(tempo.value.SetDay(day))
}

func (tempo Tempo) SetDate(year int, month int, day int) Tempo {
	return From(tempo.value.SetDate(year, month, day))
}

func (tempo Tempo) SetTime(hour int, minute int, second int, millisecond int) Tempo {
	return From(tempo.value.SetTime(hour, minute, second, millisecond))
}

func (tempo Tempo) SetTimestamp(timestamp int64) Tempo {
	return From(tempo.value.SetTimestamp(timestamp))
}

func (tempo Tempo) Midday() Tempo {
	return From(tempo.value.Midday())
}
