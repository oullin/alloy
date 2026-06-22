import { isCompleteRawKey, normalizeRawKey } from '#console/environment/raw-key/normalize';
import { parseRawKeyChunkText } from '#console/environment/raw-key/validators/chunk';
import type { RawKeyInputMode } from '#console/environment/raw-key/mode';
import type { RawKeyInput } from '#console/environment/raw-key/types';

type RawKeySessionHandlers = {
	onData(chunk: unknown): void;
	onEnd(): void;
	onError(error: Error): void;
};

export const createRawKeySession = (input: RawKeyInput, mode: RawKeyInputMode, resolve: (value: null | string) => void, reject: (reason?: unknown) => void): RawKeySessionHandlers => {
	let buffer = '';
	let escapeTimer: NodeJS.Timeout | undefined;
	let settled = false;

	const cleanup = (): void => {
		clearTimeout(escapeTimer);
		input.off('data', handlers.onData);
		input.off('end', handlers.onEnd);
		input.off('error', handlers.onError);
		mode.restore();
	};

	const resolveWithCleanup = (value: null | string): void => {
		if (settled) {
			return;
		}

		settled = true;

		try {
			cleanup();
		} catch (error) {
			reject(error);

			return;
		}

		resolve(value);
	};

	const rejectWithCleanup = (error: unknown): void => {
		if (settled) {
			return;
		}

		settled = true;

		try {
			cleanup();
		} catch (cleanupError) {
			reject(cleanupError);

			return;
		}

		reject(error);
	};

	const resolveBufferedKey = (): void => {
		resolveWithCleanup(normalizeRawKey(buffer));
	};

	const handlers: RawKeySessionHandlers = {
		onData(chunk) {
			try {
				buffer += parseRawKeyChunkText(chunk);
			} catch (error) {
				rejectWithCleanup(error);

				return;
			}

			if (isCompleteRawKey(buffer)) {
				resolveBufferedKey();

				return;
			}

			clearTimeout(escapeTimer);
			escapeTimer = setTimeout(resolveBufferedKey, 25);
			escapeTimer.unref?.();
		},
		onEnd() {
			resolveWithCleanup(buffer.length > 0 ? normalizeRawKey(buffer) : null);
		},
		onError(error) {
			rejectWithCleanup(error);
		},
	};

	return handlers;
};
