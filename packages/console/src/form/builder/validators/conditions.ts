import { z } from 'zod';
import type { FormResponses, FormStepCondition } from '#console/form/types';
import type { MaybePromise } from '#console/types';

type FormStepConditionCallback = (responses: FormResponses) => MaybePromise<boolean>;

const formStepConditionCallbackSchema = z.function();

export const isFormStepConditionCallback = (condition: FormStepCondition): condition is FormStepConditionCallback => {
	return formStepConditionCallbackSchema.safeParse(condition).success;
};

export const resolveFormStepCondition = (condition: FormStepCondition, responses: FormResponses): MaybePromise<boolean> => {
	return isFormStepConditionCallback(condition) ? condition(responses) : condition;
};
