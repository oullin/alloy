import type { Definition } from '#workflow/definition';
import type { Marking } from '#workflow/marking';
import type { Transition } from '#workflow/transition';
import { TransitionNotFoundError } from '#workflow/errors';
import { buildNextMarking } from '#workflow/machine/markings';

export interface TransitionPlan {
	transition: Transition;
	next: Marking;
}

export const resolveTransition = (definition: Definition, transitionName: string): Transition => {
	const transition = definition.transition(transitionName);

	if (transition === undefined) {
		throw new TransitionNotFoundError(transitionName);
	}

	return transition;
};

export const createTransitionPlan = (marking: Marking, transition: Transition): TransitionPlan => ({
	transition,
	next: buildNextMarking(marking, transition),
});
