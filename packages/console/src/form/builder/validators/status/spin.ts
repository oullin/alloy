import { isStatusLabel, parseStatusCallback, parseStatusLabel } from '#console/form/builder/validators/status/common';
import type { MaybePromise } from '#console/types';

export type ResolvedSpinFormArguments<T> =
	| {
			callback: () => MaybePromise<T>;
			kind: 'message';
			message: string;
			name?: string;
	  }
	| {
			callback: () => MaybePromise<T>;
			kind: 'callback';
			message: string;
			name?: string;
	  };

export const resolveSpinFormArguments = <T>(
	callbackOrMessage: (() => MaybePromise<T>) | string,
	messageOrCallback: string | (() => MaybePromise<T>) = '',
	name?: string,
): ResolvedSpinFormArguments<T> => {
	if (isStatusLabel(callbackOrMessage)) {
		return {
			callback: parseStatusCallback<T>(messageOrCallback),
			kind: 'message',
			message: callbackOrMessage,
			name,
		};
	}

	return {
		callback: parseStatusCallback<T>(callbackOrMessage),
		kind: 'callback',
		message: parseStatusLabel(messageOrCallback) ?? '',
		name,
	};
};
