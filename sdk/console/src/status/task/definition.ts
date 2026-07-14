import { isTaskTitle, parseTaskCallback, parseTaskDefinition } from '#console/status/task/validators/definition';
import type { Logger } from '#console/status/task/logger';
import type { MaybePromise } from '#console/types';

export type TaskDefinition<T> = {
	keepSummary?: boolean;
	subLabel?: string;
	title: string;
	limit?: number;
	task: (logger: Logger) => MaybePromise<T>;
};

export type ResolvedTaskDefinition<T> = {
	keepSummary: boolean;
	limit: number;
	run: (logger: Logger) => MaybePromise<T>;
	subLabel: string;
	title: string;
};

export const resolveTaskDefinition = <T>(
	definitionOrLabel: TaskDefinition<T> | string,
	callback?: (logger: Logger) => MaybePromise<T>,
	limit = 10,
	keepSummary = false,
	subLabel?: string,
): ResolvedTaskDefinition<T> => {
	if (isTaskTitle(definitionOrLabel)) {
		return {
			keepSummary,
			limit,
			run: parseTaskCallback(callback),
			subLabel: subLabel ?? '',
			title: definitionOrLabel,
		};
	}

	const definition = parseTaskDefinition<T>(definitionOrLabel);

	return {
		keepSummary: definition.keepSummary ?? keepSummary,
		limit: definition.limit ?? limit,
		run: definition.task,
		subLabel: definition.subLabel ?? subLabel ?? '',
		title: definition.title,
	};
};
