package parser

import "strconv"

func firstPresent(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := values[key]; value != "" {
			return value
		}
	}

	return ""
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

func rightPad(input string, length int) string {
	for len(input) < length {
		input += "0"
	}

	if len(input) > length {
		return input[:length]
	}

	return input
}
