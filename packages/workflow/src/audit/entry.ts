import type { Marking } from '../marking.js';
import type { WorkflowContext } from '../types.js';

export class AuditEntry<T> {
	public readonly machine: string;
	public readonly transition: string;
	public readonly subject: T;
	public readonly marking: Marking;
	public readonly context: WorkflowContext;

	public constructor(input: { machine: string; transition: string; subject: T; marking: Marking; context: WorkflowContext }) {
		this.machine = input.machine;
		this.transition = input.transition;
		this.subject = input.subject;
		this.marking = input.marking.clone();
		this.context = { ...input.context };
	}
}
