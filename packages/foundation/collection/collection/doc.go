// Package collection provides a fluent, generic wrapper for working with slices of data.
//
// The core type is:
//
//   - [List] — a generic wrapper around a slice with a rich set of chainable methods
//     for filtering, sorting, transforming, and aggregating data.
//
// # Related packages
//
//   - [alloy.dev/foundation/collection/lazy] — lazily evaluated sequences backed by iter.Seq
//   - [alloy.dev/foundation/collection/collectible] — ordered map with fluent key-value API
//   - [alloy.dev/foundation/collection/support] — shared types (Pair, Numeric) and errors
//   - [alloy.dev/foundation/collection/arr] — generic slice helpers
//   - [alloy.dev/foundation/collection/kv] — map helpers with dot-notation support
package collection
