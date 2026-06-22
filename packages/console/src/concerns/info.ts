import { isInfoCallback } from '#console/concerns/validators/info';

export type InfoResolver<T> = string | ((value: T | null) => string | null | undefined) | undefined;

export const resolveInfo = <T>(info: InfoResolver<T>, value: T | null): string => {
	if (isInfoCallback(info)) {
		return info(value) ?? '';
	}

	return info ?? '';
};

export const joinedInfoDetails = (...parts: string[]): string => {
	return parts.filter((part) => part.length > 0).join(' · ');
};
