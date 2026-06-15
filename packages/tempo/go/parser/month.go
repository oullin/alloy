package parser

import "strings"

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
