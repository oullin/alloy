import type { Validator } from '#console/contracts/base';

export type BasePromptOptions<T> = {
	message: string;
	default?: T;
	required?: boolean | string;
	hint?: string;
	validate?: Validator<T>;
};
