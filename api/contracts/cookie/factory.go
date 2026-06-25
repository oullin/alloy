package cookie

import "net/http"

// Factory creates cookies.
type Factory interface {
	Make(name, value string, opts Options) *http.Cookie
	Forever(name, value string, opts Options) *http.Cookie
	Forget(name string, opts Options) *http.Cookie
}

// QueueingFactory extends Factory with a queue for deferred attachment to HTTP
// responses.
type QueueingFactory interface {
	Factory
	Queue(c *http.Cookie) error
	Unqueue(name string, path ...string)
	HasQueued(name string, path ...string) bool
	Queued(name string, path ...string) *http.Cookie
	GetQueued() []*http.Cookie
	Flush()
}
