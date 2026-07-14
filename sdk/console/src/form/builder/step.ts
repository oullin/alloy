import type { MaybePromise } from '#console/types';

export const sideEffectStep = (callback: () => MaybePromise<void>): (() => MaybePromise<null>) => {
	return async () => {
		await callback();

		return null;
	};
};
