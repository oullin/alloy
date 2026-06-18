package parser

import "time"

type Parser struct {
	location *time.Location
}

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
