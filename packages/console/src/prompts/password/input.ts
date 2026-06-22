import { promptEnvironment } from '#console/environment';
import { cancelPrompt } from '#console/prompt';
import { readPasswordFallbackValue } from '#console/prompts/password/fallback';
import { createPasswordReaderSession } from '#console/prompts/password/session';
import type { PasswordInputOptions, PasswordReadResult } from '#console/prompts/password/types';

export const readPasswordValue = async (message: string, options: PasswordInputOptions = {}): Promise<PasswordReadResult> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return readPasswordFallbackValue(message, options);
	}

	const session = createPasswordReaderSession(message, options);

	while (true) {
		const key = await environment.input.readKey();

		if (key === null) {
			return {
				cancelled: false,
				frame: session.frame(),
				value: session.value(),
			};
		}

		const next = session.applyTypedInput(key);

		if (next.cancelled) {
			session.cancel();

			return {
				cancelled: true,
				value: await cancelPrompt(session.value()),
			};
		}

		if (next.submitted) {
			return {
				cancelled: false,
				frame: session.frame(),
				value: session.value(),
			};
		}
	}
};
