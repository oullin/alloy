export type VisibleLineWindow = {
	lines: string[];
	start: number;
	total: number;
};

export type VisibleTextWindow = VisibleLineWindow & {
	cursor: number;
	text: string;
};
