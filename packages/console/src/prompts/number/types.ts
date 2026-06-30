export type NumberInputOptions = {
	default?: number | string;
	hasDefault?: boolean;
	hint?: string;
	integer?: boolean;
	max?: number;
	min?: number;
	placeholder?: string;
	step?: number;
};

export type NumberReadResult = {
	cancelled: boolean;
	frame?: string;
	value: string;
};
