package parser

import (
	"fmt"
	"time"
)

func ParseInLocation(input string, location *time.Location) (time.Time, error) {
	return ParseInLocationStrict(input, location, false)
}

func ParseInLocationStrict(input string, location *time.Location, strict bool) (time.Time, error) {
	if matches := dateOnlyPattern.FindStringSubmatch(input); matches != nil {
		year := mustInt(matches[1])
		month := mustInt(matches[2])
		day := mustInt(matches[3])
		value := time.Date(year, time.Month(month), day, 0, 0, 0, 0, location)

		if strict && !componentsMatchTime(Components{Year: year, Month: month, Day: day}, value, location) {
			return time.Time{}, fmt.Errorf("invalid tempo local date/time components")
		}

		return value, nil
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
			value := time.Date(year, time.Month(month), day, hour, minute, second, millisecond*int(time.Millisecond), location)

			if strict && !componentsMatchTime(Components{
				Year:        year,
				Month:       month,
				Day:         day,
				Hour:        hour,
				Minute:      minute,
				Second:      second,
				Millisecond: millisecond,
			}, value, location) {
				return time.Time{}, fmt.Errorf("invalid tempo local date/time components")
			}

			return value, nil
		}
	}

	parsed, err := time.Parse(time.RFC3339Nano, input)

	if err != nil {
		return time.Time{}, fmt.Errorf("parse tempo: %w", err)
	}

	return parsed, nil
}
