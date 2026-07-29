package tempo

import (
	"time"

	"hara.sh/alloy/tempo/parser"
)

func parseInLocation(input string, location *time.Location) (time.Time, error) {
	return parser.ParseInLocation(input, location)
}

func parseFromPattern(input string, pattern string, location *time.Location) (time.Time, error) {
	return parser.ParseFromPattern(input, pattern, location)
}
