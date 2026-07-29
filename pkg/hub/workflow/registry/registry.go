// Package registry holds multiple named workflows and resolves the right one
// for a given subject via SupportStrategy predicates.
package registry

import (
	"fmt"
	"sync"

	"hara.sh/alloy/workflow"
)

// SupportStrategy reports whether a workflow applies to a subject.
type SupportStrategy[T any] func(T) bool

// Entry registers a workflow with an optional name and support predicate.
type Entry[T any] struct {
	Name     string
	Machine  workflow.Engine[T]
	Supports SupportStrategy[T]
}

// Store is a thread-safe set of workflow Entries.
type Store[T any] struct {
	mu      sync.RWMutex
	entries []Entry[T]
}

func New[T any]() *Store[T] {
	return &Store[T]{}
}

// Add registers a workflow entry. Nil-Machine entries are ignored.
func (r *Store[T]) Add(entry Entry[T]) {
	if entry.Machine == nil {
		return
	}

	r.mu.Lock()

	defer r.mu.Unlock()

	r.entries = append(r.entries, entry)
}

// Get returns the first workflow whose name matches (if provided) and whose
// SupportStrategy accepts the subject.
func (r *Store[T]) Get(subject T, name string) (workflow.Engine[T], error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	for _, entry := range r.entries {
		if name != "" && entry.Name != name {
			continue
		}

		if entry.Supports != nil && !entry.Supports(subject) {
			continue
		}

		return entry.Machine, nil
	}

	if name != "" {
		return nil, fmt.Errorf("workflow %q not found", name)
	}

	return nil, fmt.Errorf("no workflow matched subject")
}
