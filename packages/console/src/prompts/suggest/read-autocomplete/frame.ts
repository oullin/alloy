import { eraseRenderedFrame } from '#console/status/frame';
import { renderAutocomplete } from '#console/prompts/suggest/render-autocomplete';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TypedValueState } from '#console/typed-value/types';

type AutocompleteFrameRenderOptions = {
	highlighted: number;
	matches: string[];
	state: TypedValueState;
};

type AutocompleteFrameRenderer = {
	current(): string;
	render(options: AutocompleteFrameRenderOptions): void;
};

export const createAutocompleteFrameRenderer = (options: SuggestOptions): AutocompleteFrameRenderer => {
	let frame = '';

	return {
		current() {
			return frame;
		},
		render({ highlighted, matches, state }) {
			if (frame.length > 0) {
				eraseRenderedFrame(frame);
			}

			frame = renderAutocomplete(options.message, state, matches, highlighted, options.hint, options.placeholder, options.info);
		},
	};
};
