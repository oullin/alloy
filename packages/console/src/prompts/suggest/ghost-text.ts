import { dim, inverse } from '#console/theme/styles';
import { characterLength, characters, fromCharacters } from '#console/typed-value/characters';
import { valueWithCursor } from '#console/typed-value/cursor';
import type { TypedValueState } from '#console/typed-value/types';

const startsWithInput = (match: string, value: string): boolean => {
	const normalizedValue = value.toLowerCase();

	return match.toLowerCase().startsWith(normalizedValue);
};

export const autocompleteGhostText = (value: string, match: string | undefined): string => {
	if (value.length === 0 || match === undefined || !startsWithInput(match, value) || characterLength(match) <= characterLength(value)) {
		return '';
	}

	return fromCharacters(characters(match).slice(characterLength(value)));
};

export const autocompleteDisplayValue = (state: TypedValueState, match: string | undefined, placeholder = ''): string => {
	if (state.value.length === 0) {
		return placeholder;
	}

	if (state.cursor < characterLength(state.value)) {
		return valueWithCursor(state.value, state.cursor);
	}

	const ghostText = autocompleteGhostText(state.value, match);

	if (ghostText.length === 0) {
		return valueWithCursor(state.value, state.cursor);
	}

	const [cursorCharacter = '', ...remainingGhost] = characters(ghostText);

	return `${state.value}${inverse(cursorCharacter)}${dim(fromCharacters(remainingGhost))}`;
};
