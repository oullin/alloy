// Package callbacks holds the container's lifecycle-callback bookkeeping: for
// each event, one global list plus a list per abstract key.
//
// The types here are lock-free by contract. App owns the only mutex in the
// container and serialises every call into this package; adding a second lock
// here would create a lock-ordering problem where none exists today.
//
// Registry is generic over the callback type on purpose. The container's
// callback types name *App, so a Registry that declared them would have to
// import container, which imports this package back. Staying generic keeps the
// dependency one-way.
package callbacks

import "slices"

// Registry stores callbacks of type T: one global list, plus a list per
// abstract key.
//
// Registry is NOT safe for concurrent use. The caller must hold the lock
// guarding it. Do not add a mutex.
type Registry[T any] struct {
	global []T
	byKey  map[string][]T
}

// NewRegistry returns an empty, fully initialised Registry.
func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{byKey: make(map[string][]T)}
}

// AddGlobal appends a callback that fires for every abstract.
func (r *Registry[T]) AddGlobal(cb T) {
	r.global = append(r.global, cb)
}

// Add appends a callback that fires only for the given abstract.
func (r *Registry[T]) Add(key string, cb T) {
	r.byKey[key] = append(r.byKey[key], cb)
}

// Snapshot returns copies of the global and key-specific callback lists.
//
// It exists so callers can copy under the lock and fire after releasing it —
// never invoke a callback while holding a lock, since callbacks re-enter the
// container. Both results are nil when the corresponding list is empty, which
// ranges as a no-op.
func (r *Registry[T]) Snapshot(key string) (global, specific []T) {
	return slices.Clone(r.global), slices.Clone(r.byKey[key])
}

// Reset returns the Registry to its empty state.
func (r *Registry[T]) Reset() {
	r.global = nil
	r.byKey = make(map[string][]T)
}
