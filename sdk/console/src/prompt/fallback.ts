import { AsyncLocalStorage } from 'node:async_hooks';
import { parseFallbackHandler, resolveFallbackCondition } from '#console/prompt/validators/fallback';
import type { MaybePromise } from '#console/types';

export type PromptFallbackKind =
	| 'autocomplete'
	| 'confirm'
	| 'datatable'
	| 'multisearch'
	| 'multiselect'
	| 'number'
	| 'password'
	| 'pause'
	| 'search'
	| 'select'
	| 'suggest'
	| 'text'
	| 'textarea';

export type PromptFallbackHandler<TOptions = unknown, TResult = unknown> = (options: TOptions) => MaybePromise<TResult>;

export type PromptFallbackCondition = boolean | (() => boolean);

type PromptFallbackState = {
	condition: PromptFallbackCondition;
	handlers: Map<PromptFallbackKind, PromptFallbackHandler>;
};

const scopedFallbackState = new AsyncLocalStorage<PromptFallbackState>();

const currentFallbackState: PromptFallbackState = {
	condition: false,
	handlers: new Map(),
};

const cloneFallbackState = (state: PromptFallbackState): PromptFallbackState => ({
	condition: state.condition,
	handlers: new Map(state.handlers),
});

const promptFallbackState = (): PromptFallbackState => scopedFallbackState.getStore() ?? currentFallbackState;

export const fallbackWhen = (condition: PromptFallbackCondition): void => {
	promptFallbackState().condition = condition;
};

export const fallbackUsing = <TOptions, TResult>(kind: PromptFallbackKind, handler?: PromptFallbackHandler<TOptions, TResult> | null): void => {
	const state = promptFallbackState();

	if (!handler) {
		state.handlers.delete(kind);

		return;
	}

	state.handlers.set(kind, parseFallbackHandler<unknown, unknown>(handler));
};

export const shouldFallback = (kind: PromptFallbackKind): boolean => {
	const state = promptFallbackState();
	const enabled = resolveFallbackCondition(state.condition);

	return enabled && state.handlers.has(kind);
};

export const promptWithFallback = async <TOptions, TResult>(kind: PromptFallbackKind, options: TOptions, run: () => MaybePromise<TResult>): Promise<TResult> => {
	if (!shouldFallback(kind)) {
		return run();
	}

	const handlers = promptFallbackState().handlers;

	const fallback = handlers.has(kind) ? parseFallbackHandler<TOptions, TResult>(handlers.get(kind)) : undefined;

	if (!fallback) {
		return run();
	}

	return fallback(options);
};

export const withPromptFallbackScope = <T>(callback: () => MaybePromise<T>): Promise<T> => {
	return Promise.resolve(scopedFallbackState.run(cloneFallbackState(promptFallbackState()), callback));
};
