import type { Marking } from '../marking.js';
import type { Transition } from '../transition.js';

export const buildNextMarking = (marking: Marking, transition: Transition): Marking => {
	const next = marking.clone();

	for (const place of transition.from) {
		next.remove(place, 1);
	}

	for (const place of transition.to) {
		next.add(place, 1);
	}

	return next;
};

export const transitionEnabled = (marking: Marking, transition: Transition): boolean => transition.from.every((place) => marking.has(place));

export const markingBeforeEnter = (next: Marking, transition: Transition): Marking => {
	const current = next.clone();

	for (const place of transition.to) {
		current.remove(place, 1);
	}

	return current;
};

export const restoreFromPlaces = (marking: Marking, transition: Transition): Marking => {
	const current = marking.clone();

	for (const place of [...transition.from].reverse()) {
		current.add(place, 1);
	}

	return current;
};
