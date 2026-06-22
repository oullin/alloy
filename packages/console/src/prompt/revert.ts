import { promptEnvironment } from '#console/environment';
import { renderError } from '#console/theme';

export const rejectPromptRevert = (): void => {
	promptEnvironment().error.write(renderError('This cannot be reverted.'));
};
