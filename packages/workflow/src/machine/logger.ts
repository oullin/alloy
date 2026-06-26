import type { Sink } from '#workflow/types';

export class MachineLogger {
	#logger?: Sink;

	public set(logger: Sink): void {
		this.#logger = logger;
	}

	public debug(message: string, ...args: unknown[]): void {
		this.#logger?.debug(message, ...args);
	}

	public info(message: string, ...args: unknown[]): void {
		this.#logger?.info(message, ...args);
	}

	public error(message: string, ...args: unknown[]): void {
		this.#logger?.error(message, ...args);
	}
}
