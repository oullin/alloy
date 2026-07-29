package workflow

import (
	"fmt"

	"hara.sh/alloy/workflow/events"
)

func (w *Machine[T]) transitionState(subject T, transitionName string, context map[string]any) error {
	transition, ok := w.definition.Transition(transitionName)

	if !ok {
		return fmt.Errorf("%w: %s", ErrTransitionNotFound, transitionName)
	}

	marking, err := w.GetMarking(subject)

	if err != nil {
		return err
	}

	return w.transitionStateForMarking(subject, transition, marking, context)
}

func (w *Machine[T]) transitionStateForMarking(subject T, transition Transition, marking Marking, context map[string]any) error {
	if !w.transitionEnabled(marking, transition) {
		return &TransitionError{Machine: w.name, Transition: transition.Name}
	}

	guard := w.newGuardEvent(subject, transition, marking, context)
	w.dispatchGuardEvents(guard)

	if guard.Blocked() {
		return &TransitionError{
			Machine:    w.name,
			Transition: transition.Name,
			Blockers:   blockersFromEvents(guard.Blockers()),
		}
	}

	return nil
}

func (w *Machine[T]) newGuardEvent(subject T, transition Transition, marking Marking, context map[string]any) *events.GuardEvent[T] {
	return &events.GuardEvent[T]{
		Base: w.baseEvent(subject, transition, marking.Clone(), context),
	}
}

func (w *Machine[T]) dispatchGuardEvents(event *events.GuardEvent[T]) {
	w.dispatcher.Dispatch(EventNameGuard(w.name), event)
	w.dispatcher.Dispatch(EventNameGuardNamed(w.name, event.Transition().Name), event)
}
