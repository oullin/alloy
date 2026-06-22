import { spin } from '#console/status';
import { resolveSpinFormArguments } from '#console/form/builder/validators/status';
import type { FormBuilder } from '#console/form/builder/index';
import type { MaybePromise } from '#console/types';

export function spinFormStep<T>(this: FormBuilder, callback: () => MaybePromise<T>, message?: string, name?: string): FormBuilder;

export function spinFormStep<T>(this: FormBuilder, message: string, callback: () => MaybePromise<T>, name?: string): FormBuilder;

export function spinFormStep<T>(this: FormBuilder, callbackOrMessage: (() => MaybePromise<T>) | string, messageOrCallback: string | (() => MaybePromise<T>) = '', name?: string): FormBuilder {
	const resolved = resolveSpinFormArguments(callbackOrMessage, messageOrCallback, name);

	if (resolved.kind === 'message') {
		return this.add(() => spin(resolved.message, resolved.callback), resolved.name, true);
	}

	return this.add(() => spin(resolved.callback, { message: resolved.message }), resolved.name, true);
}
