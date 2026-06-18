package tempo

import (
	"fmt"
	"strconv"
	"strings"
)

func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}

func absInt64(value int64) int64 {
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

func selectedPrecision(precision []TimeStringPrecision) TimeStringPrecision {
	if len(precision) > 0 && precision[0] == MillisecondPrecision {
		return MillisecondPrecision
	}

	return SecondPrecision
}

func monthNumberFromName(input string) (int, bool) {
	normalized := strings.TrimSuffix(strings.ToLower(input), ".")
	names := map[string]int{
		"jan":       1,
		"january":   1,
		"feb":       2,
		"february":  2,
		"mar":       3,
		"march":     3,
		"apr":       4,
		"april":     4,
		"may":       5,
		"jun":       6,
		"june":      6,
		"jul":       7,
		"july":      7,
		"aug":       8,
		"august":    8,
		"sep":       9,
		"sept":      9,
		"september": 9,
		"oct":       10,
		"october":   10,
		"nov":       11,
		"november":  11,
		"dec":       12,
		"december":  12,
	}
	month, ok := names[normalized]

	return month, ok
}

func formatOffset(offsetMinutes int, separator string) string {
	sign := "+"

	if offsetMinutes < 0 {
		sign = "-"
		offsetMinutes = -offsetMinutes
	}

	return fmt.Sprintf("%s%s%s%s", sign, pad(offsetMinutes/60, 2), separator, pad(offsetMinutes%60, 2))
}

func parseOffsetMinutes(input string) (int, error) {
	if input == "Z" {
		return 0, nil
	}

	clean := strings.ReplaceAll(input, ":", "")

	if len(clean) != 5 {
		return 0, fmt.Errorf("invalid tempo offset: %s", input)
	}

	sign := 1

	if clean[0] == '-' {
		sign = -1
	} else if clean[0] != '+' {
		return 0, fmt.Errorf("invalid tempo offset: %s", input)
	}

	hours, err := strconv.Atoi(clean[1:3])

	if err != nil {
		return 0, fmt.Errorf("invalid tempo offset: %s", input)
	}

	minutes, err := strconv.Atoi(clean[3:5])

	if err != nil {
		return 0, fmt.Errorf("invalid tempo offset: %s", input)
	}

	return sign * (hours*60 + minutes), nil
}

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

func ternary(condition bool, left string, right string) string {
	if condition {
		return left
	}

	return right
}
