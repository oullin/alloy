export type PromptSubmissionState<T> = {
	capture(submitted: boolean, cancelled: boolean, value: T): void;
	render(renderSubmitted: (value: T) => void): void;
	reset(): void;
};

export const createPromptSubmissionState = <T>(emptyValue: T): PromptSubmissionState<T> => {
	let shouldRender = false;
	let value = emptyValue;

	return {
		capture(submitted, cancelled, nextValue) {
			shouldRender = submitted && !cancelled;
			value = nextValue;
		},
		render(renderSubmitted) {
			if (shouldRender) {
				renderSubmitted(value);
			}
		},
		reset() {
			shouldRender = false;
			value = emptyValue;
		},
	};
};
