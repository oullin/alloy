import { z } from 'zod';
import { functionSchema } from '#console/validators/function';
import type { TaskDefinition } from '#console/status/task/definition';
import type { Logger } from '#console/status/task/logger';
import type { MaybePromise } from '#console/types';

const taskCallbackSchema = <T>(): z.ZodType<(logger: Logger) => MaybePromise<T>> => functionSchema<(logger: Logger) => MaybePromise<T>>();

const taskDefinitionSchema = <T>(): z.ZodType<TaskDefinition<T>> =>
	z
		.object({
			keepSummary: z.boolean().optional(),
			limit: z.number().optional(),
			subLabel: z.string().optional(),
			task: taskCallbackSchema<T>(),
			title: z.string(),
		})
		.passthrough() as z.ZodType<TaskDefinition<T>>;

const taskTitleSchema = z.string();

export const isTaskTitle = (value: unknown): value is string => {
	return taskTitleSchema.safeParse(value).success;
};

export const parseTaskCallback = <T>(value: unknown): ((logger: Logger) => MaybePromise<T>) => {
	const parsed = taskCallbackSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new Error('A task callback is required.');
	}

	return parsed.data;
};

export const parseTaskDefinition = <T>(value: unknown): TaskDefinition<T> => {
	const parsed = taskDefinitionSchema<T>().safeParse(value);

	if (parsed.success) {
		return parsed.data;
	}

	const callback = z.object({ task: taskCallbackSchema<T>() }).passthrough().safeParse(value);

	if (!callback.success) {
		throw new Error('A task callback is required.');
	}

	throw new Error('A task title is required.');
};
