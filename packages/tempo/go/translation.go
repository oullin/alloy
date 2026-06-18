package tempo

import "github.com/oullin/alloy/tempo/runtime"

func replaceTranslationTokens(message string, replacements map[string]string) string {
	return runtime.ReplaceTokens(message, replacements)
}
