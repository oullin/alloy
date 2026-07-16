package cookie_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/contracts/cookie"
)

func TestDefaultOptionsSecureByDefault(t *testing.T) {
	t.Parallel()

	opts := cookie.DefaultOptions()

	if opts.Secure == nil {
		t.Fatal("expected DefaultOptions Secure to be set, got nil")
	}

	if !*opts.Secure {
		t.Fatal("expected DefaultOptions Secure to default to true")
	}
}

func TestMakeSecureTriState(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		secure *bool
		want   bool
	}{
		{name: "nil defaults to secure", secure: nil, want: true},
		{name: "explicit true honored", secure: cookie.BoolPtr(true), want: true},
		{name: "explicit false honored (dev opt-out)", secure: cookie.BoolPtr(false), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := cookie.Make("session", "value", cookie.Options{Secure: tc.secure})

			if c.Secure != tc.want {
				t.Fatalf("Secure = %v, want %v", c.Secure, tc.want)
			}
		})
	}
}
