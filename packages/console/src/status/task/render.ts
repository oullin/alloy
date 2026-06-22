import { renderSpinnerFrame } from '#console/status/spinner/render';
import { parseLogLimit } from '#console/status/validators/limit';
import type { StableTaskMessage } from '#console/status/task/messages';

const TASK_DIVIDER_WIDTH = 60;

type TaskFrameOptions = {
	frameCount?: number;
	finished?: boolean;
	keepSummary?: boolean;
	label: string;
	limit: number;
	lines: string[];
	stableMessages: StableTaskMessage[];
	subLabel?: string;
};

const stableMessageSymbol = (type: StableTaskMessage['type']): string => {
	if (type === 'success') {
		return '✔';
	}

	if (type === 'error') {
		return '✘';
	}

	return '⚠';
};

const renderStableMessage = (message: StableTaskMessage): string => {
	return `   ${stableMessageSymbol(message.type)} ${message.message}`;
};

export const renderTaskFrame = ({ finished = false, frameCount, keepSummary = false, label, limit, lines, stableMessages, subLabel }: TaskFrameOptions): string => {
	const visibleLimit = parseLogLimit(limit, 10);

	if (finished && keepSummary && stableMessages.length > 0) {
		return [` • ${label}`, ...stableMessages.map(renderStableMessage), ''].join('\n');
	}

	const output = [renderSpinnerFrame(label, frameCount).trimEnd()];

	if (subLabel) {
		output.push(`   ${subLabel}`);
	}

	output.push(...stableMessages.map(renderStableMessage));

	if (stableMessages.length > 0 || lines.length > 0) {
		output.push(` ${'─'.repeat(TASK_DIVIDER_WIDTH)}`);
	} else {
		output.push('');
	}

	const visibleLines = lines.slice(-visibleLimit);

	output.push(...visibleLines.map((line) => ` ${line}`));

	for (let remaining = visibleLimit - visibleLines.length; remaining > 0; remaining -= 1) {
		output.push('');
	}

	return `${output.join('\n')}\n`;
};
