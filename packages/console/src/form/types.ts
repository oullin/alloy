import type { MaybePromise } from '#console/types';

export interface FormResponses {
	[name: string | number]: unknown;
}

export type FormStepCondition = boolean | ((responses: FormResponses) => MaybePromise<boolean>);

export type FormStep = {
	condition: FormStepCondition;
	ignoreWhenReverting: boolean;
	name?: string;
	run: (responses: FormResponses, previous: unknown, name?: string) => MaybePromise<unknown>;
};
