import { Progress } from '#console/status/progress/progress';
import { runProgressSteps } from '#console/status/progress/run';
import { progressValues } from '#console/status/progress/steps';
import type { MaybePromise } from '#console/types';

import { isProgressTotalArgument, progressMessageArgument, progressStepsArgument } from '#console/status/progress/validators/arguments';

export { Progress };
export type { ProgressSignalTarget } from '#console/status/progress/progress';

/** Creates a {@link Progress} bar, or drives one over steps when a callback is given. */
export function progress(total: number, message?: string): Progress;

export function progress<T>(label: string, steps: Iterable<T> | number): Progress;

export function progress<T, R>(label: string, steps: Iterable<T> | number, callback: (step: T | number, progress: Progress) => MaybePromise<R>, hint?: string): Promise<R[]>;

export function progress<T, R>(
	labelOrTotal: string | number,
	stepsOrMessage?: Iterable<T> | number | string,
	callback?: (step: T | number, progress: Progress) => MaybePromise<R>,
	hint = '',
): Progress | Promise<R[]> {
	if (isProgressTotalArgument(labelOrTotal)) {
		return new Progress(labelOrTotal, progressMessageArgument(stepsOrMessage));
	}

	const values = progressValues(progressStepsArgument(stepsOrMessage));

	const bar = new Progress(values.length, labelOrTotal, hint);

	if (!callback) {
		return bar;
	}

	return runProgressSteps(bar, values, callback);
}
