import { promptEnvironment } from '#console/environment';
import { eraseRenderedFrame } from '#console/status/frame';
import { renderLoggerTaskFrame } from '#console/status/task/frame';
import type { Logger } from '#console/status/task/logger';

type TaskLifecycleRenderOptions = {
	finished?: boolean;
	keepSummary?: boolean;
};

export type TaskLifecycleRenderer = {
	current(): string;
	render(options?: TaskLifecycleRenderOptions): void;
};

export const createTaskLifecycleRenderer = (logger: Logger): TaskLifecycleRenderer => {
	const output = promptEnvironment().output;

	let frame = '';

	return {
		current() {
			return frame;
		},
		render(options = {}) {
			const finished = options.finished ?? false;

			if (frame.length > 0) {
				eraseRenderedFrame(frame);
			}

			const nextFrame = renderLoggerTaskFrame(logger, options);

			if (frame.length === 0 || finished) {
				output.write(nextFrame);
			}

			frame = nextFrame;
		},
	};
};
