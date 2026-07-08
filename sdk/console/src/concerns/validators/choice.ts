import { z } from 'zod';
import type { Choice, ChoiceOptions } from '#console/types';

const choiceSchema = z
	.object({
		disabled: z.union([z.boolean(), z.string()]).optional(),
		hint: z.string().optional(),
		label: z.string(),
		value: z.unknown(),
	})
	.passthrough();

const choiceRecordSchema = z.record(z.string(), z.string());
const choiceOptionsListSchema = z.array(z.unknown());
const typedChoiceSchema = <T>(): z.ZodType<Choice<T>> => choiceSchema as z.ZodType<Choice<T>>;
const typedChoiceOptionsListSchema = <T>(): z.ZodType<Array<Choice<T> | T>> => choiceOptionsListSchema as z.ZodType<Array<Choice<T> | T>>;
const choiceValueSchema = <T>(): z.ZodType<T> => z.unknown() as z.ZodType<T>;

export const choiceOptionsSchema = <T>(): z.ZodType<ChoiceOptions<T>> => z.union([choiceOptionsListSchema, choiceRecordSchema]) as z.ZodType<ChoiceOptions<T>>;

export type ChoiceRecordEntry = {
	label: string;
	value: string;
};

export type ParsedChoiceOptions<T> =
	| {
			kind: 'list';
			options: Array<Choice<T> | T>;
	  }
	| {
			kind: 'record';
			options: Record<string, string>;
	  };

export const parseChoice = <T>(value: unknown): Choice<T> | null => {
	const parsed = typedChoiceSchema<T>().safeParse(value);

	return parsed.success ? parsed.data : null;
};

export const parseChoiceRecord = (value: unknown): Record<string, string> | null => {
	const parsed = choiceRecordSchema.safeParse(value);

	return parsed.success ? parsed.data : null;
};

export const parseChoiceRecordEntries = (value: unknown): ChoiceRecordEntry[] => {
	const parsed = choiceRecordSchema.safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Choice records must map string values to string labels.');
	}

	return Object.entries(parsed.data).map(([value, label]) => ({ label, value }));
};

export const parseChoiceOptions = <T>(value: ChoiceOptions<T>): ParsedChoiceOptions<T> => {
	const record = choiceRecordSchema.safeParse(value);

	if (record.success) {
		return { kind: 'record', options: record.data };
	}

	const list = typedChoiceOptionsListSchema<T>().safeParse(value);

	if (!list.success) {
		throw new TypeError('Choice options must be an array or record.');
	}

	return { kind: 'list', options: list.data };
};

export const parseChoiceValue = <T>(value: unknown): T => {
	const parsed = choiceValueSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Choice values must resolve to a typed value.');
	}

	return parsed.data;
};
