import { Marking } from '../marking.js';
import { Transition } from '../transition.js';
import { Definition } from './definition.js';
import type { MutableDefinition } from './types.js';

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

	public addPlace(place: string): this {
		if (place !== '' && !this.#definition.places.includes(place)) {
			this.#definition.places.push(place);
		}

		return this;
	}

	public setInitialPlaces(...places: string[]): this {
		this.#definition.initialMarking = Marking.fromPlaces(...places);

		return this;
	}

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
