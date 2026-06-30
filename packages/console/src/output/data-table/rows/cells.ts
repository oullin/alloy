import type { TableCell } from '#console/types';

export const stringifyDataTableCell = (value: TableCell): string => {
	return value === null || value === undefined ? '' : String(value);
};
