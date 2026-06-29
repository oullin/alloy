import { Key } from '#console/key';
import { characterLength, characters } from '#console/typed-value/characters';
import { deleteTypedValueKey } from '#console/typed-value/edit/deletion';
import { insertPrintableKey } from '#console/typed-value/edit/insertion';
import { moveTypedValueCursor } from '#console/typed-value/edit/navigation';
import { typedKeyResult } from '#console/typed-value/edit/result';
import { parseTypedValueCursor } from '#console/typed-value/validators/cursor';
import type { AppliedTypedKey, TypedValueState } from '#console/typed-value/types';

export const applyTypedKey = (state: TypedValueState, key: string, allowNewLine = false, wrapWidth?: number): AppliedTypedKey => {
	const value = characters(state.value);

	let cursor = parseTypedValueCursor(state.cursor, value.length);

	if (key === Key.ctrlC) {
		return typedKeyResult(value, cursor, { cancelled: true });
	}

	if (key === Key.ctrlD && allowNewLine) {
		return typedKeyResult(value, cursor, { submitted: true });
	}

	if (key === Key.enter) {
		if (!allowNewLine) {
			return typedKeyResult(value, cursor, { submitted: true });
		}

		value.splice(cursor, 0, '\n');
		cursor += 1;

		return typedKeyResult(value, cursor);
	}

	const moved = moveTypedValueCursor(value, cursor, key, allowNewLine, wrapWidth);

	if (moved !== undefined) {
		return typedKeyResult(value, moved);
	}

	const deleted = deleteTypedValueKey(value, cursor, key, allowNewLine);

	if (deleted !== undefined) {
		return typedKeyResult(value, deleted);
	}

	const inserted = insertPrintableKey(value, cursor, key);

	if (inserted !== undefined) {
		cursor = inserted;
	}

	return typedKeyResult(value, cursor);
};

export const initialTypedValueState = (value = ''): TypedValueState => {
	return {
		cursor: characterLength(value),
		value,
	};
};
