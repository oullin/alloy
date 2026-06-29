import { AsyncLocalStorage } from 'node:async_hooks';
import { defaultEnvironment } from '#console/environment/runtime';
import { mergePromptEnvironment } from '#console/environment/scope/merge';
import { withPromptCancelScope } from '#console/prompt/cancel';
import { withPromptFallbackScope } from '#console/prompt/fallback';
import { withPromptGlobalValidationScope } from '#console/prompt/validate-using';
import type { MaybePromise, PromptEnvironment } from '#console/types';

let currentEnvironment: PromptEnvironment = defaultEnvironment;

const scopedEnvironment = new AsyncLocalStorage<PromptEnvironment>();

export const promptEnvironment = (): PromptEnvironment => scopedEnvironment.getStore() ?? currentEnvironment;

export const configurePrompts = (environment: Partial<PromptEnvironment>): void => {
	const scoped = scopedEnvironment.getStore();
	const next = mergePromptEnvironment(scoped ?? currentEnvironment, environment);

	if (scoped === undefined) {
		currentEnvironment = next;

		return;
	}

	scopedEnvironment.enterWith(next);
};

export const withPromptEnvironment = async <T>(environment: Partial<PromptEnvironment>, callback: () => MaybePromise<T>): Promise<T> => {
	const next = mergePromptEnvironment(promptEnvironment(), environment);

	return scopedEnvironment.run(next, () => withPromptCancelScope(() => withPromptFallbackScope(() => withPromptGlobalValidationScope(callback))));
};
