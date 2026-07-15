package stats_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/queue/drivers/internal/stats"
)

func TestParseInt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   map[string]string
		key  string
		want int64
	}{
		{"valid", map[string]string{"count": "42"}, "count", 42},
		{"missing key", map[string]string{}, "count", 0},
		{"nil map", nil, "count", 0},
		{"empty value", map[string]string{"count": ""}, "count", 0},
		{"non-numeric", map[string]string{"count": "abc"}, "count", 0},
		{"zero", map[string]string{"count": "0"}, "count", 0},
		{"negative", map[string]string{"count": "-7"}, "count", -7},
		{"overflows int64", map[string]string{"count": "99999999999999999999"}, "count", 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := stats.ParseInt(tc.in, tc.key); got != tc.want {
				t.Fatalf("ParseInt(%v, %q) = %d, want %d", tc.in, tc.key, got, tc.want)
			}
		})
	}
}
