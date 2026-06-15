package comparison

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

func (tempo Tempo) Before(other tempopkg.Tempo, units ...tempopkg.Unit) bool {
	return tempo.value.Before(other, units...)
}

func (tempo Tempo) After(other tempopkg.Tempo, units ...tempopkg.Unit) bool {
	return tempo.value.After(other, units...)
}

func (tempo Tempo) Same(other tempopkg.Tempo, units ...tempopkg.Unit) bool {
	return tempo.value.Same(other, units...)
}

func (tempo Tempo) Between(start tempopkg.Tempo, end tempopkg.Tempo, inclusivity ...string) bool {
	return tempo.value.Between(start, end, inclusivity...)
}

func (tempo Tempo) Clamp(minimum tempopkg.Tempo, maximum tempopkg.Tempo) (Tempo, error) {
	next, err := tempo.value.Clamp(minimum, maximum)

	if err != nil {
		return Tempo{}, err
	}

	return From(next), nil
}

func (tempo Tempo) Average(other tempopkg.Tempo) Tempo {
	return From(tempo.value.Average(other))
}

func (tempo Tempo) Closest(first tempopkg.Tempo, rest ...tempopkg.Tempo) Tempo {
	return From(tempo.value.Closest(first, rest...))
}

func (tempo Tempo) Farthest(first tempopkg.Tempo, rest ...tempopkg.Tempo) Tempo {
	return From(tempo.value.Farthest(first, rest...))
}

func (tempo Tempo) Min(other tempopkg.Tempo) Tempo {
	return From(tempo.value.Min(other))
}

func (tempo Tempo) Max(other tempopkg.Tempo) Tempo {
	return From(tempo.value.Max(other))
}
