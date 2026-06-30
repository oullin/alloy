import { fromCharacters } from '#console/typed-value/characters';
import type { AppliedTypedKey } from '#console/typed-value/types';

type TypedKeyResultOptions = {
	cancelled?: boolean;
	submitted?: boolean;
};

export const typedKeyResult = (value: string[], cursor: number, options: TypedKeyResultOptions = {}): AppliedTypedKey => ({
	cancelled: options.cancelled ?? false,
	cursor,
	submitted: options.submitted ?? false,
	value: fromCharacters(value),
});
