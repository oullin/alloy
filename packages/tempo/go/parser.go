package tempo

import "time"

type Parser struct {
	location *time.Location
	runtime  Runtime
}

func NewParser(options ...Option) (Parser, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Parser{}, err
	}

	return newParser(cfg.location, cfg.runtime), nil
}

func newParser(location *time.Location, runtime Runtime) Parser {
	return Parser{location: location, runtime: runtime}
}

func (parser Parser) Parse(input string) (Tempo, error) {
	parsed, err := parseInLocation(input, parser.location)
	if err != nil {
		return Tempo{}, err
	}

	return newTempo(parsed, parser.location, parser.runtime), nil
}

func (parser Parser) FromFormat(input string, pattern string) (Tempo, error) {
	parsed, err := parseFromPattern(input, pattern, parser.location)
	if err != nil {
		return Tempo{}, err
	}

	return newTempo(parsed, parser.location, parser.runtime), nil
}
