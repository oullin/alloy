// Each ValidatorInterface implementation evaluates a single dimension of a
// candidate route against an incoming request. The router calls all four
// validators in sequence; a route matches when every validator returns true.
package matching

import (
	"regexp"

	crouting "github.com/oullin/alloy/packages/foundation/httpx/routing/contracts"
)

// MatchableRoute is the surface a Route must expose so matching/* validators
// can interrogate it.
//
// use an interface so the matching package has no import cycle on the parent
// routing package.
type MatchableRoute = crouting.MatchableRoute

// MatchableRequest is the surface a Request must expose for matching.
type MatchableRequest = crouting.MatchableRequest

type ValidatorInterface = crouting.ValidatorInterface

// All returns the four standard validators in the order the upstream Router
// applies them: URI, Method, Scheme, Host. The order matters: cheaper checks
// run first so the common case (a path mismatch) short-circuits quickly.
func All() []ValidatorInterface {
	return []ValidatorInterface{
		UriValidator{},
		MethodValidator{},
		SchemeValidator{},
		HostValidator{},
	}
}

// hostRegexCache memoizes compiled host regexes the same way Symfony's
// dumper would. Validators consult it so repeated dispatch loops do not
// re-parse the same regex.
var hostRegexCache = struct{ get func(string) *regexp.Regexp }{
	get: func(s string) *regexp.Regexp {
		// Compilation already happens inside CompiledRoute; this stub remains
		// for symmetry with future caching extensions.
		return nil
	},
}
