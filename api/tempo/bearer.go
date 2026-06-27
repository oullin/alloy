package tempo

import (
	"time"

	"alloy.dev/api/tempo/core"
)

// Compile-time guarantee that Time and *MutableTime both satisfy the
// generic core.Bearer contract feature packages target.
var (
	_ core.Bearer[Time]         = Time{}
	_ core.Bearer[*MutableTime] = (*MutableTime)(nil)
)

func (tempo Time) State() core.State {
	return core.State{Value: tempo.value, Location: tempo.location}
}

func (tempo Time) With(value time.Time) Time {
	return newTempoWithPolicy(value, tempo.location, tempo.Context(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat)
}

func (mutable *MutableTime) State() core.State {
	return core.State{Value: mutable.value, Location: mutable.location}
}

func (mutable *MutableTime) With(value time.Time) *MutableTime {
	return mutable.replace(newTempoWithPolicy(value, mutable.location, mutable.runtime, mutable.settingsSnapshot(), mutable.serializer, mutable.toStringFormat))
}
