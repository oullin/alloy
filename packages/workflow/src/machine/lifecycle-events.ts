import { EventNames } from '../event-names.js';
import { AnnounceEvent, BaseEvent, CompletedEvent, type Dispatcher, EnteredEvent, EnterEvent, LeaveEvent, TransitionEvent } from '../events.js';
import type { Marking } from '../marking.js';
import { snapshotTransition, type Transition } from '../transition.js';
import type { WorkflowContext } from '../types.js';
import { markingBeforeEnter, restoreFromPlaces } from './markings.js';

export class WorkflowLifecycleDispatcher<T> {
	readonly #machine: string;
	readonly #dispatcher: Dispatcher<T>;

	public constructor(machine: string, dispatcher: Dispatcher<T>) {
		this.#machine = machine;
		this.#dispatcher = dispatcher;
	}

	public dispatchLeaveEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		const current = restoreFromPlaces(markingBeforeEnter(next, transition), transition);

		for (const place of transition.from) {
			const event = new LeaveEvent(this.baseEvent(subject, transition, current.clone(), context), place);

			this.#dispatcher.dispatch(EventNames.leave(this.#machine), event);
			this.#dispatcher.dispatch(EventNames.leavePlace(this.#machine, place), event);
			current.remove(place, 1);
		}
	}

	public dispatchTransitionEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		const current = markingBeforeEnter(next, transition);
		const base = this.baseEvent(subject, transition, current.clone(), context);

		this.#dispatcher.dispatch(EventNames.transition(this.#machine), new TransitionEvent(base));
		this.#dispatcher.dispatch(EventNames.transitionNamed(this.#machine, transition.name), new TransitionEvent(base));
	}

	public dispatchEnterEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		const current = markingBeforeEnter(next, transition);

		for (const place of transition.to) {
			const event = new EnterEvent(this.baseEvent(subject, transition, current.clone(), context), place);

			this.#dispatcher.dispatch(EventNames.enter(this.#machine), event);
			this.#dispatcher.dispatch(EventNames.enterPlace(this.#machine, place), event);
			current.add(place, 1);
		}
	}

	public dispatchEnteredEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		for (const place of transition.to) {
			const event = new EnteredEvent(this.baseEvent(subject, transition, next.clone(), context), place);

			this.#dispatcher.dispatch(EventNames.entered(this.#machine), event);
			this.#dispatcher.dispatch(EventNames.enteredPlace(this.#machine, place), event);
		}
	}

	public dispatchCompletionEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext, enabledTransitions: Transition[]): void {
		const completed = new CompletedEvent(this.baseEvent(subject, transition, next.clone(), context));

		this.#dispatcher.dispatch(EventNames.completed(this.#machine), completed);
		this.#dispatcher.dispatch(EventNames.completedNamed(this.#machine, transition.name), completed);

		const announce = new AnnounceEvent(this.baseEvent(subject, transition, next.clone(), context), enabledTransitions.map(snapshotTransition));

		this.#dispatcher.dispatch(EventNames.announce(this.#machine), announce);
		this.#dispatcher.dispatch(EventNames.announceNamed(this.#machine, transition.name), announce);
	}

	private baseEvent(subject: T, transition: Transition, marking: Marking, context: WorkflowContext): BaseEvent<T> {
		return new BaseEvent(this.#machine, subject, snapshotTransition(transition), marking.toJSON(), { ...context });
	}
}
