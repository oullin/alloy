import { TransitionBlocker, type TransitionBlockerShape } from '#workflow/errors/transition-blocker';

export class TransitionBlockerList {
	readonly #blockers: TransitionBlocker[] = [];

	public add(blocker: TransitionBlocker | TransitionBlockerShape): this {
		this.#blockers.push(TransitionBlocker.from(blocker));

		return this;
	}

	public all(): TransitionBlocker[] {
		return this.#blockers.map((blocker) => TransitionBlocker.from(blocker));
	}

	public empty(): boolean {
		return this.#blockers.length === 0;
	}
}
