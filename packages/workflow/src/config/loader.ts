import { DefinitionBuilder, type Definition } from '#workflow/definition';
import { coerceStringMap, coerceStringSlice, coerceTransitions, isRecord } from '#workflow/config/coercion';
import { WorkflowConfigSource } from '#workflow/config/source';

export class WorkflowConfigLoader {
	readonly #source: WorkflowConfigSource;

	public constructor(source: Record<string, unknown>) {
		this.#source = new WorkflowConfigSource(source);
	}

	public load(): Definition {
		return this.loadAt('workflow');
	}

	public loadAt(rootKey: string): Definition {
		if (rootKey === '') {
			throw new Error('config: root key is required');
		}

		const root = this.#source.rootAt(rootKey);

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
}
