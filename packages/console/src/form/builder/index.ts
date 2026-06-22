import { createFormResponses } from '#console/form/builder/responses';
import { sideEffectStep } from '#console/form/builder/step';
import { submitFormSteps } from '#console/form/builder/submit';
import { applyFormBuilderMethods, FormBuilderMethodSurface } from '#console/form/builder/methods';
import type { FormResponses, FormStep, FormStepCondition } from '#console/form/types';
import type { MaybePromise } from '#console/types';

export class FormBuilder extends FormBuilderMethodSurface {
	readonly #steps: FormStep[] = [];
	readonly #responses: FormResponses = createFormResponses();

	constructor() {
		super();
		applyFormBuilderMethods(this);
	}

	add(step: (responses: FormResponses, previous: unknown, name?: string) => MaybePromise<unknown>, name?: string, ignoreWhenReverting = false): this {
		this.#steps.push({
			condition: true,
			ignoreWhenReverting,
			name,
			run: step,
		});

		return this;
	}

	addIf(condition: FormStepCondition, step: (responses: FormResponses, previous: unknown, name?: string) => MaybePromise<unknown>, name?: string, ignoreWhenReverting = false): this {
		this.#steps.push({
			condition,
			ignoreWhenReverting,
			name,
			run: step,
		});

		return this;
	}

	addSideEffect(step: () => MaybePromise<void>, name?: string): this {
		return this.add(sideEffectStep(step), name, true);
	}

	addSideEffectIf(condition: FormStepCondition, step: () => MaybePromise<void>, name?: string): this {
		return this.addIf(condition, sideEffectStep(step), name, true);
	}

	async submit(): Promise<FormResponses> {
		return submitFormSteps(this.#steps, this.#responses);
	}
}
