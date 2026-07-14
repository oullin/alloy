import { EventNames } from '#workflow/event-names';
import { TransitionError } from '#workflow/errors';
import { BaseEvent, type Dispatcher, GuardEvent } from '#workflow/events';
import type { Marking } from '#workflow/marking';
import { snapshotTransition, type Transition } from '#workflow/transition';
import type { WorkflowContext } from '#workflow/types';
import { transitionEnabled } from '#workflow/machine/markings';

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
