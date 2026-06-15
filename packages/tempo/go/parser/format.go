package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func ParseFromPattern(input string, pattern string, location *time.Location) (time.Time, error) {
	groups := make([]string, 0)

	var expression strings.Builder

	expression.WriteString("^")

	for index := 0; index < len(pattern); {
		if pattern[index] == '[' {
			end := strings.IndexByte(pattern[index:], ']')

			if end >= 0 {
				expression.WriteString(regexp.QuoteMeta(pattern[index+1 : index+end]))
				index += end + 1

				continue
			}
		}

		matched := ""

		for _, token := range parserSettings.FormatTokens {
			if strings.HasPrefix(pattern[index:], token) {
				matched = token

				break
			}
		}

		if matched == "" {
			expression.WriteString(regexp.QuoteMeta(pattern[index : index+1]))
			index++

			continue
		}

		groups = append(groups, matched)

		switch matched {
		case "A", "a":
			expression.WriteString(`(AM|PM|am|pm)`)
		case "MMM", "MMMM", "ddd", "dddd":
			expression.WriteString(`([\p{L}.]+)`)
		case "Do":
			expression.WriteString(`(\d{1,2})(?:st|nd|rd|th)`)
		case "Z":
			expression.WriteString(`(Z|[+-]\d{2}:\d{2})`)
		case "ZZ":
			expression.WriteString(`(Z|[+-]\d{4})`)
		case "YYYY":
			expression.WriteString(`(\d{4})`)
		case "YY", "MM", "DD", "HH", "hh", "mm", "ss":
			expression.WriteString(`(\d{2})`)
		case "SSS":
			expression.WriteString(`(\d{1,3})`)
		default:
			expression.WriteString(`(\d{1,2})`)
		}

		index += len(matched)
	}

	expression.WriteString("$")
	match := regexp.MustCompile(expression.String()).FindStringSubmatch(input)

	if match == nil {
		return time.Time{}, fmt.Errorf("input does not match tempo format: %s", input)
	}

	values := make(map[string]string, len(groups))

	for index, token := range groups {
		values[token] = match[index+1]
	}

	year := 1970

	if value := values["YYYY"]; value != "" {
		year = mustInt(value)
	} else if value := values["YY"]; value != "" {
		year = 2000 + mustInt(value)
	}

	hour := mustInt(firstPresent(values, "HH", "H", "hh", "h"))
	meridiem := firstPresent(values, "A", "a")

	if meridiem != "" {
		switch strings.ToLower(meridiem) {
		case "pm":
			if hour < 12 {
				hour += 12
			}
		case "am":
			if hour == 12 {
				hour = 0
			}
		}
	}

	components := Components{
		Year:        year,
		Month:       mustInt(defaultString(firstPresent(values, "MM", "M"), "1")),
		Day:         mustInt(defaultString(firstPresent(values, "DD", "Do", "D"), "1")),
		Hour:        hour,
		Minute:      mustInt(defaultString(firstPresent(values, "mm", "m"), "0")),
		Second:      mustInt(defaultString(firstPresent(values, "ss", "s"), "0")),
		Millisecond: mustInt(rightPad(defaultString(values["SSS"], "0"), 3)),
	}

	if month, ok := monthNumberFromName(firstPresent(values, "MMMM", "MMM")); ok {
		components.Month = month
	}

	if offset := firstPresent(values, "Z", "ZZ"); offset != "" {
		offsetMinutes, err := parseOffsetMinutes(offset)

		if err != nil {
			return time.Time{}, err
		}

		utc := time.Date(
			components.Year,
			time.Month(components.Month),
			components.Day,
			components.Hour,
			components.Minute,
			components.Second,
			components.Millisecond*int(time.Millisecond),
			time.UTC,
		)

		return utc.Add(-time.Duration(offsetMinutes) * time.Minute), nil
	}

	return timeFromComponents(components, location), nil
}
