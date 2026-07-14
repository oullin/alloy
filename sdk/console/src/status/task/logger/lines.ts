import { sanitizeTaskLine } from '#console/status/task/sanitize';
import { parseTaskLogMessageLines } from '#console/status/task/logger/validators/message';
import { parseLogLimit } from '#console/status/validators/limit';

export type PartialTaskLogState = {
	startIndex: number | null;
	value: string;
};

const parsedTaskLogLines = (message: string): string[] => parseTaskLogMessageLines(message).map((line) => sanitizeTaskLine(line));

const trimTaskLogLines = (lines: string[], limit: number): number => {
	const visibleLimit = parseLogLimit(limit, 10);
	const removed = Math.max(0, lines.length - visibleLimit);

	if (removed > 0) {
		lines.splice(0, removed);
	}

	return removed;
};

export const appendTaskLogLines = (lines: string[], message: string, limit: number): void => {
	for (const line of parsedTaskLogLines(message)) {
		lines.push(line);
	}

	trimTaskLogLines(lines, limit);
};

export const appendPartialTaskLog = (lines: string[], partial: PartialTaskLogState, chunk: string, limit: number): PartialTaskLogState => {
	const value = `${partial.value}${chunk}`;
	const startIndex = Math.min(partial.startIndex ?? lines.length, lines.length);
	const parsed = parsedTaskLogLines(value);

	lines.splice(startIndex);
	lines.push(...parsed);

	const removed = trimTaskLogLines(lines, limit);
	const nextStartIndex = parsed.length === 0 || lines.length === 0 ? null : Math.max(0, startIndex - removed);

	return {
		startIndex: nextStartIndex,
		value,
	};
};
