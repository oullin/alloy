package collection

import (
	"cmp"
	"math/rand/v2"
	"sort"
)

// Sort returns a new collection sorted using the provided comparison function.
func (c *List[T]) Sort(less func(a, b T) bool) *List[T] {
	result := make([]T, len(c.items))
	copy(result, c.items)

	sort.SliceStable(result, func(i, j int) bool {
		return less(result[i], result[j])
	})

	return Collect(result)
}

// SortBy returns a new collection sorted in ascending order by the given key function.
func SortBy[T any, K cmp.Ordered](c *List[T], keyFunc func(T) K) *List[T] {
	result := make([]T, len(c.items))
	copy(result, c.items)

	sort.SliceStable(result, func(i, j int) bool {
		return keyFunc(result[i]) < keyFunc(result[j])
	})

	return Collect(result)
}

// SortByDesc returns a new collection sorted in descending order by the given key function.
func SortByDesc[T any, K cmp.Ordered](c *List[T], keyFunc func(T) K) *List[T] {
	result := make([]T, len(c.items))
	copy(result, c.items)

	sort.SliceStable(result, func(i, j int) bool {
		return keyFunc(result[i]) > keyFunc(result[j])
	})

	return Collect(result)
}

// SortDesc returns a new collection sorted in descending order using the provided comparison function.
func (c *List[T]) SortDesc(less func(a, b T) bool) *List[T] {
	return c.Sort(func(a, b T) bool {
		return less(b, a)
	})
}

// Shuffle returns a new collection with items in random order.
func (c *List[T]) Shuffle() *List[T] {
	result := make([]T, len(c.items))
	copy(result, c.items)

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return Collect(result)
}

// Random returns a new collection with the specified number of randomly selected items.
func (c *List[T]) Random(counts ...int) *List[T] {
	count := 1

	if len(counts) > 0 {
		count = counts[0]
	}

	shuffled := c.Shuffle()

	if count >= len(shuffled.items) {
		return shuffled
	}

	result := make([]T, count)
	copy(result, shuffled.items[:count])

	return Collect(result)
}
