export class RetryPolicy {
	public readonly maxTries: number;
	public readonly backoff: number[];
	public readonly timeout: number;
	/**
	 * maxExceptions, when > 0, caps the total number of exceptions the handler
	 * may throw before the retry loop gives up — counting every thrown error,
	 * not distinct error types. It binds independently of {@link maxTries}:
	 * whichever limit is reached first stops the loop. Zero disables the cap.
	 */
	public readonly maxExceptions: number;

	public constructor(input: { maxTries?: number; backoff?: number[]; timeout?: number; maxExceptions?: number } = {}) {
		this.maxTries = input.maxTries !== undefined && input.maxTries > 0 ? input.maxTries : 1;
		this.backoff = [...(input.backoff ?? [])];
		this.timeout = input.timeout ?? 0;
		this.maxExceptions = input.maxExceptions !== undefined && input.maxExceptions > 0 ? input.maxExceptions : 0;
	}

	public backoffFor(attempt: number): number {
		if (this.backoff.length === 0) {
			return 0;
		}

		return this.backoff[Math.min(attempt, this.backoff.length - 1)] ?? 0;
	}
}

export const runWithRetry = async (
	signal: AbortSignal,
	policy: RetryPolicy | undefined,
	handler: (signal: AbortSignal) => unknown,
): Promise<{ value: unknown; attempts: number; error?: unknown }> => {
	const activePolicy = policy ?? new RetryPolicy();

	let lastError: unknown;
	let exceptions = 0;
	let attempts = 0;

	for (let attempt = 0; attempt < activePolicy.maxTries; attempt++) {
		const attemptSignal = createAttemptSignal(signal, activePolicy.timeout);

		attempts = attempt + 1;

		try {
			const value = await handler(attemptSignal.signal);

			attemptSignal.cleanup();

			return { value, attempts };
		} catch (error) {
			attemptSignal.cleanup();
			exceptions++;
			lastError = signal.aborted ? signal.reason : error;

			const exhaustedExceptions = activePolicy.maxExceptions > 0 && exceptions >= activePolicy.maxExceptions;

			if (attempt + 1 >= activePolicy.maxTries || signal.aborted || exhaustedExceptions) {
				break;
			}

			const backoff = activePolicy.backoffFor(attempt);

			if (backoff > 0) {
				try {
					await sleep(backoff, signal);
				} catch (sleepError) {
					// The run signal aborted mid-backoff; surface it through the
					// normal result contract instead of letting the rejection
					// escape runWithRetry uncaught.
					lastError = signal.aborted ? signal.reason : sleepError;

					break;
				}
			}
		}
	}

	return { value: undefined, attempts, error: lastError };
};

const createAttemptSignal = (parent: AbortSignal, timeout: number): { signal: AbortSignal; cleanup: () => void } => {
	const controller = new AbortController();
	const abortFromParent = (): void => controller.abort(parent.reason);
	const timer = timeout > 0 ? setTimeout(() => controller.abort(new Error('operation timed out')), timeout) : undefined;

	if (parent.aborted) {
		abortFromParent();
	} else {
		parent.addEventListener('abort', abortFromParent, { once: true });
	}

	return {
		signal: controller.signal,
		cleanup: () => {
			parent.removeEventListener('abort', abortFromParent);

			if (timer !== undefined) {
				clearTimeout(timer);
			}
		},
	};
};

const sleep = async (ms: number, signal: AbortSignal): Promise<void> => {
	await new Promise<void>((resolve, reject) => {
		let timer: NodeJS.Timeout | undefined;

		const abort = (): void => {
			if (timer !== undefined) {
				clearTimeout(timer);
			}

			signal.removeEventListener('abort', abort);
			reject(signal.reason ?? new Error('operation aborted'));
		};

		if (signal.aborted) {
			abort();

			return;
		}

		timer = setTimeout(() => {
			signal.removeEventListener('abort', abort);
			resolve();
		}, ms);

		signal.addEventListener('abort', abort, { once: true });
	});
};
