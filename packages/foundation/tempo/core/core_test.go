package core_test

import (
	"testing"
	"time"

	"github.com/oullin/alloy/packages/foundation/tempo"
	"github.com/oullin/alloy/packages/foundation/tempo/core"
)

// Compile-time guarantees that Time and *MutableTime continue to satisfy
// the Bearer contract feature packages depend on.
var (
	_ core.Bearer[tempo.Time]         = tempo.Time{}
	_ core.Bearer[*tempo.MutableTime] = (*tempo.MutableTime)(nil)
)

func TestStateRoundTripsThroughBearer(t *testing.T) {
	parsed, err := tempo.Parse("2024-05-15T10:34:45.600Z")

	if err != nil {
		t.Fatalf("parse base: %v", err)
	}

	state := parsed.State()

	if state.Value.IsZero() {
		t.Fatalf("State().Value is zero, want populated instant")
	}

	if state.Location == nil {
		t.Fatalf("State().Location = nil, want timezone")
	}

	replacement := state.Value.Add(time.Hour)
	bumped := parsed.With(replacement)

	if got := bumped.State().Value; !got.Equal(replacement) {
		t.Fatalf("With().State() = %v, want %v", got, replacement)
	}

	if !parsed.State().Value.Equal(state.Value) {
		t.Fatalf("With() mutated the source bearer: state drifted to %v", parsed.State().Value)
	}
}

func TestMutableBearerUpdatesInPlace(t *testing.T) {
	mutable, err := tempo.ParseMutable("2024-05-15T10:34:45.600Z")

	if err != nil {
		t.Fatalf("parse mutable: %v", err)
	}

	original := mutable.State().Value
	replacement := original.Add(time.Hour)
	returned := mutable.With(replacement)

	if returned != mutable {
		t.Fatalf("Mutable With() returned a different pointer, want same identity")
	}

	if !mutable.State().Value.Equal(replacement) {
		t.Fatalf("Mutable State() = %v, want %v after With", mutable.State().Value, replacement)
	}
}
