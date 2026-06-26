import type { Definition } from '../definition.js';
import type { Marking } from '../marking.js';
import type { Transition } from '../transition.js';
import type { MetadataStore, WorkflowContext } from '../types.js';

export interface WorkflowEngine<T> {
	name(): string;
	definition(): Definition;
	metadataStore(): MetadataStore;
	getMarking(subject: T): Marking;
	can(subject: T, transition: string): boolean;
	cannot(subject: T, transition: string): boolean;
	enabledTransitions(subject: T): Transition[];
	disabledTransitions(subject: T): Transition[];
	apply(subject: T, transition: string, context?: WorkflowContext): Marking;
}
