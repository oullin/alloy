package workflow

import "fmt"

// Apply fires the named transition for the subject and returns the new marking.
//
// The read (GetMarking), guard check, and write (SetMarking) run as one atomic
// critical section under the machine lock so concurrent Applies cannot both
// validate against a stale marking and double-consume tokens. Transition events
// are dispatched only after the write succeeds — a failed write emits no
// leave/transition/enter/entered events, so it can never leave phantom
// transitions in listeners. Events are dispatched outside the lock so listeners
// may safely apply further transitions without deadlocking.
func (w *Machine[T]) Apply(subject T, transitionName string, context map[string]any) (Marking, error) {
	w.logDebug("applying transition", "transition", transitionName, "workflow", w.name)

	transition, next, err := w.commitApply(subject, transitionName, context)

	if err != nil {
		return Marking{}, err
	}

	w.logInfo("transition applied successfully", "transition", transitionName, "workflow", w.name)

	w.dispatchLeaveEvents(subject, transition, next, context)
	w.dispatchTransitionEvents(subject, transition, next, context)
	w.dispatchEnterEvents(subject, transition, next, context)
	w.dispatchEnteredEvents(subject, transition, next, context)

	if err := w.dispatchCompletionEvents(subject, transition, next, context); err != nil {
		w.logError("failed to dispatch completion events", "transition", transitionName, "workflow", w.name, "error", err)

		// The transition itself committed; only the post-commit completion
		// dispatch failed. Return the committed marking alongside the error so
		// callers can tell the store state advanced.
		return next, err
	}

	return next, nil
}

// commitApply runs the read-guard-write step under the machine lock so the
// guard decision and the marking write are atomic with respect to other Applies
// on this machine. It performs no event dispatch beyond the guard (which is part
// of the guard decision itself).
func (w *Machine[T]) commitApply(subject T, transitionName string, context map[string]any) (Transition, Marking, error) {
	w.applyMu.Lock()

	defer w.applyMu.Unlock()

	transition, next, err := w.prepareApply(subject, transitionName, context)

	if err != nil {
		w.logError("failed to prepare transition", "transition", transitionName, "workflow", w.name, "error", err)

		return Transition{}, Marking{}, err
	}

	if err := w.store.SetMarking(subject, next, w.definition, context); err != nil {
		w.logError("failed to set marking", "transition", transitionName, "workflow", w.name, "error", err)

		return Transition{}, Marking{}, err
	}

	return transition, next, nil
}

func (w *Machine[T]) prepareApply(subject T, transitionName string, context map[string]any) (Transition, Marking, error) {
	transition, ok := w.definition.Transition(transitionName)

	if !ok {
		return Transition{}, Marking{}, fmt.Errorf("%w: %s", ErrTransitionNotFound, transitionName)
	}

	marking, err := w.GetMarking(subject)

	if err != nil {
		return Transition{}, Marking{}, err
	}

	if err := w.transitionStateForMarking(subject, transition, marking, context); err != nil {
		return Transition{}, Marking{}, err
	}

	return transition, buildNextMarking(marking, transition), nil
}

func buildNextMarking(marking Marking, transition Transition) Marking {
	next := marking.Clone()

	for _, place := range transition.From {
		next.Remove(place, 1)
	}

	for _, place := range transition.To {
		next.Add(place, 1)
	}

	return next
}
