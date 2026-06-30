import type { AutocompletePromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices/autocomplete';
import type { ConfirmPromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices/confirm';
import type { MultiSearchPromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices/multisearch';
import type { MultiSelectPromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices/multiselect';
import type { PausePromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices/pause';
import type { SearchPromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices/search';
import type { SelectPromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices/select';
import type { SuggestPromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices/suggest';

export type ChoicePromptBuilderMethods = AutocompletePromptBuilderMethods &
	ConfirmPromptBuilderMethods &
	MultiSearchPromptBuilderMethods &
	MultiSelectPromptBuilderMethods &
	PausePromptBuilderMethods &
	SearchPromptBuilderMethods &
	SelectPromptBuilderMethods &
	SuggestPromptBuilderMethods;
