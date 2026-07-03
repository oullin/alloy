/**
 * Handler shapes attached by raw-key reads: `data` listeners receive the raw
 * chunk, `error` listeners receive the failure, and `end` listeners receive
 * nothing (see `raw-key.ts` and `raw-key/session.ts`).
 */
type RawKeyListener = ((chunk: unknown) => void) | ((error: Error) => void) | (() => void);

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
