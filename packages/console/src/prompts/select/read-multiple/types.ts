export type MultipleChoicesReadResult<T> = {
	cancelled: boolean;
	frame?: string;
	submitted: boolean;
	submittedLabels: string[];
	value: T[];
};
