import type { MetadataStore } from '#workflow/types';
import type { Definition } from '#workflow/definition';

export class DefinitionMetadataStore implements MetadataStore {
	readonly #definition: Definition;

	public constructor(definition: Definition) {
		this.#definition = definition;
	}

	public workflowMetadata(key: string): unknown {
		return this.#definition.metadataValue(key);
	}

	public hasWorkflowMetadata(key: string): boolean {
		return this.#definition.hasMetadataValue(key);
	}

	public placeMetadata(place: string, key: string): unknown {
		return this.#definition.placeMetadataValue(place, key);
	}

	public hasPlaceMetadata(place: string, key: string): boolean {
		return this.#definition.hasPlaceMetadataValue(place, key);
	}

	public transitionMetadata(transition: string, key: string): unknown {
		return this.#definition.transitionMetadataValue(transition, key);
	}

	public hasTransitionMetadata(transition: string, key: string): boolean {
		return this.#definition.hasTransitionMetadataValue(transition, key);
	}
}
