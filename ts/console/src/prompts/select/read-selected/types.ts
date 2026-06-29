export type SelectedChoiceReadResult<T> = {
	cancelled: boolean;
	frame?: string;
	submitted: boolean;
	submittedLabel: string;
	value: T;
};
