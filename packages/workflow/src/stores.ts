import type { Definition } from '#workflow/definition';
import { Marking } from '#workflow/marking';
import type { WorkflowContext } from '#workflow/types';

export interface MarkingStore<T> {
	getMarking(subject: T, definition: Definition): Marking;
	setMarking(subject: T, marking: Marking, definition: Definition, context: WorkflowContext): void;
}

export class SingleStateStore<T> implements MarkingStore<T> {
	readonly #getter: (subject: T) => string;
	readonly #setter: (subject: T, place: string) => void;

	public constructor(getter: (subject: T) => string, setter: (subject: T, place: string) => void) {
		this.#getter = getter;
		this.#setter = setter;
	}

	public getMarking(subject: T): Marking {
		const place = this.#getter(subject);

		if (place === '') {
			return new Marking();
		}

		return Marking.fromPlaces(place);
	}

	public setMarking(subject: T, marking: Marking): void {
		const places = marking.activePlaces();

		if (places.length > 1) {
			throw new Error(`single state store cannot persist ${places.length} active places`);
		}

		this.#setter(subject, places[0] ?? '');
	}
}

export class MultiStateStore<T> implements MarkingStore<T> {
	readonly #getter: (subject: T) => string[];
	readonly #setter: (subject: T, places: string[]) => void;

	public constructor(getter: (subject: T) => string[], setter: (subject: T, places: string[]) => void) {
		this.#getter = getter;
		this.#setter = setter;
	}

	public getMarking(subject: T): Marking {
		return Marking.fromPlaces(...this.#getter(subject));
	}

	public setMarking(subject: T, marking: Marking): void {
		this.#setter(subject, marking.activePlaces());
	}
}
