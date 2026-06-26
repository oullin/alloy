import { CompletedEvent, EventNames, type Dispatcher, type WorkflowEvent } from '../events.js';
import { Marking } from '../marking.js';
import { AuditEntry } from './entry.js';

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
