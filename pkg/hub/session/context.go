package session

import "context"

// storeContextKey is the context key under which the per-request Store is
// stored. It is an unexported struct type so it cannot collide with keys
// defined in other packages.
type storeContextKey struct{}

// NewContext returns a copy of ctx carrying the given Store. Integrators
// writing their own session middleware can use it to expose the request's
// Store to downstream handlers.
func NewContext(ctx context.Context, store *Store) context.Context {
	return context.WithValue(ctx, storeContextKey{}, store)
}

// FromContext returns the Store carried by ctx, if any. The boolean is false
// when no Store is present.
func FromContext(ctx context.Context) (*Store, bool) {
	store, ok := ctx.Value(storeContextKey{}).(*Store)

	return store, ok
}
