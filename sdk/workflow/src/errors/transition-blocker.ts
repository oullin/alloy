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
