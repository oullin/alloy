import { CompileError } from '#workflow/multisteps/errors';
import { CompiledGraph } from '#workflow/multisteps/graph';
import type { JobSpec } from '#workflow/multisteps/jobs';

/**
 * A named collection of job specs that compiles into a dependency graph for
 * multi-step execution (see `@alloy/sdk/workflow/multisteps`).
 */
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
