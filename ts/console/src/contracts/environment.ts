import type { MaybePromise } from '#console/contracts/base';

export type PromptOutput = {
	write(content: string): void;
};

export type PromptInput = {
	readKey?(): MaybePromise<string | null>;
	readLine?(message: string): MaybePromise<string>;
};

export type PromptEnvironment = {
	input: PromptInput;
	output: PromptOutput;
	error: PromptOutput;
	interactive: boolean;
};
