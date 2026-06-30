import { z } from 'zod';

const typedValueRowsSchema = z.number().finite().positive();

export const parseTypedValueRows = (rows: unknown): number | undefined => {
	const parsed = typedValueRowsSchema.safeParse(rows);

	if (!parsed.success) {
		return undefined;
	}

	return Math.floor(parsed.data);
};
