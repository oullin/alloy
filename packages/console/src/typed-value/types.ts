export type TypedValueState = {
	cursor: number;
	value: string;
};

export type TypedValueOptions = {
	allowNewLine?: boolean;
	default?: string;
	hint?: string;
	placeholder?: string;
	rows?: number;
};

export type AppliedTypedKey = TypedValueState & {
	cancelled: boolean;
	submitted: boolean;
};
