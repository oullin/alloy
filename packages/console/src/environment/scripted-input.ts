import { parseScriptedInputLines } from '#console/environment/validators/scripted-input';
import type { PromptInput } from '#console/types';

export const createScriptedInput = (lines: string[]): PromptInput => {
	const queued = parseScriptedInputLines(lines);

	return {
		async readKey(): Promise<string | null> {
			return queued.shift() ?? null;
		},
		async readLine(): Promise<string> {
			return queued.shift() ?? '';
		},
	};
};
