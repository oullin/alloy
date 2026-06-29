export const SPINNER_FRAMES = ['⠂', '⠒', '⠐', '⠰', '⠠', '⠤', '⠄', '⠆'] as const;

export const STATIC_SPINNER_FRAME = '⠶';

export const spinnerFrame = (count: number): string => {
	const index = ((count % SPINNER_FRAMES.length) + SPINNER_FRAMES.length) % SPINNER_FRAMES.length;

	return SPINNER_FRAMES[index] ?? STATIC_SPINNER_FRAME;
};

export const renderSpinnerFrame = (message: string, count?: number): string => {
	const frame = count === undefined ? STATIC_SPINNER_FRAME : spinnerFrame(count);

	return ` ${frame} ${message}\n`;
};
