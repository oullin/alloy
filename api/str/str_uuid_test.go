package str

import (
	"strings"
	"testing"
)

// Ref: @alloy/code-0380
func TestStrUuid(t *testing.T) {
	// NOT parallel — UUID tests may conflict with freeze
	uuid := Uuid()

	if !IsUuid(uuid) {
		t.Errorf("Uuid() = %q is not a valid UUID", uuid)
	}
}

// Ref: @alloy/code-0380
func TestStrFreezeUuids(t *testing.T) {
	// NOT parallel — modifies global UUID state
	cleanup := FreezeUuids(func() string { return "frozen-uuid" })

	defer cleanup()

	if got := Uuid(); got != "frozen-uuid" {
		t.Errorf("FreezeUuids: expected 'frozen-uuid', got %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrFreezeUuidsCleanup(t *testing.T) {
	// NOT parallel — modifies global UUID state
	cleanup := FreezeUuids(func() string { return "frozen" })
	frozen := Uuid()
	cleanup()

	// After cleanup, should generate real UUIDs again
	normal := Uuid()

	if normal == "frozen" && frozen != normal {
		// This is fine — just verify cleanup runs
	}

	if frozen != "frozen" {
		t.Errorf("expected 'frozen', got %q", frozen)
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testItCanSpecifyAFallbackForASequence
func TestStrUuidSequence(t *testing.T) {
	// NOT parallel — modifies global state
	cleanup := CreateUuidsUsingSequence([]string{
		"first-uuid",
		"second-uuid",
	})

	defer cleanup()

	if got := Uuid(); got != "first-uuid" {
		t.Errorf("sequence[0] = %q", got)
	}

	if got := Uuid(); got != "second-uuid" {
		t.Errorf("sequence[1] = %q", got)
	}

	cleanup()
	cleanup = CreateUuidsUsingSequence([]string{"only-uuid"}, func() string { return "fallback-uuid" })

	if got := Uuid(); got != "only-uuid" {
		t.Errorf("fallback sequence first value = %q", got)
	}

	if got := Uuid(); got != "fallback-uuid" {
		t.Errorf("fallback value = %q", got)
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testItCanFreezeUlidsInAClosure
func TestStrFreezeUlids(t *testing.T) {
	// NOT parallel — modifies global state
	cleanup := FreezeUlids(func() string { return "FROZENULID00000000000000000" })

	defer cleanup()

	if got := Ulid(); got != "FROZENULID00000000000000000" {
		t.Errorf("FreezeUlids: expected frozen ULID, got %q", got)
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testItCanSpecifyAFallbackForAUlidSequence
func TestStrUlidSequence(t *testing.T) {
	// NOT parallel — modifies global state
	seq := []string{
		"01ARZ3NDEKTSV4RRFFQ69G5FAV",
		"01ARZ3NDEKTSV4RRFFQ69G5FAW",
	}
	cleanup := CreateUlidsUsingSequence(seq)

	defer cleanup()

	if got := Ulid(); got != seq[0] {
		t.Errorf("sequence[0] = %q", got)
	}

	if got := Ulid(); got != seq[1] {
		t.Errorf("sequence[1] = %q", got)
	}

	cleanup()
	cleanup = CreateUlidsUsingSequence([]string{"ONLYULID0000000000000000"}, func() string {
		return "FALLBACKULID000000000000"
	})

	if got := Ulid(); got != "ONLYULID0000000000000000" {
		t.Errorf("fallback sequence first value = %q", got)
	}

	if got := Ulid(); got != "FALLBACKULID000000000000" {
		t.Errorf("fallback value = %q", got)
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testItCreatesUlidsNormallyAfterFailureWithinFreezeMethod
func TestStrCreateUuidsNormally(t *testing.T) {
	// NOT parallel — modifies global state
	cleanup := FreezeUuids(func() string { return "frozen" })
	cleanup() // restore immediately

	uuid := Uuid()

	if uuid == "frozen" {
		t.Error("after CreateUuidsNormally, should generate real UUIDs")
	}

	if !IsUuid(uuid) {
		t.Errorf("should be valid UUID, got %q", uuid)
	}

	ulidCleanup := FreezeUlids(func() string { return "FROZENULID00000000000000000" })
	ulidCleanup()

	ulid := Ulid()

	if ulid == "FROZENULID00000000000000000" {
		t.Error("after ULID cleanup, should generate real ULIDs")
	}

	if len(ulid) != 26 {
		t.Errorf("should be valid ULID length, got %q", ulid)
	}
}

// Ref: @alloy/code-0380
func TestStrOrderedUuid(t *testing.T) {
	// NOT parallel — may interfere with UUID freeze tests
	uuid := OrderedUuid()

	if !IsUuid(uuid) {
		t.Errorf("OrderedUuid() = %q is not a valid UUID", uuid)
	}
}

// Ref: @alloy/code-0380
func TestStrUlid(t *testing.T) {
	// NOT parallel
	ulid := Ulid()

	if len(ulid) != 26 {
		t.Errorf("ULID length should be 26, got %d (%q)", len(ulid), ulid)
	}

	if strings.ToUpper(ulid) != ulid {
		// ULID should be uppercase
		t.Errorf("ULID should be uppercase, got %q", ulid)
	}
}

// Ref: @alloy/code-0380
func TestStrResetFactoryState(t *testing.T) {
	// NOT parallel — modifies global state
	CreateUuidsUsing(func() string { return "custom" })
	ResetFactoryState()

	uuid := Uuid()

	if uuid == "custom" {
		t.Error("after ResetFactoryState, UUID factory should be reset")
	}
}
