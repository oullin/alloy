package duration

import "strconv"

func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}

func pad(value int, length int) string {
	result := strconv.Itoa(value)

	if value < 0 {
		result = strconv.Itoa(-value)
	}

	for len(result) < length {
		result = "0" + result
	}

	return result
}

func mustInt(input string) int {
	value, _ := strconv.Atoi(input)

	return value
}

func defaultString(input string, fallback string) string {
	if input == "" {
		return fallback
	}

	return input
}
