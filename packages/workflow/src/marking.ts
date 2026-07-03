/**
 * The set of places a subject currently occupies, with a token count per
 * place (petri-net style). State machines hold exactly one active place.
 */
export class Marking {
	public readonly places: Record<string, number>;

	public constructor(places: Record<string, number> = {}) {
		this.places = { ...places };
	}

	public static fromPlaces(...places: string[]): Marking {
		const marking = new Marking();

		for (const place of places) {
			if (place !== '') {
				marking.add(place, 1);
			}
		}

		return marking;
	}

	public clone(): Marking {
		return new Marking(this.places);
	}

	public has(place: string): boolean {
		return (this.places[place] ?? 0) > 0;
	}

	public tokens(place: string): number {
		return this.places[place] ?? 0;
	}

	public add(place: string, count = 1): this {
		if (count <= 0) {
			return this;
		}

		this.places[place] = (this.places[place] ?? 0) + count;

		return this;
	}

	public remove(place: string, count = 1): this {
		if (count <= 0) {
			return this;
		}

		const next = (this.places[place] ?? 0) - count;

		if (next <= 0) {
			delete this.places[place];

			return this;
		}

		this.places[place] = next;

		return this;
	}

	public activePlaces(): string[] {
		return Object.entries(this.places)
			.filter(([, count]) => count > 0)
			.map(([place]) => place)
			.sort();
	}

	public toJSON(): Record<string, number> {
		return { ...this.places };
	}
}
