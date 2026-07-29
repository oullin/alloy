package parser_test

import (
	"testing"
	"time"

	"hara.sh/alloy/tempo/parser"
)

// BenchmarkParseFromPatternBulk parses many inputs against a single repeated
// layout, the pattern the format-regex cache is designed to amortize.
func BenchmarkParseFromPatternBulk(b *testing.B) {
	const layout = "YYYY-MM-DD HH:mm:ss"

	inputs := []string{
		"2026-07-15 09:30:00",
		"1999-12-31 23:59:59",
		"2000-01-01 00:00:00",
		"2021-06-05 12:00:01",
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]

		if _, err := parser.ParseFromPattern(input, layout, time.UTC); err != nil {
			b.Fatalf("ParseFromPattern(%q) error: %v", input, err)
		}
	}
}
