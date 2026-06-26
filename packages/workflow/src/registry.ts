import type { WorkflowEngine } from '#workflow/machine';

export type SupportStrategy<T> = (subject: T) => boolean;

export class WorkflowRegistryEntry<T> {
	public readonly name: string;
	public readonly machine: WorkflowEngine<T>;
	public readonly supports?: SupportStrategy<T>;

	public constructor(input: { name?: string; machine: WorkflowEngine<T>; supports?: SupportStrategy<T> }) {
		this.name = input.name ?? '';
		this.machine = input.machine;
		this.supports = input.supports;
	}
}

export class WorkflowRegistry<T> {
	readonly #entries: WorkflowRegistryEntry<T>[] = [];

	public add(entry: WorkflowRegistryEntry<T>): this {
		this.#entries.push(entry);

		return this;
	}

	public get(subject: T, name = ''): WorkflowEngine<T> {
		for (const entry of this.#entries) {
			if (name !== '' && entry.name !== name) {
				continue;
			}

			if (entry.supports !== undefined && !entry.supports(subject)) {
				continue;
			}

			return entry.machine;
		}

		if (name !== '') {
			throw new Error(`workflow "${name}" not found`);
		}

		throw new Error('no workflow matched subject');
	}
}
