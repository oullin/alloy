import { symbols } from '#console/theme/symbols';

export const renderQuestion = (message: string, hint?: string): string => {
	const suffix = hint ? ` ${hint}` : '';

	return `${symbols.question} ${message}${suffix} `;
};
