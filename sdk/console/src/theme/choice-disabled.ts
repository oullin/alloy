import { parseDisabledReason } from '#console/theme/validators/disabled';

export const choiceDisabledSuffix = (disabled: boolean | string | undefined): string => {
	if (!disabled) {
		return '';
	}

	return ` (${parseDisabledReason(disabled) ?? 'disabled'})`;
};
