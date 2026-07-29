import { fromCharacters, stringIndexToCharacterIndex } from '#console/typed-value/characters';
import { moveToLineBoundary } from '#console/typed-value/lines';

export const deleteNextCharacter = (value: string[], cursor: number): number => {
	value.splice(cursor, 1);

	return cursor;
};

export const deleteToLineStart = (value: string[], cursor: number, allowNewLine: boolean): number => {
	const start = allowNewLine ? moveToLineBoundary(value, cursor, 'start') : 0;

	value.splice(start, cursor - start);

	return start;
};

export const deletePreviousCharacter = (value: string[], cursor: number): number => {
	if (cursor === 0) {
		return cursor;
	}

	value.splice(cursor - 1, 1);

	return cursor - 1;
};

export const deletePreviousWord = (value: string[], cursor: number): number => {
	const before = fromCharacters(value.slice(0, cursor));

	const match = before.match(/(?:[\p{L}\p{M}\p{N}]+|[^\p{L}\p{M}\p{N}\s]+)\s*$/u);
	const start = match?.index === undefined ? 0 : stringIndexToCharacterIndex(before, match.index);

	value.splice(start, cursor - start);

	return start;
};
