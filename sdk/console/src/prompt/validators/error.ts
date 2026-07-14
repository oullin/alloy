import { z } from 'zod';
import { PromptValidationError } from '#console/prompt/error';

const promptValidationErrorSchema = z.instanceof(PromptValidationError);

export const isPromptValidationError = (error: unknown): error is PromptValidationError => {
	return promptValidationErrorSchema.safeParse(error).success;
};
