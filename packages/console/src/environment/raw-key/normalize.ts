import { Key } from '#console/key';
import { parseKnownKeyValues } from '#console/key/validators/value';

const knownRawKeys = parseKnownKeyValues();

const rawKeyAliases = new Map<string, string>([
	['\r', Key.enter],
	['\u0008', Key.backspace],
]);

export const normalizeRawKey = (value: string): string => {
	return rawKeyAliases.get(value) ?? value;
};

export const isCompleteRawKey = (value: string): boolean => {
	return rawKeyAliases.has(value) || !isPartialRawKey(value);
};

export const isPartialRawKey = (value: string): boolean => {
	return knownRawKeys.some((key) => key !== value && key.startsWith(value));
};
