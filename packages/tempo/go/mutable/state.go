package mutable

import "time"

type State[R any] struct {
	Value    time.Time
	Location *time.Location
	Runtime  R
}

func Replace[R any](_ State[R], next State[R]) State[R] {
	return next
}
