import { truncate, visibleWidth } from '#console/strings';
import { terminalSize } from '#console/terminal';
import { parseBoxWidth } from '#console/theme/validators/box';

const DEFAULT_BOX_WIDTH = 60;
const TERMINAL_PADDING = 6;

type BoxOptions = {
	body: string;
	borderStyle?: (value: string) => string;
	info?: string;
	title?: string;
	width?: number;
};

const padVisible = (value: string, width: number): string => `${value}${' '.repeat(Math.max(0, width - visibleWidth(value)))}`;

const defaultBoxWidth = (): number => Math.min(DEFAULT_BOX_WIDTH, Math.max(0, terminalSize().columns - TERMINAL_PADDING));

export const renderBox = ({ body, borderStyle = (value) => value, info = '', title = '', width }: BoxOptions): string => {
	const bodyLines = body.split('\n');
	const requestedWidth = parseBoxWidth(width, defaultBoxWidth());
	const contentWidth = Math.max(requestedWidth, visibleWidth(title), ...bodyLines.map(visibleWidth));
	const titleWidth = visibleWidth(title);
	const titleLabel = titleWidth > 0 ? ` ${title} ` : '';
	const topBorder = '─'.repeat(contentWidth - titleWidth + (titleWidth > 0 ? 0 : 2));
	const renderedInfo = info ? truncate(info, contentWidth - 1) : '';
	const bottomBorder = '─'.repeat(renderedInfo ? contentWidth - visibleWidth(renderedInfo) : contentWidth + 2);

	return [
		`${borderStyle(' ┌')}${titleLabel}${borderStyle(`${topBorder}┐`)}`,
		...bodyLines.map((line) => `${borderStyle(' │')} ${padVisible(line, contentWidth)} ${borderStyle('│')}`),
		borderStyle(` └${bottomBorder}${renderedInfo ? ` ${renderedInfo} ` : ''}┘`),
	].join('\n');
};
