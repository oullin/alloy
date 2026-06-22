type RawKeyListener = (...args: any[]) => void;

export type RawKeyInput = {
	isPaused?(): boolean;
	isRaw?: boolean;
	isTTY?: boolean;
	off(eventName: string | symbol, listener: RawKeyListener): unknown;
	on(eventName: string | symbol, listener: RawKeyListener): unknown;
	once(eventName: string | symbol, listener: RawKeyListener): unknown;
	pause(): unknown;
	resume(): unknown;
	setRawMode?(mode: boolean): unknown;
};
