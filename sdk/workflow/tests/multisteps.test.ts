import { describe, expect, it } from 'vite-plus/test';

import {
	AsyncJob,
	CompileError,
	type Driver,
	LiteralArg,
	MultiStepEngine,
	MultiStepResult,
	MultiStepWorkflow,
	ResponseArg,
	RetryPolicy,
	SyncJob,
	type Task,
	UnresolvedResponseError,
	VariableArg,
	WorkflowError,
} from '@alloy/sdk/workflow/multisteps';

class ParallelDriver implements Driver {
	public async run(_signal: AbortSignal, tasks: Task[]): Promise<unknown[]> {
		return Promise.all(tasks.map((task) => task()));
	}
}

class SerialDriver implements Driver {
	public async run(signal: AbortSignal, tasks: Task[]): Promise<unknown[]> {
		const results: unknown[] = [];

		for (const task of tasks) {
			if (signal.aborted) {
				throw signal.reason;
			}

			results.push(await task());
		}

		return results;
	}
}

describe('multisteps workflow', () => {
	it('runs a sync job before async siblings and resolves variables/responses', async () => {
		const order: string[] = [];

		const workflow = MultiStepWorkflow.machine(
			'signup',
			new SyncJob('create', ({ resolved }) => {
				order.push('create');

				return { id: `u-${String(resolved.name)}` };
			}).withArgs({ name: new VariableArg('name') }),
			new AsyncJob('email', async ({ resolved }) => {
				order.push(`email:${String(resolved.userId)}`);

				return { sent_to: resolved.userId };
			}).withArgs({ userId: new ResponseArg('create', 'id') }),
			new AsyncJob('notify', async ({ resolved }) => {
				order.push(`notify:${String(resolved.userId)}`);

				return { notified: resolved.userId };
			}).withArgs({ userId: new ResponseArg('create', 'id') }),
		);

		const result = await new MultiStepEngine({ driver: new ParallelDriver() }).run(workflow, { name: 'Jane' });

		expect(order[0]).toBe('create');
		expect(result.responses.create).toEqual({ id: 'u-Jane' });
		expect(result.responses.email).toEqual({ sent_to: 'u-Jane' });
		expect(result.responses.notify).toEqual({ notified: 'u-Jane' });
	});

	it('retries until success and wraps exhausted retries', async () => {
		let attempts = 0;

		const okWorkflow = MultiStepWorkflow.machine(
			'retry',
			new SyncJob('flaky', () => {
				attempts += 1;

				if (attempts < 3) {
					throw new Error('transient');
				}

				return 'ok';
			}).withRetry(3, 1),
		);

		await expect(new MultiStepEngine().run(okWorkflow)).resolves.toMatchObject({ responses: { flaky: 'ok' } });

		expect(attempts).toBe(3);

		const brokenWorkflow = MultiStepWorkflow.machine(
			'fail',
			new SyncJob('broken', () => {
				throw new Error('permanent');
			}).withRetry(2, 1),
		);

		await expect(new MultiStepEngine().run(brokenWorkflow)).rejects.toBeInstanceOf(WorkflowError);
	});

	it('removes retry backoff abort listeners after successful sleeps', async () => {
		const controller = new AbortController();
		const signal = controller.signal;
		const originalAddEventListener = signal.addEventListener.bind(signal);
		const originalRemoveEventListener = signal.removeEventListener.bind(signal);

		let activeAbortListeners = 0;

		signal.addEventListener = ((type: string, listener: EventListenerOrEventListenerObject, options?: boolean | AddEventListenerOptions): void => {
			if (type === 'abort') {
				activeAbortListeners += 1;
			}

			originalAddEventListener(type, listener, options);
		}) as AbortSignal['addEventListener'];
		signal.removeEventListener = ((type: string, listener: EventListenerOrEventListenerObject, options?: boolean | EventListenerOptions): void => {
			if (type === 'abort') {
				activeAbortListeners -= 1;
			}

			originalRemoveEventListener(type, listener, options);
		}) as AbortSignal['removeEventListener'];

		let attempts = 0;

		const workflow = MultiStepWorkflow.machine(
			'retry-cleanup',
			new SyncJob('flaky', () => {
				attempts += 1;

				if (attempts < 3) {
					throw new Error('transient');
				}

				return 'ok';
			}).withRetry(3, 1),
		);

		await expect(new MultiStepEngine().run(workflow, {}, signal)).resolves.toMatchObject({ responses: { flaky: 'ok' } });

		expect(activeAbortListeners).toBe(0);
	});

	it('skips run-if jobs and lets downstream dependencies continue', async () => {
		let notifyRan = false;

		const workflow = MultiStepWorkflow.machine(
			'conditional',
			new SyncJob('create', () => ({ id: 'u1' })),
			new SyncJob('notify', () => {
				notifyRan = true;

				return 'notified';
			})
				.withArgs({ id: new ResponseArg('create', 'id') })
				.withRunIf(({ vars }) => vars.env === 'prod'),
			new SyncJob('after', () => 'done').dependsOn('notify'),
		);

		const result = await new MultiStepEngine().run(workflow, { env: 'staging' });

		expect(notifyRan).toBe(false);
		expect(result.skipped).toEqual(['notify']);
		expect(result.responses.after).toBe('done');
	});

	it('detects cycles and dangling response dependencies', () => {
		expect(() => MultiStepWorkflow.machine('cyclic', new SyncJob('a', () => undefined).dependsOn('b'), new SyncJob('b', () => undefined).dependsOn('a')).compile()).toThrow(CompileError);

		expect(() => MultiStepWorkflow.machine('dangling', new SyncJob('a', () => undefined).withArgs({ x: new ResponseArg('ghost', 'id') })).compile()).toThrow(CompileError);
	});

	it('cancels async siblings in fail-fast mode and lets them finish in lenient mode', async () => {
		let cancelled = false;

		const failFast = MultiStepWorkflow.machine(
			'failfast',
			new AsyncJob('fast', async () => {
				await delay(5);

				throw new Error('boom');
			}),
			new AsyncJob('slow', async ({ signal }) => {
				await waitForAbort(signal);

				cancelled = true;
			}),
		);

		await expect(new MultiStepEngine({ driver: new ParallelDriver() }).run(failFast)).rejects.toBeInstanceOf(WorkflowError);

		expect(cancelled).toBe(true);

		let completed = false;

		const lenient = MultiStepWorkflow.machine(
			'lenient',
			new AsyncJob('fast', async () => {
				throw new Error('boom');
			}),
			new AsyncJob('slow', async () => {
				await delay(5);

				completed = true;

				return 'ok';
			}),
		);

		await expect(new MultiStepEngine({ driver: new ParallelDriver(), continueOnError: true }).run(lenient)).rejects.toBeInstanceOf(WorkflowError);

		expect(completed).toBe(true);
	});

	it('supports serial drivers, literals, and typed result extraction', async () => {
		const order: string[] = [];

		const workflow = MultiStepWorkflow.machine(
			'ordered',
			new AsyncJob('a', () => {
				order.push('a');

				return { id: 7 };
			}).withArgs({ literal: new LiteralArg(42) }),
			new AsyncJob('b', () => {
				order.push('b');

				return 'ok';
			}),
		);

		const result = await new MultiStepEngine({ driver: new SerialDriver() }).run(workflow);

		expect(order).toEqual(['a', 'b']);
		expect(result.as<number>('a', 'id')).toBe(7);
		expect(new MultiStepResult({ job: { a: 1 } }).as<number>('job', 'a')).toBe(1);
		expect(() => new MultiStepResult({ job: { a: 1 } }).as<number>('job', 'missing')).toThrow(UnresolvedResponseError);
	});

	it('stops retrying once maxExceptions is reached', async () => {
		const maxExceptions = 3;

		let calls = 0;

		const workflow = MultiStepWorkflow.machine(
			'cap',
			new SyncJob('always-fails', () => {
				calls += 1;

				throw new Error('boom');
			}).withRetryPolicy(new RetryPolicy({ maxTries: 10, maxExceptions })),
		);

		await expect(new MultiStepEngine().run(workflow)).rejects.toMatchObject({ attempts: maxExceptions });

		// The cap binds before maxTries: the handler runs exactly maxExceptions times.
		expect(calls).toBe(maxExceptions);
	});

	it('lets maxTries bind independently when it is the tighter limit', async () => {
		let calls = 0;

		const workflow = MultiStepWorkflow.machine(
			'tries',
			new SyncJob('always-fails', () => {
				calls += 1;

				throw new Error('boom');
			}).withRetryPolicy(new RetryPolicy({ maxTries: 2, maxExceptions: 5 })),
		);

		await expect(new MultiStepEngine().run(workflow)).rejects.toMatchObject({ attempts: 2 });

		expect(calls).toBe(2);
	});

	it('lets an abort signal win over the remaining retry budget', async () => {
		const controller = new AbortController();

		let calls = 0;

		const workflow = MultiStepWorkflow.machine(
			'abort',
			new SyncJob('flaky', () => {
				calls += 1;

				controller.abort(new Error('cancelled'));

				throw new Error('boom');
			}).withRetryPolicy(new RetryPolicy({ maxTries: 5, maxExceptions: 5 })),
		);

		await expect(new MultiStepEngine().run(workflow, {}, controller.signal)).rejects.toBeInstanceOf(WorkflowError);

		// Abort short-circuits the loop despite budget for four more attempts.
		expect(calls).toBe(1);
	});
});

const delay = async (ms: number): Promise<void> => {
	await new Promise((resolve) => setTimeout(resolve, ms));
};

const waitForAbort = async (signal: AbortSignal): Promise<void> => {
	await new Promise<void>((resolve) => {
		if (signal.aborted) {
			resolve();

			return;
		}

		signal.addEventListener('abort', () => resolve(), { once: true });
	});
};
