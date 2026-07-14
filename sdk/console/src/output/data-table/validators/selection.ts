import { z } from 'zod';

const dataTableRowTotalSchema = z.number().finite().nonnegative();
const dataTableSelectedIndexSchema = z.number().finite().nonnegative();

export const parseDataTableRowTotal = (total: unknown): number => {
	const parsed = dataTableRowTotalSchema.safeParse(total);

	return parsed.success ? Math.floor(parsed.data) : 0;
};

export const parseDataTableSelectedIndex = (selected: unknown, total: unknown): number => {
	const rowTotal = parseDataTableRowTotal(total);

	if (rowTotal === 0) {
		return 0;
	}

	const parsed = dataTableSelectedIndexSchema.safeParse(selected);

	if (!parsed.success) {
		return 0;
	}

	return Math.min(Math.floor(parsed.data), rowTotal - 1);
};
