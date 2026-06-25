import { CompletedEvent, EventNames, type Dispatcher, type WorkflowEvent } from './events.js';
import { Marking } from './marking.js';
import type { WorkflowContext } from './types.js';

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

export class AuditTrail<T> {
	public readonly entries: AuditEntry<T>[] = [];

	public attach(workflowName: string, dispatcher: Dispatcher<T>): this {
		dispatcher.on(EventNames.completed(workflowName), (event: WorkflowEvent<T>) => {
			if (!(event instanceof CompletedEvent)) {
				return;
			}

			this.entries.push(
				new AuditEntry({
					machine: event.workflowName(),
					transition: event.transition().name,
					subject: event.subject(),
					marking: new Marking(event.marking()),
					context: event.context(),
				}),
			);
		});

		return this;
	}
}
