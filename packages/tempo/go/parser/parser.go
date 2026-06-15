package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Parser struct {
	location *time.Location
}

type Components struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Millisecond int
}

var (
	dateOnlyPattern = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
	localPattern    = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})(?:[T\s](\d{2})(?::?(\d{2}))?(?::?(\d{2})(?:\.(\d{1,9}))?)?)?$`)
	zonePattern     = regexp.MustCompile(`(?:Z|[+-]\d{2}:?\d{2})$`)
)

func New(location *time.Location) Parser {
	if location == nil {
		location = time.UTC
	}

	return Parser{location: location}
}

func (parser Parser) Parse(input string) (time.Time, error) {
	return ParseInLocation(input, parser.location)
}

func (parser Parser) FromFormat(input string, pattern string) (time.Time, error) {
	return ParseFromPattern(input, pattern, parser.location)
}

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

func ParseFromPattern(input string, pattern string, location *time.Location) (time.Time, error) {
	tokens := []string{"YYYY", "MMMM", "dddd", "MMM", "ddd", "SSS", "Do", "YY", "ZZ", "MM", "DD", "HH", "hh", "mm", "ss", "Z", "M", "D", "H", "h", "m", "s", "A", "a"}
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
		for _, token := range tokens {
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

func timeFromComponents(components Components, location *time.Location) time.Time {
	month := components.Month
	if month == 0 {
		month = 1
	}

	day := components.Day
	if day == 0 {
		day = 1
	}

	return time.Date(
		components.Year,
		time.Month(month),
		day,
		components.Hour,
		components.Minute,
		components.Second,
		components.Millisecond*int(time.Millisecond),
		location,
	).UTC()
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
