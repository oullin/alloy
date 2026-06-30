import { numberFormStep, passwordFormStep, textFormStep, textareaFormStep } from '#console/form/builder/prompts/basic/index';
import type { FormBuilder } from '#console/form/builder/index';
import type { BasicPromptBuilderMethods } from '#console/form/builder/prompts/types';

export const basicPromptBuilderMethods: BasicPromptBuilderMethods & ThisType<FormBuilder> = {
	number: numberFormStep,
	password: passwordFormStep,
	text: textFormStep,
	textarea: textareaFormStep,
};
