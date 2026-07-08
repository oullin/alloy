import { isRecord } from '#workflow/config/coercion';

export class WorkflowConfigSource {
	readonly #source: Record<string, unknown>;

	public constructor(source: Record<string, unknown>) {
		this.#source = source;
	}

	public rootAt(rootKey: string): unknown {
		const nested = this.valueAt(rootKey);

		if (isRecord(nested)) {
			return nested;
		}

		const prefix = `${rootKey}.`;
		const dotted: Record<string, unknown> = {};

		for (const [key, value] of Object.entries(this.#source)) {
			if (key.startsWith(prefix)) {
				dotted[key.slice(prefix.length)] = value;
			}
		}

		if (Object.keys(dotted).length > 0) {
			return dotted;
		}

		return nested;
	}

	private valueAt(path: string): unknown {
		const direct = this.#source[path];

		if (direct !== undefined) {
			return direct;
		}

		let current: unknown = this.#source;

		for (const segment of path.split('.')) {
			if (!isRecord(current)) {
				return undefined;
			}

			current = current[segment];
		}

		return current;
	}
}
