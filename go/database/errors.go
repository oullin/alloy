package database

import "errors"

var (
	// ErrRecordNotFound is returned when a query expecting a result finds none.
	ErrRecordNotFound = errors.New("database: record not found")
)
