import type { ProgressController } from '#console/status/progress/controller';

export const setProgressLabel = (controller: ProgressController, value: string): void => {
	controller.state.label(value);
};

export const setProgressHint = (controller: ProgressController, value: string): void => {
	controller.state.hint(value);
};
