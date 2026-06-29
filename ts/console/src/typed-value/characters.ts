export const characters = (value: string): string[] => Array.from(value);

export const fromCharacters = (value: string[]): string => value.join('');

export const characterLength = (value: string): number => characters(value).length;

export const stringIndexToCharacterIndex = (value: string, index: number): number => characterLength(value.slice(0, index));

export const isPrintable = (key: string): boolean => {
	if (key.length === 0) {
		return false;
	}

	return characters(key).every((character) => {
		const code = character.codePointAt(0) ?? 0;

		return code >= 32 && code !== 127;
	});
};
