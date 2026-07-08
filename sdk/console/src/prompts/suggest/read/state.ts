import { createSuggestHighlightState } from '#console/prompts/suggest/read/state/highlight';
import { createSuggestQueryState } from '#console/prompts/suggest/state/query';
import type { SuggestNavigationAction } from '#console/prompts/suggest/keys';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TypedValueState } from '#console/typed-value/types';

export type SuggestReaderState = {
	applyTypedInput(key: string): Promise<{ cancelled: boolean; submitted: boolean }>;
	clearHighlight(): void;
	highlighted(): number | null;
	matches(): string[];
	move(action: SuggestNavigationAction): Promise<void>;
	value(): TypedValueState;
};

export const createSuggestReaderState = async (options: SuggestOptions): Promise<SuggestReaderState> => {
	const query = await createSuggestQueryState(options);

	const highlighted = createSuggestHighlightState();

	return {
		async applyTypedInput(key) {
			const next = await query.applyTypedInput(key);

			if (next.submitted || next.cancelled) {
				return { cancelled: next.cancelled, submitted: next.submitted };
			}

			if (next.changed) {
				highlighted.clear();
			}

			return { cancelled: false, submitted: false };
		},
		clearHighlight() {
			highlighted.clear();
		},
		highlighted() {
			return highlighted.value();
		},
		matches() {
			return query.matches();
		},
		async move(action) {
			await query.resolve();

			highlighted.move(query.matches(), action, options.scroll);
		},
		value() {
			return query.value();
		},
	};
};
