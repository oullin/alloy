import { renderBox } from '#console/theme/box';
import { clampedProgressCurrent, progressRatio } from '#console/status/progress/state/value';
import { progressFrameWidth } from '#console/status/progress/validators/frame';

export type ProgressFrameState = 'active' | 'cancel' | 'error' | 'submit';

export type ProgressFrameOptions = {
	current: number;
	hint: string;
	label: string;
	state?: ProgressFrameState;
	total: number;
	width?: number;
};

export const progressPercentage = (current: number, total: number): number => {
	return progressRatio(current, total);
};

const formatProgressNumber = (value: number): string => new Intl.NumberFormat('en-US').format(value);

export const formatProgressFraction = (current: number, total: number): string => {
	return `${formatProgressNumber(clampedProgressCurrent(current, total))} / ${formatProgressNumber(total)}`;
};

export const renderProgressFrame = ({ current, hint, label, state = 'active', total, width = 60 }: ProgressFrameOptions): string => {
	const fillWidth = progressFrameWidth(width);
	const filled = Math.ceil(fillWidth * progressPercentage(current, total));
	const body = '█'.repeat(filled);

	const lines = [renderBox({ body, info: formatProgressFraction(current, total), title: label, width: fillWidth })];

	if (state === 'active') {
		lines.push(hint ? `  ${hint}` : '');
	}

	return `${lines.join('\n')}\n`;
};
