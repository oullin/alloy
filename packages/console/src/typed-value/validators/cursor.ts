import { z } from 'zod';

const typedValueCursorSchema = z.number().finite().nonnegative();
const typedValueLengthSchema = z.number().finite().nonnegative();

export const parseTypedValueCursor = (cursor: unknown, length: unknown): number => {
	const parsedLength = typedValueLengthSchema.safeParse(length);
	const max = parsedLength.success ? Math.floor(parsedLength.data) : 0;
	const parsedCursor = typedValueCursorSchema.safeParse(cursor);

	if (!parsedCursor.success) {
		return 0;
	}

	return Math.min(Math.floor(parsedCursor.data), max);
};
