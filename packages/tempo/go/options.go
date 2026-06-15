package tempo

func WithTimezone(name string) Option {
	return func(cfg *config) error {
		location, err := loadLocation(name)
		if err != nil {
			return err
		}

		cfg.location = location
		return nil
	}
}

func WithRuntime(runtime Runtime) Option {
	return func(cfg *config) error {
		cfg.runtime = runtime
		return nil
	}
}

func WithTranslator(translator Translator) Option {
	return func(cfg *config) error {
		cfg.runtime = cfg.runtime.With(RuntimeTranslator(translator))
		return nil
	}
}

func WithLocale(locale string) Option {
	return func(cfg *config) error {
		cfg.runtime = cfg.runtime.With(RuntimeLocale(locale))
		return nil
	}
}

func WithFallbackLocale(locale string) Option {
	return func(cfg *config) error {
		cfg.runtime = cfg.runtime.With(RuntimeFallbackLocale(locale))
		return nil
	}
}
