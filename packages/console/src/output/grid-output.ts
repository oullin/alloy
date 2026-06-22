import { promptEnvironment } from '#console/environment';
import { renderGrid } from '#console/output/grid';

export const grid = (items: Array<string | number | boolean> = [], maxWidth?: number): void => {
	const rendered = renderGrid(items, maxWidth);

	if (rendered === '') {
		return;
	}

	promptEnvironment().output.write(`${rendered}\n`);
};
