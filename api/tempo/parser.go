package tempo

import (
	"time"

	dateparser "github.com/oullin/alloy/api/tempo/parser"
)

type Parser struct {
	location       *time.Location
	runtime        Context
	settings       Settings
	serializer     Serializer
	toStringFormat string
}

func NewParser(options ...Option) (Parser, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Parser{}, err
	}

	return newParserWithPolicy(cfg.location, cfg.runtime, cfg.settings, cfg.serializer, cfg.toStringFormat), nil
}

func newParser(location *time.Location, runtime Context) Parser {
	return newParserWithPolicy(location, runtime, defaultSettings(), nil, "")
}

func newParserWithPolicy(location *time.Location, runtime Context, settings Settings, serializer Serializer, toStringFormat string) Parser {
	return Parser{
		location:       location,
		runtime:        runtime,
		settings:       cloneSettings(normalizeSettings(settings)),
		serializer:     serializer,
		toStringFormat: toStringFormat,
	}
}

func (parser Parser) Parse(input string) (Time, error) {
	parsed, err := dateparser.ParseInLocationStrict(input, parser.location, parser.settings.StrictMode)

	if err != nil {
		return Time{}, err
	}

	return newTempoWithPolicy(parsed, parser.location, parser.runtime, parser.settings, parser.serializer, parser.toStringFormat), nil
}

func (parser Parser) FromFormat(input string, pattern string) (Time, error) {
	parsed, err := dateparser.ParseFromPatternStrict(input, pattern, parser.location, parser.settings.StrictMode)

	if err != nil {
		return Time{}, err
	}

	return newTempoWithPolicy(parsed, parser.location, parser.runtime, parser.settings, parser.serializer, parser.toStringFormat), nil
}
