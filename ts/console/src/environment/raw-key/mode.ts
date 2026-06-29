import type { RawKeyInput } from '#console/environment/raw-key/types';

export type RawKeyInputMode = {
	activate(): void;
	restore(): void;
};

export const createRawKeyInputMode = (input: RawKeyInput): RawKeyInputMode => {
	const wasRaw = Boolean(input.isRaw);
	const wasPaused = input.isPaused?.() ?? true;

	return {
		activate() {
			if (!input.isTTY) {
				return;
			}

			input.setRawMode?.(true);
			input.resume();
		},
		restore() {
			if (!input.isTTY) {
				return;
			}

			try {
				input.setRawMode?.(wasRaw);
			} finally {
				if (wasPaused) {
					input.pause();
				} else {
					input.resume();
				}
			}
		},
	};
};
