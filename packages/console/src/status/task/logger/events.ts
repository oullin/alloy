export type TaskLoggerChangeHandler = () => void;

export const notifyTaskLoggerChanged = (handler?: TaskLoggerChangeHandler): void => {
	handler?.();
};
