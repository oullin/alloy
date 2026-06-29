import { red } from '#console/theme/styles';
import { symbols } from '#console/theme/symbols';

export const renderError = (message: string): string => `${red(`${symbols.error} ${message}`)}\n`;
