import { promptEnvironment } from '#console/environment';
import { eraseRenderedFrame } from '#console/status/frame';
import { StatusSignalCleanup } from '#console/status/signals';
import { renderSpinnerFrame } from '#console/status/spinner/render';
import { hideCursor, showCursor } from '#console/terminal';
import type { MaybePromise } from '#console/types';

export const runSpinnerLifecycle = async <T>(message: string, callback: () => MaybePromise<T>): Promise<T> => {
	const frame = renderSpinnerFrame(message);

	let rendered = false;

	const cleanup = new StatusSignalCleanup(() => {
		try {
			if (rendered) {
				eraseRenderedFrame(frame);
			}
		} finally {
			showCursor();
		}
	});

	try {
		hideCursor();
		cleanup.attach();
		promptEnvironment().output.write(frame);
		rendered = true;
	} catch (error) {
		cleanup.detach();
		showCursor();

		throw error;
	}

	try {
		return await callback();
	} finally {
		cleanup.restore();
	}
};
