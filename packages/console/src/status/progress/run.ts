import type { MaybePromise } from '#console/types';

type ProgressRunner = {
	advance(): void;
	fail(): void;
	finish(): void;
	start(): void;
};

export const runProgressSteps = async <T, R, TProgress extends ProgressRunner>(
	progress: TProgress,
	values: Array<T | number>,
	callback: (step: T | number, progress: TProgress) => MaybePromise<R>,
): Promise<R[]> => {
	const results: R[] = [];

	progress.start();

	try {
		for (const value of values) {
			results.push(await callback(value, progress));

			progress.advance();
		}
	} catch (error) {
		progress.fail();

		throw error;
	}

	progress.finish();

	return results;
};
