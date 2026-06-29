package tempo

import "alloy.dev/go/tempo/runtime"

func replaceTranslationTokens(message string, replacements map[string]string) string {
	return runtime.ReplaceTokens(message, replacements)
}
