import type { Definition } from '../definition.js';
import { Dispatcher } from '../events.js';
import type { Marking } from '../marking.js';
import { DefinitionMetadataStore } from '../metadata.js';
import type { MarkingStore } from '../stores.js';
import type { Transition } from '../transition.js';
import type { MetadataStore, Sink, WorkflowContext } from '../types.js';
import type { WorkflowEngine } from './engine.js';
import { WorkflowLifecycleDispatcher } from './lifecycle-events.js';
import { MachineLogger } from './logger.js';
import { assertTransitionAllowed } from './transition-gate.js';
import { createTransitionPlan, resolveTransition } from './transition-plan.js';

export class Machine<T> implements WorkflowEngine<T> {
	readonly #name: string;
	readonly #definition: Definition;
	readonly #store: MarkingStore<T>;
	readonly #dispatcher: Dispatcher<T>;
	readonly #metadata: DefinitionMetadataStore;
	readonly #lifecycle: WorkflowLifecycleDispatcher<T>;
	readonly #logger = new MachineLogger();

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
		this.#lifecycle = new WorkflowLifecycleDispatcher(name, dispatcher);
	}

	public setLogger(logger: Sink): this {
		this.#logger.set(logger);

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
		this.#logger.debug('applying transition', 'transition', transitionName, 'workflow', this.#name);

		try {
			const { transition, next } = this.prepareApply(subject, transitionName, context);

			this.#lifecycle.dispatchLeaveEvents(subject, transition, next, context);
			this.#lifecycle.dispatchTransitionEvents(subject, transition, next, context);
			this.#lifecycle.dispatchEnterEvents(subject, transition, next, context);
			this.#store.setMarking(subject, next.clone(), this.#definition, { ...context });
			this.#logger.info('transition applied successfully', 'transition', transitionName, 'workflow', this.#name);
			this.#lifecycle.dispatchEnteredEvents(subject, transition, next, context);
			this.#lifecycle.dispatchCompletionEvents(subject, transition, next, context, this.enabledTransitions(subject));

			return next;
		} catch (error) {
			this.#logger.error('failed to apply transition', 'transition', transitionName, 'workflow', this.#name, 'error', error);
			throw error;
		}
	}

	private prepareApply(subject: T, transitionName: string, context: WorkflowContext): { transition: Transition; next: Marking } {
		const transition = resolveTransition(this.#definition, transitionName);
		const marking = this.getMarking(subject);

		this.transitionStateForMarking(subject, transition, marking, context);

		return createTransitionPlan(marking, transition);
	}

	private transitionState(subject: T, transitionName: string, context: WorkflowContext): void {
		this.transitionStateForMarking(subject, resolveTransition(this.#definition, transitionName), this.getMarking(subject), context);
	}

	private transitionStateForMarking(subject: T, transition: Transition, marking: Marking, context: WorkflowContext): void {
		assertTransitionAllowed({
			machine: this.#name,
			dispatcher: this.#dispatcher,
			subject,
			transition,
			marking,
			context,
		});
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
}
