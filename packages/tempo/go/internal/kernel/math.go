package kernel

const (
	maxInt64 = int64(^uint64(0) >> 1)
	minInt64 = -maxInt64 - 1
)

func AbsInt64(value int64) int64 {
	if value < 0 {
		if value == minInt64 {
			return maxInt64
		}

		return -value
	}

	return value
}

func AverageMilliseconds(startMs int64, endMs int64) int64 {
	if (startMs < 0) != (endMs < 0) {
		return (startMs + endMs) / 2
	}

	return startMs/2 + endMs/2 + (startMs%2+endMs%2)/2
}

func DistanceInt64(left int64, right int64) uint64 {
	if left >= 0 && right >= 0 {
		if left >= right {
			return uint64(left - right)
		}

		return uint64(right - left)
	}

	if left < 0 && right < 0 {
		if left >= right {
			return uint64(left - right)
		}

		return uint64(right - left)
	}

	if left < 0 {
		return uint64(-(left + 1)) + 1 + uint64(right)
	}

	return uint64(left) + uint64(-(right + 1)) + 1
}

func DifferenceInt64(end int64, start int64) int64 {
	distance := DistanceInt64(end, start)

	if end >= start {
		if distance > uint64(maxInt64) {
			return maxInt64
		}

		return int64(distance)
	}

	if distance > uint64(maxInt64)+1 {
		return minInt64
	}

	if distance == uint64(maxInt64)+1 {
		return minInt64
	}

	return -int64(distance)
}
