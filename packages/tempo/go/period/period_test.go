package period_test

import (
	"testing"

	"github.com/oullin/alloy/tempo/period"
)

func TestForwardBoundsContains(t *testing.T) {
	bounds := period.Bounds{StartMs: 100, EndMs: 200, IncludeEnd: true}

	if !bounds.Forward() {
		t.Fatalf("Forward() = false, want true")
	}

	if !bounds.Contains(100) {
		t.Fatalf("Contains(start) = false, want true")
	}

	if !bounds.Contains(150) {
		t.Fatalf("Contains(inside) = false, want true")
	}

	if !bounds.Contains(200) {
		t.Fatalf("Contains(end, IncludeEnd=true) = false, want true")
	}

	if bounds.Contains(201) {
		t.Fatalf("Contains(past end) = true, want false")
	}

	excludeEnd := period.Bounds{StartMs: 100, EndMs: 200, IncludeEnd: false}

	if excludeEnd.Contains(200) {
		t.Fatalf("Contains(end, IncludeEnd=false) = true, want false")
	}

	if !excludeEnd.Contains(199) {
		t.Fatalf("Contains(just before end) = false, want true")
	}
}

func TestReverseBoundsContains(t *testing.T) {
	bounds := period.Bounds{StartMs: 200, EndMs: 100, IncludeEnd: true}

	if bounds.Forward() {
		t.Fatalf("Forward() = true, want false")
	}

	if !bounds.Contains(150) {
		t.Fatalf("Contains(inside reverse) = false, want true")
	}

	if !bounds.Contains(100) {
		t.Fatalf("Contains(end reverse, IncludeEnd=true) = false, want true")
	}

	if !bounds.Contains(200) {
		t.Fatalf("Contains(start reverse) = false, want true")
	}

	if bounds.Contains(99) {
		t.Fatalf("Contains(past reverse end) = true, want false")
	}

	excludeEnd := period.Bounds{StartMs: 200, EndMs: 100, IncludeEnd: false}

	if excludeEnd.Contains(100) {
		t.Fatalf("Contains(end reverse, IncludeEnd=false) = true, want false")
	}

	if !excludeEnd.Contains(101) {
		t.Fatalf("Contains(just past reverse end) = false, want true")
	}
}
