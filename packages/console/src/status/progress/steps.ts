import { parseProgressValuesInput } from '#console/status/validators/progress';

export const progressValues = <T>(steps: Iterable<T> | number): Array<T | number> => {
	const parsed = parseProgressValuesInput(steps);

	return parsed.kind === 'total' ? Array.from({ length: parsed.total }, (_, index) => index) : Array.from(parsed.steps);
};
