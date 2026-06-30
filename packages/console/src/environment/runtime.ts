import { stderr as defaultStderr, stdin as defaultStdin, stdout as defaultStdout } from 'node:process';
import { defaultInput } from '#console/environment/default-input';
import { outputFromStream } from '#console/environment/stream-output';
import type { PromptEnvironment } from '#console/types';

export const defaultEnvironment: PromptEnvironment = {
	input: defaultInput,
	output: outputFromStream(defaultStdout),
	error: outputFromStream(defaultStderr),
	interactive: Boolean(defaultStdin.isTTY),
};
