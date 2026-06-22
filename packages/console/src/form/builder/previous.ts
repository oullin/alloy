import { parsePreviousArray, parsePreviousBoolean, parsePreviousNumber, parsePreviousString, parsePreviousValue } from '#console/form/builder/validators/previous';

export const previousString = (previous: unknown, defaultValue: string): string => {
	return parsePreviousString(previous, defaultValue);
};

export const previousNumber = (previous: unknown, defaultValue: number | string): number | string => {
	return parsePreviousNumber(previous, defaultValue);
};

export const previousArray = <T>(previous: unknown, defaultValue: T[]): T[] => {
	return parsePreviousArray(previous, defaultValue);
};

export const previousBoolean = (previous: unknown, defaultValue: boolean): boolean => {
	return parsePreviousBoolean(previous, defaultValue);
};

export const previousValue = <T>(previous: unknown, defaultValue: T): T => {
	return parsePreviousValue(previous, defaultValue);
};
