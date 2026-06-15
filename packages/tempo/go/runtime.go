package tempo

import "strings"

type Runtime struct {
	locale         string
	fallbackLocale string
	translator     Translator
}

func NewRuntime(options ...RuntimeOption) Runtime {
	runtime := Runtime{locale: "en-US", fallbackLocale: "en-US"}
	for _, option := range options {
		option(&runtime)
	}

	return runtime
}

type RuntimeOption func(*Runtime)

func RuntimeLocale(locale string) RuntimeOption {
	return func(runtime *Runtime) {
		if strings.TrimSpace(locale) != "" {
			runtime.locale = locale
		}
	}
}

func RuntimeFallbackLocale(locale string) RuntimeOption {
	return func(runtime *Runtime) {
		if strings.TrimSpace(locale) != "" {
			runtime.fallbackLocale = locale
		}
	}
}

func RuntimeTranslator(translator Translator) RuntimeOption {
	return func(runtime *Runtime) {
		runtime.translator = translator
	}
}

func (runtime Runtime) With(options ...RuntimeOption) Runtime {
	next := runtime
	for _, option := range options {
		option(&next)
	}

	return next
}

func (runtime Runtime) Locale() string { return runtime.locale }

func (runtime Runtime) FallbackLocale() string { return runtime.fallbackLocale }

func (runtime Runtime) HasTranslator() bool { return runtime.translator != nil }

func (runtime Runtime) Translator() Translator { return runtime.translator }

func (runtime Runtime) Message(key string) (any, bool) {
	if runtime.translator != nil {
		if value, ok := runtime.translator.Message(key); ok {
			return value, true
		}
	}

	switch key {
	case "day_of_first_week_of_year":
		return 4, true
	case "first_day_of_week":
		return 1, true
	case "locale":
		return runtime.locale, true
	default:
		return nil, false
	}
}

func (runtime Runtime) Translate(key string, replacements map[string]string) (string, bool) {
	if runtime.translator != nil {
		if value, ok := runtime.translator.Translate(key, replacements); ok {
			return value, true
		}
	}
	if value, ok := runtime.Message(key); ok {
		if message, ok := value.(string); ok {
			return replaceTranslationTokens(message, replacements), true
		}
	}

	return "", false
}
