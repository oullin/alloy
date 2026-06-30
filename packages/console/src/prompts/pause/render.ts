import { renderBox } from '#console/theme/box';
import { cyan } from '#console/theme/styles';

export const renderPause = (message: string): string => {
	return renderBox({ body: '', borderStyle: cyan, title: cyan(message) });
};
