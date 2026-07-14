import { promptWithFallback } from '#console/prompt';
import { readPauseConfirmation } from '#console/prompts/pause/read';

/** Waits for the user to press enter before continuing. */
export const pause = async (message = 'Press enter to continue...'): Promise<boolean> => {
	return promptWithFallback('pause', { message }, async () => readPauseConfirmation(message));
};
