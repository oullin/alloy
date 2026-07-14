import type { FormBuilder } from '#console/form/builder/index';
import type { ChoicePromptBuilderMethods } from '#console/form/builder/prompts/types';

import {
	autocompleteFormStep,
	confirmFormStep,
	multisearchFormStep,
	multiselectFormStep,
	pauseFormStep,
	searchFormStep,
	selectFormStep,
	suggestFormStep,
} from '#console/form/builder/prompts/choice/index';

export const choicePromptBuilderMethods: ChoicePromptBuilderMethods & ThisType<FormBuilder> = {
	autocomplete: autocompleteFormStep,
	confirm: confirmFormStep,
	multisearch: multisearchFormStep,
	multiselect: multiselectFormStep,
	pause: pauseFormStep,
	search: searchFormStep,
	select: selectFormStep,
	suggest: suggestFormStep,
};
