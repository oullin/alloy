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

export class MultiStepResult {
	public readonly responses: Record<string, unknown>;
	public readonly skipped: string[];

	public constructor(responses: Record<string, unknown> = {}, skipped: string[] = []) {
		this.responses = { ...responses };
		this.skipped = [...skipped];
	}

	public as<T>(job: string, field = ''): T {
		const raw = this.responses[job];

		if (!Object.hasOwn(this.responses, job)) {
			throw new Error(`response "${job}" not available`);
		}

		const value = field === '' ? raw : resolveField(raw, field);

		return value as T;
	}
}

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

export abstract class Arg {
	public abstract kind(): string;

	public dependsOnJob(): string {
		return '';
	}

	public abstract resolve(vars: Record<string, unknown>, responses: Record<string, unknown>): unknown;
}

export class LiteralArg extends Arg {
	readonly #value: unknown;

	public constructor(value: unknown) {
		super();
		this.#value = value;
	}

	public kind(): string {
		return 'literal';
	}

	public resolve(): unknown {
		return this.#value;
	}
}

export class VariableArg extends Arg {
	public readonly name: string;

	public constructor(name: string) {
		super();
		this.name = name;
	}

	public kind(): string {
		return 'variable';
	}

	public resolve(vars: Record<string, unknown>): unknown {
		if (!Object.hasOwn(vars, this.name)) {
			throw new Error(`variable "${this.name}" not provided`);
		}

		return vars[this.name];
	}
}

export class ResponseArg extends Arg {
	public readonly job: string;
	public readonly field: string;

	public constructor(job: string, field = '') {
		super();
		this.job = job;
		this.field = field;
	}

	public kind(): string {
		return 'response';
	}

	public dependsOnJob(): string {
		return this.job;
	}

	public resolve(_vars: Record<string, unknown>, responses: Record<string, unknown>): unknown {
		if (!Object.hasOwn(responses, this.job)) {
			throw new Error(`response "${this.job}" not available`);
		}

		const raw = responses[this.job];

		if (this.field === '') {
			return raw;
		}

		return resolveField(raw, this.field, this.job);
	}
}

export class RetryPolicy {
	public readonly maxTries: number;
	public readonly backoff: number[];
	public readonly timeout: number;
	public readonly maxExceptions: number;

	public constructor(input: { maxTries?: number; backoff?: number[]; timeout?: number; maxExceptions?: number } = {}) {
		this.maxTries = input.maxTries !== undefined && input.maxTries > 0 ? input.maxTries : 1;
		this.backoff = [...(input.backoff ?? [])];
		this.timeout = input.timeout ?? 0;
		this.maxExceptions = input.maxExceptions ?? 0;
	}

	public backoffFor(attempt: number): number {
		if (this.backoff.length === 0) {
			return 0;
		}

		return this.backoff[Math.min(attempt, this.backoff.length - 1)] ?? 0;
	}
}

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

export class MultiStepWorkflow {
	readonly #name: string;
	readonly #jobs: JobSpec[];
	#compiled?: CompiledGraph;

	public constructor(name: string, jobs: JobSpec[] = []) {
		this.#name = name;
		this.#jobs = [...jobs];
	}

	public static machine(name: string, ...jobs: JobSpec[]): MultiStepWorkflow {
		return new MultiStepWorkflow(name, jobs);
	}

	public name(): string {
		return this.#name;
	}

	public jobs(): JobSpec[] {
		return [...this.#jobs];
	}

	public compile(): void {
		this.#compiled = new CompiledGraph(this.#name, this.#jobs);
	}

	public compiledGraph(): CompiledGraph {
		if (this.#compiled === undefined) {
			this.compile();
		}

		if (this.#compiled === undefined) {
			throw new CompileError('workflow did not compile');
		}

		return this.#compiled;
	}
}

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

export class CompiledGraph {
	public readonly jobs: JobSpec[];
	readonly #byName = new Map<string, JobSpec>();
	readonly #parents = new Map<string, string[]>();
	readonly #children = new Map<string, string[]>();

	public constructor(workflowName: string, jobs: JobSpec[]) {
		if (jobs.length === 0) {
			throw new CompileError(`workflow ${workflowName} declares no jobs`);
		}

		this.jobs = [...jobs];
		this.indexJobs();
		this.indexEdges();
		this.detectCycle();
	}

	public byName(name: string): JobSpec {
		const spec = this.#byName.get(name);

		if (spec === undefined) {
			throw new CompileError(`unknown job ${name}`);
		}

		return spec;
	}

	public parents(name: string): string[] {
		return [...(this.#parents.get(name) ?? [])];
	}

	private indexJobs(): void {
		for (const [index, spec] of this.jobs.entries()) {
			if (spec.name === '') {
				throw new CompileError(`job at index ${index} has empty name`);
			}

			if (this.#byName.has(spec.name)) {
				throw new CompileError(`duplicate job name: ${spec.name}`);
			}

			this.#byName.set(spec.name, spec);
		}
	}

	private indexEdges(): void {
		for (const spec of this.jobs) {
			const deps = new Set<string>();

			for (const arg of Object.values(spec.args())) {
				const dependency = arg.dependsOnJob();

				if (dependency !== '') {
					deps.add(dependency);
				}
			}

			for (const dependency of spec.dependencies()) {
				deps.add(dependency);
			}

			for (const dependency of deps) {
				if (!this.#byName.has(dependency)) {
					throw new CompileError(`job ${spec.name} depends on unknown job ${dependency}`);
				}

				this.#parents.set(spec.name, [...(this.#parents.get(spec.name) ?? []), dependency]);
				this.#children.set(dependency, [...(this.#children.get(dependency) ?? []), spec.name]);
			}
		}
	}

	private detectCycle(): void {
		const states = new Map<string, 'done' | 'visiting'>();

		const visit = (name: string): void => {
			const state = states.get(name);

			if (state === 'visiting') {
				throw new CompileError(`cycle detected involving job ${name}`);
			}

			if (state === 'done') {
				return;
			}

			states.set(name, 'visiting');

			for (const child of this.#children.get(name) ?? []) {
				visit(child);
			}

			states.set(name, 'done');
		};

		for (const spec of this.jobs) {
			visit(spec.name);
		}
	}
}

class RunState {
	readonly #graph: CompiledGraph;
	readonly #vars: Record<string, unknown>;
	readonly #responses: Record<string, unknown> = {};
	readonly #completed = new Set<string>();
	readonly #skipped = new Set<string>();

	public constructor(graph: CompiledGraph, vars: Record<string, unknown>) {
		this.#graph = graph;
		this.#vars = { ...vars };
	}

	public vars(): Record<string, unknown> {
		return { ...this.#vars };
	}

	public responsesSnapshot(): Record<string, unknown> {
		return { ...this.#responses };
	}

	public recordResult(name: string, value: unknown, error: unknown): void {
		if (error === undefined && !this.#skipped.has(name)) {
			this.#responses[name] = value;
		}

		this.#completed.add(name);
	}

	public markSkipped(name: string): void {
		this.#skipped.add(name);
		this.#responses[name] = undefined;
	}

	public nextWave(): string[] {
		return this.#graph.jobs
			.filter((spec) => !this.#completed.has(spec.name))
			.filter((spec) => this.#graph.parents(spec.name).every((parent) => this.#completed.has(parent)))
			.map((spec) => spec.name);
	}

	public done(): boolean {
		return this.#completed.size >= this.#graph.jobs.length;
	}

	public result(): MultiStepResult {
		return new MultiStepResult(
			this.#responses,
			this.#graph.jobs.filter((spec) => this.#skipped.has(spec.name)).map((spec) => spec.name),
		);
	}
}

const runWithRetry = async (
	signal: AbortSignal,
	policy: RetryPolicy | undefined,
	handler: (signal: AbortSignal) => unknown,
): Promise<{ value: unknown; attempts: number; error?: unknown }> => {
	const activePolicy = policy ?? new RetryPolicy();

	let lastError: unknown;

	for (let attempt = 0; attempt < activePolicy.maxTries; attempt++) {
		const attemptSignal = createAttemptSignal(signal, activePolicy.timeout);

		try {
			const value = await handler(attemptSignal.signal);

			attemptSignal.cleanup();

			return { value, attempts: attempt + 1 };
		} catch (error) {
			attemptSignal.cleanup();
			lastError = signal.aborted ? signal.reason : error;

			if (attempt + 1 >= activePolicy.maxTries || signal.aborted) {
				break;
			}

			const backoff = activePolicy.backoffFor(attempt);

			if (backoff > 0) {
				await sleep(backoff, signal);
			}
		}
	}

	return { value: undefined, attempts: activePolicy.maxTries, error: lastError };
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

const resolveField = (raw: unknown, field: string, job = ''): unknown => {
	if (raw === null || raw === undefined) {
		throw new UnresolvedResponseError(field, job);
	}

	if (typeof raw === 'object' || typeof raw === 'function') {
		const record = raw as Record<string, unknown>;

		if (Object.hasOwn(record, field)) {
			return record[field];
		}
	}

	throw new UnresolvedResponseError(field, job);
};

const formatCause = (cause: unknown): string => {
	if (cause instanceof Error) {
		return cause.message;
	}

	return String(cause);
};
