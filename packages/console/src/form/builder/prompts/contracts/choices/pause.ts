import type { FormBuilder } from '#console/form/builder/index';

export type PausePromptBuilderMethods = {
	pause(this: FormBuilder, message?: string, name?: string): FormBuilder;
};
