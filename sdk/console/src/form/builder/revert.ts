import { promptEnvironment, withPromptEnvironment } from '#console/environment';
import { Key } from '#console/key';
import { rejectPromptRevert } from '#console/prompt/revert';
import type { MaybePromise, PromptInput } from '#console/types';

export class FormRevertedError extends Error {
	constructor() {
		super('Form reverted.');
		this.name = 'FormRevertedError';
	}
}

export const runWithFormRevert = async <T>(canRevert: boolean, callback: () => MaybePromise<T>): Promise<T> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return callback();
	}

	const previousInput = environment.input;

	const input: PromptInput = {
		...previousInput,
		async readKey(): Promise<string | null> {
			while (true) {
				const key = await previousInput.readKey?.();

				if (key !== Key.ctrlU) {
					return key ?? null;
				}

				if (canRevert) {
					throw new FormRevertedError();
				}

				rejectPromptRevert();
			}
		},
	};

	return withPromptEnvironment({ input }, callback);
};
