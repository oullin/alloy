import type { CompiledGraph } from './graph.js';
import { MultiStepResult } from './result.js';

export class RunState {
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
