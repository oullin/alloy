import { renderDataTableFrame } from '#console/output/data-table/render';
import { eraseRenderedFrame } from '#console/status/frame';
import type { DataTableReaderState } from '#console/output/data-table/reader/state';
import type { DataTableReadOptions } from '#console/output/data-table/reader/types';

export const renderDataTableReaderSessionFrame = <T>(options: DataTableReadOptions<T>, headers: string[], state: DataTableReaderState<T>, currentFrame: string): string => {
	if (currentFrame.length > 0) {
		eraseRenderedFrame(currentFrame);
	}

	const rendered = renderDataTableFrame({
		allRows: options.rows,
		headers,
		message: options.message,
		mode: state.mode(),
		query: state.query(),
		rows: state.rows(),
		scroll: options.scroll,
		selected: state.selected(),
	});

	state.setSelected(rendered.selected);

	return rendered.frame;
};
