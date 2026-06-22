import { isStatusLabel, parseStatusLabel } from '#console/form/builder/validators/status/common';
import { parseTaskCallback } from '#console/status/task/validators/definition';
import type { MaybePromise } from '#console/types';
import type { TaskDefinition } from '#console/status';
import type { Logger } from '#console/status/task/logger';

export type ResolvedTaskFormArguments<T> =
	| {
			callback: (logger: Logger) => MaybePromise<T>;
			keepSummary: boolean;
			kind: 'label';
			label: string;
			limit: number;
			name?: string;
			subLabel: string;
	  }
	| {
			definition: TaskDefinition<T>;
			kind: 'definition';
			name?: string;
	  };

export const resolveTaskFormArguments = <T>(
	definitionOrLabel: TaskDefinition<T> | string,
	callbackOrName?: ((logger: Logger) => MaybePromise<T>) | string,
	limit = 10,
	keepSummary = false,
	subLabel = '',
	name?: string,
): ResolvedTaskFormArguments<T> => {
	if (isStatusLabel(definitionOrLabel)) {
		return {
			callback: parseTaskCallback<T>(callbackOrName),
			keepSummary,
			kind: 'label',
			label: definitionOrLabel,
			limit,
			name,
			subLabel,
		};
	}

	return {
		definition: definitionOrLabel,
		kind: 'definition',
		name: parseStatusLabel(callbackOrName),
	};
};
