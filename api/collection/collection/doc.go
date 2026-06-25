// Package collection provides a fluent, generic wrapper for working with slices of data.
//
// The core type is:
//
//   - [List] — a generic wrapper around a slice with a rich set of chainable methods
//     for filtering, sorting, transforming, and aggregating data.
//
// # Related packages
//
//   - [github.com/oullin/alloy/collection/lazy] — lazily evaluated sequences backed by iter.Seq
//   - [github.com/oullin/alloy/collection/collectible] — ordered map with fluent key-value API
//   - [github.com/oullin/alloy/collection/support] — shared types (Pair, Numeric) and errors
//   - [github.com/oullin/alloy/collection/arr] — generic slice helpers
//   - [github.com/oullin/alloy/collection/kv] — map helpers with dot-notation support
package collection
