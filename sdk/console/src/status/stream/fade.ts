import { backgroundColor, foregroundColor, supportsTrueColor } from '#console/terminal';
import { dim, foregroundRgb } from '#console/theme/styles';
import { parseStreamFadeColor, parseStreamFadeSteps, parseStreamFadeTrueColor } from '#console/status/stream/validators/fade';
import type { TerminalColor } from '#console/terminal/capabilities';

export type StreamFadeOptions = {
	background?: TerminalColor;
	foreground?: TerminalColor;
	steps?: number;
	trueColor?: boolean;
};

export type StreamFadeStyle = (value: string) => string;

const interpolate = (foreground: TerminalColor, background: TerminalColor, factor: number): TerminalColor => {
	const red = Math.trunc(background[0] + (foreground[0] - background[0]) * factor);
	const green = Math.trunc(background[1] + (foreground[1] - background[1]) * factor);
	const blue = Math.trunc(background[2] + (foreground[2] - background[2]) * factor);

	return [red, green, blue];
};

export const streamFadeStyles = (options: StreamFadeOptions = {}): StreamFadeStyle[] => {
	const trueColor = parseStreamFadeTrueColor(options.trueColor, supportsTrueColor());

	if (!trueColor) {
		return [(value) => value, dim];
	}

	const steps = parseStreamFadeSteps(options.steps, 10);

	const foreground = parseStreamFadeColor(options.foreground, foregroundColor());

	const background = parseStreamFadeColor(options.background, backgroundColor());

	return Array.from({ length: steps }, (_, step) => {
		const [red, green, blue] = interpolate(foreground, background, 1 - step / steps);

		return (value: string) => foregroundRgb(value, red, green, blue);
	});
};
