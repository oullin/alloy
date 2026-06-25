export class TransitionNotFoundError extends Error {
	public readonly transition: string;

	public constructor(transition: string) {
		super(`transition not found: ${transition}`);
		this.name = 'TransitionNotFoundError';
		this.transition = transition;
	}
}

export interface TransitionBlockerShape {
	message: string;
	code?: string;
}

export class TransitionBlocker {
	public readonly message: string;
	public readonly code: string;

	public constructor(message: string, code = '') {
		this.message = message;
		this.code = code;
	}

	public static from(input: TransitionBlocker | TransitionBlockerShape): TransitionBlocker {
		if (input instanceof TransitionBlocker) {
			return new TransitionBlocker(input.message, input.code);
		}

		return new TransitionBlocker(input.message, input.code ?? '');
	}
}

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

export class TransitionError extends Error {
	public readonly machine: string;
	public readonly transition: string;
	public readonly blockers: TransitionBlocker[];

	public constructor(machine: string, transition: string, blockers: TransitionBlocker[] = []) {
		const suffix = blockers.length > 0 ? `: ${blockers.map((blocker) => blocker.message).join('; ')}` : '';

		super(`cannot apply transition "${transition}" on workflow "${machine}"${suffix}`);
		this.name = 'TransitionError';
		this.machine = machine;
		this.transition = transition;
		this.blockers = blockers.map((blocker) => TransitionBlocker.from(blocker));
	}
}
