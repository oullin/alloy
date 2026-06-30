import { erasePreviousLines } from '#console/terminal';

export const renderedFrameLineCount = (frame: string): number => {
	if (frame.length === 0) {
		return 0;
	}

	const lines = frame.split('\n').length;

	return frame.endsWith('\n') ? lines - 1 : lines;
};

export const eraseRenderedFrame = (frame: string): void => {
	erasePreviousLines(renderedFrameLineCount(frame));
};
