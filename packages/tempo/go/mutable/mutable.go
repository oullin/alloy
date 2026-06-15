package mutable

import tempopkg "github.com/oullin/alloy/tempo/tempo"

type MutableTempo struct {
	value *tempopkg.MutableTempo
}

func From(value tempopkg.Tempo) MutableTempo {
	return MutableTempo{value: tempopkg.NewMutable(value)}
}

func FromMutable(value *tempopkg.MutableTempo) MutableTempo {
	return MutableTempo{value: value}
}

func (mutable MutableTempo) Mutable() *tempopkg.MutableTempo {
	return mutable.value
}

func (mutable MutableTempo) Tempo() tempopkg.Tempo {
	return mutable.value.Tempo()
}

func (mutable MutableTempo) AddDays(days int) MutableTempo {
	return FromMutable(mutable.value.AddDays(days))
}

func (mutable MutableTempo) SubDays(days int) MutableTempo {
	return FromMutable(mutable.value.SubDays(days))
}

func (mutable MutableTempo) SetDate(year int, month int, day int) MutableTempo {
	return FromMutable(mutable.value.SetDate(year, month, day))
}

func (mutable MutableTempo) SetTime(hour int, minute int, second int, millisecond int) MutableTempo {
	return FromMutable(mutable.value.SetTime(hour, minute, second, millisecond))
}

func (mutable MutableTempo) SetTimezone(name string) (MutableTempo, error) {
	next, err := mutable.value.SetTimezone(name)

	if err != nil {
		return MutableTempo{}, err
	}

	return FromMutable(next), nil
}

func (mutable MutableTempo) ISOString() string {
	return mutable.value.ISOString()
}
