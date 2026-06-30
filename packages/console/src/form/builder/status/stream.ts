import { stream } from '#console/status';
import { resolveStreamFormArguments } from '#console/form/builder/validators/status';
import type { FormBuilder } from '#console/form/builder/index';

export function streamFormStep(this: FormBuilder, name?: string): FormBuilder;

export function streamFormStep(this: FormBuilder): FormBuilder;

export function streamFormStep(this: FormBuilder, source: AsyncIterable<string> | Iterable<string>, name?: string): FormBuilder;

export function streamFormStep(this: FormBuilder, sourceOrName?: AsyncIterable<string> | Iterable<string> | string, name?: string): FormBuilder {
	const resolved = resolveStreamFormArguments(sourceOrName, name);

	if (resolved.kind === 'manual') {
		return this.add(() => stream(), resolved.name);
	}

	return this.addSideEffect(() => stream(resolved.source), resolved.name);
}
