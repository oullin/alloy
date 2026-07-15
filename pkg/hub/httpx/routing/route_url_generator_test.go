package routing

import (
	"testing"
)

func TestStringify_Floats(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want string
	}{
		{"fractional", 3.14, "3.14"},
		{"negative_fractional", -0.5, "-0.5"},
		{"whole_valued_float", 2.0, "2"},
		{"large", 1e20, "100000000000000000000"},
		{"small_fraction", 0.001, "0.001"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stringify(tc.in); got != tc.want {
				t.Errorf("stringify(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestStringify_UnknownTypeNotDropped(t *testing.T) {
	// An unhandled type must not be silently dropped (returned as ""); it
	// should fall back to a best-effort string representation.
	type custom struct{ A int }

	if got := stringify(custom{A: 7}); got == "" {
		t.Errorf("stringify(unknown) returned empty string; value was silently dropped")
	}

	// uint is not in the explicit switch; it must still render its digits.
	if got := stringify(uint(42)); got != "42" {
		t.Errorf("stringify(uint(42)) = %q, want %q", got, "42")
	}
}

func TestStringify_CommonTypesPreserved(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"string", "abc", "abc"},
		{"int", 5, "5"},
		{"int64", int64(-9), "-9"},
		{"bool_true", true, "1"},
		{"bool_false", false, "0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stringify(tc.in); got != tc.want {
				t.Errorf("stringify(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
