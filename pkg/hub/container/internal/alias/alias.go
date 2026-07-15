// Package alias holds the container's alias bookkeeping: the alias -> abstract
// mapping and the reverse index of aliases registered against an abstract.
//
// The types here are lock-free by contract. App owns the only mutex in the
// container and serialises every call into this package; adding a second lock
// here would create a lock-ordering problem where none exists today.
package alias

import "slices"

// Table maps alias names to abstracts and keeps a reverse index.
//
// Table is NOT safe for concurrent use. Every method assumes the caller already
// holds the lock guarding it. Do not add a mutex.
//
// Resolve and Has must stay strictly read-only — no memoisation, no path
// compression, no lazy map initialisation. App calls both from RLock-only
// paths (Resolved, IsShared, rebound, GetAlias), so any hidden write would turn
// a reader into a writer and introduce a data race.
type Table struct {
	to   map[string]string   // alias -> abstract, one hop
	from map[string][]string // abstract -> aliases registered against it
}

// NewTable returns an empty, fully initialised Table.
func NewTable() *Table {
	return &Table{
		to:   make(map[string]string),
		from: make(map[string][]string),
	}
}

// Add registers name as an alias of abstract.
func (t *Table) Add(abstract, name string) {
	t.to[name] = abstract
	t.from[abstract] = append(t.from[abstract], name)
}

// Resolve walks the alias chain to the underlying abstract. A name that is not
// an alias is returned unchanged.
//
// Resolve does not terminate on a cyclic chain (Alias("a","b") plus
// Alias("b","a")). That matches the behaviour of the code this replaced: the
// container guards only against self-aliasing, so a cycle hangs here today just
// as it did before.
func (t *Table) Resolve(name string) string {
	for {
		target, ok := t.to[name]

		if !ok {
			return name
		}

		name = target
	}
}

// Has reports whether name is a registered alias.
func (t *Table) Has(name string) bool {
	_, ok := t.to[name]

	return ok
}

// Drop removes name's forward entry and deliberately leaves the reverse index
// alone. It backs the container's stale-binding cleanup, where an abstract is
// being rebound and only the alias pointing at it is invalid.
//
// Drop and Remove are not interchangeable; see Remove.
func (t *Table) Drop(name string) {
	delete(t.to, name)
}

// Remove deletes name's forward entry and purges it from every reverse-index
// entry. It backs registering a concrete instance, where the name stops being
// an alias entirely and must not resurface through the reverse index.
//
// The extra reverse-index purge is the only thing separating Remove from Drop.
// Collapsing the two changes behaviour.
func (t *Table) Remove(name string) {
	for abs, aliases := range t.from {
		for i, a := range aliases {
			if a == name {
				t.from[abs] = slices.Delete(aliases, i, i+1)

				break
			}
		}
	}

	delete(t.to, name)
}

// Reset returns the Table to its empty state.
func (t *Table) Reset() {
	t.to = make(map[string]string)
	t.from = make(map[string][]string)
}
