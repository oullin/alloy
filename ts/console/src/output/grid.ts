import { createGridLayout } from '#console/output/grid/layout';
import { boxedGridLines } from '#console/output/grid/lines';
import { parseGridWidth } from '#console/output/grid/validators/layout';
import { terminalSize } from '#console/terminal';

export const renderGrid = (items: Array<string | number | boolean> = [], maxWidth?: number): string => {
	if (items.length === 0) {
		return '';
	}

	const values = items.map(String);
	const width = parseGridWidth(maxWidth, terminalSize().columns);
	const availableWidth = width - 2;
	const layout = createGridLayout(values, availableWidth);

	return boxedGridLines(layout)
		.map((line) => ` ${line}`)
		.join('\n');
};
