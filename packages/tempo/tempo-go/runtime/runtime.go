// Package runtime carries the per-Tempo locale, fallback locale and
// optional translator that drive locale-aware rendering. It is split
// out of the tempo umbrella so feature packages can depend on Runtime
// without pulling in the rest of tempo, and so applications can build
// and pass Runtime values directly.
package runtime

import "strings"

// Translator is implemented by anything that resolves locale messages
// and template translations. Implementations should be side-effect
// free and goroutine-safe.
type Translator interface {
	Message(key string) (any, bool)
	Translate(key string, replacements map[string]string) (string, bool)
}

type Runtime struct {
	locale         string
	fallbackLocale string
	translator     Translator
}

type Option func(*Runtime)

func New(options ...Option) Runtime {
	runtime := Runtime{locale: "en-US", fallbackLocale: "en-US"}

	for _, option := range options {
		option(&runtime)
	}

	return runtime
}

func Locale(locale string) Option {
	return func(runtime *Runtime) {
		if strings.TrimSpace(locale) != "" {
			runtime.locale = locale
		}
	}
}

func FallbackLocale(locale string) Option {
	return func(runtime *Runtime) {
		if strings.TrimSpace(locale) != "" {
			runtime.fallbackLocale = locale
		}
	}
}

func WithTranslator(translator Translator) Option {
	return func(runtime *Runtime) {
		runtime.translator = translator
	}
}

func (runtime Runtime) With(options ...Option) Runtime {
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
			return ReplaceTokens(message, replacements), true
		}
	}

	return "", false
}

func ReplaceTokens(message string, replacements map[string]string) string {
	output := message

	for key, value := range replacements {
		output = strings.ReplaceAll(output, ":"+key, value)
	}

	return output
}
