import { eraseRenderedFrame } from '#console/status/frame';
import { StatusSignalCleanup } from '#console/status/signals';
import { Logger } from '#console/status/task/logger';
import { captureTaskProcessOutput } from '#console/status/task/process-output';
import { createTaskLifecycleRenderer } from '#console/status/task/lifecycle/rendering';
import { hideCursor, showCursor } from '#console/terminal';
import type { ResolvedTaskDefinition } from '#console/status/task/definition';
import type { TaskLifecycleRenderer } from '#console/status/task/lifecycle/rendering';

export const runTaskLifecycle = async <T>(definition: ResolvedTaskDefinition<T>): Promise<T> => {
	let renderer: TaskLifecycleRenderer | null = null;

	const logger = new Logger(definition.limit, definition.title, definition.subLabel, () => {
		renderer?.render();
	});

	let processOutput: ReturnType<typeof captureTaskProcessOutput>;

	try {
		hideCursor();
		renderer = createTaskLifecycleRenderer(logger);
		renderer.render();
		processOutput = captureTaskProcessOutput(logger);
	} catch (error) {
		showCursor();

		throw error;
	}

	const cleanup = new StatusSignalCleanup(() => {
		try {
			processOutput.stop();
			eraseRenderedFrame(renderer?.current() ?? '');
		} finally {
			showCursor();
		}
	}).attach();

	try {
		const result = await definition.run(logger);

		processOutput.stop();
		renderer.render({ finished: true, keepSummary: definition.keepSummary });

		return result;
	} catch (error) {
		processOutput.stop();
		renderer.render();

		throw error;
	} finally {
		cleanup.detach();
		showCursor();
	}
};
