package workflow

import (
	"maps"
	"sort"
)

// Sink is the optional structured logger interface the workflow engine writes to.
type Sink interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Engine is the public Petri-Net workflow contract.
type Engine[T any] interface {
	Name() string
	GetMarking(subject T) (Marking, error)
	Can(subject T, transition string) bool
	CanNot(subject T, transition string) bool
	EnabledTransitions(subject T) ([]Transition, error)
	DisabledTransitions(subject T) ([]Transition, error)
	Apply(subject T, transition string, context map[string]any) (Marking, error)
}

// MarkingStore reads and writes the marking for a subject.
type MarkingStore[T any] interface {
	GetMarking(subject T, definition any) (Marking, error)
	SetMarking(subject T, marking Marking, definition any, context map[string]any) error
}

// MetadataStore exposes workflow definition metadata.
type MetadataStore interface {
	WorkflowMetadata(key string) (any, bool)
	PlaceMetadata(place string, key string) (any, bool)
	TransitionMetadata(transition string, key string) (any, bool)
}

// Marking tracks active places and token counts.
type Marking struct {
	Places map[string]int
}

// NewMarking constructs a marking from active places.

// Transition is an arc in the Petri-Net moving tokens between places.
type Transition struct {
	Name     string
	From     []string
	To       []string
	Metadata map[string]any
}

// NewTransition constructs a Transition.

// TransitionBlocker carries a reason a guard rejected a transition.
type TransitionBlocker struct {
	Message string
	Code    string
}

func NewMarking(places ...string) Marking {
	m := Marking{Places: make(map[string]int, len(places))}

	for _, place := range places {
		if place == "" {
			continue
		}

		m.Places[place]++
	}

	return m
}

func (m Marking) Clone() Marking {
	cloned := Marking{Places: make(map[string]int, len(m.Places))}

	maps.Copy(cloned.Places, m.Places)

	return cloned
}

func (m Marking) Has(place string) bool   { return m.Places[place] > 0 }
func (m Marking) Tokens(place string) int { return m.Places[place] }

func (m Marking) Add(place string, count int) {
	if count <= 0 {
		return
	}

	if m.Places == nil {
		m.Places = map[string]int{}
	}

	m.Places[place] += count
}

func (m Marking) Remove(place string, count int) {
	if count <= 0 || m.Places == nil {
		return
	}

	next := m.Places[place] - count

	if next <= 0 {
		delete(m.Places, place)

		return
	}

	m.Places[place] = next
}

func (m Marking) ActivePlaces() []string {
	places := make([]string, 0, len(m.Places))

	for place, count := range m.Places {
		if count > 0 {
			places = append(places, place)
		}
	}

	sort.Strings(places)

	return places
}

func NewTransition(name string, from []string, to []string) Transition {
	return Transition{
		Name: name,
		From: append([]string(nil), from...),
		To:   append([]string(nil), to...),
	}
}
