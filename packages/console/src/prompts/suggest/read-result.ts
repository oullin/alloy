import { characterLength } from '#console/typed-value/characters';
import type { TypedValueState } from '#console/typed-value';

export type TextSuggestionReadResult = {
	cancelled: boolean;
	frame?: string;
	rendered: boolean;
	value: string;
};

export const initialSuggestionState = (value = ''): TypedValueState => ({
	cursor: characterLength(value),
	value,
});

export const suggestionReadResult = (value: string, rendered: boolean, cancelled = false, frame?: string): TextSuggestionReadResult => ({
	cancelled,
	frame,
	rendered,
	value,
});
