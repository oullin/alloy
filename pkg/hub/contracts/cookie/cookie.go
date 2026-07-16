package cookie

import (
	"net/http"
	"time"
)

// Options configures cookie attributes.
type Options struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   *bool
	HTTPOnly *bool
	SameSite http.SameSite
	Raw      *bool
}

const (
	SameSiteDefault = http.SameSiteDefaultMode
	SameSiteLax     = http.SameSiteLaxMode
	SameSiteStrict  = http.SameSiteStrictMode
	SameSiteNone    = http.SameSiteNoneMode
)

// BoolPtr returns a pointer to v.
func BoolPtr(v bool) *bool { return &v }

// DefaultOptions returns sensible production defaults. Secure defaults to true
// so cookies are never emitted over plaintext HTTP unless a caller explicitly
// opts out for local development via BoolPtr(false).
func DefaultOptions() Options {
	return Options{
		Path:     "/",
		Secure:   BoolPtr(true),
		HTTPOnly: BoolPtr(true),
		SameSite: SameSiteLax,
	}
}

// Make creates an *http.Cookie from the given options.
func Make(name, value string, opts Options) *http.Cookie {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     opts.Path,
		Domain:   opts.Domain,
		MaxAge:   opts.MaxAge,
		SameSite: opts.SameSite,
	}

	// Secure-by-default: a nil Secure is treated as true so a bare Options{}
	// never downgrades to a plaintext-transmittable cookie. Local dev must opt
	// out explicitly with BoolPtr(false).
	if opts.Secure != nil {
		c.Secure = *opts.Secure
	} else {
		c.Secure = true
	}

	if opts.HTTPOnly != nil {
		c.HttpOnly = *opts.HTTPOnly
	}

	return c
}

// Forever creates a long-lived cookie.
func Forever(name, value string, opts Options) *http.Cookie {
	opts.MaxAge = int((400 * 24 * time.Hour).Seconds())

	return Make(name, value, opts)
}

// Forget creates a cookie deletion response.
func Forget(name string, opts Options) *http.Cookie {
	opts.MaxAge = -1

	return Make(name, "", opts)
}
