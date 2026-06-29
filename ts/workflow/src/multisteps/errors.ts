export class WorkflowError extends Error {
	public readonly job: string;
	public readonly attempts: number;
	public readonly cause: unknown;

	public constructor(job: string, attempts: number, cause: unknown) {
		super(`multisteps: job "${job}" failed after ${attempts} attempt(s): ${formatCause(cause)}`);
		this.name = 'WorkflowError';
		this.job = job;
		this.attempts = attempts;
		this.cause = cause;
	}
}

export class UnresolvedResponseError extends Error {
	public readonly job: string;
	public readonly field: string;

	public constructor(field: string, job = '') {
		super(job === '' ? `multisteps: response field "${field}" not resolvable` : `multisteps: response field "${field}" on job "${job}" not resolvable`);
		this.name = 'UnresolvedResponseError';
		this.job = job;
		this.field = field;
	}
}

export class CompileError extends Error {
	public readonly reason: string;

	public constructor(reason: string) {
		super(`multisteps: compile error: ${reason}`);
		this.name = 'CompileError';
		this.reason = reason;
	}
}

const formatCause = (cause: unknown): string => {
	if (cause instanceof Error) {
		return cause.message;
	}

	return String(cause);
};
