package kernel

func AbsInt64(value int64) int64 {
	if value < 0 {
		return -value
	}

	return value
}

func AverageMilliseconds(startMs int64, endMs int64) int64 {
	return (startMs + endMs) / 2
}
