import { z } from 'zod';
import { parseTerminalColor } from '#console/terminal/validators/color';
import type { TerminalColor } from '#console/terminal/capabilities';

const streamFadeStepsSchema = z.number().finite().positive();

const streamFrameWidthSchema = z.number().finite().positive();

const streamFadeTrueColorSchema = z.boolean();

export const parseStreamFadeSteps = (steps: unknown, defaultValue: number): number => {
	const parsed = streamFadeStepsSchema.safeParse(steps);

	if (!parsed.success) {
		return defaultValue;
	}

	return Math.max(1, Math.floor(parsed.data));
};

export const parseStreamFrameWidth = (width: unknown, defaultValue: number): number => {
	const parsed = streamFrameWidthSchema.safeParse(width);

	if (!parsed.success) {
		return defaultValue;
	}

	return Math.max(1, Math.floor(parsed.data));
};

export const parseStreamFadeColor = (value: unknown, fallback: TerminalColor): TerminalColor => {
	return parseTerminalColor(value, fallback);
};

export const parseStreamFadeTrueColor = (value: unknown, fallback: boolean): boolean => {
	const parsed = streamFadeTrueColorSchema.safeParse(value);

	return parsed.success ? parsed.data : fallback;
};
