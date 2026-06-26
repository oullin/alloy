import { TransitionBlocker } from './transition-blocker.js';

export class TransitionError extends Error {
	public readonly machine: string;
	public readonly transition: string;
	public readonly blockers: TransitionBlocker[];

	public constructor(machine: string, transition: string, blockers: TransitionBlocker[] = []) {
		const suffix = blockers.length > 0 ? `: ${blockers.map((blocker) => blocker.message).join('; ')}` : '';

		super(`cannot apply transition "${transition}" on workflow "${machine}"${suffix}`);
		this.name = 'TransitionError';
		this.machine = machine;
		this.transition = transition;
		this.blockers = blockers.map((blocker) => TransitionBlocker.from(blocker));
	}
}
