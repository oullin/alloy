import { WorkflowError } from './errors.js';
import type { CompiledGraph } from './graph.js';
import type { MultiStepResult } from './result.js';
import { runWithRetry } from './retry.js';
import { RunState } from './run-state.js';
import type { Driver, JobInput, Task } from './types.js';
import type { MultiStepWorkflow } from './workflow.js';

export class MultiStepEngine {
	readonly #driver?: Driver;
	readonly #continueOnError: boolean;

	public constructor(options: { driver?: Driver; continueOnError?: boolean } = {}) {
		this.#driver = options.driver;
		this.#continueOnError = options.continueOnError ?? false;
	}

	public async run(workflow: MultiStepWorkflow, vars: Record<string, unknown> = {}, signal?: AbortSignal): Promise<MultiStepResult> {
		const graph = workflow.compiledGraph();
		const state = new RunState(graph, vars);
		const rootController = new AbortController();

		const abortFromParent = (): void => rootController.abort(signal?.reason);

		if (signal?.aborted === true) {
			abortFromParent();
		} else {
			signal?.addEventListener('abort', abortFromParent, { once: true });
		}

		try {
			while (!state.done()) {
				const wave = state.nextWave();

				if (wave.length === 0) {
					break;
				}

				await this.runWave(rootController.signal, graph, state, wave);
			}

			return state.result();
		} finally {
			signal?.removeEventListener('abort', abortFromParent);
		}
	}

	private async runWave(signal: AbortSignal, graph: CompiledGraph, state: RunState, wave: string[]): Promise<void> {
		const syncJobs = wave.filter((name) => !graph.byName(name).isAsync());
		const asyncJobs = wave.filter((name) => graph.byName(name).isAsync());

		for (const name of syncJobs) {
			await this.runJob(signal, graph, state, name);
		}

		if (asyncJobs.length === 0) {
			return;
		}

		if (this.#driver === undefined) {
			for (const name of asyncJobs) {
				await this.runJob(signal, graph, state, name);
			}

			return;
		}

		if (this.#continueOnError) {
			await this.runAsyncLenient(signal, graph, state, asyncJobs);

			return;
		}

		await this.runAsyncFailFast(signal, graph, state, asyncJobs);
	}

	private async runAsyncFailFast(signal: AbortSignal, graph: CompiledGraph, state: RunState, names: string[]): Promise<void> {
		const controller = new AbortController();
		const abortFromParent = (): void => controller.abort(signal.reason);

		if (signal.aborted) {
			abortFromParent();
		} else {
			signal.addEventListener('abort', abortFromParent, { once: true });
		}

		const results = Array.from<unknown>({ length: names.length });
		const errors = Array.from<unknown>({ length: names.length });

		const tasks = names.map(
			(name, index): Task =>
				async () => {
					try {
						const value = await this.invokeJob(controller.signal, graph, state, name);

						results[index] = value;

						return value;
					} catch (error) {
						errors[index] = error;
						controller.abort(error);

						throw error;
					}
				},
		);

		try {
			await this.#driver?.run(controller.signal, tasks);
		} catch (driverError) {
			for (const [index, name] of names.entries()) {
				state.recordResult(name, results[index], errors[index]);
			}

			for (const [index, error] of errors.entries()) {
				if (error !== undefined) {
					throw this.wrapWorkflowError(names[index] ?? '', error, 1);
				}
			}

			throw driverError;
		} finally {
			signal.removeEventListener('abort', abortFromParent);
		}

		for (const [index, name] of names.entries()) {
			state.recordResult(name, results[index], errors[index]);
		}
	}

	private async runAsyncLenient(signal: AbortSignal, graph: CompiledGraph, state: RunState, names: string[]): Promise<void> {
		const results = Array.from<unknown>({ length: names.length });
		const errors = Array.from<unknown>({ length: names.length });

		const tasks = names.map(
			(name, index): Task =>
				async () => {
					try {
						const value = await this.invokeJob(signal, graph, state, name);

						results[index] = value;

						return value;
					} catch (error) {
						errors[index] = error;

						return undefined;
					}
				},
		);

		await Promise.all(tasks.map((task) => task()));

		for (const [index, name] of names.entries()) {
			state.recordResult(name, results[index], errors[index]);
		}

		for (const [index, error] of errors.entries()) {
			if (error !== undefined) {
				throw this.wrapWorkflowError(names[index] ?? '', error, 1);
			}
		}
	}

	private async runJob(signal: AbortSignal, graph: CompiledGraph, state: RunState, name: string): Promise<void> {
		try {
			state.recordResult(name, await this.invokeJob(signal, graph, state, name), undefined);
		} catch (error) {
			state.recordResult(name, undefined, error);
			throw error;
		}
	}

	private async invokeJob(signal: AbortSignal, graph: CompiledGraph, state: RunState, name: string): Promise<unknown> {
		const spec = graph.byName(name);
		const responses = state.responsesSnapshot();
		const resolved: Record<string, unknown> = {};

		for (const [key, arg] of Object.entries(spec.args())) {
			try {
				resolved[key] = arg.resolve(state.vars(), responses);
			} catch (error) {
				throw this.wrapWorkflowError(name, error, 0);
			}
		}

		const input: JobInput = {
			signal,
			args: spec.args(),
			resolved,
			vars: state.vars(),
			responses: state.responsesSnapshot(),
		};

		if (spec.runIfPredicate()?.(input) === false) {
			state.markSkipped(name);

			return undefined;
		}

		const { value, attempts, error } = await runWithRetry(signal, spec.retryPolicy(), async (attemptSignal) => {
			input.signal = attemptSignal;

			return spec.handler(input);
		});

		if (error !== undefined) {
			throw this.wrapWorkflowError(name, error, attempts);
		}

		return value;
	}

	private wrapWorkflowError(job: string, error: unknown, attempts: number): WorkflowError {
		if (error instanceof WorkflowError) {
			return error;
		}

		return new WorkflowError(job, attempts, error);
	}
}
