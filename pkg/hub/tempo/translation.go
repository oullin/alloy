package tempo

import "github.com/oullin/alloy/pkg/hub/tempo/runtime"

func replaceTranslationTokens(message string, replacements map[string]string) string {
	return runtime.ReplaceTokens(message, replacements)
}
