import type { ProgressFrameState } from '#console/status/progress/render';

export type ProgressStateSnapshot = {
	current: number;
	hint: string;
	label: string;
	state: ProgressFrameState;
	total: number;
};

export const progressStateSnapshot = (current: number, total: number, label: string, hint: string, state: ProgressFrameState): ProgressStateSnapshot => ({
	current,
	hint,
	label,
	state,
	total,
});
