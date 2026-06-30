package parser

import "time"

type Reader struct {
	location *time.Location
}

func New(location *time.Location) Reader {
	if location == nil {
		location = time.UTC
	}

	return Reader{location: location}
}

func (parser Reader) Parse(input string) (time.Time, error) {
	return ParseInLocation(input, parser.location)
}

func (parser Reader) FromFormat(input string, pattern string) (time.Time, error) {
	return ParseFromPattern(input, pattern, parser.location)
}
