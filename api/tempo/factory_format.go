package tempo

func FromFormat(input string, pattern string, options ...Option) (Time, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Time{}, err
	}

	return newParserWithPolicy(cfg.location, cfg.runtime, cfg.settings, cfg.serializer, cfg.toStringFormat).FromFormat(input, pattern)
}
