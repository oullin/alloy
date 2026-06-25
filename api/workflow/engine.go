package workflow

import (
	"fmt"

	cworkflow "github.com/oullin/alloy/api/contracts/workflow"
	"github.com/oullin/alloy/workflow/events"
)

// Sink is the optional structured logger interface the engine writes to.
type Sink = cworkflow.Sink

// Engine is the public Petri-Net workflow contract.
type Engine[T any] interface {
	Name() string
	Definition() *Definition
	MetadataStore() MetadataStore
	GetMarking(subject T) (Marking, error)
	Can(subject T, transition string) bool
	CanNot(subject T, transition string) bool
	EnabledTransitions(subject T) ([]Transition, error)
	DisabledTransitions(subject T) ([]Transition, error)
	Apply(subject T, transition string, context map[string]any) (Marking, error)
}

// Machine is the default Engine implementation.
type Machine[T any] struct {
	name       string
	definition *Definition
	store      MarkingStore[T]
	dispatcher *events.Dispatcher[T]
	metadata   *DefinitionMetadataStore
	logger     Sink
}

// New constructs a Machine engine. A nil dispatcher yields an empty default.
func New[T any](name string, definition *Definition, store MarkingStore[T], dispatcher *events.Dispatcher[T]) (*Machine[T], error) {
	if name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}

	if definition == nil {
		return nil, fmt.Errorf("definition is required")
	}

	if err := definition.Validate(); err != nil {
		return nil, err
	}

	if store == nil {
		return nil, fmt.Errorf("marking store is required")
	}

	if dispatcher == nil {
		dispatcher = events.NewDispatcher[T]()
	}

	return &Machine[T]{
		name:       name,
		definition: definition.Clone(),
		store:      store,
		dispatcher: dispatcher,
		metadata:   NewMetadataStore(definition),
	}, nil
}

func (w *Machine[T]) SetLogger(logger Sink) {
	w.logger = logger
}

func (w *Machine[T]) logDebug(msg string, args ...any) {
	if w.logger != nil {
		w.logger.Debug(msg, args...)
	}
}

func (w *Machine[T]) logInfo(msg string, args ...any) {
	if w.logger != nil {
		w.logger.Info(msg, args...)
	}
}

func (w *Machine[T]) logError(msg string, args ...any) {
	if w.logger != nil {
		w.logger.Error(msg, args...)
	}
}

func (w *Machine[T]) Name() string { return w.name }

func (w *Machine[T]) Definition() *Definition {
	return w.definition.Clone()
}

func (w *Machine[T]) MetadataStore() MetadataStore {
	return w.metadata
}

func (w *Machine[T]) EventDispatcher() *events.Dispatcher[T] {
	return w.dispatcher
}

func (w *Machine[T]) GetMarking(subject T) (Marking, error) {
	marking, err := w.store.GetMarking(subject, w.definition)

	if err != nil {
		return Marking{}, err
	}

	if len(marking.Places) == 0 {
		return w.definition.InitialMarking.Clone(), nil
	}

	return marking, nil
}
