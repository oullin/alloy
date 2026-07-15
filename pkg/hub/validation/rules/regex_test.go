package rules

import "testing"

func TestRegexValid(t *testing.T) {
	t.Parallel()

	if !validateRegex("value", "123", []string{`^\d+$`}, nil) {
		t.Fatal("expected regex to match digits")
	}
	if validateRegex("value", "abc", []string{`^\d+$`}, nil) {
		t.Fatal("expected regex to reject non-digits")
	}
	if validateNotRegex("value", "123", []string{`^\d+$`}, nil) {
		t.Fatal("expected not_regex to reject a matching value")
	}
	if !validateNotRegex("value", "abc", []string{`^\d+$`}, nil) {
		t.Fatal("expected not_regex to accept a non-matching value")
	}
}

func TestRegexMalformedFailsClosed(t *testing.T) {
	t.Parallel()

	if validateRegex("value", "anything", []string{"("}, nil) {
		t.Fatal("expected regex to reject a malformed pattern")
	}
	if validateNotRegex("value", "anything", []string{"("}, nil) {
		t.Fatal("expected not_regex to reject a malformed pattern")
	}
}

func TestRegexFlags(t *testing.T) {
	t.Parallel()

	if !validateRegex("value", "123", []string{`/^\d+$/i`}, nil) {
		t.Fatal("expected delimited digit pattern to match")
	}
	if !validateRegex("value", "ABC", []string{`/^[a-z]+$/i`}, nil) {
		t.Fatal("expected case-insensitive pattern to match")
	}
	if !validateRegex("value", "a\nb", []string{`/^a.b$/is`}, nil) {
		t.Fatal("expected dotall pattern to match a newline")
	}
	if !validateRegex("value", "123", []string{`^\d+$`}, nil) {
		t.Fatal("expected undelimited pattern to match")
	}
}

func TestRegexInternalSlashes(t *testing.T) {
	t.Parallel()

	if !validateRegex("value", "a/b", []string{`/^a\/b$/`}, nil) {
		t.Fatal("expected escaped internal slash pattern to match")
	}
	if !validateRegex("value", "A/B", []string{`/^a\/b$/i`}, nil) {
		t.Fatal("expected escaped internal slash pattern with flags to match")
	}
}

func TestRegexUnknownFlagRejects(t *testing.T) {
	t.Parallel()

	if validateRegex("value", "123", []string{`/^\d+$/z`}, nil) {
		t.Fatal("expected regex to reject an unknown flag")
	}
	if validateNotRegex("value", "123", []string{`/^\d+$/z`}, nil) {
		t.Fatal("expected not_regex to reject an unknown flag")
	}
}
