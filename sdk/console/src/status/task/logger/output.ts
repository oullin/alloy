import { appendStableTaskMessage } from '#console/status/task/logger/stable';
import type { StableTaskMessage } from '#console/status/task/messages';

export const resetTaskLoggerLines = (lines: string[]): void => {
	lines.splice(0);
};

export const appendStableTaskLoggerOutput = (messages: StableTaskMessage[], lines: string[], type: StableTaskMessage['type'], message: string, limit: number): void => {
	appendStableTaskMessage(messages, type, message, limit);
	resetTaskLoggerLines(lines);
};
