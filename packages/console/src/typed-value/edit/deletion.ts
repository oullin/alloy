import { Key } from '#console/key';
import { deleteNextCharacter, deletePreviousCharacter, deletePreviousWord, deleteToLineStart } from '#console/typed-value/delete';

export const deleteTypedValueKey = (value: string[], cursor: number, key: string, allowNewLine: boolean): number | undefined => {
	if (key === Key.delete) {
		return deleteNextCharacter(value, cursor);
	}

	if (key === Key.ctrlU) {
		return deleteToLineStart(value, cursor, allowNewLine);
	}

	if (key === Key.backspace || key === Key.ctrlH) {
		return deletePreviousCharacter(value, cursor);
	}

	if (key === Key.optionBackspace) {
		return deletePreviousWord(value, cursor);
	}

	return undefined;
};
