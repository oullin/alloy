import { parseOptionalScrollSize } from '#console/concerns/validators/scroll';

export const fixedVisualDataTableRows = (rows: string[][], scroll?: number): string[][] => {
	const height = parseOptionalScrollSize(scroll);

	if (height === undefined || rows.length >= height) {
		return rows;
	}

	const columnCount = Math.max(0, ...rows.map((row) => row.length));
	const paddedRows = [...rows];

	while (paddedRows.length < height) {
		paddedRows.push(Array.from({ length: columnCount }, () => ''));
	}

	return paddedRows;
};
