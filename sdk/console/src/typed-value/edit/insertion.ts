import { characters, isPrintable } from '#console/typed-value/characters';

export const insertPrintableKey = (value: string[], cursor: number, key: string): number | undefined => {
	if (!isPrintable(key)) {
		return undefined;
	}

	const inserted = characters(key);

	value.splice(cursor, 0, ...inserted);

	return cursor + inserted.length;
};
