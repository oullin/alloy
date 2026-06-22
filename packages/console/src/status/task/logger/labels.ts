export type TaskLoggerLabels = {
	label: string;
	subLabel: string;
};

export const createTaskLoggerLabels = (label: string, subLabel: string): TaskLoggerLabels => ({
	label,
	subLabel,
});

export const setTaskLoggerLabel = (labels: TaskLoggerLabels, label: string): TaskLoggerLabels => ({
	...labels,
	label,
});

export const setTaskLoggerSubLabel = (labels: TaskLoggerLabels, subLabel: string): TaskLoggerLabels => ({
	...labels,
	subLabel,
});
