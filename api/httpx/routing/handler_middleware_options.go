package routing

// HandlerMiddlewareOptions is a parity alias for [MiddlewareOptions].
//
// The PHP class is named HandlerMiddlewareOptions; the Go shorter name
// reads more naturally inside handler bodies. Both names refer to the
// same underlying type so test ports remain syntactically identical.
type HandlerMiddlewareOptions = MiddlewareOptions
