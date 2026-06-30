import type { BasicPromptBuilderMethods } from '#console/form/builder/prompts/contracts/basic';
import type { ChoicePromptBuilderMethods } from '#console/form/builder/prompts/contracts/choices';

export type PromptBuilderMethods = BasicPromptBuilderMethods & ChoicePromptBuilderMethods;
