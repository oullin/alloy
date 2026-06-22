import type { MaybePromise } from '#console/types';

export interface FormResponses extends Array<unknown> {
	[name: string]: unknown;
}

export type FormStepCondition = boolean | ((responses: FormResponses) => MaybePromise<boolean>);

export type FormStep = {
	condition: FormStepCondition;
	ignoreWhenReverting: boolean;
	name?: string;
	run: (responses: FormResponses, previous: unknown, name?: string) => MaybePromise<unknown>;
};
