package comparison_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/tempo"
	"github.com/oullin/alloy/pkg/hub/tempo/comparison"
	"github.com/oullin/alloy/pkg/hub/tempo/duration"
)

const (
	maxTestInt64 = int64(^uint64(0) >> 1)
	minTestInt64 = -maxTestInt64 - 1
)

func mustParse(t *testing.T, value string) tempo.Time {
	t.Helper()

	parsed, err := tempo.Parse(value)

	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}

	return parsed
}

func mustTimestampMs(t *testing.T, value int64) tempo.Time {
	t.Helper()

	parsed, err := tempo.FromTimestampMs(value)

	if err != nil {
		t.Fatalf("timestamp %d: %v", value, err)
	}

	return parsed
}

func TestBeforeAfterSameWithUnits(t *testing.T) {
	base := mustParse(t, "2024-05-15T10:00:00Z")
	later := mustParse(t, "2024-05-15T23:59:00Z")
	earlier := mustParse(t, "2024-05-14T10:00:00Z")

	if !comparison.Before(base, later.State()) {
		t.Fatalf("Before(later) = false, want true")
	}

	if !comparison.After(base, earlier.State()) {
		t.Fatalf("After(earlier) = false, want true")
	}

	if comparison.Same(base, later.State()) {
		t.Fatalf("Same(later) = true, want false (millisecond unit)")
	}

	if !comparison.Same(base, later.State(), duration.Day) {
		t.Fatalf("Same(later, Day) = false, want true")
	}

	if !comparison.SameOrBefore(base, base.State()) {
		t.Fatalf("SameOrBefore(self) = false, want true")
	}

	if !comparison.SameOrAfter(base, base.State()) {
		t.Fatalf("SameOrAfter(self) = false, want true")
	}
}

func TestBetweenInclusivityModes(t *testing.T) {
	base := mustParse(t, "2024-05-15T00:00:00Z")
	start := mustParse(t, "2024-05-15T00:00:00Z")
	end := mustParse(t, "2024-05-20T00:00:00Z")

	if !comparison.Between(base, start.State(), end.State()) {
		t.Fatalf("Between(default []) = false, want true")
	}

	if comparison.Between(base, start.State(), end.State(), "()") {
		t.Fatalf("Between(()) on boundary = true, want false")
	}

	if !comparison.Between(base, start.State(), end.State(), "[)") {
		t.Fatalf("Between([)) on start boundary = false, want true")
	}

	if comparison.Between(base, end.State(), end.State(), "(]") {
		t.Fatalf("Between((]) on collapsed range with base != end = unexpected true")
	}

	if !comparison.Between(base, end.State(), start.State()) {
		t.Fatalf("Between(reversed range) = false, want true (auto-normalized)")
	}
}

func TestClampMinMaxClosestFarthest(t *testing.T) {
	base := mustParse(t, "2024-05-15T12:00:00Z")
	minimum := mustParse(t, "2024-05-10T00:00:00Z")
	maximum := mustParse(t, "2024-05-20T00:00:00Z")
	before := mustParse(t, "2024-05-01T00:00:00Z")
	after := mustParse(t, "2024-05-30T00:00:00Z")

	if got, err := comparison.Clamp(before, minimum.State(), maximum.State()); err != nil || got.ISOString() != "2024-05-10T00:00:00.000Z" {
		t.Fatalf("Clamp(before) = %q, %v", got.ISOString(), err)
	}

	if got, err := comparison.Clamp(base, minimum.State(), maximum.State()); err != nil || got.ISOString() != "2024-05-15T12:00:00.000Z" {
		t.Fatalf("Clamp(base) = %q, %v", got.ISOString(), err)
	}

	if got, err := comparison.Clamp(after, minimum.State(), maximum.State()); err != nil || got.ISOString() != "2024-05-20T00:00:00.000Z" {
		t.Fatalf("Clamp(after) = %q, %v", got.ISOString(), err)
	}

	if _, err := comparison.Clamp(base, maximum.State(), minimum.State()); err == nil {
		t.Fatalf("Clamp(min > max) error = nil, want error")
	}

	if got := comparison.Min(base, minimum.State()).ISOString(); got != "2024-05-10T00:00:00.000Z" {
		t.Fatalf("Min() = %q, want minimum", got)
	}

	if got := comparison.Max(base, maximum.State()).ISOString(); got != "2024-05-20T00:00:00.000Z" {
		t.Fatalf("Max() = %q, want maximum", got)
	}

	near := mustParse(t, "2024-05-16T00:00:00Z")
	far := mustParse(t, "2024-05-22T00:00:00Z")

	if got := comparison.Closest(base, minimum.State(), near.State(), far.State()).ISOString(); got != "2024-05-16T00:00:00.000Z" {
		t.Fatalf("Closest() = %q, want near", got)
	}

	if got := comparison.Farthest(base, minimum.State(), near.State(), far.State()).ISOString(); got != "2024-05-22T00:00:00.000Z" {
		t.Fatalf("Farthest() = %q, want far", got)
	}
}

func TestAveragePicksMidpoint(t *testing.T) {
	left := mustParse(t, "2024-05-15T00:00:00Z")
	right := mustParse(t, "2024-05-17T00:00:00Z")

	if got := comparison.Average(left, right.State()).ISOString(); got != "2024-05-16T00:00:00.000Z" {
		t.Fatalf("Average() = %q, want midpoint", got)
	}
}

func TestSelectionHandlesExtremeMillisecondDistances(t *testing.T) {
	base := mustTimestampMs(t, 0)
	minimum := mustTimestampMs(t, minTestInt64)
	maximum := mustTimestampMs(t, maxTestInt64)
	near := mustTimestampMs(t, 1)

	if got := comparison.Closest(base, minimum.State(), maximum.State(), near.State()).TimestampMs(); got != 1 {
		t.Fatalf("Closest(extremes).TimestampMs() = %d, want 1", got)
	}

	if got := comparison.Farthest(base, near.State(), minimum.State(), maximum.State()).TimestampMs(); got != minTestInt64 {
		t.Fatalf("Farthest(extremes).TimestampMs() = %d, want min int64", got)
	}

	if got := comparison.Average(maximum, maximum.State()).TimestampMs(); got != maxTestInt64 {
		t.Fatalf("Average(max,max).TimestampMs() = %d, want max int64", got)
	}
}
