package cookie

import (
	"net/http"

	ccookie "alloy.dev/go/contracts/cookie"
)

// SameSite matches the http.SameSite constants for convenient use.

// Options configures cookie attributes. Boolean fields use *bool so that
// callers can distinguish "not set" (nil) from an explicit false, allowing
// defaults to be overridden in either direction.
type Options = ccookie.Options

// BoolPtr returns a pointer to v, useful for setting Options boolean fields.
//
//go:fix inline
func BoolPtr(v bool) *bool { return ccookie.BoolPtr(v) }

const (
	SameSiteDefault = ccookie.SameSiteDefault
	SameSiteLax     = ccookie.SameSiteLax
	SameSiteStrict  = ccookie.SameSiteStrict
	SameSiteNone    = ccookie.SameSiteNone
)

// DefaultOptions returns sensible production defaults.
func DefaultOptions() Options {
	return ccookie.DefaultOptions()
}

// Make creates an *http.Cookie from the given options.
func Make(name, value string, opts Options) *http.Cookie {
	return ccookie.Make(name, value, opts)
}

// Forever creates a cookie that expires in 400 days (~the upstream "forever").
func Forever(name, value string, opts Options) *http.Cookie {
	return ccookie.Forever(name, value, opts)
}

// Forget creates a cookie with a negative max-age to instruct the browser
// to delete it.
func Forget(name string, opts Options) *http.Cookie {
	return ccookie.Forget(name, opts)
}
