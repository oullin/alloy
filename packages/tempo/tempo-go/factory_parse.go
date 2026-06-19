package tempo

import "strconv"

func Parse(input string, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	return newParserWithPolicy(cfg.location, cfg.runtime, cfg.settings, cfg.serializer, cfg.toStringFormat).Parse(input)
}

func FromSerialized(input string, options ...Option) (Tempo, error) {
	value, err := strconv.Unquote(input)

	if err != nil {
		return Tempo{}, err
	}

	return Parse(value, options...)
}
