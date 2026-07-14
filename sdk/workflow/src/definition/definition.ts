import { Marking } from '#workflow/marking';
import { Transition } from '#workflow/transition';
import type { Metadata, NestedMetadata } from '#workflow/types';
import { cloneNestedRecord, cloneRecord } from '#workflow/types';
import type { DefinitionInput } from '#workflow/definition/types';
import { validateDefinition } from '#workflow/definition/validation';

/**
 * Immutable description of a workflow: its places, transitions, initial
 * marking, and metadata. Usually assembled with `DefinitionBuilder` and
 * consumed by `Machine`/`StateMachine`.
 */
export class Definition {
	public readonly places: string[];
	public readonly transitions: Transition[];
	public readonly initialMarking: Marking;
	public readonly metadata: Metadata;
	public readonly placeMetadata: NestedMetadata;
	public readonly transitionMetadata: NestedMetadata;
	#placesSet?: Set<string>;

	public constructor(input: DefinitionInput = {}) {
		this.places = [...(input.places ?? [])];
		this.transitions = (input.transitions ?? []).map((transition) => transition.clone());
		this.initialMarking = input.initialMarking?.clone() ?? new Marking();
		this.metadata = cloneRecord(input.metadata);
		this.placeMetadata = cloneNestedRecord(input.placeMetadata);
		this.transitionMetadata = cloneNestedRecord(input.transitionMetadata);
	}

	public clone(): Definition {
		return new Definition({
			places: this.places,
			transitions: this.transitions,
			initialMarking: this.initialMarking,
			metadata: this.metadata,
			placeMetadata: this.placeMetadata,
			transitionMetadata: this.transitionMetadata,
		});
	}

	/** Finds a transition by name, returning a metadata-enriched clone or undefined. */
	public transition(name: string): Transition | undefined {
		const transition = this.transitions.find((item) => item.name === name);

		if (transition === undefined) {
			return undefined;
		}

		if (Object.keys(transition.metadata).length === 0 && this.transitionMetadata[name] !== undefined) {
			return new Transition(transition.name, transition.from, transition.to, this.transitionMetadata[name]);
		}

		return transition.clone();
	}

	public hasPlace(place: string): boolean {
		this.#placesSet ??= new Set(this.places);

		return this.#placesSet.has(place);
	}

	public metadataValue(key: string): unknown {
		return this.metadata[key];
	}

	public hasMetadataValue(key: string): boolean {
		return Object.hasOwn(this.metadata, key);
	}

	public placeMetadataValue(place: string, key: string): unknown {
		return this.placeMetadata[place]?.[key];
	}

	public hasPlaceMetadataValue(place: string, key: string): boolean {
		return Object.hasOwn(this.placeMetadata[place] ?? {}, key);
	}

	public transitionMetadataValue(transition: string, key: string): unknown {
		return this.transitionMetadata[transition]?.[key];
	}

	public hasTransitionMetadataValue(transition: string, key: string): boolean {
		return Object.hasOwn(this.transitionMetadata[transition] ?? {}, key);
	}

	/**
	 * Checks structural integrity: known places, a non-empty initial marking,
	 * and no unreachable places.
	 *
	 * @throws Error when the definition is invalid.
	 */
	public validate(): void {
		validateDefinition(this);
	}
}
