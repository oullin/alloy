import { createAutocompleteFrameRenderer } from '#console/prompts/suggest/read-autocomplete/frame';
import { createAutocompleteReaderState } from '#console/prompts/suggest/read-autocomplete/state';
import type { TypedValueState } from '#console/typed-value/types';
import type { SuggestOptions } from '#console/prompts/suggest/options';

export type AutocompleteReaderSession = {
	acceptHighlighted(requireGrowth: boolean): Promise<void>;
	applyTypedInput(key: string): Promise<{ cancelled: boolean; submitted: boolean }>;
	frame(): string;
	move(direction: 1 | -1): Promise<void>;
	render(): void;
	state(): TypedValueState;
};

export const createAutocompleteReaderSession = async (options: SuggestOptions): Promise<AutocompleteReaderSession> => {
	const state = await createAutocompleteReaderState(options);

	const frame = createAutocompleteFrameRenderer(options);

	function render(): void {
		frame.render({ highlighted: state.highlighted(), matches: state.matches(), state: state.value() });
	}

	return {
		async acceptHighlighted(requireGrowth) {
			await state.acceptHighlighted(requireGrowth);

			render();
		},
		async applyTypedInput(key) {
			const next = await state.applyTypedInput(key);

			if (next.submitted || next.cancelled) {
				return next;
			}

			render();

			return next;
		},
		frame() {
			return frame.current();
		},
		async move(direction) {
			await state.move(direction);

			render();
		},
		render,
		state() {
			return state.value();
		},
	};
};
