// Package collection provides a fluent, generic wrapper for working with slices of data.
//
// The core type is:
//
//   - [List] — a generic wrapper around a slice with a rich set of chainable methods
//     for filtering, sorting, transforming, and aggregating data.
//
// # Related packages
//
//   - [hara.sh/alloy/collection/lazy] — lazily evaluated sequences backed by iter.Seq
//   - [hara.sh/alloy/collection/collectible] — ordered map with fluent key-value API
//   - [hara.sh/alloy/collection/support] — shared types (Pair, Numeric) and errors
//   - [hara.sh/alloy/collection/arr] — generic slice helpers
//   - [hara.sh/alloy/collection/kv] — map helpers with dot-notation support
package collection
