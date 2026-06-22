import { pause } from '#console/prompts/choices';
import type { FormBuilder } from '#console/form/builder/index';

export const pauseFormStep = function pauseFormStep(this: FormBuilder, message = 'Press enter to continue...', name?: string): FormBuilder {
	return this.add(() => pause(message), name);
};
