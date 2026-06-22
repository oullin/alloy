export { FormBuilder } from '#console/form/builder/index';
export type { FormResponses, FormStep, FormStepCondition } from '#console/form/types';

import { FormBuilder } from '#console/form/builder/index';
import type { MaybePromise } from '#console/types';

export function form(): FormBuilder;

export function form<T>(builder: (form: FormBuilder) => MaybePromise<T>): Promise<T>;

export function form<T>(builder?: (form: FormBuilder) => MaybePromise<T>): FormBuilder | Promise<T> {
	const instance = new FormBuilder();

	return builder ? Promise.resolve(builder(instance)) : instance;
}
