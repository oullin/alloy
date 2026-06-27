package tempo

import "alloy.dev/api/tempo/internal/kernel"

func AverageMilliseconds(startMs int64, endMs int64) int64 {
	return kernel.AverageMilliseconds(startMs, endMs)
}

func EarlierIndex(values []int64) int {
	if len(values) == 0 {
		return -1
	}

	selected := 0

	for index, value := range values[1:] {
		if value < values[selected] {
			selected = index + 1
		}
	}

	return selected
}

func LaterIndex(values []int64) int {
	if len(values) == 0 {
		return -1
	}

	selected := 0

	for index, value := range values[1:] {
		if value > values[selected] {
			selected = index + 1
		}
	}

	return selected
}
