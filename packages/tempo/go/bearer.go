package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/core"
)

// Compile-time guarantee that Tempo and *MutableTempo both satisfy the
// generic core.Bearer contract feature packages target.
var (
	_ core.Bearer[Tempo]         = Tempo{}
	_ core.Bearer[*MutableTempo] = (*MutableTempo)(nil)
)

func (tempo Tempo) State() core.State {
	return core.State{Value: tempo.value, Location: tempo.location}
}

func (tempo Tempo) With(value time.Time) Tempo {
	return newTempoWithPolicy(value, tempo.location, tempo.Runtime(), tempo.settingsSnapshot(), tempo.serializer, tempo.toStringFormat)
}

func (mutable *MutableTempo) State() core.State {
	return core.State{Value: mutable.value, Location: mutable.location}
}

func (mutable *MutableTempo) With(value time.Time) *MutableTempo {
	return mutable.replace(newTempoWithPolicy(value, mutable.location, mutable.runtime, mutable.settingsSnapshot(), mutable.serializer, mutable.toStringFormat))
}
