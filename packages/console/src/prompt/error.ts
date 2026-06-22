export class PromptValidationError extends Error {
	readonly value: unknown;

	constructor(message: string, value?: unknown) {
		super(message);
		this.name = 'PromptValidationError';
		this.value = value;
	}
}
