import { z } from 'zod';

const choiceAnswerIndexSchema = z
	.string()
	.trim()
	.regex(/^\d+$/u)
	.transform((value) => Number.parseInt(value, 10));

const choiceAnswerListSchema = z.string().transform((value) =>
	value
		.split(',')
		.map((part) => part.trim())
		.filter((part) => part.length > 0),
);

const choiceRecordKeySchema = z
	.string()
	.regex(/^-?\d+$/u)
	.transform((value) => Number.parseInt(value, 10));

const choiceRecordValueSchema = <T>(): z.ZodType<T> => z.unknown() as z.ZodType<T>;

export const parseChoiceAnswerIndex = (answer: string): number | null => {
	const parsed = choiceAnswerIndexSchema.safeParse(answer);

	return parsed.success ? parsed.data : null;
};

export const parseChoiceAnswerList = (answer: unknown): string[] => {
	const parsed = choiceAnswerListSchema.safeParse(answer);

	if (!parsed.success) {
		throw new TypeError('Choice answers must be a comma-separated string.');
	}

	return parsed.data;
};

export const parseChoiceRecordKey = (key: string): string | number => {
	const parsed = choiceRecordKeySchema.safeParse(key);

	return parsed.success ? parsed.data : key;
};

export const parseChoiceRecordValue = <T>(key: string): T => {
	const parsed = choiceRecordValueSchema<T>().safeParse(parseChoiceRecordKey(key));

	if (!parsed.success) {
		throw new TypeError('Choice record keys must resolve to a choice value.');
	}

	return parsed.data;
};
