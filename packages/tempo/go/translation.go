package tempo

import "strings"

func replaceTranslationTokens(message string, replacements map[string]string) string {
	output := message
	for key, value := range replacements {
		output = strings.ReplaceAll(output, ":"+key, value)
	}

	return output
}
