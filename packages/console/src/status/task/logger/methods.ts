import { appendPartialTaskLog, appendTaskLogLines } from '#console/status/task/logger/lines';
import { appendStableTaskLoggerOutput } from '#console/status/task/logger/output';
import type { TaskLoggerLimits } from '#console/status/task/logger/limits';
import type { PartialTaskLogState } from '#console/status/task/logger/lines';
import type { StableTaskMessage } from '#console/status/task/messages';

export const writeTaskLoggerLine = (lines: string[], message: string, limit: number): void => {
	appendTaskLogLines(lines, message.trimEnd(), limit);
};

export const writeTaskLoggerPartial = (lines: string[], partial: PartialTaskLogState, chunk: string, limit: number): PartialTaskLogState => {
	return appendPartialTaskLog(lines, partial, chunk, limit);
};

export const clearTaskLoggerPartial = (): PartialTaskLogState => ({
	startIndex: null,
	value: '',
});

export const writeStableTaskLoggerMessage = (stableMessages: StableTaskMessage[], lines: string[], type: StableTaskMessage['type'], message: string, limits: TaskLoggerLimits): PartialTaskLogState => {
	appendStableTaskLoggerOutput(stableMessages, lines, type, message, limits.stable);

	return clearTaskLoggerPartial();
};
