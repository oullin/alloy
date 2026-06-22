import { promptEnvironment } from '#console/environment';
import { Key } from '#console/key';
import { cancelPrompt } from '#console/prompt';
import { createNumberReaderSession } from '#console/prompts/number/session';
import type { NumberInputOptions, NumberReadResult } from '#console/prompts/number/types';

export const readInteractiveNumberValue = async (message: string, options: NumberInputOptions): Promise<NumberReadResult> => {
	const environment = promptEnvironment();
	const session = createNumberReaderSession(message, options);

	while (true) {
		const key = await environment.input.readKey?.();

		if (key === null || key === undefined) {
			return {
				cancelled: false,
				frame: session.frame(),
				value: session.value(),
			};
		}

		if (key === Key.up || key === Key.upArrow) {
			session.step(1);
			continue;
		}

		if (key === Key.down || key === Key.downArrow) {
			session.step(-1);
			continue;
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
