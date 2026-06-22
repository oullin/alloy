import { shouldIgnoreStepWhenReverting, shouldRunStep } from '#console/form/builder/conditions';
import { runWithFormRevert } from '#console/form/builder/revert';
import { isFormRevertedError } from '#console/form/builder/validators/revert';
import type { FormResponses, FormStep } from '#console/form/types';

export const submitFormSteps = async (steps: FormStep[], responses: FormResponses): Promise<FormResponses> => {
	let index = 0;
	let wasReverted = false;

	while (index < steps.length) {
		const step = steps[index];

		if (step === undefined) {
			break;
		}

		if (wasReverted && index > 0 && (await shouldIgnoreStepWhenReverting(step, responses))) {
			index -= 1;
			continue;
		}

		wasReverted = false;

		const key = step.name ?? index;

		if (!(await shouldRunStep(step, responses))) {
			responses[key] = null;
			index += 1;
			continue;
		}

		try {
			responses[key] = await runWithFormRevert(index > 0, () => step.run(responses, responses[key], step.name));
		} catch (error) {
			if (!isFormRevertedError(error)) {
				throw error;
			}

			wasReverted = true;
		}

		index += wasReverted ? -1 : 1;
	}

	return responses;
};
