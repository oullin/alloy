import { CompileError } from './errors.js';
import type { JobSpec } from './jobs.js';

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
