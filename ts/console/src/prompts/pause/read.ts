import { promptEnvironment } from '#console/environment';
import { Key } from '#console/key';
import { ask, cancelPrompt } from '#console/prompt';
import { renderPause } from '#console/prompts/pause/render';

export const readPauseConfirmation = async (message: string): Promise<boolean> => {
	const environment = promptEnvironment();

	if (!environment.interactive) {
		return false;
	}

	if (!environment.input.readKey) {
		await ask(message);

		return true;
	}

	environment.output.write(renderPause(message));

	while (true) {
		const key = await environment.input.readKey();

		if (key === null || key === Key.enter) {
			environment.output.write('\n');

			return true;
		}

		if (key === Key.ctrlC) {
			environment.output.write('\n');

			return cancelPrompt(false);
		}
	}
};
