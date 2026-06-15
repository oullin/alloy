package tempo

func FromFormat(input string, pattern string, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	return newParser(cfg.location, cfg.runtime).FromFormat(input, pattern)
}

func CreateFromFormat(input string, pattern string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func CreateFromIsoFormat(input string, pattern string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func CreateFromLocaleFormat(input string, pattern string, _ string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func CreateFromLocaleIsoFormat(input string, pattern string, _ string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func RawCreateFromFormat(input string, pattern string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func TryFromFormat(input string, pattern string, options ...Option) (Tempo, bool) {
	tempo, err := FromFormat(input, pattern, options...)

	return tempo, err == nil
}

func HasFormat(input string, pattern string, options ...Option) bool {
	_, ok := TryFromFormat(input, pattern, options...)

	return ok
}

func CanBeCreatedFromFormat(input string, pattern string, options ...Option) bool {
	return HasFormat(input, pattern, options...)
}

func HasFormatWithModifiers(input string, pattern string, options ...Option) bool {
	return HasFormat(input, pattern, options...)
}
