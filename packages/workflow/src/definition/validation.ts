import type { Definition } from '#workflow/definition/definition';

export const validateDefinition = (definition: Definition): void => {
	if (definition.places.length === 0) {
		throw new Error('definition requires at least one place');
	}

	for (const transition of definition.transitions) {
		if (transition.name === '') {
			throw new Error('transition name cannot be empty');
		}

		if (transition.from.length === 0) {
			throw new Error(`transition "${transition.name}" requires at least one from place`);
		}

		if (transition.to.length === 0) {
			throw new Error(`transition "${transition.name}" requires at least one to place`);
		}

		for (const place of [...transition.from, ...transition.to]) {
			if (!definition.hasPlace(place)) {
				throw new Error(`transition "${transition.name}" references unknown place "${place}"`);
			}
		}
	}

	if (definition.initialMarking.activePlaces().length === 0) {
		throw new Error('definition requires an initial marking');
	}

	for (const place of definition.initialMarking.activePlaces()) {
		if (!definition.hasPlace(place)) {
			throw new Error(`initial marking references unknown place "${place}"`);
		}
	}

	validateReachability(definition);
};

const validateReachability = (definition: Definition): void => {
	const reachable = new Set<string>();
	const queue = definition.initialMarking.activePlaces();

	for (const place of queue) {
		reachable.add(place);
	}

	while (queue.length > 0) {
		const current = queue.shift();

		if (current === undefined) {
			break;
		}

		for (const transition of definition.transitions) {
			if (!transition.from.includes(current)) {
				continue;
			}

			for (const place of transition.to) {
				if (!reachable.has(place)) {
					reachable.add(place);
					queue.push(place);
				}
			}
		}
	}

	for (const place of definition.places) {
		if (!reachable.has(place)) {
			throw new Error(`place "${place}" is unreachable from the initial marking`);
		}
	}
};
