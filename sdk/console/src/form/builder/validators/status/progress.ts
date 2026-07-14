import { isProgressTotal, isStatusLabel, parseRequiredProgressCallback, parseStatusLabel } from '#console/form/builder/validators/status/common';
import { progressStepsArgument } from '#console/status/progress/validators/arguments';
import type { MaybePromise } from '#console/types';
import type { Progress } from '#console/status';

export type ResolvedProgressFormArguments<T, R> =
	| {
			kind: 'total';
			message?: string;
			name?: string;
			total: number;
	  }
	| {
			callback: (step: T | number, bar: Progress) => MaybePromise<R>;
			hint: string;
			kind: 'label';
			label: string;
			name?: string;
			steps: Iterable<T> | number;
	  };

export const resolveProgressFormArguments = <T, R>(
	labelOrTotal: string | number,
	stepsOrMessage?: Iterable<T> | number | string,
	callbackOrName?: ((step: T | number, bar: Progress) => MaybePromise<R>) | string,
	hint = '',
	name?: string,
): ResolvedProgressFormArguments<T, R> => {
	if (isProgressTotal(labelOrTotal)) {
		return {
			kind: 'total',
			message: isStatusLabel(stepsOrMessage) ? stepsOrMessage : undefined,
			name: parseStatusLabel(callbackOrName),
			total: labelOrTotal,
		};
	}

	return {
		callback: parseRequiredProgressCallback<T, R>(callbackOrName),
		hint,
		kind: 'label',
		label: labelOrTotal,
		name,
		steps: progressStepsArgument<T>(stepsOrMessage),
	};
};
