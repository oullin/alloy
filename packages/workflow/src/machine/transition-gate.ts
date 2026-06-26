import { EventNames } from '../event-names.js';
import { TransitionError } from '../errors.js';
import { BaseEvent, type Dispatcher, GuardEvent } from '../events.js';
import type { Marking } from '../marking.js';
import { snapshotTransition, type Transition } from '../transition.js';
import type { WorkflowContext } from '../types.js';
import { transitionEnabled } from './markings.js';

export const assertTransitionAllowed = <T>(input: { machine: string; dispatcher: Dispatcher<T>; subject: T; transition: Transition; marking: Marking; context: WorkflowContext }): void => {
	if (!transitionEnabled(input.marking, input.transition)) {
		throw new TransitionError(input.machine, input.transition.name);
	}

	const guard = new GuardEvent(
		new BaseEvent(input.machine, input.subject, snapshotTransition(input.transition), input.marking.clone().toJSON(), {
			...input.context,
		}),
	);

	input.dispatcher.dispatch(EventNames.guard(input.machine), guard);
	input.dispatcher.dispatch(EventNames.guardNamed(input.machine, input.transition.name), guard);

	if (guard.blocked()) {
		throw new TransitionError(input.machine, input.transition.name, guard.blockers());
	}
};
