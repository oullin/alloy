package tempo

import "alloy.dev/foundation/tempo/runtime"

func replaceTranslationTokens(message string, replacements map[string]string) string {
	return runtime.ReplaceTokens(message, replacements)
}
