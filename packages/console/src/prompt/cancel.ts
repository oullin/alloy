import { AsyncLocalStorage } from 'node:async_hooks';
import { parseCancelValue } from '#console/prompt/validators/cancel';
import type { MaybePromise } from '#console/types';

export type PromptCancelHandler = () => MaybePromise<unknown>;

type PromptCancelState = {
	handler: PromptCancelHandler | null;
};

const scopedCancelState = new AsyncLocalStorage<PromptCancelState>();
const currentCancelState: PromptCancelState = { handler: null };

const promptCancelState = (): PromptCancelState => scopedCancelState.getStore() ?? currentCancelState;

export const cancelUsing = (handler?: PromptCancelHandler | null): void => {
	promptCancelState().handler = handler ?? null;
};

export const cancelPrompt = async <T>(fallback: T): Promise<T> => {
	const handler = promptCancelState().handler;

	if (handler === null) {
		return fallback;
	}

	return parseCancelValue<T>(await handler());
};

export const withPromptCancelScope = <T>(callback: () => MaybePromise<T>): Promise<T> => {
	return Promise.resolve(scopedCancelState.run({ ...promptCancelState() }, callback));
};
