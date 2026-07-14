import { clear, title } from '#console/output';
import type { FormBuilder } from '#console/form/builder/index';

export const clearFormStep = function clearFormStep(this: FormBuilder, name?: string): FormBuilder {
	return this.addSideEffect(() => clear(), name);
};

export const titleFormStep = function titleFormStep(this: FormBuilder, value: string, name?: string): FormBuilder {
	return this.addSideEffect(() => title(value), name);
};
