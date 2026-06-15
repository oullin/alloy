package parser

import (
	"fmt"
	"time"
)

func ParseInLocation(input string, location *time.Location) (time.Time, error) {
	if matches := dateOnlyPattern.FindStringSubmatch(input); matches != nil {
		year := mustInt(matches[1])
		month := mustInt(matches[2])
		day := mustInt(matches[3])

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, location), nil
	}

	if !zonePattern.MatchString(input) {
		if matches := localPattern.FindStringSubmatch(input); matches != nil {
			year := mustInt(matches[1])
			month := mustInt(matches[2])
			day := mustInt(matches[3])
			hour := mustInt(defaultString(matches[4], "0"))
			minute := mustInt(defaultString(matches[5], "0"))
			second := mustInt(defaultString(matches[6], "0"))
			millisecond := mustInt(rightPad(defaultString(matches[7], "0"), 3))

			return time.Date(year, time.Month(month), day, hour, minute, second, millisecond*int(time.Millisecond), location), nil
		}
	}

	parsed, err := time.Parse(time.RFC3339Nano, input)

	if err != nil {
		return time.Time{}, fmt.Errorf("parse tempo: %w", err)
	}

	return parsed, nil
}
