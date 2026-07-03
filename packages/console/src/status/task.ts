import { resolveTaskDefinition } from '#console/status/task/definition';
import { runTaskLifecycle } from '#console/status/task/lifecycle';
import { Logger } from '#console/status/task/logger';
import type { TaskDefinition } from '#console/status/task/definition';
import type { MaybePromise } from '#console/types';

export { Logger };
export type { TaskDefinition } from '#console/status/task/definition';

/** Runs a labelled task with a scoped {@link Logger} that renders live output lines. */
export async function task<T>(definition: TaskDefinition<T>): Promise<T>;

export async function task<T>(label: string, callback: (logger: Logger) => MaybePromise<T>, limit?: number, keepSummary?: boolean, subLabel?: string): Promise<T>;

export async function task<T>(definitionOrLabel: TaskDefinition<T> | string, callback?: (logger: Logger) => MaybePromise<T>, limit = 10, keepSummary = false, subLabel?: string): Promise<T> {
	return runTaskLifecycle(resolveTaskDefinition(definitionOrLabel, callback, limit, keepSummary, subLabel));
}
