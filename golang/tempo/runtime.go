package tempo

import "github.com/oullin/alloy/tempo/runtime"

type Runtime = runtime.Runtime

type RuntimeOption = runtime.Option

func NewRuntime(options ...RuntimeOption) Runtime {
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
