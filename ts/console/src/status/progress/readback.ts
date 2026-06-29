import type { ProgressController } from '#console/status/progress/controller';

export const progressCurrent = (controller: ProgressController): number => {
	return controller.state.current();
};

export const progressPercentageValue = (controller: ProgressController): number => {
	return controller.state.percentage();
};

export const progressValue = (): boolean => true;
