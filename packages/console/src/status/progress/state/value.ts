import { parseProgressStep } from '#console/status/validators/progress';

export const nextProgressCurrent = (current: number, total: number, step: number): number => {
	return Math.max(0, Math.min(current + parseProgressStep(step), total));
};

export const clampedProgressCurrent = (current: number, total: number): number => {
	return Math.max(0, Math.min(current, total));
};

export const progressRatio = (current: number, total: number): number => {
	if (total <= 0) {
		return 0;
	}

	return clampedProgressCurrent(current, total) / total;
};
