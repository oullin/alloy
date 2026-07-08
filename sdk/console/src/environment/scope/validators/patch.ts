import { z } from 'zod';
import { functionSchema } from '#console/validators/function';
import type { PromptEnvironment, PromptInput, PromptOutput } from '#console/types';

const promptInputPatchSchema = z
	.object({
		readKey: functionSchema<NonNullable<PromptInput['readKey']>>().optional(),
		readLine: functionSchema<NonNullable<PromptInput['readLine']>>().optional(),
	})
	.passthrough() as unknown as z.ZodType<PromptInput>;

const promptOutputPatchSchema = z
	.object({
		write: functionSchema<PromptOutput['write']>(),
	})
	.passthrough() as unknown as z.ZodType<PromptOutput>;

const promptEnvironmentPatchSchema = z
	.object({
		error: promptOutputPatchSchema.optional(),
		input: promptInputPatchSchema.optional(),
		interactive: z.boolean().optional(),
		output: promptOutputPatchSchema.optional(),
	})
	.passthrough() as unknown as z.ZodType<Partial<PromptEnvironment>>;

export const parsePromptEnvironmentPatch = (environment: Partial<PromptEnvironment>): Partial<PromptEnvironment> => {
	const parsed = promptEnvironmentPatchSchema.safeParse(environment);

	if (!parsed.success) {
		throw new Error('Prompt environment patches must include valid input, output, error, or interactive values.');
	}

	return environment;
};
