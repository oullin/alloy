import { Marking } from './marking.js';
import { Transition } from './transition.js';
import type { Metadata, NestedMetadata } from './types.js';
import { cloneNestedRecord, cloneRecord } from './types.js';

export class Definition {
	public readonly places: string[];
	public readonly transitions: Transition[];
	public readonly initialMarking: Marking;
	public readonly metadata: Metadata;
	public readonly placeMetadata: NestedMetadata;
	public readonly transitionMetadata: NestedMetadata;
	#placesSet?: Set<string>;

	public constructor(
		input: {
			places?: string[];
			transitions?: Transition[];
			initialMarking?: Marking;
			metadata?: Metadata;
			placeMetadata?: NestedMetadata;
			transitionMetadata?: NestedMetadata;
		} = {},
	) {
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

	public validate(): void {
		if (this.places.length === 0) {
			throw new Error('definition requires at least one place');
		}

		for (const transition of this.transitions) {
			if (transition.name === '') {
				throw new Error('transition name cannot be empty');
			}

			if (transition.from.length === 0) {
				throw new Error(`transition "${transition.name}" requires at least one from place`);
			}

			if (transition.to.length === 0) {
				throw new Error(`transition "${transition.name}" requires at least one to place`);
			}

			for (const place of [...transition.from, ...transition.to]) {
				if (!this.hasPlace(place)) {
					throw new Error(`transition "${transition.name}" references unknown place "${place}"`);
				}
			}
		}

		if (this.initialMarking.activePlaces().length === 0) {
			throw new Error('definition requires an initial marking');
		}

		for (const place of this.initialMarking.activePlaces()) {
			if (!this.hasPlace(place)) {
				throw new Error(`initial marking references unknown place "${place}"`);
			}
		}

		this.validateReachability();
	}

	private validateReachability(): void {
		const reachable = new Set<string>();
		const queue = this.initialMarking.activePlaces();

		for (const place of queue) {
			reachable.add(place);
		}

		while (queue.length > 0) {
			const current = queue.shift();

			if (current === undefined) {
				break;
			}

			for (const transition of this.transitions) {
				if (!transition.from.includes(current)) {
					continue;
				}

				for (const place of transition.to) {
					if (!reachable.has(place)) {
						reachable.add(place);
						queue.push(place);
					}
				}
			}
		}

		for (const place of this.places) {
			if (!reachable.has(place)) {
				throw new Error(`place "${place}" is unreachable from the initial marking`);
			}
		}
	}
}

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

interface MutableDefinition {
	places: string[];
	transitions: Transition[];
	initialMarking: Marking;
	metadata: Metadata;
	placeMetadata: NestedMetadata;
	transitionMetadata: NestedMetadata;
}
