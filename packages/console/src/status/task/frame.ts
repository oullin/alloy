import { renderTaskFrame } from '#console/status/task/render';
import type { Logger } from '#console/status/task/logger';

export type LoggerTaskFrameOptions = {
	finished?: boolean;
	keepSummary?: boolean;
};

export const renderLoggerTaskFrame = (logger: Logger, options: LoggerTaskFrameOptions = {}): string => {
	const finished = options.finished ?? false;
	const keepSummary = options.keepSummary ?? false;

	return renderTaskFrame({
		finished,
		keepSummary,
		label: logger.labelValue,
		limit: logger.limitValue,
		lines: logger.lines,
		stableMessages: finished && !keepSummary ? [] : logger.stableMessages,
		subLabel: logger.subLabelValue,
	});
};
