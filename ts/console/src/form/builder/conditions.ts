import { resolveFormStepCondition } from '#console/form/builder/validators/conditions';
import type { FormResponses, FormStep } from '#console/form/types';

export const shouldRunStep = async (step: FormStep, responses: FormResponses): Promise<boolean> => {
	return resolveFormStepCondition(step.condition, responses);
};

export const shouldIgnoreStepWhenReverting = async (step: FormStep, responses: FormResponses): Promise<boolean> => {
	if (!(await shouldRunStep(step, responses))) {
		return true;
	}

	return step.ignoreWhenReverting;
};
