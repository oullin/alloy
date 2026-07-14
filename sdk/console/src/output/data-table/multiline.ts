export const splitMultilineDataTableCell = (cell: string): string[] => cell.split(/\r?\n/u);

export const expandMultilineDataTableRows = (rows: string[][]): string[][] => {
	return rows.flatMap((row) => {
		const columns = row.map(splitMultilineDataTableCell);
		const height = Math.max(1, ...columns.map((lines) => lines.length));

		return Array.from({ length: height }, (_, lineIndex) => {
			return columns.map((lines, columnIndex) => {
				if (columnIndex === 0 && lineIndex > 0) {
					return '';
				}

				return lines[lineIndex] ?? '';
			});
		});
	});
};
