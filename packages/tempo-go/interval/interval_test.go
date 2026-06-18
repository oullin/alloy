package interval_test

import (
	"testing"
	"time"

	"github.com/oullin/alloy/tempo/interval"
)

const (
	maxTestInt64 = int64(^uint64(0) >> 1)
	minTestInt64 = -maxTestInt64 - 1
)

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339Nano, value)

	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}

	return parsed.UTC()
}

func TestSpanDurationAccessors(t *testing.T) {
	span := interval.New(
		mustTime(t, "2024-01-01T00:00:00Z"),
		mustTime(t, "2024-01-03T12:00:00Z"),
	)

	if got := span.Hours(); got != 60 {
		t.Fatalf("Hours() = %d, want 60", got)
	}

	if got := span.Days(); got != 2 {
		t.Fatalf("Days() = %d, want 2", got)
	}

	if got := span.Minutes(); got != 60*60 {
		t.Fatalf("Minutes() = %d, want %d", got, 60*60)
	}

	if span.Inverted() {
		t.Fatalf("Inverted() = true, want false")
	}
}

func TestSpanMillisecondsSaturatesExtremeDifference(t *testing.T) {
	forward := interval.Span{StartMs: minTestInt64, EndMs: maxTestInt64}

	if got := forward.Milliseconds(); got != int(maxTestInt64) {
		t.Fatalf("forward extreme Milliseconds() = %d, want saturated max", got)
	}

	backward := interval.Span{StartMs: maxTestInt64, EndMs: minTestInt64}

	if got := backward.Milliseconds(); got != int(minTestInt64) {
		t.Fatalf("backward extreme Milliseconds() = %d, want saturated min", got)
	}
}

func TestSpanInvertAndAbs(t *testing.T) {
	span := interval.New(
		mustTime(t, "2024-01-03T12:00:00Z"),
		mustTime(t, "2024-01-01T00:00:00Z"),
	)

	if !span.Inverted() {
		t.Fatalf("Inverted() = false, want true")
	}

	if got := span.Hours(); got != -60 {
		t.Fatalf("inverted Hours() = %d, want -60", got)
	}

	abs := span.Abs()

	if abs.Inverted() {
		t.Fatalf("Abs().Inverted() = true, want false")
	}

	if got := abs.Hours(); got != 60 {
		t.Fatalf("Abs().Hours() = %d, want 60", got)
	}

	if abs := span.Invert(); abs.Inverted() {
		t.Fatalf("Invert().Inverted() = true, want false")
	}
}

func TestSpanContainsRespectsInclusivity(t *testing.T) {
	span := interval.New(
		mustTime(t, "2024-01-01T00:00:00Z"),
		mustTime(t, "2024-01-03T00:00:00Z"),
	)

	endMs := mustTime(t, "2024-01-03T00:00:00Z").UnixMilli()
	startMs := mustTime(t, "2024-01-01T00:00:00Z").UnixMilli()
	insideMs := mustTime(t, "2024-01-02T00:00:00Z").UnixMilli()

	if !span.Contains(insideMs, "[]") {
		t.Fatalf("Contains(inside) = false, want true")
	}

	if !span.Contains(endMs, "[]") {
		t.Fatalf("Contains(end, []) = false, want true")
	}

	if span.Contains(endMs, "[)") {
		t.Fatalf("Contains(end, [)) = true, want false")
	}

	if span.Contains(startMs, "(]") {
		t.Fatalf("Contains(start, (]) = true, want false")
	}
}

func TestSpanOverlapUnionIntersection(t *testing.T) {
	left := interval.New(
		mustTime(t, "2024-01-01T00:00:00Z"),
		mustTime(t, "2024-01-05T00:00:00Z"),
	)
	right := interval.New(
		mustTime(t, "2024-01-04T00:00:00Z"),
		mustTime(t, "2024-01-10T00:00:00Z"),
	)
	disjoint := interval.New(
		mustTime(t, "2024-02-01T00:00:00Z"),
		mustTime(t, "2024-02-05T00:00:00Z"),
	)

	if !left.Overlaps(right) {
		t.Fatalf("Overlaps(right) = false, want true")
	}

	if left.Overlaps(disjoint) {
		t.Fatalf("Overlaps(disjoint) = true, want false")
	}

	intersection, ok := left.Intersection(right)

	if !ok {
		t.Fatalf("Intersection(right) ok = false, want true")
	}

	if got := intersection.Days(); got != 1 {
		t.Fatalf("Intersection(right).Days() = %d, want 1", got)
	}

	if _, ok := left.Intersection(disjoint); ok {
		t.Fatalf("Intersection(disjoint) ok = true, want false")
	}

	union := left.Union(right)

	if got := union.Days(); got != 9 {
		t.Fatalf("Union(right).Days() = %d, want 9", got)
	}
}
