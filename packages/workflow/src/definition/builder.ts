import { Marking } from '#workflow/marking';
import { Transition } from '#workflow/transition';
import { Definition } from '#workflow/definition/definition';
import type { MutableDefinition } from '#workflow/definition/types';

/**
 * Fluent builder for {@link Definition} instances: declare places, initial
 * places, transitions, and metadata, then call {@link build} to validate and
 * freeze the result.
 */
export class DefinitionBuilder {
	readonly #definition: MutableDefinition;

	public constructor() {
		this.#definition = {
			places: [],
			transitions: [],
			initialMarking: new Marking(),
			metadata: {},
			placeMetadata: {},
			transitionMetadata: {},
		};
	}

	/** Registers a place (state); empty names and duplicates are ignored. */
	public addPlace(place: string): this {
		if (place !== '' && !this.#definition.places.includes(place)) {
			this.#definition.places.push(place);
		}

		return this;
	}

	/** Sets the places a fresh subject starts in. */
	public setInitialPlaces(...places: string[]): this {
		this.#definition.initialMarking = Marking.fromPlaces(...places);

		return this;
	}

	/** Adds a named transition from one set of places to another. */
	public addTransition(name: string, from: string[], to: string[]): this {
		this.#definition.transitions.push(new Transition(name, from, to));

		return this;
	}

	public setMetadata(key: string, value: unknown): this {
		this.#definition.metadata[key] = value;

		return this;
	}

	public setPlaceMetadata(place: string, key: string, value: unknown): this {
		this.#definition.placeMetadata[place] ??= {};
		this.#definition.placeMetadata[place][key] = value;

		return this;
	}

	public setTransitionMetadata(transition: string, key: string, value: unknown): this {
		this.#definition.transitionMetadata[transition] ??= {};
		this.#definition.transitionMetadata[transition][key] = value;

		return this;
	}

	/**
	 * Builds and validates the definition, merging transition metadata.
	 *
	 * @throws Error when the definition is invalid (unknown places, missing initial marking, unreachable places).
	 */
	public build(): Definition {
		const transitions = this.#definition.transitions.map((transition) => {
			const metadata = this.#definition.transitionMetadata[transition.name];

			if (metadata === undefined) {
				return transition.clone();
			}

			return new Transition(transition.name, transition.from, transition.to, { ...transition.metadata, ...metadata });
		});

		const definition = new Definition({
			...this.#definition,
			transitions,
		});

		definition.validate();

		return definition;
	}
}
