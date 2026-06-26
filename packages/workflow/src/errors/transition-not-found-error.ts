export class TransitionNotFoundError extends Error {
	public readonly transition: string;

	public constructor(transition: string) {
		super(`transition not found: ${transition}`);
		this.name = 'TransitionNotFoundError';
		this.transition = transition;
	}
}
