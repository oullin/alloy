// PHP 8 attributes have no direct Go equivalent. The strict-parity story here
// is "the value object is identical, but you attach it via the
// [handlermiddleware.Provider] interface instead of an attribute on the class".
// The struct itself is exposed for code that wants to construct a middleware
// declaration out-of-band (e.g. when porting a handler mechanically).
package attributes

import handlermiddleware "github.com/oullin/alloy/packages/foundation/httpx/handlerx/middleware"

// Middleware is a value object describing a handler-attached middleware.
// It is the parity counterpart of #[Middleware] in PHP.
type Middleware = handlermiddleware.Entry
