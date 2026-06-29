import { z } from 'zod';

const previousStringValue = (value: unknown): string => {
	if (typeof value === 'string') {
		return value;
	}

	if (typeof value === 'number' || typeof value === 'boolean' || typeof value === 'bigint' || typeof value === 'symbol') {
		return String(value);
	}

	if (Array.isArray(value)) {
		return value.map((item) => (item === null || item === undefined ? '' : previousStringValue(item))).join(',');
	}

	if (typeof value === 'object' && value !== null) {
		const customToString = Reflect.get(value, 'toString') as unknown;

		if (typeof customToString === 'function' && customToString !== Object.prototype.toString) {
			return customToString.call(value);
		}

		return Object.prototype.toString.call(value);
	}

	return '';
};

const presentPreviousValueSchema = z.unknown().refine((value) => value !== undefined && value !== null);
const previousStringSchema = presentPreviousValueSchema.transform(previousStringValue);
const previousNumberSchema = z.union([z.number(), z.string()]);
const previousArraySchema = <T>(): z.ZodType<T[]> => z.array(z.unknown()) as z.ZodType<T[]>;
const previousBooleanSchema = z.boolean();
const previousValueSchema = <T>(): z.ZodType<T> => presentPreviousValueSchema as z.ZodType<T>;

export const parsePreviousString = (previous: unknown, defaultValue: string): string => {
	const parsed = previousStringSchema.safeParse(previous);

	return parsed.success ? parsed.data : defaultValue;
};

export const hasPreviousResponse = (previous: unknown): boolean => {
	return presentPreviousValueSchema.safeParse(previous).success;
};

export const parsePreviousNumber = (previous: unknown, defaultValue: number | string): number | string => {
	const parsed = previousNumberSchema.safeParse(previous);

	return parsed.success ? parsed.data : defaultValue;
};

export const parsePreviousArray = <T>(previous: unknown, defaultValue: T[]): T[] => {
	const parsed = previousArraySchema<T>().safeParse(previous);

	return parsed.success ? parsed.data : defaultValue;
};

export const parsePreviousBoolean = (previous: unknown, defaultValue: boolean): boolean => {
	const parsed = previousBooleanSchema.safeParse(previous);

	return parsed.success ? parsed.data : defaultValue;
};

export const parsePreviousValue = <T>(previous: unknown, defaultValue: T): T => {
	const parsed = previousValueSchema<T>().safeParse(previous);

	return parsed.success ? parsed.data : defaultValue;
};
