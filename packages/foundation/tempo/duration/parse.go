package duration

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

var durationPattern = regexp.MustCompile(`^(-)?P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)

func Parse(input string) (Span, error) {
	matches := durationPattern.FindStringSubmatch(input)

	if matches == nil {
		return Span{}, fmt.Errorf("invalid tempo duration: %s", input)
	}

	sign := 1

	if matches[1] == "-" {
		sign = -1
	}

	seconds := 0
	milliseconds := 0

	if matches[8] != "" {
		value, err := strconv.ParseFloat(matches[8], 64)

		if err != nil {
			return Span{}, fmt.Errorf("invalid tempo duration seconds: %w", err)
		}

		seconds = int(math.Trunc(value))
		milliseconds = int(math.Round((value - math.Trunc(value)) * 1000))
	}

	parsed := Span{
		Years:        sign * mustInt(defaultString(matches[2], "0")),
		Months:       sign * mustInt(defaultString(matches[3], "0")),
		Weeks:        sign * mustInt(defaultString(matches[4], "0")),
		Days:         sign * mustInt(defaultString(matches[5], "0")),
		Hours:        sign * mustInt(defaultString(matches[6], "0")),
		Minutes:      sign * mustInt(defaultString(matches[7], "0")),
		Seconds:      sign * seconds,
		Milliseconds: sign * milliseconds,
	}

	return parsed.Normalize(), nil
}
