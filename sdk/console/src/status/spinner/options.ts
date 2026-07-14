import { isSpinnerCallback, isSpinnerMessage, parseSpinnerCallback, parseSpinnerOptions } from '#console/status/spinner/validators/options';
import type { MaybePromise, StatusOptions } from '#console/types';

export type ResolvedSpinnerOptions<T> = {
	callback: () => MaybePromise<T>;
	message: string;
};

export const resolveSpinnerOptions = <T>(callbackOrMessage: (() => MaybePromise<T>) | string, optionsOrCallback: StatusOptions | (() => MaybePromise<T>)): ResolvedSpinnerOptions<T> => {
	if (isSpinnerMessage(callbackOrMessage)) {
		return {
			callback: parseSpinnerCallback<T>(optionsOrCallback),
			message: callbackOrMessage,
		};
	}

	return {
		callback: parseSpinnerCallback<T>(callbackOrMessage),
		message: isSpinnerCallback<T>(optionsOrCallback) ? '' : parseSpinnerOptions(optionsOrCallback).message,
	};
};
