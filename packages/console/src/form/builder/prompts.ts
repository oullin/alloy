import { basicPromptBuilderMethods } from '#console/form/builder/prompts/basic';
import { choicePromptBuilderMethods } from '#console/form/builder/prompts/choices';
import type { FormBuilder } from '#console/form/builder/index';
import type { PromptBuilderMethods } from '#console/form/builder/prompts/types';

export type { PromptBuilderMethods };

export const promptBuilderMethods: PromptBuilderMethods & ThisType<FormBuilder> = {
	...basicPromptBuilderMethods,
	...choicePromptBuilderMethods,
};
