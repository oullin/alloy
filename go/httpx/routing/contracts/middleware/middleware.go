package middleware

// RateLimiter is the surface throttle middleware needs from a rate limiter.
type RateLimiter interface {
	TooManyAttempts(key string, maxAttempts int) bool
	Hit(key string, decay int) int
	AvailableIn(key string) int
	RetriesLeft(key string, maxAttempts int) int
	Clear(key string)
}

// ThrottleRequest is the surface throttle middleware needs from requests.
type ThrottleRequest interface {
	IP() string
	UserID() string
	Path() string
}

// SignatureValidator is the surface signature middleware needs from requests.
type SignatureValidator interface {
	HasValidSignatureWhileIgnoring(ignore []string, absolute bool) bool
}
