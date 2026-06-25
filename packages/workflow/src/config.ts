import { DefinitionBuilder, type Definition } from './definition.js';

export class WorkflowConfigLoader {
	readonly #source: Record<string, unknown>;

	public constructor(source: Record<string, unknown>) {
		this.#source = source;
	}

	public load(): Definition {
		return this.loadAt('workflow');
	}

	public loadAt(rootKey: string): Definition {
		if (rootKey === '') {
			throw new Error('config: root key is required');
		}

		const root = this.rootAt(rootKey);

		if (!isRecord(root)) {
			throw new Error(`config: no workflow definition at "${rootKey}"`);
		}

		const builder = new DefinitionBuilder();
		const places = coerceStringSlice(root.places);

		if (places.length === 0) {
			throw new Error(`config: workflow "${rootKey}" requires at least one place`);
		}

		for (const place of places) {
			builder.addPlace(place);
		}

		const initial = coerceStringSlice(root.initial);

		if (initial.length === 0) {
			throw new Error(`config: workflow "${rootKey}" requires an initial place list`);
		}

		builder.setInitialPlaces(...initial);

		for (const transition of coerceTransitions(root.transitions)) {
			builder.addTransition(transition.name, transition.from, transition.to);
		}

		for (const [key, value] of Object.entries(coerceStringMap(root.metadata))) {
			builder.setMetadata(key, value);
		}

		for (const [place, raw] of Object.entries(coerceStringMap(root.places_metadata))) {
			for (const [key, value] of Object.entries(coerceStringMap(raw))) {
				builder.setPlaceMetadata(place, key, value);
			}
		}

		for (const [transition, raw] of Object.entries(coerceStringMap(root.transitions_metadata))) {
			for (const [key, value] of Object.entries(coerceStringMap(raw))) {
				builder.setTransitionMetadata(transition, key, value);
			}
		}

		return builder.build();
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

	private rootAt(rootKey: string): unknown {
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
}

interface TransitionDeclaration {
	name: string;
	from: string[];
	to: string[];
}

const isRecord = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null && !Array.isArray(value);

const coerceTransitions = (raw: unknown): TransitionDeclaration[] => {
	if (raw === undefined || raw === null) {
		return [];
	}

	if (!Array.isArray(raw)) {
		throw new Error(`config: transitions must be a list, got ${typeof raw}`);
	}

	return raw.map((item, index) => {
		const entry = coerceStringMap(item);

		if (Object.keys(entry).length === 0 && !isRecord(item)) {
			throw new Error(`config: transition[${index}] must be an object`);
		}

		return {
			name: typeof entry.name === 'string' ? entry.name : '',
			from: coerceStringSlice(entry.from),
			to: coerceStringSlice(entry.to),
		};
	});
};

const coerceStringSlice = (raw: unknown): string[] => {
	if (raw === undefined || raw === null) {
		return [];
	}

	if (typeof raw === 'string') {
		return [raw];
	}

	if (!Array.isArray(raw)) {
		return [];
	}

	return raw.filter((value): value is string => typeof value === 'string');
};

const coerceStringMap = (raw: unknown): Record<string, unknown> => {
	if (!isRecord(raw)) {
		return {};
	}

	return raw;
};
