// Package collection provides a fluent, generic wrapper for working with slices of data.
package collection

import "iter"

// List wraps a slice and provides a fluent API for working with arrays of data.
type List[T any] struct {
	items []T
}

// New creates a new List from the given items.
func New[T any](items ...T) *List[T] {
	if items == nil {
		items = make([]T, 0)
	}

	return &List[T]{items: items}
}

// Collect creates a new List from a slice.
func Collect[T any](items []T) *List[T] {
	if items == nil {
		items = make([]T, 0)
	}

	return &List[T]{items: items}
}

// Empty creates an empty List.
func Empty[T any]() *List[T] {
	return &List[T]{items: make([]T, 0)}
}

// Wrap wraps the given value in a collection if it is not already one.
func Wrap[T any](value any) *List[T] {
	switch v := value.(type) {
	case *List[T]:
		return v
	case []T:
		return Collect(v)
	default:
		if item, ok := value.(T); ok {
			return New(item)
		}

		return Empty[T]()
	}
}

// Unwrap returns the underlying items from a List, or the value itself if it is a slice.
func Unwrap[T any](value any) []T {
	if c, ok := value.(*List[T]); ok {
		return c.All()
	}

	if items, ok := value.([]T); ok {
		return items
	}

	return nil
}

// Times create a new collection by invoking the callback a given number of times.
func Times[T any](number int, callback func(int) T) *List[T] {
	if number < 1 {
		return Empty[T]()
	}

	items := make([]T, number)

	for i := range number {
		items[i] = callback(i + 1)
	}

	return Collect(items)
}

// Range creates a collection of consecutive integers from start to end (inclusive).
func Range(from, to int) *List[int] {
	if from > to {
		items := make([]int, 0, from-to+1)

		for i := from; i >= to; i-- {
			items = append(items, i)
		}

		return Collect(items)
	}

	items := make([]int, 0, to-from+1)

	for i := from; i <= to; i++ {
		items = append(items, i)
	}

	return Collect(items)
}

// All returns all items in the collection as a slice.
func (c *List[T]) All() []T {
	return c.items
}

// Count returns the total number of items in the collection.
func (c *List[T]) Count() int {
	return len(c.items)
}

// IsEmpty reports whether the collection contains no items.
func (c *List[T]) IsEmpty() bool {
	return len(c.items) == 0
}

// IsNotEmpty reports whether the collection contains at least one item.
func (c *List[T]) IsNotEmpty() bool {
	return len(c.items) > 0
}

// ContainsOneItem reports whether the collection contains exactly one item.
func (c *List[T]) ContainsOneItem() bool {
	return len(c.items) == 1
}

// ContainsManyItems reports whether the collection contains more than one item.
func (c *List[T]) ContainsManyItems() bool {
	return len(c.items) > 1
}

// HasMany is an alias for ContainsManyItems.
func (c *List[T]) HasMany() bool {
	return c.ContainsManyItems()
}

// Len returns the number of items, implementing sort.Interface.
func (c *List[T]) Len() int {
	return len(c.items)
}

// ToBase returns the collection itself.
func (c *List[T]) ToBase() *List[T] {
	return c
}

// Iter returns an iter.Seq[T] that yields each item in the collection.
func (c *List[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range c.items {
			if !yield(item) {
				return
			}
		}
	}
}

// PairIter returns an iter.Seq2[int, T] that yields each index-item pair in the collection.
func (c *List[T]) PairIter() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, item := range c.items {
			if !yield(i, item) {
				return
			}
		}
	}
}
