import { task } from '#console/status';
import { resolveTaskFormArguments } from '#console/form/builder/validators/status';
import type { FormBuilder } from '#console/form/builder/index';
import type { MaybePromise } from '#console/types';
import type { Logger, TaskDefinition } from '#console/status';

export function taskFormStep<T>(this: FormBuilder, definition: TaskDefinition<T>, name?: string): FormBuilder;

export function taskFormStep<T>(
	this: FormBuilder,
	label: string,
	callback: (logger: Logger) => MaybePromise<T>,
	limit?: number,
	keepSummary?: boolean,
	subLabel?: string,
	name?: string,
): FormBuilder;

export function taskFormStep<T>(
	this: FormBuilder,
	definitionOrLabel: TaskDefinition<T> | string,
	callbackOrName?: ((logger: Logger) => MaybePromise<T>) | string,
	limit = 10,
	keepSummary = false,
	subLabel = '',
	name?: string,
): FormBuilder {
	const resolved = resolveTaskFormArguments(definitionOrLabel, callbackOrName, limit, keepSummary, subLabel, name);

	if (resolved.kind === 'label') {
		return this.add(() => task(resolved.label, resolved.callback, resolved.limit, resolved.keepSummary, resolved.subLabel), resolved.name);
	}

	return this.add(() => task(resolved.definition), resolved.name);
}
