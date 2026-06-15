package tempo

func GetLastErrors() error { return defaultConfig.LastError }

func ExecuteWithLocale[T any](locale string, callback func() T) T {
	previous := defaultConfig.Settings.Locale
	defaultConfig.Settings.Locale = locale

	defer func() { defaultConfig.Settings.Locale = previous }()

	return callback()
}

func SerializeUsing(next Serializer) { defaultConfig.Serializer = next }

func SetToStringFormat(pattern string) { defaultConfig.ToStringFormat = pattern }

func ResetToStringFormat() { defaultConfig.ToStringFormat = "" }
