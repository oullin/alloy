package core

import "time"

// Bearer is the contract feature packages target so a single generic
// implementation serves both the immutable Tempo and the mutable Mutable.
//
//   - State exposes the current snapshot (read-only).
//   - With produces the new bearer given a replacement time.Time.
//     Immutable bearers return a fresh value; mutable bearers update their
//     own state in place and return themselves.
//
// T is the concrete bearer type so With's return type tracks the caller —
// e.g. arithmetic.AddDays[Tempo] returns Tempo, AddDays[*Mutable] returns
// *Mutable, with no type assertions at the call site.
type Bearer[T any] interface {
	State() State
	With(time.Time) T
}
