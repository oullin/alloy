package events

import "maps"

// Transition is a snapshot used inside workflow event payloads.
type Transition struct {
	Name string
	From []string
	To   []string
}

// Event is implemented by all workflow lifecycle events.
type Event[T any] interface {
	WorkflowName() string
	Subject() T
	Transition() Transition
	Marking() map[string]int
	Context() map[string]any
}

// Base carries fields shared by every lifecycle event.
type Base[T any] struct {
	Machine    string
	SubjectVal T
	Step       Transition
	Tokens     map[string]int
	Ctx        map[string]any
}

type TransitionBlocker struct {
	Message string
	Code    string
}

func (e Base[T]) WorkflowName() string   { return e.Machine }
func (e Base[T]) Subject() T             { return e.SubjectVal }
func (e Base[T]) Transition() Transition { return e.Step }
func (e Base[T]) Marking() map[string]int {
	out := make(map[string]int, len(e.Tokens))
	maps.Copy(out, e.Tokens)

	return out
}

func (e Base[T]) Context() map[string]any {
	out := make(map[string]any, len(e.Ctx))
	maps.Copy(out, e.Ctx)

	return out
}
