import type { Definition } from '#workflow/definition';
import type { Marking } from '#workflow/marking';
import type { Transition } from '#workflow/transition';
import type { MetadataStore, WorkflowContext } from '#workflow/types';

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
