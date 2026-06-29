package tempo

import "alloy.dev/backend/tempo/runtime"

type Context = runtime.Context

type RuntimeOption = runtime.Option

func NewRuntime(options ...RuntimeOption) Context {
	return runtime.New(options...)
}

func RuntimeLocale(locale string) RuntimeOption {
	return runtime.Locale(locale)
}

func RuntimeFallbackLocale(locale string) RuntimeOption {
	return runtime.FallbackLocale(locale)
}

func RuntimeTranslator(translator Translator) RuntimeOption {
	return runtime.WithTranslator(translator)
}
