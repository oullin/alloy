import { createSuggestFrameRenderer } from '#console/prompts/suggest/read/frame';
import { createSuggestReaderState } from '#console/prompts/suggest/read/state';
import type { SuggestNavigationAction } from '#console/prompts/suggest/keys';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TypedValueState } from '#console/typed-value/types';

export type SuggestReaderSession = {
	applyTypedInput(key: string): Promise<{ cancelled: boolean; submitted: boolean }>;
	clearHighlight(): void;
	frame(): string;
	highlighted(): number | null;
	matches(): string[];
	move(action: SuggestNavigationAction): Promise<void>;
	render(): void;
	state(): TypedValueState;
};

export const createSuggestReaderSession = async (options: SuggestOptions): Promise<SuggestReaderSession> => {
	const state = await createSuggestReaderState(options);

	const frame = createSuggestFrameRenderer(options);

	function render(): void {
		frame.render({ highlighted: state.highlighted(), matches: state.matches(), state: state.value() });
	}

	return {
		async applyTypedInput(key) {
			const next = await state.applyTypedInput(key);

			if (next.submitted || next.cancelled) {
				return next;
			}

			render();

			return next;
		},
		clearHighlight() {
			state.clearHighlight();
			render();
		},
		frame() {
			return frame.current();
		},
		highlighted() {
			return state.highlighted();
		},
		matches() {
			return state.matches();
		},
		async move(action) {
			await state.move(action);

			render();
		},
		render,
		state() {
			return state.value();
		},
	};
};
