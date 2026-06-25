package parser_test

import (
	"testing"
	"time"

	"github.com/oullin/alloy/tempo/parser"
)

func TestParseISO(t *testing.T) {
	p := parser.New(time.UTC)

	got, err := p.Parse("2024-05-15T09:30:00Z")

	if err != nil {
		t.Fatalf("Parse ISO8601: unexpected error %v", err)
	}

	want := time.Date(2024, time.May, 15, 9, 30, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Fatalf("Parse ISO8601 = %v, want %v", got, want)
	}
}

func TestParseDateOnly(t *testing.T) {
	p := parser.New(time.UTC)

	got, err := p.Parse("2024-05-15")

	if err != nil {
		t.Fatalf("Parse date-only: unexpected error %v", err)
	}

	if got.Year() != 2024 || got.Month() != time.May || got.Day() != 15 {
		t.Fatalf("Parse date-only landed on %v", got)
	}
}

func TestParseRespectsLocation(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")

	if err != nil {
		t.Skipf("America/New_York tzdata unavailable: %v", err)
	}

	p := parser.New(ny)

	got, err := p.Parse("2024-05-15T09:30:00")

	if err != nil {
		t.Fatalf("Parse naive in NY: unexpected error %v", err)
	}

	// Naive 09:30 in NY (EDT = -04:00) is 13:30 UTC.
	want := time.Date(2024, time.May, 15, 13, 30, 0, 0, time.UTC)

	if !got.UTC().Equal(want) {
		t.Fatalf("Parse naive in NY = %v, want %v UTC", got.UTC(), want)
	}
}

func TestParseInvalidInput(t *testing.T) {
	p := parser.New(time.UTC)

	if _, err := p.Parse("not a date"); err == nil {
		t.Fatalf("Parse(\"not a date\") should fail")
	}
}

func TestFromFormat(t *testing.T) {
	p := parser.New(time.UTC)

	got, err := p.FromFormat("15/05/2024", "DD/MM/YYYY")

	if err != nil {
		t.Fatalf("FromFormat DD/MM/YYYY: unexpected error %v", err)
	}

	if got.Year() != 2024 || got.Month() != time.May || got.Day() != 15 {
		t.Fatalf("FromFormat parsed wrong date: %v", got)
	}
}

func TestNewNilLocationDefaultsToUTC(t *testing.T) {
	p := parser.New(nil)

	got, err := p.Parse("2024-05-15T09:30:00")

	if err != nil {
		t.Fatalf("Parse with nil-location parser: %v", err)
	}

	// With nil → UTC, naive 09:30 stays 09:30 UTC.
	if got.UTC().Hour() != 9 || got.UTC().Minute() != 30 {
		t.Fatalf("nil-location parser did not default to UTC: %v", got.UTC())
	}
}
