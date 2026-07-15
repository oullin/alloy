package rules

import (
	"fmt"
	"regexp"
)

func init_regex() {
	Register("Regex", validateRegex)
	Register("NotRegex", validateNotRegex)
}

func validateRegex(_ string, value any, params []string, _ RuleContext) bool {
	if len(params) == 0 {
		return true
	}

	re, err := compilePattern(params[0])
	if err != nil {
		return false
	}

	return re.MatchString(regexValue(value))
}

func validateNotRegex(_ string, value any, params []string, _ RuleContext) bool {
	if len(params) == 0 {
		return true
	}

	re, err := compilePattern(params[0])
	if err != nil {
		return false
	}

	return !re.MatchString(regexValue(value))
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if len(pattern) < 2 || pattern[0] != '/' {
		return regexp.Compile(pattern)
	}

	closing := lastUnescapedSlash(pattern)
	if closing == 0 {
		return regexp.Compile(pattern)
	}

	flags := pattern[closing+1:]
	if !allASCIILetters(flags) {
		return regexp.Compile(pattern)
	}

	for _, flag := range flags {
		if !isSupportedRegexFlag(flag) {
			return nil, fmt.Errorf("unsupported regex flag %q", flag)
		}
	}

	compiled := pattern[1:closing]
	if flags != "" {
		compiled = "(?" + flags + ")" + compiled
	}

	return regexp.Compile(compiled)
}

func regexValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}

	return stringify(value)
}

func lastUnescapedSlash(pattern string) int {
	for i := len(pattern) - 1; i > 0; i-- {
		if pattern[i] != '/' {
			continue
		}

		backslashes := 0
		for j := i - 1; j >= 0 && pattern[j] == '\\'; j-- {
			backslashes++
		}
		if backslashes%2 == 0 {
			return i
		}
	}

	return 0
}

func allASCIILetters(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}

	return true
}

func isSupportedRegexFlag(flag rune) bool {
	return flag == 'i' || flag == 's' || flag == 'm' || flag == 'U'
}
