import { alert, error, grid, info, intro, note, outro, warning } from '#console/output';
import type { FormBuilder } from '#console/form/builder/index';

export const alertFormStep = function alertFormStep(this: FormBuilder, message: string, name?: string): FormBuilder {
	return this.addSideEffect(() => alert(message), name);
};

export const errorFormStep = function errorFormStep(this: FormBuilder, message: string, name?: string): FormBuilder {
	return this.addSideEffect(() => error(message), name);
};

export const gridFormStep = function gridFormStep(this: FormBuilder, items: Array<string | number | boolean> = [], maxWidth?: number, name?: string): FormBuilder {
	return this.addSideEffect(() => grid(items, maxWidth), name);
};

export const infoFormStep = function infoFormStep(this: FormBuilder, message: string, name?: string): FormBuilder {
	return this.addSideEffect(() => info(message), name);
};

export const introFormStep = function introFormStep(this: FormBuilder, message: string, name?: string): FormBuilder {
	return this.addSideEffect(() => intro(message), name);
};

export const noteFormStep = function noteFormStep(this: FormBuilder, message: string, type: string | null = null, name?: string): FormBuilder {
	return this.addSideEffect(() => note(message, type), name);
};

export const outroFormStep = function outroFormStep(this: FormBuilder, message: string, name?: string): FormBuilder {
	return this.addSideEffect(() => outro(message), name);
};

export const warningFormStep = function warningFormStep(this: FormBuilder, message: string, name?: string): FormBuilder {
	return this.addSideEffect(() => warning(message), name);
};
