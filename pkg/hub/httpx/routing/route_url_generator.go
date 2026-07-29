package routing

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	contracts "hara.sh/alloy/httpx/routing/contracts"
)

// RouteUrlGenerator builds a URL for a single named route by substituting
// parameter values into the route's URI template.
//
// (rather than a method on UrlGenerator) so signed URL generation can call it
// recursively without circular wiring.
type RouteUrlGenerator struct {
	url     *UrlGenerator
	request URLRequest
}

// URLRequest is the minimum surface UrlGenerator needs from a request.
//
// foundation.Request will satisfy this in M11; tests provide a fake.
type URLRequest = contracts.URLRequest

// NewRouteUrlGenerator returns a route URL generator bound to a parent
// [UrlGenerator] and the current request.
func NewRouteUrlGenerator(u *UrlGenerator, request URLRequest) *RouteUrlGenerator {
	return &RouteUrlGenerator{url: u, request: request}
}

// To produces a URL for route with the supplied parameter values.
func (g *RouteUrlGenerator) To(route *Route, parameters map[string]any, absolute bool) string {
	domain := route.GetDomain()
	uri := route.Uri
	uri = "/" + strings.TrimLeft(uri, "/")

	// Substitute path parameters in declaration order.
	consumed := map[string]struct{}{}
	uri = substituteParameters(uri, parameters, consumed)

	// Strip any unconsumed optional placeholders (the trailing "?").
	uri = optionalRe.ReplaceAllString(uri, "")

	// Build the query string from the unconsumed parameters, sorted by key.
	queryString := buildQuery(parameters, consumed)

	if queryString != "" {
		uri += "?" + queryString
	}

	if !absolute {
		return uri
	}

	scheme := "http"
	host := ""

	if g.request != nil {
		scheme = g.request.Scheme()
		host = g.request.Host()
	}

	if domain != "" {
		host = domain
	}

	if g.url != nil && g.url.forcedScheme != "" {
		scheme = g.url.forcedScheme
	}

	return scheme + "://" + host + uri
}

var (
	parameterRe = regexp.MustCompile(`\{([\w]+)\??\}`)
	optionalRe  = regexp.MustCompile(`/\{[\w]+\?\}`)
)

func substituteParameters(uri string, parameters map[string]any, consumed map[string]struct{}) string {
	return parameterRe.ReplaceAllStringFunc(uri, func(match string) string {
		// Strip { } and trailing ?
		name := strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(match, "{"), "}"), "?")
		v, ok := parameters[name]

		if !ok {
			return match
		}

		consumed[name] = struct{}{}

		return url.PathEscape(stringify(v))
	})
}

// buildQuery returns the canonical query string for parameters not consumed
// by the URI, sorted by key.
// flat associative arrays (key=value joined by "&", values URL-encoded with
// '+' for spaces, RFC 1738).
func buildQuery(parameters map[string]any, consumed map[string]struct{}) string {
	values := url.Values{}

	for k, v := range parameters {
		if _, ok := consumed[k]; ok {
			continue
		}

		values.Set(k, stringify(v))
	}

	// Route through the shared canonical encoder so the signing side (here)
	// and the verifying side (HasCorrectSignature -> canonicalizeQuery)
	// produce byte-for-byte identical HMAC inputs.
	return canonicalQueryEncode(values)
}

func stringify(v any) string {
	// nil keeps its historical empty-string rendering; letting it reach the
	// fmt.Sprint fallback would leak the literal "<nil>" into generated URLs.
	if v == nil {
		return ""
	}

	switch x := v.(type) {
	case string:
		return x
	case int:
		return intToString(int64(x))
	case int64:
		return intToString(x)
	case float64:
		return floatToString(x)
	case bool:
		if x {
			return "1"
		}

		return "0"
	case BackedEnum:
		return x.BackingValue()
	case contracts.UrlRoutable:
		return stringify(x.GetRouteKey())
	}

	// Best-effort fallback for unhandled types. Dropping the value (returning
	// "") would silently omit a route/query parameter, which is worse than
	// emitting a reasonable string representation the caller can inspect.
	return fmt.Sprint(v)
}

func intToString(i int64) string {
	// Avoid importing strconv twice; use a tiny inline formatter.
	if i == 0 {
		return "0"
	}

	negative := false

	if i < 0 {
		negative = true
		i = -i
	}

	var buf [20]byte
	pos := len(buf)

	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}

	if negative {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}

func floatToString(f float64) string {
	// 'f' (decimal, no exponent) with precision -1 emits the shortest string
	// that round-trips back to f: 3.14 -> "3.14", -0.5 -> "-0.5", and large
	// magnitudes print in full ("100000000000000000000") rather than being
	// truncated to an int64 (which overflowed and dropped the fraction). This
	// keeps values friendly for URL route/query params and constraint regexes.
	return strconv.FormatFloat(f, 'f', -1, 64)
}
