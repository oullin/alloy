package interval

import (
	"time"

	"alloy.dev/foundation/tempo/internal/kernel"
)

type Span struct {
	StartMs int64
	EndMs   int64
}

func New(start time.Time, end time.Time) Span {
	return Span{StartMs: start.UnixMilli(), EndMs: end.UnixMilli()}
}

func (span Span) Inverted() bool {
	return span.StartMs > span.EndMs
}

func (span Span) Milliseconds() int {
	return int(kernel.DifferenceInt64(span.EndMs, span.StartMs))
}

func (span Span) Seconds() int {
	return span.Milliseconds() / int(time.Second/time.Millisecond)
}

func (span Span) Minutes() int {
	return span.Milliseconds() / int(time.Minute/time.Millisecond)
}

func (span Span) Hours() int {
	return span.Milliseconds() / int(time.Hour/time.Millisecond)
}

func (span Span) Days() int {
	return span.Milliseconds() / int((24*time.Hour)/time.Millisecond)
}

func (span Span) Weeks() int {
	return span.Days() / 7
}

func (span Span) Abs() Span {
	if !span.Inverted() {
		return span
	}

	return span.Invert()
}

func (span Span) Invert() Span {
	return Span{StartMs: span.EndMs, EndMs: span.StartMs}
}

func (span Span) Contains(inputMs int64, inclusivity string) bool {
	normalized := span.Abs()
	includeStart := len(inclusivity) < 1 || inclusivity[0] == '['
	includeEnd := len(inclusivity) < 2 || inclusivity[1] == ']'

	afterStart := inputMs > normalized.StartMs || (includeStart && inputMs == normalized.StartMs)
	beforeEnd := inputMs < normalized.EndMs || (includeEnd && inputMs == normalized.EndMs)

	return afterStart && beforeEnd
}

func (span Span) Overlaps(other Span) bool {
	left := span.Abs()
	right := other.Abs()

	return left.StartMs <= right.EndMs && right.StartMs <= left.EndMs
}

func (span Span) Intersection(other Span) (Span, bool) {
	if !span.Overlaps(other) {
		return Span{}, false
	}

	left := span.Abs()
	right := other.Abs()

	return Span{
		StartMs: max(left.StartMs, right.StartMs),
		EndMs:   min(left.EndMs, right.EndMs),
	}, true
}

func (span Span) Union(other Span) Span {
	left := span.Abs()
	right := other.Abs()

	return Span{
		StartMs: min(left.StartMs, right.StartMs),
		EndMs:   max(left.EndMs, right.EndMs),
	}
}
