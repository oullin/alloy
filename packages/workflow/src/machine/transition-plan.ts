import type { Definition } from '../definition.js';
import type { Marking } from '../marking.js';
import type { Transition } from '../transition.js';
import { TransitionNotFoundError } from '../errors.js';
import { buildNextMarking } from './markings.js';

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
