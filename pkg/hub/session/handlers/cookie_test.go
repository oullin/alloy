package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/alloy/pkg/hub/session/handlers"
)

// identityEncrypter passes payloads through unchanged for cookie flag tests.
type identityEncrypter struct{}

func (identityEncrypter) Encrypt(plaintext string) (string, error)  { return plaintext, nil }
func (identityEncrypter) Decrypt(ciphertext string) (string, error) { return ciphertext, nil }

func sessionCookie(t *testing.T, rr *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()

	for _, c := range rr.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}

	t.Fatalf("expected %q cookie in Set-Cookie headers", name)

	return nil
}

func TestCookieHandlerSecureTriState(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(h *handlers.CookieHandler)
		want  bool
	}{
		{name: "unset defaults to secure", setup: func(*handlers.CookieHandler) {}, want: true},
		{name: "explicit true honored", setup: func(h *handlers.CookieHandler) { h.SetSecure(true) }, want: true},
		{name: "explicit false honored (dev opt-out)", setup: func(h *handlers.CookieHandler) { h.SetSecure(false) }, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := handlers.NewCookieHandler(identityEncrypter{}, "sess")
			tc.setup(h)

			// Write must emit the session cookie with the resolved Secure flag.
			rr := httptest.NewRecorder()
			h.SetWriter(rr)

			if err := h.Write(context.Background(), "id", "payload"); err != nil {
				t.Fatalf("Write: %v", err)
			}

			if c := sessionCookie(t, rr, "sess"); c.Secure != tc.want {
				t.Fatalf("Write cookie Secure = %v, want %v", c.Secure, tc.want)
			}

			// Destroy's expiring cookie must carry the same Secure flag.
			rr = httptest.NewRecorder()
			h.SetWriter(rr)

			if err := h.Destroy(context.Background(), "id"); err != nil {
				t.Fatalf("Destroy: %v", err)
			}

			c := sessionCookie(t, rr, "sess")

			if c.Secure != tc.want {
				t.Fatalf("Destroy cookie Secure = %v, want %v", c.Secure, tc.want)
			}

			if c.MaxAge != -1 {
				t.Fatalf("Destroy cookie MaxAge = %d, want -1", c.MaxAge)
			}
		})
	}
}
