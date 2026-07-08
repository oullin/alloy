import type { Arg } from '#workflow/multisteps/args';

export type Task = () => Promise<unknown>;

export interface Driver {
	run(signal: AbortSignal, tasks: Task[]): Promise<unknown[]>;
}

export type JobHandler = (input: JobInput) => unknown;

export type RunIfPredicate = (input: JobInput) => boolean;

export type ArgMap = Record<string, Arg>;

export interface JobInput {
	signal: AbortSignal;
	args: ArgMap;
	resolved: Record<string, unknown>;
	vars: Record<string, unknown>;
	responses: Record<string, unknown>;
}
