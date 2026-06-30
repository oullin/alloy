package parser

import (
	"fmt"
	"strconv"
	"strings"
)

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
