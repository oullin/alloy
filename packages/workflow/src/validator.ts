import type { Definition } from './definition.js';

export class WorkflowValidator {
	public validateDefinition(definition: Definition | null | undefined): void {
		if (definition === null || definition === undefined) {
			throw new Error('definition is required');
		}

		definition.validate();
	}

	public validateStateMachine(definition: Definition | null | undefined): void {
		this.validateDefinition(definition);

		if (definition === null || definition === undefined) {
			throw new Error('definition is required');
		}

		if (definition.initialMarking.activePlaces().length !== 1) {
			throw new Error('state machine requires exactly one initial place');
		}

		for (const transition of definition.transitions) {
			if (transition.to.length !== 1) {
				throw new Error(`state machine transition "${transition.name}" must target exactly one place`);
			}
		}
	}
}
