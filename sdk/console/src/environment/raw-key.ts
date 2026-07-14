import { createRawKeyInputMode } from '#console/environment/raw-key/mode';
import { createRawKeySession } from '#console/environment/raw-key/session';
import type { RawKeyInput } from '#console/environment/raw-key/types';

export type { RawKeyInput };

export const readRawKey = async (input: RawKeyInput): Promise<null | string> =>
	new Promise((resolve, reject) => {
		const mode = createRawKeyInputMode(input);
		const session = createRawKeySession(input, mode, resolve, reject);

		input.on('data', session.onData);
		input.once('end', session.onEnd);
		input.once('error', session.onError);

		if (input.isTTY) {
			try {
				mode.activate();
			} catch (error) {
				input.off('data', session.onData);
				input.off('end', session.onEnd);
				input.off('error', session.onError);
				mode.restore();
				reject(error);
			}
		}
	});
