import { z } from 'zod';

const outputStepNameSchema = z.string();
const outputScrollSchema = z.number();
const nullOutputRowsSchema = z.null();

export const isNullOutputRows = (value: unknown): value is null => {
	return nullOutputRowsSchema.safeParse(value).success;
};

export const parseOutputStepName = (value: unknown): string | undefined => {
	const parsed = outputStepNameSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

export const parseOutputScroll = (value: unknown, defaultValue: number): number => {
	const parsed = outputScrollSchema.safeParse(value);

	return parsed.success ? parsed.data : defaultValue;
};
