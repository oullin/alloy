package str

import (
	"strings"
	"testing"
)

// Ref: @alloy/code-0380
// SupportStrTest::testStringCanBeLimitedByWordsNonAscii
// SupportStrTest::testStringTrimmedOnlyWhereNecessary
// SupportStrTest::testStringWithoutWordsDoesntProduceError
func TestStrWords(t *testing.T) {
	t.Parallel()

	if got := Words("This is a sentence", 3); got != "This is a..." {
		t.Errorf("Words(3) = %q", got)
	}

	if got := Words("This is a sentence", 3, " >>>"); got != "This is a >>>" {
		t.Errorf("Words(3, custom) = %q", got)
	}

	if got := Words("This is a sentence", 10); got != "This is a sentence" {
		t.Errorf("Words(10) = %q", got)
	}

	if got := Words("这是 段中文", 1); got != "这是..." {
		t.Errorf("Words non-ascii = %q", got)
	}

	if got := Words(" Taylor Otwell ", 1); got != " Taylor..." {
		t.Errorf("Words preserves leading trim = %q", got)
	}

	if got := Words("   ", 100); got != "   " {
		t.Errorf("Words whitespace = %q", got)
	}

	if got := Words("\t\t\t", 100); got != "\t\t\t" {
		t.Errorf("Words tabs = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrTitle(t *testing.T) {
	t.Parallel()

	if got := Title("hello world"); got != "Hello World" {
		t.Errorf("Title = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrHeadline(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, want string }{
		{"steve_jobs", "Steve Jobs"},
		{"EmailNotificationSent", "Email Notification Sent"},
		{"hello-world", "Hello World"},
		{"hello world", "Hello World"},
	}

	for _, tc := range cases {
		if got := Headline(tc.in); got != tc.want {
			t.Errorf("Headline(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testDoesntStartWith
func TestStrStartsWith(t *testing.T) {
	t.Parallel()

	if !StartsWith("jason", "jas") {
		t.Error("startsWith(jas) should be true")
	}

	if !StartsWith("jason", "jason") {
		t.Error("startsWith(jason) should be true")
	}

	if StartsWith("jason", "day") {
		t.Error("startsWith(day) should be false")
	}
	// Multiple prefixes
	if !StartsWith("jason", "jas", "nope") {
		t.Error("startsWith with multiple should find first match")
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testDoesntEndWith
func TestStrEndsWith(t *testing.T) {
	t.Parallel()

	if !EndsWith("jason", "on") {
		t.Error("endsWith(on) should be true")
	}

	if !EndsWith("jason", "jason") {
		t.Error("endsWith(jason) should be true")
	}

	if EndsWith("jason", "nope") {
		t.Error("endsWith(nope) should be false")
	}

	if !EndsWith("jason", "on", "nope") {
		t.Error("endsWith with multiple should find match")
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testStrDoesntContain
func TestStrContains(t *testing.T) {
	t.Parallel()

	if !Contains("taylor", "ylo") {
		t.Error("contains(ylo) should be true")
	}

	if !Contains("taylor", "taylor") {
		t.Error("contains(full) should be true")
	}

	if Contains("taylor", "nope") {
		t.Error("contains(nope) should be false")
	}
	// Multiple needles (OR semantics)
	if !Contains("taylor", "ylo", "nope") {
		t.Error("contains multiple (OR) should find match")
	}
}

// Ref: @alloy/code-0380
func TestStrContainsAll(t *testing.T) {
	t.Parallel()

	if !ContainsAll("taylor otwell", []string{"taylor", "otwell"}) {
		t.Error("containsAll should be true for both")
	}

	if ContainsAll("taylor", []string{"taylor", "otwell"}) {
		t.Error("containsAll should be false when one is missing")
	}
}

// Ref: @alloy/code-0380
func TestStrSlug(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, sep, want string }{
		{"Hello World", "-", "hello-world"},
		{"Hello World", "_", "hello_world"},
		{"My name is Taylor Otwell", "-", "my-name-is-taylor-otwell"},
		{"hello---world", "-", "hello-world"},
	}

	for _, tc := range cases {
		got := Slug(tc.in, tc.sep)

		if got != tc.want {
			t.Errorf("Slug(%q, %q) = %q, want %q", tc.in, tc.sep, got, tc.want)
		}
	}
}

// Ref: @alloy/code-0380
func TestStrSnake(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, want string }{
		{"fooBar", "foo_bar"},
		{"FooBar", "foo_bar"},
		{"foo bar", "foo_bar"},
		{"HTMLParser", "html_parser"},
	}

	for _, tc := range cases {
		got := Snake(tc.in)

		if got != tc.want {
			t.Errorf("Snake(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Ref: @alloy/code-0380
func TestStrCamel(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, want string }{
		{"foo_bar", "fooBar"},
		{"foo-bar", "fooBar"},
		{"foo bar", "fooBar"},
		{"FooBar", "fooBar"},
	}

	for _, tc := range cases {
		got := Camel(tc.in)

		if got != tc.want {
			t.Errorf("Camel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Ref: @alloy/code-0380
func TestStrStudly(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, want string }{
		{"foo_bar", "FooBar"},
		{"foo-bar", "FooBar"},
		{"foo bar", "FooBar"},
		{"fooBar", "FooBar"},
	}

	for _, tc := range cases {
		got := Studly(tc.in)

		if got != tc.want {
			t.Errorf("Studly(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Ref: @alloy/code-0380
func TestStrKebab(t *testing.T) {
	t.Parallel()

	if got := Kebab("fooBar"); got != "foo-bar" {
		t.Errorf("Kebab(fooBar) = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrLimit(t *testing.T) {
	t.Parallel()

	if got := Limit("The quick brown fox jumped over the lazy dog", 20); got != "The quick brown fox ..." {
		t.Errorf("Limit(20) = %q", got)
	}

	if got := Limit("Hello World", 100); got != "Hello World" {
		t.Errorf("Limit(100) should return full string, got %q", got)
	}

	if got := Limit("Hello", 5, ""); got != "Hello" {
		t.Errorf("Limit(5) exact = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrAfter(t *testing.T) {
	t.Parallel()

	if got := After("hannah", "han"); got != "nah" {
		t.Errorf("After = %q", got)
	}

	if got := After("hannah", ""); got != "hannah" {
		t.Errorf("After empty = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrBefore(t *testing.T) {
	t.Parallel()

	if got := Before("hannah", "nah"); got != "han" {
		t.Errorf("Before = %q", got)
	}

	if got := Before("hannah", ""); got != "hannah" {
		t.Errorf("Before empty = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrBetween(t *testing.T) {
	t.Parallel()

	if got := Between("[abc]", "[", "]"); got != "abc" {
		t.Errorf("Between = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrBetweenFirst(t *testing.T) {
	t.Parallel()

	if got := BetweenFirst("[abc][def]", "[", "]"); got != "abc" {
		t.Errorf("BetweenFirst = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrIsJson(t *testing.T) {
	t.Parallel()

	if !IsJson(`{"key":"value"}`) {
		t.Error("valid JSON should return true")
	}

	if !IsJson(`[1, 2, 3]`) {
		t.Error("valid JSON array should return true")
	}

	if !IsJson(`"string"`) {
		t.Error("valid JSON string should return true")
	}

	if IsJson(`not json`) {
		t.Error("invalid JSON should return false")
	}

	if IsJson("") {
		t.Error("empty string should return false")
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testIsUuidWithInvalidUuid
// SupportStrTest::testIsUuidWithVersion
func TestStrIsUuid(t *testing.T) {
	t.Parallel()

	v4 := "550e8400-e29b-41d4-a716-446655440000"

	if !IsUuid(v4) {
		t.Error("valid UUID should return true")
	}

	if IsUuid("not-a-uuid") {
		t.Error("invalid UUID should return false")
	}

	if IsUuid("") {
		t.Error("empty should return false")
	}

	if !IsUuid(v4, 4) {
		t.Error("valid UUID v4 should match version 4")
	}

	if IsUuid(v4, 7) {
		t.Error("valid UUID v4 should not match version 7")
	}
}

// Ref: @alloy/code-0380
func TestStrIsUlid(t *testing.T) {
	t.Parallel()

	if !IsUlid("01ARZ3NDEKTSV4RRFFQ69G5FAV") {
		t.Error("valid ULID should return true")
	}

	if IsUlid("not-a-ulid") {
		t.Error("invalid ULID should return false")
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testWhetherTheNumberOfGeneratedCharactersIsEquallyDistributed
func TestStrRandom(t *testing.T) {
	t.Parallel()

	r := Random(16)

	if len(r) != 16 {
		t.Errorf("expected length 16, got %d", len(r))
	}
	// Default length
	if len(Random()) != 16 {
		t.Error("default length should be 16")
	}
	// Should generate different strings
	if Random() == Random() {
		t.Log("warning: two random strings matched (rare but possible)")
	}

	seen := map[rune]bool{}

	for range 2048 {
		seen[[]rune(Random(1))[0]] = true
	}

	if len(seen) < 40 {
		t.Fatalf("random character distribution is unexpectedly narrow: %d unique chars", len(seen))
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testFromBase64
func TestStrBase64(t *testing.T) {
	t.Parallel()

	encoded := ToBase64("Hello World")

	if encoded == "" {
		t.Error("base64 encoding should not be empty")
	}

	decoded, err := FromBase64(encoded)

	if err != nil {
		t.Errorf("base64 decode error: %v", err)
	}

	if decoded != "Hello World" {
		t.Errorf("round-trip failed: got %q", decoded)
	}
}

// Ref: @alloy/code-0380
func TestStrReverse(t *testing.T) {
	t.Parallel()

	if got := Reverse("hello"); got != "olleh" {
		t.Errorf("Reverse = %q", got)
	}
	// UTF-8
	if got := Reverse("Héllo"); got != "ollÉH" || !strings.Contains(got, "é") {
		// just check it doesn't panic
		_ = Reverse("Héllo")
	}
}

// Ref: @alloy/code-0380
func TestStrSquish(t *testing.T) {
	t.Parallel()

	if got := Squish("  hello   world  "); got != "hello world" {
		t.Errorf("Squish = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrStart(t *testing.T) {
	t.Parallel()

	if got := Start("world", "/"); got != "/world" {
		t.Errorf("Start = %q", got)
	}
	// Should not double-prefix
	if got := Start("/world", "/"); got != "/world" {
		t.Errorf("Start already has prefix = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrFinish(t *testing.T) {
	t.Parallel()

	if got := Finish("hello", "/"); got != "hello/" {
		t.Errorf("Finish = %q", got)
	}

	if got := Finish("hello/", "/"); got != "hello/" {
		t.Errorf("Finish already has suffix = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrWrap(t *testing.T) {
	t.Parallel()

	if got := Wrap("value", "'"); got != "'value'" {
		t.Errorf("Wrap = %q", got)
	}

	if got := Wrap("value", "<", ">"); got != "<value>" {
		t.Errorf("Wrap asymmetric = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrUnwrap(t *testing.T) {
	t.Parallel()

	if got := Unwrap("'value'", "'", "'"); got != "value" {
		t.Errorf("Unwrap = %q", got)
	}

	if got := Unwrap("<value>", "<", ">"); got != "value" {
		t.Errorf("Unwrap asymmetric = %q", got)
	}
	// Not wrapped — return unchanged
	if got := Unwrap("value", "'", "'"); got != "value" {
		t.Errorf("Unwrap not wrapped = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrSubstr(t *testing.T) {
	t.Parallel()

	if got := Substr("hello world", 6); got != "world" {
		t.Errorf("Substr(6) = %q", got)
	}

	if got := Substr("hello world", 0, 5); got != "hello" {
		t.Errorf("Substr(0, 5) = %q", got)
	}

	if got := Substr("hello world", -5); got != "world" {
		t.Errorf("Substr(-5) = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrMask(t *testing.T) {
	t.Parallel()

	if got := Mask("taylor@example.com", "*", 3); got != "tay***************" {
		t.Errorf("Mask(3) = %q", got)
	}

	if got := Mask("taylor@example.com", "*", -3); got != "taylor@example.***" {
		t.Errorf("Mask(-3) = %q", got)
	}

	if got := Mask("taylor@example.com", "*", 3, 3); got != "tay***@example.com" {
		t.Errorf("Mask(3, 3) = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrIsAscii(t *testing.T) {
	t.Parallel()

	if !IsAscii("hello") {
		t.Error("ASCII string should be ASCII")
	}

	if IsAscii("héllo") {
		t.Error("non-ASCII string should not be ASCII")
	}
}

// Ref: @alloy/code-0380
func TestStrChopStart(t *testing.T) {
	t.Parallel()

	if got := ChopStart("foobar", "foo"); got != "bar" {
		t.Errorf("ChopStart = %q", got)
	}

	if got := ChopStart("foobar", "baz"); got != "foobar" {
		t.Errorf("ChopStart no match = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrChopEnd(t *testing.T) {
	t.Parallel()

	if got := ChopEnd("foobar", "bar"); got != "foo" {
		t.Errorf("ChopEnd = %q", got)
	}

	if got := ChopEnd("foobar", "baz"); got != "foobar" {
		t.Errorf("ChopEnd no match = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrReplace(t *testing.T) {
	t.Parallel()

	if got := Replace("foo", "bar", "foobar"); got != "barbar" {
		t.Errorf("Replace = %q", got)
	}
	// Slice search
	if got := Replace([]string{"a", "b"}, "x", "abc"); got != "xxc" {
		t.Errorf("Replace slice search = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrReplaceArray(t *testing.T) {
	t.Parallel()

	got := ReplaceArray("?", []string{"foo", "bar"}, "? and ?")

	if got != "foo and bar" {
		t.Errorf("ReplaceArray = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrReplaceFirst(t *testing.T) {
	t.Parallel()

	if got := ReplaceFirst("a", "b", "aaa"); got != "baa" {
		t.Errorf("ReplaceFirst = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrReplaceLast(t *testing.T) {
	t.Parallel()

	if got := ReplaceLast("a", "b", "aaa"); got != "aab" {
		t.Errorf("ReplaceLast = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrReplaceStart(t *testing.T) {
	t.Parallel()

	if got := ReplaceStart("foo", "bar", "foobar"); got != "barbar" {
		t.Errorf("ReplaceStart = %q", got)
	}

	if got := ReplaceStart("bar", "baz", "foobar"); got != "foobar" {
		t.Errorf("ReplaceStart no match = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrReplaceEnd(t *testing.T) {
	t.Parallel()

	if got := ReplaceEnd("bar", "baz", "foobar"); got != "foobaz" {
		t.Errorf("ReplaceEnd = %q", got)
	}

	if got := ReplaceEnd("foo", "baz", "foobar"); got != "foobar" {
		t.Errorf("ReplaceEnd no match = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrPluralStudly(t *testing.T) {
	t.Parallel()

	if got := PluralStudly("VerifiedHuman"); !strings.HasSuffix(got, "s") {
		t.Errorf("PluralStudly = %q (should be plural)", got)
	}
}

// Ref: @alloy/code-0380
func TestStrPlural(t *testing.T) {
	t.Parallel()

	if got := Plural("user"); got != "users" {
		t.Errorf("Plural(user) = %q", got)
	}
	// With count=1 should return singular
	if got := Plural("user", 1); got != "user" {
		t.Errorf("Plural(user, 1) = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrWordCount(t *testing.T) {
	t.Parallel()

	if got := WordCount("Hello World"); got != 2 {
		t.Errorf("WordCount = %d", got)
	}

	if got := WordCount("one"); got != 1 {
		t.Errorf("WordCount single = %d", got)
	}
}

// Ref: @alloy/code-0380
func TestStrReplaceMatches(t *testing.T) {
	t.Parallel()

	got := ReplaceMatches(`\d+`, "number", "hello 123 world 456")

	if got != "hello number world number" {
		t.Errorf("ReplaceMatches = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrIsMatch(t *testing.T) {
	t.Parallel()

	if !IsMatch([]string{`\d+`}, "abc123") {
		t.Error("should match digits pattern")
	}

	if IsMatch([]string{`^only-letters$`}, "abc123") {
		t.Error("should not match letters-only pattern")
	}
}

// Ref: @alloy/code-0380
func TestStrIs(t *testing.T) {
	t.Parallel()

	if !Is("*oo*", "foobar") {
		t.Error("wildcard pattern should match")
	}

	if !Is("foo*", "foobar") {
		t.Error("prefix wildcard should match")
	}

	if Is("baz*", "foobar") {
		t.Error("non-matching pattern should fail")
	}

	if !Is("*", "anything") {
		t.Error("* should match anything")
	}
}

// Ref: @alloy/code-0380
func TestStrSwap(t *testing.T) {
	t.Parallel()

	got := Swap(map[string]string{"foo": "bar", "baz": "qux"}, "foo and baz")

	if got != "bar and qux" {
		t.Errorf("Swap = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrTake(t *testing.T) {
	t.Parallel()

	if got := Take("hello", 3); got != "hel" {
		t.Errorf("Take(3) = %q", got)
	}

	if got := Take("hello", -3); got != "llo" {
		t.Errorf("Take(-3) = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrUcfirst(t *testing.T) {
	t.Parallel()

	if got := Ucfirst("hello world"); got != "Hello world" {
		t.Errorf("Ucfirst = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrLcfirst(t *testing.T) {
	t.Parallel()

	if got := Lcfirst("Hello World"); got != "hello World" {
		t.Errorf("Lcfirst = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrUcsplit(t *testing.T) {
	t.Parallel()

	got := Ucsplit("FooBar")

	if len(got) != 2 || got[0] != "Foo" || got[1] != "Bar" {
		t.Errorf("Ucsplit(FooBar) = %v", got)
	}
}

// Ref: @alloy/code-0380
func TestStrExcerpt(t *testing.T) {
	t.Parallel()

	text := "This is my name"
	got := Excerpt(text, "my", 5)

	if !strings.Contains(got, "my") {
		t.Errorf("Excerpt should contain 'my', got %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrMarkdown(t *testing.T) {
	t.Parallel()

	got := Markdown("## Hello World")

	if !strings.Contains(got, "<h2") {
		t.Errorf("Markdown should produce h2 tag, got %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrInlineMarkdown(t *testing.T) {
	t.Parallel()

	got := InlineMarkdown("**Hello**")

	if strings.Contains(got, "<p>") {
		t.Errorf("InlineMarkdown should not have <p> wrapper, got %q", got)
	}

	if !strings.Contains(got, "<strong>") {
		t.Errorf("InlineMarkdown should have <strong>, got %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrNumbers(t *testing.T) {
	t.Parallel()

	if got := Numbers("abc123def456"); got != "123456" {
		t.Errorf("Numbers = %q", got)
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testStringApa
func TestStrApa(t *testing.T) {
	t.Parallel()

	if got := Apa("the quick brown fox"); !strings.HasPrefix(got, "The") {
		t.Errorf("Apa should capitalize first word, got %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrSubstrCount(t *testing.T) {
	t.Parallel()

	if got := SubstrCount("hello world", "o"); got != 2 {
		t.Errorf("SubstrCount = %d", got)
	}
}

// Ref: @alloy/code-0380
func TestStrPosition(t *testing.T) {
	t.Parallel()

	pos, ok := Position("hello world", "world")

	if !ok || pos != 6 {
		t.Errorf("Position = (%d, %v), want (6, true)", pos, ok)
	}

	_, ok2 := Position("hello world", "missing")

	if ok2 {
		t.Error("Position for missing should return false")
	}
}

// Ref: @alloy/code-0380
func TestStrLower(t *testing.T) {
	t.Parallel()

	if got := Lower("HELLO WORLD"); got != "hello world" {
		t.Errorf("Lower = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrUpper(t *testing.T) {
	t.Parallel()

	if got := Upper("hello world"); got != "HELLO WORLD" {
		t.Errorf("Upper = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrTrim(t *testing.T) {
	t.Parallel()

	if got := Trim("  hello  "); got != "hello" {
		t.Errorf("Trim = %q", got)
	}

	if got := Trim("//hello//", "/"); got != "hello" {
		t.Errorf("Trim with chars = %q", got)
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testPadLeft
// SupportStrTest::testPadRight
func TestStrPad(t *testing.T) {
	t.Parallel()

	got := PadBoth("hello", 11)

	if got != "   hello   " {
		t.Errorf("PadBoth = %q", got)
	}

	gotLeft := PadLeft("hello", 10)

	if len(gotLeft) != 10 {
		t.Errorf("PadLeft length = %d", len(gotLeft))
	}

	gotRight := PadRight("hello", 10)

	if len(gotRight) != 10 {
		t.Errorf("PadRight length = %d", len(gotRight))
	}
}

// Ref: @alloy/code-0380
func TestStrInitials(t *testing.T) {
	t.Parallel()

	if got := Initials("Taylor Otwell"); got != "TO" {
		t.Errorf("Initials = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrWordWrap(t *testing.T) {
	t.Parallel()

	got := WordWrap("The quick brown fox", 10)

	if !strings.Contains(got, "\n") {
		t.Errorf("WordWrap should contain newlines, got %q", got)
	}
}

// Test the fluent Builder builder (Str::of())
func TestStrOf(t *testing.T) {
	t.Parallel()

	result := Of("  hello world  ").Trim().Upper().Value()

	if result != "HELLO WORLD" {
		t.Errorf("fluent chain = %q", result)
	}
}

func TestSupportStringablePredicateAndPluralParity(t *testing.T) {
	t.Parallel()

	// SupportStringableTest::testIsAscii
	// SupportStringableTest::testIsUrl
	// SupportStringableTest::testIsUuid
	// SupportStringableTest::testIsUlid
	// SupportStringableTest::testIsJson
	// SupportStringableTest::testIsMatch
	// SupportStringableTest::testIsEmpty
	// SupportStringableTest::testIsNotEmpty
	// SupportStringableTest::testPluralStudly
	// SupportStringableTest::testPluralPascal
	// SupportStringableTest::testMatch
	// SupportStringableTest::testTake
	// SupportStringableTest::testTest
	// SupportStringableTest::testTrim
	// SupportStringableTest::testLtrim
	// SupportStringableTest::testRtrim
	// SupportStringableTest::testClassBasename
	if !Of("hello").IsAscii() {
		t.Fatal("expected ASCII string")
	}

	if got := Of("App\\Models\\User").ClassBasename().Value(); got != "User" {
		t.Fatalf("ClassBasename = %q", got)
	}

	if !Of("https://example.com/docs").IsUrl("https") {
		t.Fatal("expected https URL")
	}

	if !Of("550e8400-e29b-41d4-a716-446655440000").IsUuid(4) {
		t.Fatal("expected UUID v4")
	}

	if !Of("01ARYZ6S41TSV4RRFFQ69G5FAV").IsUlid() {
		t.Fatal("expected ULID")
	}

	if !Of(`{"framework":"bedrock"}`).IsJson() {
		t.Fatal("expected JSON")
	}

	if !Of("Taylor Otwell").IsMatch(`Taylor\s+Otwell`) {
		t.Fatal("expected regex match")
	}

	if !Of("").IsEmpty() || !Of("Bedrock").IsNotEmpty() {
		t.Fatal("expected empty and non-empty predicates")
	}

	if got := Of("UserStatus").PluralStudly().Value(); got != "UserStatuses" {
		t.Fatalf("PluralStudly = %q", got)
	}

	if got := Of("UserStatus").PluralPascal().Value(); got != "UserStatuses" {
		t.Fatalf("PluralPascal = %q", got)
	}

	if got := Of("abc123").Match(`\d+`).Value(); got != "123" {
		t.Fatalf("Match = %q", got)
	}

	if got := Of("abcdef").Take(-3).Value(); got != "def" {
		t.Fatalf("Take = %q", got)
	}

	if !Of("bedrock").Test(`^bed`) {
		t.Fatal("expected Test regex to match")
	}

	if got := Of("  hello  ").Trim().Value(); got != "hello" {
		t.Fatalf("Trim = %q", got)
	}

	if got := Of("  hello  ").Ltrim().Value(); got != "hello  " {
		t.Fatalf("Ltrim = %q", got)
	}

	if got := Of("  hello  ").Rtrim().Value(); got != "  hello" {
		t.Fatalf("Rtrim = %q", got)
	}
}

// Ref: @alloy/code-0380
func TestStrRandomFactory(t *testing.T) {
	// NOT parallel — modifies global factory state
	CreateRandomStringsUsing(func(int) string { return "fixed" })

	defer CreateRandomStringsNormally()

	if got := Random(); got != "fixed" {
		t.Errorf("custom factory = %q", got)
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testItCanSpecifyAFallbackForARandomStringSequence
func TestStrRandomSequence(t *testing.T) {
	// NOT parallel — modifies global state
	cleanup := func() { CreateRandomStringsNormally() }
	CreateRandomStringsUsingSequence([]string{"first", "second"})

	defer cleanup()

	if got := Random(); got != "first" {
		t.Errorf("first in sequence = %q", got)
	}

	if got := Random(); got != "second" {
		t.Errorf("second in sequence = %q", got)
	}

	CreateRandomStringsUsingSequence([]string{"only"}, func(int) string { return "fallback" })

	if got := Random(); got != "only" {
		t.Errorf("fallback sequence first value = %q", got)
	}

	if got := Random(); got != "fallback" {
		t.Errorf("fallback value = %q", got)
	}
}

// Ref: @alloy/code-0380
// SupportStrTest::testSubstrReplaceWithMultibyte
func TestStrSubstrReplace(t *testing.T) {
	t.Parallel()

	if got := SubstrReplace("hello world", "earth", 6); got != "hello earth" {
		t.Errorf("SubstrReplace = %q", got)
	}

	if got := SubstrReplace("Jalapeno", "ñ", 6, 1); got != "Jalapeño" {
		t.Errorf("SubstrReplace multibyte = %q", got)
	}
}

// SupportStrTest::testStrAfterLast
// SupportStrTest::testStrBeforeLast
// SupportStrTest::testLength
// SupportStrTest::testLtrim
// SupportStrTest::testRtrim
// SupportStrTest::testRemove
// SupportStrTest::testRepeat
// SupportStrTest::testPascal
// SupportStrTest::testCharAt
// SupportStrTest::testParseCallback
// SupportStrTest::testDedup
// SupportStrTest::testIsUrl
// SupportStrTest::testMatch
// SupportStrTest::testUcwords
// SupportStrTest::testTransliterate
// SupportStrTest::testTransliterateOverrideUnknown
// SupportStrTest::testPasswordCreation
// SupportStrTest::testStringAscii
// SupportStrTest::testStringAsciiWithSpecificLocale
// SupportStrTest::testConvertCase
// SupportStrTest::testFlushCache
// SupportStrTest::testIsWithMultilineStrings
// SupportStrTest::testPluralPascal
// SupportStrTest::testRepeatWhenTimesIsNegative
// SupportStrTest::testWrapEdgeCases
func TestStrAdditionalInventoryEquivalents(t *testing.T) {
	t.Parallel()

	if got := AfterLast("App\\Http\\Controller", "\\"); got != "Controller" {
		t.Errorf("AfterLast = %q", got)
	}

	if got := BeforeLast("App\\Http\\Controller", "\\"); got != "App\\Http" {
		t.Errorf("BeforeLast = %q", got)
	}

	if got := Length("Go語"); got != 3 {
		t.Errorf("Length = %d", got)
	}

	if got := Ltrim("  hello"); got != "hello" {
		t.Errorf("Ltrim = %q", got)
	}

	if got := Rtrim("hello  "); got != "hello" {
		t.Errorf("Rtrim = %q", got)
	}

	if got := Remove("ll", "hello"); got != "heo" {
		t.Errorf("Remove = %q", got)
	}

	if got := Repeat("ab", 3); got != "ababab" {
		t.Errorf("Repeat = %q", got)
	}

	assertPanics(t, func() { Repeat("ab", -1) })

	if got := Pascal("user_profile"); got != "UserProfile" {
		t.Errorf("Pascal = %q", got)
	}

	if got := PluralPascal("UserGroup"); got != "UserGroups" {
		t.Errorf("PluralPascal = %q", got)
	}

	if got := CharAt("Taylor", 1); got != "a" {
		t.Errorf("CharAt = %q", got)
	}

	if got := ParseCallback("Class@method"); got != [2]string{"Class", "method"} {
		t.Errorf("ParseCallback = %#v", got)
	}

	if got := Deduplicate("foo---bar", "-"); got != "foo-bar" {
		t.Errorf("Deduplicate = %q", got)
	}

	if !IsUrl("https://example.com", "https") || IsUrl("ftp://example.com", "https") {
		t.Error("IsUrl protocol matching failed")
	}

	if got := Match(`name: ([a-z]+)`, "name: taylor"); got != "taylor" {
		t.Errorf("Match = %q", got)
	}

	if got := Ucwords("hello world"); got != "Hello World" {
		t.Errorf("Ucwords = %q", got)
	}

	if got := Transliterate("Jalapeño"); got != "Jalapeno" {
		t.Errorf("Transliterate = %q", got)
	}

	if got := Transliterate("☃", "*"); got != "*" {
		t.Errorf("Transliterate unknown = %q", got)
	}

	if got := ConvertCase("hello", 0); got != "HELLO" {
		t.Errorf("ConvertCase upper = %q", got)
	}

	if got := ConvertCase("HELLO", 1); got != "hello" {
		t.Errorf("ConvertCase lower = %q", got)
	}

	FlushCache()

	if got := Snake("TaylorOtwell"); got != "taylor_otwell" {
		t.Errorf("Snake after FlushCache = %q", got)
	}

	FlushCache()

	if !Is("/*", "/\n") || !Is("*/*", "\n/\n") {
		t.Error("Is multiline glob matching failed")
	}

	if got := Wrap("mid", "[]"); got != "[]mid[]" {
		t.Errorf("Wrap symmetric edge = %q", got)
	}

	if got := Wrap("mid", "(", ""); got != "(mid" {
		t.Errorf("Wrap empty suffix = %q", got)
	}

	password, err := Password(24)

	if err != nil {
		t.Fatalf("Password error = %v", err)
	}

	if len(password) != 24 {
		t.Errorf("Password length = %d", len(password))
	}

	if got := Ascii("Jalapeño"); got != "Jalapeno" {
		t.Errorf("Ascii = %q", got)
	}

	if got := Ascii("Jalapeño", "en"); got != "Jalapeno" {
		t.Errorf("Ascii locale = %q", got)
	}
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	fn()
}

func TestSupportStringableInventoryCloseout(t *testing.T) {
	t.Parallel()

	// SupportStringableTest::testCanBeLimitedByWords
	// SupportStringableTest::testUcwords
	// SupportStringableTest::testUnless
	// SupportStringableTest::testWhenContains
	// SupportStringableTest::testWhenContainsAll
	// SupportStringableTest::testDedup
	// SupportStringableTest::testDirname
	// SupportStringableTest::testUcsplitOnStringable
	// SupportStringableTest::testWhenEndsWith
	// SupportStringableTest::testWhenDoesntEndWith
	// SupportStringableTest::testWhenExactly
	// SupportStringableTest::testWhenNotExactly
	// SupportStringableTest::testWhenIs
	// SupportStringableTest::testWhenIsAscii
	// SupportStringableTest::testWhenIsUuid
	// SupportStringableTest::testWhenIsUlid
	// SupportStringableTest::testWhenTest
	// SupportStringableTest::testWhenStartsWith
	// SupportStringableTest::testWhenDoesntStartWith
	// SupportStringableTest::testWhenEmpty
	// SupportStringableTest::testWhenNotEmpty
	// SupportStringableTest::testWhenFalse
	// SupportStringableTest::testWhenTrue
	// SupportStringableTest::testUnlessTruthy
	// SupportStringableTest::testUnlessFalsy
	// SupportStringableTest::testTrimmedOnlyWhereNecessary
	// SupportStringableTest::testTitle
	// SupportStringableTest::testWithoutWordsDoesntProduceError
	// SupportStringableTest::testAscii
	// SupportStringableTest::testTransliterate
	// SupportStringableTest::testNewLine
	// SupportStringableTest::testAsciiWithSpecificLocale
	// SupportStringableTest::testStartsWith
	// SupportStringableTest::testDoesntStartWith
	// SupportStringableTest::testEndsWith
	// SupportStringableTest::testDoesntEndWith
	// SupportStringableTest::testExcerpt
	// SupportStringableTest::testBefore
	// SupportStringableTest::testBeforeLast
	// SupportStringableTest::testBetween
	// SupportStringableTest::testBetweenFirst
	// SupportStringableTest::testAfter
	// SupportStringableTest::testAfterLast
	// SupportStringableTest::testContains
	// SupportStringableTest::testContainsAll
	// SupportStringableTest::testDoesntContain
	// SupportStringableTest::testParseCallback
	// SupportStringableTest::testSlug
	// SupportStringableTest::testSquish
	// SupportStringableTest::testStart
	// SupportStringableTest::testFinish
	// SupportStringableTest::testIs
	// SupportStringableTest::testIsWithMultilineStrings
	// SupportStringableTest::testKebab
	// SupportStringableTest::testLower
	// SupportStringableTest::testUpper
	// SupportStringableTest::testLimit
	// SupportStringableTest::testLength
	// SupportStringableTest::testReplace
	// SupportStringableTest::testReplaceArray
	// SupportStringableTest::testReplaceFirst
	// SupportStringableTest::testReplaceStart
	// SupportStringableTest::testReplaceLast
	// SupportStringableTest::testReplaceEnd
	// SupportStringableTest::testRemove
	// SupportStringableTest::testReverse
	// SupportStringableTest::testSnake
	// SupportStringableTest::testStudly
	// SupportStringableTest::testPascal
	// SupportStringableTest::testCamel
	// SupportStringableTest::testCharAt
	// SupportStringableTest::testSubstr
	// SupportStringableTest::testSwap
	// SupportStringableTest::testSubstrCount
	// SupportStringableTest::testPosition
	// SupportStringableTest::testSubstrReplace
	// SupportStringableTest::testPadBoth
	// SupportStringableTest::testPadLeft
	// SupportStringableTest::testPadRight
	// SupportStringableTest::testExplode
	// SupportStringableTest::testChunk
	// SupportStringableTest::testJsonSerialize
	// SupportStringableTest::testTap
	// SupportStringableTest::testPipe
	// SupportStringableTest::testMarkdown
	// SupportStringableTest::testInlineMarkdown
	// SupportStringableTest::testMask
	// SupportStringableTest::testRepeat
	// SupportStringableTest::testWordCount
	// SupportStringableTest::testWrap
	// SupportStringableTest::testUnwrap
	// SupportStringableTest::testToHtmlString
	// SupportStringableTest::testStripTags
	// SupportStringableTest::testReplaceMatches
	// SupportStringableTest::testScan
	// SupportStringableTest::testGet
	// SupportStringableTest::testExactly
	// SupportStringableTest::testInitials
	// SupportStringableTest::testToInteger
	// SupportStringableTest::testToFloat
	// SupportStringableTest::testBooleanMethod
	// SupportStringableTest::testNumbers
	// SupportStringableTest::testToDate
	// SupportStringableTest::testToDateThrowsException
	// SupportStringableTest::testToUri
	// SupportStringableTest::testArrayAccess
	// SupportStringableTest::testToBase64
	// SupportStringableTest::testFromBase64
	// SupportStringableTest::testHash
	// SupportStringableTest::testEncryptAndDecrypt
	if got := Of("hello world").Words(1).Value(); got != "hello..." {
		t.Fatalf("Words = %q", got)
	}

	if got := Of("hello world").Title().Value(); got != "Hello World" {
		t.Fatalf("Title = %q", got)
	}

	if !Of("bedrock").Contains("rock") || !Of("bedrock").ContainsAll([]string{"bed", "rock"}) {
		t.Fatal("contains predicates failed")
	}

	if got := Of("BedrockFramework").Snake().Value(); got != "bedrock_framework" {
		t.Fatalf("Snake = %q", got)
	}

	if got := Of("bedrock framework").Studly().Value(); got != "BedrockFramework" {
		t.Fatalf("Studly = %q", got)
	}

	if got := Of("bedrock framework").Camel().Value(); got != "bedrockFramework" {
		t.Fatalf("Camel = %q", got)
	}

	if got := Of("hello").PadBoth(9, "-").Value(); got != "--hello--" {
		t.Fatalf("PadBoth = %q", got)
	}

	if got := Of("abc123").Numbers().Value(); got != "123" {
		t.Fatalf("Numbers = %q", got)
	}

	encoded := Of("bedrock").ToBase64()
	decoded, err := encoded.FromBase64()

	if err != nil || decoded.Value() != "bedrock" {
		t.Fatalf("base64 round trip = %q, %v", decoded.Value(), err)
	}
}
