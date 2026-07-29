package callbacks_test

import (
	"testing"

	"hara.sh/alloy/container/internal/callbacks"
)

func TestSnapshotEmptyRegistryReturnsNil(t *testing.T) {
	t.Parallel()

	r := callbacks.NewRegistry[func()]()

	global, specific := r.Snapshot("anything")

	// Both must be nil, not empty-but-non-nil: the container ranges over these
	// directly and relies on the nil case being a no-op.
	if global != nil {
		t.Fatalf("global = %v, want nil", global)
	}

	if specific != nil {
		t.Fatalf("specific = %v, want nil", specific)
	}
}

func TestSnapshotSeparatesGlobalFromKeyed(t *testing.T) {
	t.Parallel()

	r := callbacks.NewRegistry[string]()
	r.AddGlobal("g1")
	r.AddGlobal("g2")
	r.Add("db", "k1")
	r.Add("cache", "other")

	global, specific := r.Snapshot("db")

	if len(global) != 2 || global[0] != "g1" || global[1] != "g2" {
		t.Fatalf("global = %v, want [g1 g2]", global)
	}

	if len(specific) != 1 || specific[0] != "k1" {
		t.Fatalf("specific = %v, want [k1]", specific)
	}
}

func TestSnapshotUnknownKeyStillReturnsGlobals(t *testing.T) {
	t.Parallel()

	r := callbacks.NewRegistry[string]()
	r.AddGlobal("g1")

	global, specific := r.Snapshot("never-registered")

	if len(global) != 1 {
		t.Fatalf("global = %v, want [g1]", global)
	}

	if specific != nil {
		t.Fatalf("specific = %v, want nil", specific)
	}
}

// Snapshot must copy. The container fires callbacks after releasing the lock,
// so a snapshot aliasing internal state could be mutated mid-fire.
func TestSnapshotCopies(t *testing.T) {
	t.Parallel()

	r := callbacks.NewRegistry[string]()
	r.AddGlobal("g1")
	r.Add("db", "k1")

	global, specific := r.Snapshot("db")

	global[0] = "mutated"
	specific[0] = "mutated"

	global2, specific2 := r.Snapshot("db")

	if global2[0] != "g1" {
		t.Fatalf("mutating a snapshot leaked into the registry: %v", global2)
	}

	if specific2[0] != "k1" {
		t.Fatalf("mutating a snapshot leaked into the registry: %v", specific2)
	}
}

// Appending after a snapshot must not be visible to the already-taken copy.
func TestSnapshotIsPointInTime(t *testing.T) {
	t.Parallel()

	r := callbacks.NewRegistry[string]()
	r.Add("db", "k1")

	_, before := r.Snapshot("db")

	r.Add("db", "k2")

	if len(before) != 1 {
		t.Fatalf("earlier snapshot grew to %v", before)
	}

	_, after := r.Snapshot("db")

	if len(after) != 2 {
		t.Fatalf("later snapshot = %v, want both callbacks", after)
	}
}

func TestAddPreservesOrder(t *testing.T) {
	t.Parallel()

	r := callbacks.NewRegistry[string]()

	for _, v := range []string{"a", "b", "c"} {
		r.Add("db", v)
	}

	_, specific := r.Snapshot("db")

	for i, want := range []string{"a", "b", "c"} {
		if specific[i] != want {
			t.Fatalf("specific = %v, want [a b c]", specific)
		}
	}
}

func TestReset(t *testing.T) {
	t.Parallel()

	r := callbacks.NewRegistry[string]()
	r.AddGlobal("g1")
	r.Add("db", "k1")

	r.Reset()

	global, specific := r.Snapshot("db")

	if global != nil || specific != nil {
		t.Fatalf("Reset left state behind: global=%v specific=%v", global, specific)
	}

	// Reset must leave the registry usable, not just empty. A nil byKey map
	// would panic on the next Add.
	r.Add("db", "k2")
	r.AddGlobal("g2")

	global, specific = r.Snapshot("db")

	if len(global) != 1 || len(specific) != 1 {
		t.Fatalf("registry unusable after Reset: global=%v specific=%v", global, specific)
	}
}
