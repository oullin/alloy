import type { Definition } from './definition.js';
import { EventNames } from './event-names.js';
import { TransitionError, TransitionNotFoundError } from './errors.js';
import { AnnounceEvent, BaseEvent, CompletedEvent, Dispatcher, EnteredEvent, EnterEvent, GuardEvent, LeaveEvent, TransitionEvent } from './events.js';
import type { Marking } from './marking.js';
import { DefinitionMetadataStore } from './metadata.js';
import type { MarkingStore } from './stores.js';
import { snapshotTransition, type Transition } from './transition.js';
import type { MetadataStore, Sink, WorkflowContext } from './types.js';

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

export class Machine<T> implements WorkflowEngine<T> {
	readonly #name: string;
	readonly #definition: Definition;
	readonly #store: MarkingStore<T>;
	readonly #dispatcher: Dispatcher<T>;
	readonly #metadata: DefinitionMetadataStore;
	#logger?: Sink;

	public constructor(name: string, definition: Definition, store: MarkingStore<T>, dispatcher = new Dispatcher<T>()) {
		if (name === '') {
			throw new Error('workflow name is required');
		}

		definition.validate();
		this.#name = name;
		this.#definition = definition.clone();
		this.#store = store;
		this.#dispatcher = dispatcher;
		this.#metadata = new DefinitionMetadataStore(definition);
	}

	public setLogger(logger: Sink): this {
		this.#logger = logger;

		return this;
	}

	public name(): string {
		return this.#name;
	}

	public definition(): Definition {
		return this.#definition.clone();
	}

	public metadataStore(): MetadataStore {
		return this.#metadata;
	}

	public eventDispatcher(): Dispatcher<T> {
		return this.#dispatcher;
	}

	public getMarking(subject: T): Marking {
		const marking = this.#store.getMarking(subject, this.#definition);

		if (marking.activePlaces().length === 0) {
			return this.#definition.initialMarking.clone();
		}

		return marking;
	}

	public can(subject: T, transitionName: string): boolean {
		try {
			this.transitionState(subject, transitionName, {});

			return true;
		} catch {
			return false;
		}
	}

	public cannot(subject: T, transitionName: string): boolean {
		return !this.can(subject, transitionName);
	}

	public enabledTransitions(subject: T): Transition[] {
		return this.transitionsByState(subject, true);
	}

	public disabledTransitions(subject: T): Transition[] {
		return this.transitionsByState(subject, false);
	}

	public apply(subject: T, transitionName: string, context: WorkflowContext = {}): Marking {
		this.logDebug('applying transition', 'transition', transitionName, 'workflow', this.#name);

		try {
			const { transition, next } = this.prepareApply(subject, transitionName, context);

			this.dispatchLeaveEvents(subject, transition, next, context);
			this.dispatchTransitionEvents(subject, transition, next, context);
			this.dispatchEnterEvents(subject, transition, next, context);
			this.#store.setMarking(subject, next.clone(), this.#definition, { ...context });
			this.logInfo('transition applied successfully', 'transition', transitionName, 'workflow', this.#name);
			this.dispatchEnteredEvents(subject, transition, next, context);
			this.dispatchCompletionEvents(subject, transition, next, context);

			return next;
		} catch (error) {
			this.logError('failed to apply transition', 'transition', transitionName, 'workflow', this.#name, 'error', error);
			throw error;
		}
	}

	private prepareApply(subject: T, transitionName: string, context: WorkflowContext): { transition: Transition; next: Marking } {
		const transition = this.#definition.transition(transitionName);

		if (transition === undefined) {
			throw new TransitionNotFoundError(transitionName);
		}

		const marking = this.getMarking(subject);

		this.transitionStateForMarking(subject, transition, marking, context);

		return {
			transition,
			next: this.buildNextMarking(marking, transition),
		};
	}

	private buildNextMarking(marking: Marking, transition: Transition): Marking {
		const next = marking.clone();

		for (const place of transition.from) {
			next.remove(place, 1);
		}

		for (const place of transition.to) {
			next.add(place, 1);
		}

		return next;
	}

	private transitionState(subject: T, transitionName: string, context: WorkflowContext): void {
		const transition = this.#definition.transition(transitionName);

		if (transition === undefined) {
			throw new TransitionNotFoundError(transitionName);
		}

		this.transitionStateForMarking(subject, transition, this.getMarking(subject), context);
	}

	private transitionStateForMarking(subject: T, transition: Transition, marking: Marking, context: WorkflowContext): void {
		if (!this.transitionEnabled(marking, transition)) {
			throw new TransitionError(this.#name, transition.name);
		}

		const guard = new GuardEvent(this.baseEvent(subject, transition, marking.clone(), context));

		this.#dispatcher.dispatch(EventNames.guard(this.#name), guard);
		this.#dispatcher.dispatch(EventNames.guardNamed(this.#name, transition.name), guard);

		if (guard.blocked()) {
			throw new TransitionError(this.#name, transition.name, guard.blockers());
		}
	}

	private transitionsByState(subject: T, enabled: boolean): Transition[] {
		const marking = this.getMarking(subject);
		const transitions: Transition[] = [];

		for (const transition of this.#definition.transitions) {
			let isEnabled = true;

			try {
				this.transitionStateForMarking(subject, transition, marking, {});
			} catch {
				isEnabled = false;
			}

			if (isEnabled === enabled) {
				transitions.push(transition.clone());
			}
		}

		return transitions;
	}

	private transitionEnabled(marking: Marking, transition: Transition): boolean {
		return transition.from.every((place) => marking.has(place));
	}

	private dispatchLeaveEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		const current = this.restoreFromPlaces(this.markingBeforeEnter(next, transition), transition);

		for (const place of transition.from) {
			const event = new LeaveEvent(this.baseEvent(subject, transition, current.clone(), context), place);

			this.#dispatcher.dispatch(EventNames.leave(this.#name), event);
			this.#dispatcher.dispatch(EventNames.leavePlace(this.#name, place), event);
			current.remove(place, 1);
		}
	}

	private dispatchTransitionEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		const current = this.markingBeforeEnter(next, transition);
		const base = this.baseEvent(subject, transition, current.clone(), context);

		this.#dispatcher.dispatch(EventNames.transition(this.#name), new TransitionEvent(base));
		this.#dispatcher.dispatch(EventNames.transitionNamed(this.#name, transition.name), new TransitionEvent(base));
	}

	private dispatchEnterEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		const current = this.markingBeforeEnter(next, transition);

		for (const place of transition.to) {
			const event = new EnterEvent(this.baseEvent(subject, transition, current.clone(), context), place);

			this.#dispatcher.dispatch(EventNames.enter(this.#name), event);
			this.#dispatcher.dispatch(EventNames.enterPlace(this.#name, place), event);
			current.add(place, 1);
		}
	}

	private dispatchEnteredEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		for (const place of transition.to) {
			const event = new EnteredEvent(this.baseEvent(subject, transition, next.clone(), context), place);

			this.#dispatcher.dispatch(EventNames.entered(this.#name), event);
			this.#dispatcher.dispatch(EventNames.enteredPlace(this.#name, place), event);
		}
	}

	private dispatchCompletionEvents(subject: T, transition: Transition, next: Marking, context: WorkflowContext): void {
		const completed = new CompletedEvent(this.baseEvent(subject, transition, next.clone(), context));

		this.#dispatcher.dispatch(EventNames.completed(this.#name), completed);
		this.#dispatcher.dispatch(EventNames.completedNamed(this.#name, transition.name), completed);

		const announce = new AnnounceEvent(this.baseEvent(subject, transition, next.clone(), context), this.enabledTransitions(subject).map(snapshotTransition));

		this.#dispatcher.dispatch(EventNames.announce(this.#name), announce);
		this.#dispatcher.dispatch(EventNames.announceNamed(this.#name, transition.name), announce);
	}

	private markingBeforeEnter(next: Marking, transition: Transition): Marking {
		const current = next.clone();

		for (const place of transition.to) {
			current.remove(place, 1);
		}

		return current;
	}

	private restoreFromPlaces(marking: Marking, transition: Transition): Marking {
		const current = marking.clone();

		for (const place of [...transition.from].reverse()) {
			current.add(place, 1);
		}

		return current;
	}

	private baseEvent(subject: T, transition: Transition, marking: Marking, context: WorkflowContext): BaseEvent<T> {
		return new BaseEvent(this.#name, subject, snapshotTransition(transition), marking.toJSON(), { ...context });
	}

	private logDebug(message: string, ...args: unknown[]): void {
		this.#logger?.debug(message, ...args);
	}

	private logInfo(message: string, ...args: unknown[]): void {
		this.#logger?.info(message, ...args);
	}

	private logError(message: string, ...args: unknown[]): void {
		this.#logger?.error(message, ...args);
	}
}
