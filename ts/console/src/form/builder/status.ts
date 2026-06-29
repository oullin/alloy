import { progressFormStep, spinFormStep, streamFormStep, taskFormStep } from '#console/form/builder/status/index';
import type { FormBuilder } from '#console/form/builder/index';
import type { MaybePromise } from '#console/types';
import type { Logger, Progress, TaskDefinition } from '#console/status';

export type StatusBuilderMethods = {
	progress(this: FormBuilder, total: number, message?: string, name?: string): FormBuilder;
	progress<T, R>(this: FormBuilder, label: string, steps: Iterable<T> | number, callback: (step: T | number, bar: Progress) => MaybePromise<R>, hint?: string, name?: string): FormBuilder;
	spin<T>(this: FormBuilder, callback: () => MaybePromise<T>, message?: string, name?: string): FormBuilder;
	spin<T>(this: FormBuilder, message: string, callback: () => MaybePromise<T>, name?: string): FormBuilder;
	task<T>(this: FormBuilder, definition: TaskDefinition<T>, name?: string): FormBuilder;
	stream(this: FormBuilder, name?: string): FormBuilder;
	stream(this: FormBuilder): FormBuilder;
	stream(this: FormBuilder, source: AsyncIterable<string> | Iterable<string>, name?: string): FormBuilder;
	task<T>(this: FormBuilder, label: string, callback: (logger: Logger) => MaybePromise<T>, limit?: number, keepSummary?: boolean, subLabel?: string, name?: string): FormBuilder;
};

export const statusBuilderMethods: StatusBuilderMethods & ThisType<FormBuilder> = {
	progress: progressFormStep,
	spin: spinFormStep,
	stream: streamFormStep,
	task: taskFormStep,
};
