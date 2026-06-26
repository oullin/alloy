import { RetryPolicy } from './retry.js';
import type { ArgMap, JobHandler, RunIfPredicate } from './types.js';

export abstract class JobSpec {
	public readonly name: string;
	public readonly handler: JobHandler;
	readonly #args: ArgMap = {};
	#retry?: RetryPolicy;
	#runIf?: RunIfPredicate;
	readonly #dependsOn: string[] = [];

	protected constructor(name: string, handler: JobHandler) {
		this.name = name;
		this.handler = handler;
	}

	public abstract isAsync(): boolean;

	public args(): ArgMap {
		return { ...this.#args };
	}

	public retryPolicy(): RetryPolicy | undefined {
		return this.#retry;
	}

	public runIfPredicate(): RunIfPredicate | undefined {
		return this.#runIf;
	}

	public dependencies(): string[] {
		return [...this.#dependsOn];
	}

	public withArgs(args: ArgMap): this {
		Object.assign(this.#args, args);

		return this;
	}

	public withRetry(maxTries: number, delay = 0, timeout = 0): this {
		this.#retry = new RetryPolicy({ maxTries, backoff: [delay], timeout });

		return this;
	}

	public withRetryPolicy(policy: RetryPolicy): this {
		this.#retry = new RetryPolicy({
			maxTries: policy.maxTries,
			backoff: policy.backoff,
			timeout: policy.timeout,
			maxExceptions: policy.maxExceptions,
		});

		return this;
	}

	public withRunIf(predicate: RunIfPredicate): this {
		this.#runIf = predicate;

		return this;
	}

	public dependsOn(...jobs: string[]): this {
		this.#dependsOn.push(...jobs);

		return this;
	}
}

export class SyncJob extends JobSpec {
	public constructor(name: string, handler: JobHandler) {
		super(name, handler);
	}

	public isAsync(): boolean {
		return false;
	}
}

export class AsyncJob extends JobSpec {
	public constructor(name: string, handler: JobHandler) {
		super(name, handler);
	}

	public isAsync(): boolean {
		return true;
	}
}
