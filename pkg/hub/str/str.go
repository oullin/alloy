package str

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// randomMu guards package-level random string factory state.

// Casing caches (camel, studly, snake)

// After returns the portion of the string after the first occurrence of search.
// If search is not found, the full subject is returned.

// AfterLast returns the portion after the last occurrence of search.

// Before returns the portion before the first occurrence of search.

// BeforeLast returns the portion before the last occurrence of search.

// Between returns the portion between from and to.

// BetweenFirst returns the smallest portion between from and to.

// Camel converts a string to camelCase.

// Studly converts a string to StudlyCase (PascalCase).

// Replace common separators with spaces, then title-case

// studlySplit splits a string on separators (-, _, space) and camel-case boundaries.

// Replace hyphens, underscores with spaces

// Pascal converts a string to PascalCase (alias for Studly).

// Snake converts a string to snake_case.

// Insert separator before uppercase letters that follow lowercase or digits

// Kebab converts a string to kebab-case.

// Title converts a string to Title Case.

//nolint:staticcheck

// Headline converts a string to Headline Case.
// Each word is capitalized, separating camelCase and underscored strings.

// Handle camelCase by inserting spaces

// Replace separators

// Title case each word

// Apa converts a string to APA title case.

// APA: lowercase articles, conjunctions, short prepositions

// Capitalize

// Ucfirst uppercases the first character of the string.

// Lcfirst lowercases the first character of the string.

// Ucwords uppercases the first character of each word.

//nolint:staticcheck

// Word-boundary title case with custom separators

// Ucsplit splits a string by uppercase characters.

// Contains determines if the haystack contains the given needle(s).
// Case-sensitive by default.

// ContainsAll determines if the haystack contains all given needles.

// DoesntContain determines if the haystack does not contain the given needle(s).

// StartsWith determines if the subject starts with the given prefix(es).

// DoesntStartWith determines if the subject does not start with any of the given values.

// EndsWith determines if the subject ends with the given suffix(es).

// DoesntEndWith determines if the subject does not end with any of the given values.

// Is determines if the value matches the given pattern.
// An asterisk (*) may be used as a wildcard value.

// Convert glob pattern to regex

// IsMatch determines if the value matches any of the given patterns (regex).

// IsAscii determines if a string is 7-bit ASCII.

// IsJson determines if the given string is valid JSON.

// IsUrl determines if the given string is a valid URL.

// IsUuid determines if the given string is a valid UUID.

// IsUlid determines if the given string is a valid ULID.

// Length returns the length of the string in runes (characters).

// Limit truncates the string to the given number of characters.

// Words limits the number of words in a string.

// Split preserving spaces around words

// Lower converts the string to lowercase.

// Upper converts the string to uppercase.

// Reverse reverses the string (UTF-8 aware).

// Replace replaces occurrences of search in subject with replace.

// ReplaceArray replaces each search occurrence with the next value from replacements.

// ReplaceFirst replaces the first occurrence of search in subject.

// ReplaceLast replaces the last occurrence of search in subject.

// ReplaceStart replaces the search value if it appears at the start of subject.

// ReplaceEnd replaces the search value if it appears at the end of subject.

// ReplaceMatches replaces all occurrences of a regex pattern in subject.

// Remove removes all occurrences of search from subject.

// Repeat repeats the string the given number of times.

// Squish removes all extra blank space (including between words).

// Deduplicate replaces consecutive occurrences of characters with a single instance.

// Trim removes leading and trailing whitespace (or given chars).

// Ltrim removes leading whitespace (or given chars).

// Rtrim removes trailing whitespace (or given chars).

// Start prepends a single instance of value to the subject
// if it does not already start with it.

// Finish appends a single instance of cap to value
// if it does not already end with it.

// Wrap wraps the string with the given strings.

// Unwrap removes the given prefix and suffix from the string.

// ChopStart removes a single leading occurrence of needle (or the first matching needle).

// ChopEnd removes a single trailing occurrence of needle (or the first matching needle).

// Mask replaces a portion of the string with a repeated character.

// Excerpt returns a string excerpt around the given phrase.

// Match returns the first regex match in subject.

// MatchAll returns all regex matches in subject.

// Numbers extracts all numeric characters from the string.

// PadBoth pads the string on both sides to the given length.

// PadLeft pads the left side of the string to the given length.

// PadRight pads the right side of the string to the given length.

// Position finds the position of needle in haystack.
// Returns the byte position and whether it was found.

// Substr returns a substring from start for the given length.

// SubstrCount counts the occurrences of needle in haystack.

// SubstrReplace replaces a portion of the string.

// Swap swaps multiple keyword pairs in the subject.

// Build patterns ordered by length (longer first to avoid partial replacements)

// Take returns the first n characters, or from the end if n is negative.

// negative: from end

// WordCount counts the number of words in the string.

// WordWrap wraps the string to the given number of characters.

// ToBase64 encodes the string to base64.

// FromBase64 decodes a base64-encoded string.

// Try URL-safe base64

// Slug generates a URL-friendly slug from the given string.

// Convert to ASCII

// Replace @ with separator (common convention)

// Replace non-alphanumeric with separator

// Remove leading/trailing separators and deduplicate

// Ascii transliterates the string to its closest ASCII representation.

// Unknown non-ASCII chars are dropped (like the upstream behavior)

// Transliterate converts the string to its closest ASCII form.

// Initials returns the initials of the given name.

// CharAt returns the character at the given index (UTF-8 aware).

// ParseCallback parses a "Class@method" or "Class::method" callback string.
// Returns [class, method].

// Random generates a random alphanumeric string of the given length.

// fallback used when sequence exhausted

// CreateRandomStringsUsing sets a custom random string factory.

// CreateRandomStringsUsingSequence sets a sequence of strings to use for random generation.

// CreateRandomStringsNormally resets random string generation to default.

// Password generates a secure random password.

// FlushCache clears the casing caches.

// ConvertCase converts the string case using the given mode.

// MB_CASE_UPPER

// MB_CASE_LOWER

// MB_CASE_TITLE
//nolint:staticcheck

// Of creates a new Builder for fluent string manipulation.

// Builder provides a fluent interface for string manipulation.
// Ref: @alloy/code-0354
type Builder struct {
	value string
}

var (
	randomMu       sync.Mutex
	randomFactory  func(int) string
	randomSequence []string
	randomFallback func(int) string

	camelCache  sync.Map
	studlyCache sync.Map
	snakeCache  sync.Map
)

func After(subject, search string) string {
	if search == "" {
		return subject
	}

	_, after, ok := strings.Cut(subject, search)

	if !ok {
		return subject
	}

	return after
}

func AfterLast(subject, search string) string {
	if search == "" {
		return subject
	}

	idx := strings.LastIndex(subject, search)

	if idx == -1 {
		return subject
	}

	return subject[idx+len(search):]
}

func Before(subject, search string) string {
	if search == "" {
		return subject
	}

	before, _, ok := strings.Cut(subject, search)

	if !ok {
		return subject
	}

	return before
}

func BeforeLast(subject, search string) string {
	if search == "" {
		return subject
	}

	idx := strings.LastIndex(subject, search)

	if idx == -1 {
		return subject
	}

	return subject[:idx]
}

func Between(subject, from, to string) string {
	if from == "" || to == "" {
		return subject
	}

	return BeforeLast(After(subject, from), to)
}

func BetweenFirst(subject, from, to string) string {
	if from == "" || to == "" {
		return subject
	}

	return Before(After(subject, from), to)
}

func Camel(value string) string {
	if cached, ok := camelCache.Load(value); ok {
		return cached.(string)
	}

	result := strLcfirst(Studly(value))
	camelCache.Store(value, result)

	return result
}

func Studly(value string) string {
	if cached, ok := studlyCache.Load(value); ok {
		return cached.(string)
	}

	words := studlySplit(value)

	var builder strings.Builder

	for _, w := range words {
		if len(w) > 0 {
			r, size := utf8.DecodeRuneInString(w)
			builder.WriteRune(unicode.ToUpper(r))
			builder.WriteString(w[size:])
		}
	}

	result := builder.String()
	studlyCache.Store(value, result)

	return result
}

func studlySplit(value string) []string {

	value = regexp.MustCompile(`[-_\s]+`).ReplaceAllString(value, " ")

	return strings.Fields(value)
}

func Pascal(value string) string {
	return Studly(value)
}

func Snake(value string, delimiter ...string) string {
	sep := "_"

	if len(delimiter) > 0 {
		sep = delimiter[0]
	}

	// Separate value and sep with a NUL so distinct (value, sep) pairs cannot
	// collide on the cache key: e.g. Snake("a","_") and Snake("a_","").
	key := value + "\x00" + sep

	if cached, ok := snakeCache.Load(key); ok {
		return cached.(string)
	}

	var result strings.Builder
	runes := []rune(value)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
				result.WriteString(sep)
			} else if i > 0 && unicode.IsUpper(runes[i-1]) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				result.WriteString(sep)
			}

			result.WriteRune(unicode.ToLower(r))
		} else if r == '-' || r == ' ' {
			result.WriteString(sep)
		} else {
			result.WriteRune(r)
		}
	}

	res := strings.Trim(result.String(), sep)
	snakeCache.Store(key, res)

	return res
}

func Kebab(value string) string {
	return Snake(value, "-")
}

func Title(value string) string {
	return strings.Title(strings.ToLower(value))
}

func Headline(value string) string {

	re := regexp.MustCompile(`([a-z])([A-Z])`)
	value = re.ReplaceAllString(value, "$1 $2")

	value = regexp.MustCompile(`[-_]+`).ReplaceAllString(value, " ")

	words := strings.Fields(value)

	for i, w := range words {
		if len(w) == 0 {
			continue
		}

		r, size := utf8.DecodeRuneInString(w)
		words[i] = string(unicode.ToUpper(r)) + strings.ToLower(w[size:])
	}

	return strings.Join(words, " ")
}

func Apa(value string) string {

	minors := map[string]bool{
		"a": true, "an": true, "the": true,
		"and": true, "but": true, "for": true, "nor": true, "or": true, "so": true, "yet": true,
		"as": true, "at": true, "by": true, "in": true, "of": true, "on": true, "to": true, "up": true,
		"via": true,
	}

	words := strings.Fields(value)

	for i, w := range words {
		lower := strings.ToLower(w)

		if i == 0 || i == len(words)-1 || !minors[lower] {

			if len(w) > 0 {
				r, size := utf8.DecodeRuneInString(w)
				words[i] = string(unicode.ToUpper(r)) + w[size:]
			}
		} else {
			words[i] = lower
		}
	}

	return strings.Join(words, " ")
}

func Ucfirst(value string) string {
	if value == "" {
		return value
	}

	r, size := utf8.DecodeRuneInString(value)

	return string(unicode.ToUpper(r)) + value[size:]
}

func Lcfirst(value string) string {
	return strLcfirst(value)
}

func strLcfirst(value string) string {
	if value == "" {
		return value
	}

	r, size := utf8.DecodeRuneInString(value)

	return string(unicode.ToLower(r)) + value[size:]
}

func Ucwords(value string, separators ...string) string {
	seps := " \t\r\n\f\v"

	if len(separators) > 0 {
		seps = separators[0]
	}

	if seps == " " || seps == " \t\r\n\f\v" {
		return strings.Title(value)
	}

	runes := []rune(value)
	newWord := true

	for i, r := range runes {
		if strings.ContainsRune(seps, r) {
			newWord = true
		} else if newWord {
			runes[i] = unicode.ToUpper(r)
			newWord = false
		}
	}

	return string(runes)
}

func Ucsplit(value string) []string {
	re := regexp.MustCompile(`(?m)[A-Z][^A-Z]*`)
	matches := re.FindAllString(value, -1)

	if len(matches) == 0 {
		return []string{value}
	}

	return matches
}

func Contains(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if needle == "" || strings.Contains(haystack, needle) {
			return true
		}
	}

	return false
}

func ContainsAll(haystack string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			return false
		}
	}

	return true
}

func DoesntContain(haystack string, needles ...string) bool {
	return !Contains(haystack, needles...)
}

func StartsWith(subject string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(subject, prefix) {
			return true
		}
	}

	return false
}

func DoesntStartWith(subject string, prefixes ...string) bool {
	return !StartsWith(subject, prefixes...)
}

func EndsWith(subject string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(subject, suffix) {
			return true
		}
	}

	return false
}

func DoesntEndWith(subject string, suffixes ...string) bool {
	return !EndsWith(subject, suffixes...)
}

func Is(pattern, value string, ignoreCase ...bool) bool {
	ci := len(ignoreCase) > 0 && ignoreCase[0]

	if ci {
		pattern = strings.ToLower(pattern)
		value = strings.ToLower(value)
	}

	if pattern == value {
		return true
	}

	if pattern == "*" {
		return true
	}

	re := globToRegex(pattern)
	matched, err := regexp.MatchString(re, value)

	return err == nil && matched
}

func globToRegex(pattern string) string {
	var sb strings.Builder

	sb.WriteString("^")

	for _, r := range pattern {
		switch r {
		case '*':
			sb.WriteString(`[\s\S]*`)
		case '?':
			sb.WriteString(`[\s\S]`)
		case '.', '+', '(', ')', '[', ']', '{', '}', '^', '$', '|', '\\':
			sb.WriteRune('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}

	sb.WriteString("$")

	return sb.String()
}

func IsMatch(patterns []string, value string) bool {
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, value)

		if err == nil && matched {
			return true
		}
	}

	return false
}

func IsAscii(value string) bool {
	for _, r := range value {
		if r > 127 {
			return false
		}
	}

	return true
}

func IsJson(value string) bool {
	if value == "" {
		return false
	}

	var v any

	return json.Unmarshal([]byte(value), &v) == nil
}

func IsUrl(value string, protocols ...string) bool {
	u, err := url.ParseRequestURI(value)

	if err != nil {
		return false
	}

	if u.Scheme == "" || u.Host == "" {
		return false
	}

	if len(protocols) > 0 {
		return slices.Contains(protocols, u.Scheme)
	}

	return true
}

func IsUuid(value string, version ...int) bool {
	re := regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	if !re.MatchString(value) {
		return false
	}

	if len(version) == 0 {
		return true
	}

	if version[0] < 1 || version[0] > 8 {
		return false
	}

	return int(value[14]-'0') == version[0]
}

func IsUlid(value string) bool {
	if len(value) != 26 {
		return false
	}

	re := regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)

	return re.MatchString(strings.ToUpper(value))
}

func Length(value string) int {
	return utf8.RuneCountInString(value)
}

func Limit(value string, limit int, end ...string) string {
	suffix := "..."

	if len(end) > 0 {
		suffix = end[0]
	}

	runes := []rune(value)

	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit]) + suffix
}

func Words(value string, words int, end ...string) string {
	suffix := "..."

	if len(end) > 0 {
		suffix = end[0]
	}

	if words <= 0 {
		return value
	}

	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return value
	}

	wordList := regexp.MustCompile(`\s+`).Split(trimmed, -1)

	if len(wordList) <= words {
		return value
	}

	leading := regexp.MustCompile(`^\s*`).FindString(value)

	return leading + strings.Join(wordList[:words], " ") + suffix
}

func Lower(value string) string {
	return strings.ToLower(value)
}

func Upper(value string) string {
	return strings.ToUpper(value)
}

func Reverse(value string) string {
	runes := []rune(value)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func Replace(search, replace any, subject string, caseSensitive ...bool) string {
	cs := len(caseSensitive) == 0 || caseSensitive[0]

	doReplace := func(s, r string) string {
		if cs {
			return strings.ReplaceAll(subject, s, r)
		}

		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(s))

		return re.ReplaceAllString(subject, r)
	}

	switch s := search.(type) {
	case string:
		r, _ := replace.(string)
		subject = doReplace(s, r)
	case []string:
		switch r := replace.(type) {
		case []string:
			for i, sv := range s {
				rep := ""

				if i < len(r) {
					rep = r[i]
				}

				subject = doReplace(sv, rep)
			}
		case string:
			for _, sv := range s {
				subject = doReplace(sv, r)
			}
		}
	}

	return subject
}

func ReplaceArray(search string, replacements []string, subject string) string {
	idx := 0

	for {
		pos := strings.Index(subject, search)

		if pos == -1 {
			break
		}

		rep := ""

		if idx < len(replacements) {
			rep = replacements[idx]
			idx++
		}

		subject = subject[:pos] + rep + subject[pos+len(search):]
	}

	return subject
}

func ReplaceFirst(search, replace, subject string) string {
	if search == "" {
		return subject
	}

	before, after, ok := strings.Cut(subject, search)

	if !ok {
		return subject
	}

	return before + replace + after
}

func ReplaceLast(search, replace, subject string) string {
	if search == "" {
		return subject
	}

	idx := strings.LastIndex(subject, search)

	if idx == -1 {
		return subject
	}

	return subject[:idx] + replace + subject[idx+len(search):]
}

func ReplaceStart(search, replace, subject string) string {
	if search == "" || strings.HasPrefix(subject, search) {
		return replace + subject[len(search):]
	}

	return subject
}

func ReplaceEnd(search, replace, subject string) string {
	if search != "" && strings.HasSuffix(subject, search) {
		return subject[:len(subject)-len(search)] + replace
	}

	return subject
}

func ReplaceMatches(pattern, replace, subject string, limit ...int) string {
	re, err := regexp.Compile(pattern)

	if err != nil {
		return subject
	}

	n := -1

	if len(limit) > 0 {
		n = limit[0]
	}

	if n < 0 {
		return re.ReplaceAllString(subject, replace)
	}

	count := 0

	return re.ReplaceAllStringFunc(subject, func(match string) string {
		if count >= n {
			return match
		}

		count++

		return re.ReplaceAllString(match, replace)
	})
}

func Remove(search any, subject string, caseSensitive ...bool) string {
	return Replace(search, "", subject, caseSensitive...)
}

func Repeat(str string, times int) string {
	return strings.Repeat(str, times)
}

func Squish(value string) string {
	re := regexp.MustCompile(`\s+`)

	return strings.TrimSpace(re.ReplaceAllString(value, " "))
}

func Deduplicate(str string, characters ...string) string {
	chars := " "

	if len(characters) > 0 {
		chars = characters[0]
	}

	for _, ch := range chars {
		re := regexp.MustCompile(regexp.QuoteMeta(string(ch)) + `+`)
		str = re.ReplaceAllString(str, string(ch))
	}

	return str
}

func Trim(value string, chars ...string) string {
	if len(chars) > 0 {
		return strings.Trim(value, chars[0])
	}

	return strings.TrimSpace(value)
}

func Ltrim(value string, chars ...string) string {
	if len(chars) > 0 {
		return strings.TrimLeft(value, chars[0])
	}

	return strings.TrimLeftFunc(value, unicode.IsSpace)
}

func Rtrim(value string, chars ...string) string {
	if len(chars) > 0 {
		return strings.TrimRight(value, chars[0])
	}

	return strings.TrimRightFunc(value, unicode.IsSpace)
}

func Start(value, prefix string) string {
	if strings.HasPrefix(value, prefix) {
		return value
	}

	return prefix + value
}

func Finish(value, cap string) string {
	if strings.HasSuffix(value, cap) {
		return value
	}

	return value + cap
}

func Wrap(value, before string, after ...string) string {
	a := before

	if len(after) > 0 {
		a = after[0]
	}

	return before + value + a
}

func Unwrap(value, before, after string) string {
	if after == "" {
		after = before
	}

	if strings.HasPrefix(value, before) && strings.HasSuffix(value, after) {
		value = value[len(before):]

		if strings.HasSuffix(value, after) {
			value = value[:len(value)-len(after)]
		}
	}

	return value
}

func ChopStart(subject string, needles ...string) string {
	for _, needle := range needles {
		if strings.HasPrefix(subject, needle) {
			return subject[len(needle):]
		}
	}

	return subject
}

func ChopEnd(subject string, needles ...string) string {
	for _, needle := range needles {
		if strings.HasSuffix(subject, needle) {
			return subject[:len(subject)-len(needle)]
		}
	}

	return subject
}

func Mask(str, character string, index int, length ...int) string {
	runes := []rune(str)
	runeCount := len(runes)

	if index < 0 {
		index = max(runeCount+index, 0)
	}

	end := runeCount

	if len(length) > 0 && length[0] >= 0 {
		end = index + length[0]

		if end > runeCount {
			end = runeCount
		}
	}

	maskRune, _ := utf8.DecodeRuneInString(character)

	for i := index; i < end; i++ {
		runes[i] = maskRune
	}

	return string(runes)
}

func Excerpt(text, phrase string, radius int, omission ...string) string {
	om := "..."

	if len(omission) > 0 {
		om = omission[0]
	}

	if phrase == "" {
		runes := []rune(text)

		if len(runes) <= radius {
			return text
		}

		return string(runes[:radius]) + om
	}

	idx := strings.Index(strings.ToLower(text), strings.ToLower(phrase))

	if idx == -1 {
		runes := []rune(text)

		if len(runes) <= radius {
			return text
		}

		return string(runes[:radius]) + om
	}

	runes := []rune(text)
	phraseRunes := []rune(phrase)
	phraseStart := utf8.RuneCountInString(text[:idx])
	phraseEnd := phraseStart + len(phraseRunes)

	start := max(0, phraseStart-radius)
	end := min(len(runes), phraseEnd+radius)

	var sb strings.Builder

	if start > 0 {
		sb.WriteString(om)
	}

	sb.WriteString(string(runes[start:end]))

	if end < len(runes) {
		sb.WriteString(om)
	}

	return sb.String()
}

func Match(pattern, subject string) string {
	re, err := regexp.Compile(pattern)

	if err != nil {
		return ""
	}

	match := re.FindStringSubmatch(subject)

	if len(match) == 0 {
		return ""
	}

	if len(match) > 1 {
		return match[1]
	}

	return match[0]
}

func MatchAll(pattern, subject string) []string {
	re, err := regexp.Compile(pattern)

	if err != nil {
		return nil
	}

	matches := re.FindAllStringSubmatch(subject, -1)

	var result []string

	for _, m := range matches {
		if len(m) > 1 {
			result = append(result, m[1])
		} else if len(m) > 0 {
			result = append(result, m[0])
		}
	}

	return result
}

func Numbers(value string) string {
	var sb strings.Builder

	for _, r := range value {
		if unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

func PadBoth(value string, length int, pad ...string) string {
	padStr := " "

	if len(pad) > 0 {
		padStr = pad[0]
	}

	runes := []rune(value)
	total := length - len(runes)

	if total <= 0 {
		return value
	}

	leftPad := total / 2
	rightPad := total - leftPad

	return strPad(padStr, leftPad) + value + strPad(padStr, rightPad)
}

func PadLeft(value string, length int, pad ...string) string {
	padStr := " "

	if len(pad) > 0 {
		padStr = pad[0]
	}

	runes := []rune(value)
	total := length - len(runes)

	if total <= 0 {
		return value
	}

	return strPad(padStr, total) + value
}

func PadRight(value string, length int, pad ...string) string {
	padStr := " "

	if len(pad) > 0 {
		padStr = pad[0]
	}

	runes := []rune(value)
	total := length - len(runes)

	if total <= 0 {
		return value
	}

	return value + strPad(padStr, total)
}

func strPad(pad string, n int) string {
	if pad == "" || n <= 0 {
		return ""
	}

	repeated := strings.Repeat(pad, (n/len([]rune(pad)))+1)

	return string([]rune(repeated)[:n])
}

func Position(haystack, needle string, offset ...int) (int, bool) {
	off := 0

	if len(offset) > 0 {
		off = offset[0]
	}

	runes := []rune(haystack)

	if off < 0 {
		off = len(runes) + off
	}

	if off < 0 || off >= len(runes) {
		return -1, false
	}

	sub := string(runes[off:])
	before, _, ok := strings.Cut(sub, needle)

	if !ok {
		return -1, false
	}

	return off + utf8.RuneCountInString(before), true
}

func Substr(str string, start int, length ...int) string {
	runes := []rune(str)
	runeLen := len(runes)

	if start < 0 {
		start = max(runeLen+start, 0)
	}

	if start >= runeLen {
		return ""
	}

	end := runeLen

	if len(length) > 0 {
		l := length[0]

		if l < 0 {
			end = runeLen + l
		} else {
			end = start + l
		}
	}

	if end > runeLen {
		end = runeLen
	}

	if start >= end {
		return ""
	}

	return string(runes[start:end])
}

func SubstrCount(haystack, needle string, offset ...int) int {
	if needle == "" {
		return 0
	}

	off := 0

	if len(offset) > 0 {
		off = offset[0]
	}

	if off > 0 {
		runes := []rune(haystack)

		if off >= len(runes) {
			return 0
		}

		haystack = string(runes[off:])
	}

	return strings.Count(haystack, needle)
}

func SubstrReplace(str, replace string, offset int, length ...int) string {
	runes := []rune(str)
	runeLen := len(runes)

	if offset < 0 {
		offset = max(runeLen+offset, 0)
	}

	if offset > runeLen {
		offset = runeLen
	}

	end := runeLen

	if len(length) > 0 {
		l := length[0]

		if l < 0 {
			end = max(runeLen+l, offset)
		} else {
			end = min(offset+l, runeLen)
		}
	}

	result := make([]rune, 0, runeLen)
	result = append(result, runes[:offset]...)
	result = append(result, []rune(replace)...)
	result = append(result, runes[end:]...)

	return string(result)
}

func Swap(m map[string]string, subject string) string {
	if len(m) == 0 {
		return subject
	}

	for k, v := range m {
		subject = strings.ReplaceAll(subject, k, v)
	}

	return subject
}

func Take(str string, n int) string {
	runes := []rune(str)

	if n >= 0 {
		if n >= len(runes) {
			return str
		}

		return string(runes[:n])
	}

	start := len(runes) + n

	if start < 0 {
		start = 0
	}

	return string(runes[start:])
}

func WordCount(str string) int {
	return len(strings.Fields(str))
}

func WordWrap(str string, width int, breakStr ...string) string {
	brk := "\n"

	if len(breakStr) > 0 {
		brk = breakStr[0]
	}

	words := strings.Fields(str)

	if len(words) == 0 {
		return str
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len([]rune(currentLine))+1+len([]rune(word)) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	lines = append(lines, currentLine)

	return strings.Join(lines, brk)
}

func ToBase64(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func FromBase64(value string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)

	if err != nil {

		decoded, err = base64.URLEncoding.DecodeString(value)

		if err != nil {
			return "", err
		}
	}

	return string(decoded), nil
}

func Slug(title string, separator ...string) string {
	sep := "-"

	if len(separator) > 0 {
		sep = separator[0]
	}

	title = Ascii(title)
	title = strings.ToLower(title)

	title = strings.ReplaceAll(title, "@", sep+"at"+sep)

	re := regexp.MustCompile(`[^a-z0-9]+`)
	title = re.ReplaceAllString(title, sep)

	title = strings.Trim(title, sep)

	if sep != "" {
		re2 := regexp.MustCompile(regexp.QuoteMeta(sep) + `+`)
		title = re2.ReplaceAllString(title, sep)
	}

	return title
}

func Ascii(value string, language ...string) string {
	var sb strings.Builder

	for _, r := range value {
		if r <= 127 {
			sb.WriteRune(r)

			continue
		}

		if s, ok := asciiMap[r]; ok {
			sb.WriteString(s)
		}

	}

	return sb.String()
}

func Transliterate(str string, unknown ...string) string {
	unk := "?"

	if len(unknown) > 0 {
		unk = unknown[0]
	}

	var sb strings.Builder

	for _, r := range str {
		if r <= 127 {
			sb.WriteRune(r)

			continue
		}

		if s, ok := asciiMap[r]; ok {
			sb.WriteString(s)
		} else {
			sb.WriteString(unk)
		}
	}

	return sb.String()
}

func Initials(name string, delimiter ...string) string {
	sep := ""

	if len(delimiter) > 0 {
		sep = delimiter[0]
	}

	words := strings.Fields(name)

	var initials []string

	for _, w := range words {
		r, _ := utf8.DecodeRuneInString(w)
		initials = append(initials, string(unicode.ToUpper(r)))
	}

	return strings.Join(initials, sep)
}

func CharAt(subject string, index int) string {
	runes := []rune(subject)

	if index < 0 {
		index = len(runes) + index
	}

	if index < 0 || index >= len(runes) {
		return ""
	}

	return string(runes[index])
}

func ParseCallback(callback string, def ...string) [2]string {
	d := ""

	if len(def) > 0 {
		d = def[0]
	}

	for _, sep := range []string{"@", "::"} {
		if idx := strings.LastIndex(callback, sep); idx != -1 {
			return [2]string{callback[:idx], callback[idx+len(sep):]}
		}
	}

	return [2]string{callback, d}
}

func Random(length ...int) string {
	l := 16

	if len(length) > 0 {
		l = length[0]
	}

	randomMu.Lock()
	factory := randomFactory
	sequence := randomSequence
	fallback := randomFallback
	randomMu.Unlock()

	if factory != nil {
		return factory(l)
	}

	if len(sequence) > 0 {
		randomMu.Lock()
		val := randomSequence[0]
		randomSequence = randomSequence[1:]
		randomMu.Unlock()

		return val
	}

	if fallback != nil {
		return fallback(l)
	}

	return generateRandom(l)
}

func generateRandom(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))

		if err != nil {
			panic(fmt.Errorf("str: generate random: %w", err))
		}

		result[i] = charset[n.Int64()]
	}

	return string(result)
}

func CreateRandomStringsUsing(factory func(int) string) {
	randomMu.Lock()
	randomFactory = factory
	randomMu.Unlock()
}

func CreateRandomStringsUsingSequence(sequence []string, whenMissing ...func(int) string) {
	randomMu.Lock()
	randomSequence = sequence

	if len(whenMissing) > 0 {
		randomFallback = whenMissing[0]
	}

	randomMu.Unlock()
}

func CreateRandomStringsNormally() {
	randomMu.Lock()
	randomFactory = nil
	randomSequence = nil
	randomFallback = nil
	randomMu.Unlock()
}

func Password(length ...int) (string, error) {
	l := 32

	if len(length) > 0 {
		l = length[0]
	}

	const (
		letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers = "0123456789"
		symbols = "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	)
	charset := letters + numbers + symbols

	result := make([]byte, l)

	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))

		if err != nil {
			return "", err
		}

		result[i] = charset[n.Int64()]
	}

	return string(result), nil
}

func FlushCache() {
	camelCache.Clear()
	studlyCache.Clear()
	snakeCache.Clear()
}

func ConvertCase(str string, mode int) string {
	switch mode {
	case 0:
		return strings.ToUpper(str)
	case 1:
		return strings.ToLower(str)
	case 2:
		return strings.Title(str)
	default:
		return strings.ToLower(str)
	}
}

func Of(value string) *Builder {
	return &Builder{value: value}
}

func (s *Builder) String() string { return s.value }
func (s *Builder) Value() string  { return s.value }

func (s *Builder) After(search string) *Builder {
	return &Builder{value: After(s.value, search)}
}
func (s *Builder) AfterLast(search string) *Builder {
	return &Builder{value: AfterLast(s.value, search)}
}
func (s *Builder) Before(search string) *Builder {
	return &Builder{value: Before(s.value, search)}
}
func (s *Builder) BeforeLast(search string) *Builder {
	return &Builder{value: BeforeLast(s.value, search)}
}
func (s *Builder) Between(from, to string) *Builder {
	return &Builder{value: Between(s.value, from, to)}
}
func (s *Builder) BetweenFirst(from, to string) *Builder {
	return &Builder{value: BetweenFirst(s.value, from, to)}
}
func (s *Builder) Camel() *Builder {
	return &Builder{value: Camel(s.value)}
}
func (s *Builder) Studly() *Builder {
	return &Builder{value: Studly(s.value)}
}
func (s *Builder) Pascal() *Builder {
	return &Builder{value: Pascal(s.value)}
}
func (s *Builder) Snake(delimiter ...string) *Builder {
	return &Builder{value: Snake(s.value, delimiter...)}
}
func (s *Builder) Kebab() *Builder {
	return &Builder{value: Kebab(s.value)}
}
func (s *Builder) Lower() *Builder {
	return &Builder{value: Lower(s.value)}
}
func (s *Builder) Upper() *Builder {
	return &Builder{value: Upper(s.value)}
}
func (s *Builder) Title() *Builder {
	return &Builder{value: Title(s.value)}
}
func (s *Builder) Headline() *Builder {
	return &Builder{value: Headline(s.value)}
}
func (s *Builder) Apa() *Builder {
	return &Builder{value: Apa(s.value)}
}
func (s *Builder) Ucfirst() *Builder {
	return &Builder{value: Ucfirst(s.value)}
}
func (s *Builder) Lcfirst() *Builder {
	return &Builder{value: Lcfirst(s.value)}
}
func (s *Builder) Slug(separator ...string) *Builder {
	return &Builder{value: Slug(s.value, separator...)}
}
func (s *Builder) Contains(needles ...string) bool {
	return Contains(s.value, needles...)
}
func (s *Builder) ContainsAll(needles []string) bool {
	return ContainsAll(s.value, needles)
}
func (s *Builder) StartsWith(prefixes ...string) bool {
	return StartsWith(s.value, prefixes...)
}
func (s *Builder) EndsWith(suffixes ...string) bool {
	return EndsWith(s.value, suffixes...)
}
func (s *Builder) Is(pattern string, ignoreCase ...bool) bool {
	return Is(pattern, s.value, ignoreCase...)
}
func (s *Builder) ClassBasename() *Builder {
	name := strings.Trim(s.value, "\\/")

	if idx := strings.LastIndexAny(name, "\\/"); idx >= 0 {
		name = name[idx+1:]
	}

	return &Builder{value: name}
}
func (s *Builder) IsMatch(patterns ...string) bool {
	return IsMatch(patterns, s.value)
}
func (s *Builder) IsAscii() bool { return IsAscii(s.value) }
func (s *Builder) IsJson() bool  { return IsJson(s.value) }
func (s *Builder) IsUrl(protocols ...string) bool {
	return IsUrl(s.value, protocols...)
}
func (s *Builder) IsUuid(version ...int) bool {
	return IsUuid(s.value, version...)
}
func (s *Builder) IsUlid() bool     { return IsUlid(s.value) }
func (s *Builder) IsEmpty() bool    { return s.value == "" }
func (s *Builder) IsNotEmpty() bool { return s.value != "" }
func (s *Builder) Length() int      { return Length(s.value) }
func (s *Builder) Limit(limit int, end ...string) *Builder {
	return &Builder{value: Limit(s.value, limit, end...)}
}
func (s *Builder) Words(words int, end ...string) *Builder {
	return &Builder{value: Words(s.value, words, end...)}
}
func (s *Builder) Mask(character string, index int, length ...int) *Builder {
	return &Builder{value: Mask(s.value, character, index, length...)}
}
func (s *Builder) Match(pattern string) *Builder {
	return &Builder{value: Match(pattern, s.value)}
}
func (s *Builder) MatchAll(pattern string) []string {
	return MatchAll(pattern, s.value)
}
func (s *Builder) Test(pattern string) bool {
	matched, err := regexp.MatchString(pattern, s.value)

	return err == nil && matched
}
func (s *Builder) Replace(search, replace any, caseSensitive ...bool) *Builder {
	return &Builder{value: Replace(search, replace, s.value, caseSensitive...)}
}
func (s *Builder) ReplaceFirst(search, replace string) *Builder {
	return &Builder{value: ReplaceFirst(search, replace, s.value)}
}
func (s *Builder) ReplaceLast(search, replace string) *Builder {
	return &Builder{value: ReplaceLast(search, replace, s.value)}
}
func (s *Builder) ReplaceMatches(pattern, replace string) *Builder {
	return &Builder{value: ReplaceMatches(pattern, replace, s.value)}
}
func (s *Builder) Remove(search any, caseSensitive ...bool) *Builder {
	return &Builder{value: Remove(search, s.value, caseSensitive...)}
}
func (s *Builder) Reverse() *Builder {
	return &Builder{value: Reverse(s.value)}
}
func (s *Builder) Squish() *Builder {
	return &Builder{value: Squish(s.value)}
}
func (s *Builder) Trim(chars ...string) *Builder {
	return &Builder{value: Trim(s.value, chars...)}
}
func (s *Builder) Ltrim(chars ...string) *Builder {
	return &Builder{value: Ltrim(s.value, chars...)}
}
func (s *Builder) Rtrim(chars ...string) *Builder {
	return &Builder{value: Rtrim(s.value, chars...)}
}
func (s *Builder) Start(prefix string) *Builder {
	return &Builder{value: Start(s.value, prefix)}
}
func (s *Builder) Finish(cap string) *Builder {
	return &Builder{value: Finish(s.value, cap)}
}
func (s *Builder) Wrap(before string, after ...string) *Builder {
	return &Builder{value: Wrap(s.value, before, after...)}
}
func (s *Builder) Unwrap(before, after string) *Builder {
	return &Builder{value: Unwrap(s.value, before, after)}
}
func (s *Builder) ChopStart(needles ...string) *Builder {
	return &Builder{value: ChopStart(s.value, needles...)}
}
func (s *Builder) ChopEnd(needles ...string) *Builder {
	return &Builder{value: ChopEnd(s.value, needles...)}
}
func (s *Builder) Substr(start int, length ...int) *Builder {
	return &Builder{value: Substr(s.value, start, length...)}
}
func (s *Builder) Take(n int) *Builder {
	return &Builder{value: Take(s.value, n)}
}
func (s *Builder) PadLeft(length int, pad ...string) *Builder {
	return &Builder{value: PadLeft(s.value, length, pad...)}
}
func (s *Builder) PadRight(length int, pad ...string) *Builder {
	return &Builder{value: PadRight(s.value, length, pad...)}
}
func (s *Builder) PadBoth(length int, pad ...string) *Builder {
	return &Builder{value: PadBoth(s.value, length, pad...)}
}
func (s *Builder) Swap(m map[string]string) *Builder {
	return &Builder{value: Swap(m, s.value)}
}
func (s *Builder) Append(values ...string) *Builder {
	return &Builder{value: s.value + strings.Join(values, "")}
}
func (s *Builder) Prepend(values ...string) *Builder {
	return &Builder{value: strings.Join(values, "") + s.value}
}
func (s *Builder) Numbers() *Builder {
	return &Builder{value: Numbers(s.value)}
}
func (s *Builder) Ascii(language ...string) *Builder {
	return &Builder{value: Ascii(s.value, language...)}
}
func (s *Builder) ToBase64() *Builder {
	return &Builder{value: ToBase64(s.value)}
}
func (s *Builder) FromBase64() (*Builder, error) {
	v, err := FromBase64(s.value)

	return &Builder{value: v}, err
}
func (s *Builder) Plural(count ...int) *Builder {
	return &Builder{value: Plural(s.value, count...)}
}
func (s *Builder) PluralStudly(count ...int) *Builder {
	return &Builder{value: PluralStudly(s.value, count...)}
}
func (s *Builder) PluralPascal(count ...int) *Builder {
	return &Builder{value: PluralPascal(s.value, count...)}
}
func (s *Builder) Singular() *Builder {
	return &Builder{value: Singular(s.value)}
}
func (s *Builder) Initials(delimiter ...string) *Builder {
	return &Builder{value: Initials(s.value, delimiter...)}
}
func (s *Builder) WordCount() int { return WordCount(s.value) }
func (s *Builder) WordWrap(width int, breakStr ...string) *Builder {
	return &Builder{value: WordWrap(s.value, width, breakStr...)}
}
func (s *Builder) Excerpt(phrase string, radius int, omission ...string) *Builder {
	return &Builder{value: Excerpt(s.value, phrase, radius, omission...)}
}
func (s *Builder) Ucsplit() []string { return Ucsplit(s.value) }
func (s *Builder) Deduplicate(chars ...string) *Builder {
	return &Builder{value: Deduplicate(s.value, chars...)}
}
func (s *Builder) Markdown(options ...map[string]any) *Builder {
	return &Builder{value: Markdown(s.value, options...)}
}
func (s *Builder) InlineMarkdown(options ...map[string]any) *Builder {
	return &Builder{value: InlineMarkdown(s.value, options...)}
}
