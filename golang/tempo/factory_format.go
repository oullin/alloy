package tempo

func FromFormat(input string, pattern string, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	return newParserWithPolicy(cfg.location, cfg.runtime, cfg.settings, cfg.serializer, cfg.toStringFormat).FromFormat(input, pattern)
}
