import { clearTerminal, setTerminalTitle } from '#console/terminal';

export const title = (value: string): void => {
	setTerminalTitle(value);
};

export const clear = (): void => {
	clearTerminal();
};
