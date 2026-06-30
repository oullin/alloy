import { hasConfirmDefaultArgument, isSelectPromptLabel, parseConfirmDefault } from '#console/prompts/select/validators/overload';
import { hasPromptDefault } from '#console/validators/default';
import type { ConfirmPromptOptions } from '#console/types';

export type NormalizedConfirmPromptOptions = ConfirmPromptOptions & {
	default: boolean;
	hasDefault: boolean;
};

export const confirmHasDefault = (messageOrOptions: string | ConfirmPromptOptions, argumentCount: number, defaultValue?: boolean): boolean => {
	return isSelectPromptLabel(messageOrOptions) ? hasConfirmDefaultArgument(argumentCount, defaultValue) : hasPromptDefault(messageOrOptions);
};

export const preserveConfirmRetryDefault = (options: NormalizedConfirmPromptOptions, value: boolean | undefined): void => {
	if (value === undefined) {
		return;
	}

	options.default = value;
	options.hasDefault = true;
};

export const normalizeConfirmPromptOptions = (
	messageOrOptions: string | ConfirmPromptOptions,
	defaultValue: boolean,
	yes: string,
	no: string,
	required: boolean | string,
	validate: ConfirmPromptOptions['validate'],
	hint: string,
	transform: ConfirmPromptOptions['transform'],
	hasDefault: boolean,
): NormalizedConfirmPromptOptions => {
	if (!isSelectPromptLabel(messageOrOptions)) {
		return {
			...messageOrOptions,
			default: hasDefault ? parseConfirmDefault(messageOrOptions.default, true) : true,
			hasDefault,
		};
	}

	return {
		default: hasDefault ? defaultValue : true,
		hasDefault,
		hint,
		label: messageOrOptions,
		message: messageOrOptions,
		no,
		required,
		transform,
		validate,
		yes,
	};
};

export const transformConfirmValue = async (options: Pick<ConfirmPromptOptions, 'transform'>, value: boolean): Promise<boolean> => {
	return options.transform ? options.transform(value) : value;
};

export const transformedConfirmDefault = async (options: NormalizedConfirmPromptOptions): Promise<boolean> => {
	try {
		return await transformConfirmValue(options, options.default);
	} catch {
		return options.default;
	}
};
