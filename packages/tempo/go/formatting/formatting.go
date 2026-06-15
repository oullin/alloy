package formatting

import (
	"time"

	tempopkg "github.com/oullin/alloy/tempo/tempo"
)

type Tempo struct {
	value tempopkg.Tempo
}

func From(value tempopkg.Tempo) Tempo {
	return Tempo{value: value}
}

func (tempo Tempo) Tempo() tempopkg.Tempo {
	return tempo.value
}

func (tempo Tempo) Format(pattern string) string {
	return tempo.value.Format(pattern)
}

func (tempo Tempo) DateString() string {
	return tempo.value.DateString()
}

func (tempo Tempo) TimeString(precision ...tempopkg.TimeStringPrecision) string {
	return tempo.value.TimeString(precision...)
}

func (tempo Tempo) DateTimeString() string {
	return tempo.value.DateTimeString()
}

func (tempo Tempo) ISOString() string {
	return tempo.value.ISOString()
}

func (tempo Tempo) ISO8601String() string {
	return tempo.value.ISO8601String()
}

func (tempo Tempo) RFC3339String(precision ...tempopkg.TimeStringPrecision) string {
	return tempo.value.RFC3339String(precision...)
}

func (tempo Tempo) Serialize() string {
	return tempo.value.Serialize()
}

func (tempo Tempo) String() string {
	return tempo.value.String()
}

func (tempo Tempo) Time() time.Time {
	return tempo.value.Time()
}

func (tempo Tempo) ToObject() tempopkg.Object {
	return tempo.value.ToObject()
}

func (tempo Tempo) ToMap() map[string]interface{} {
	return tempo.value.ToMap()
}

func (tempo Tempo) ToArray() [7]int {
	return tempo.value.ToArray()
}
