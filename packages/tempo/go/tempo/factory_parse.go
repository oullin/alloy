package tempo

import "strconv"

func Parse(input string, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Tempo{}, err
	}

	return newParser(cfg.location, cfg.runtime).Parse(input)
}

func RawParse(input string, options ...Option) (Tempo, error) {
	return Parse(input, options...)
}

func TryParse(input string, options ...Option) (Tempo, bool) {
	tempo, err := Parse(input, options...)
	defaultConfig.LastError = err

	return tempo, err == nil
}

func CanParse(input string, options ...Option) bool {
	_, ok := TryParse(input, options...)

	return ok
}

func FromSerialized(input string, options ...Option) (Tempo, error) {
	value, err := strconv.Unquote(input)

	if err != nil {
		return Tempo{}, err
	}

	return Parse(value, options...)
}

func GetClock() *Tempo { return GetTestNow() }

func Make(input string, options ...Option) (Tempo, error) {
	return Parse(input, options...)
}

func ParseFromLocale(input string, _ string, options ...Option) (Tempo, error) {
	return Parse(input, options...)
}
