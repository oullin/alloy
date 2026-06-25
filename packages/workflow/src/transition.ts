import type { Metadata } from './types.js';
import { cloneRecord } from './types.js';

export class Transition {
	public readonly name: string;
	public readonly from: string[];
	public readonly to: string[];
	public readonly metadata: Metadata;

	public constructor(name: string, from: string[] = [], to: string[] = [], metadata: Metadata = {}) {
		this.name = name;
		this.from = [...from];
		this.to = [...to];
		this.metadata = cloneRecord(metadata);
	}

	public clone(): Transition {
		return new Transition(this.name, this.from, this.to, this.metadata);
	}
}

export interface TransitionSnapshot {
	name: string;
	from: string[];
	to: string[];
}

export const snapshotTransition = (transition: Transition): TransitionSnapshot => ({
	name: transition.name,
	from: [...transition.from],
	to: [...transition.to],
});
