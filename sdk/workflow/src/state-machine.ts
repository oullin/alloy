import type { Definition } from '#workflow/definition';
import { Dispatcher } from '#workflow/events';
import { Machine } from '#workflow/machine';
import type { MarkingStore } from '#workflow/stores';

/**
 * A {@link Machine} restricted to classic state-machine semantics: exactly
 * one initial place and single-target transitions, so a subject occupies one
 * state at a time.
 */
export class StateMachine<T> extends Machine<T> {
	/** @throws Error when the definition allows more than one active place. */
	public constructor(name: string, definition: Definition, store: MarkingStore<T>, dispatcher = new Dispatcher<T>()) {
		if (definition.initialMarking.activePlaces().length !== 1) {
			throw new Error('state machine requires exactly one initial place');
		}

		for (const transition of definition.transitions) {
			if (transition.to.length !== 1) {
				throw new Error(`state machine transition "${transition.name}" must target exactly one place`);
			}
		}

		super(name, definition, store, dispatcher);
	}
}
