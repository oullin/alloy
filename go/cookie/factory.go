package cookie

import ccookie "alloy.dev/api/contracts/cookie"

// Factory creates cookies.
type Factory = ccookie.Factory

// QueueingFactory extends Factory with a queue for deferred attachment to
// HTTP responses. Cookies are keyed by name and path, so the same cookie
// name on different paths can coexist in the queue.
type QueueingFactory = ccookie.QueueingFactory
