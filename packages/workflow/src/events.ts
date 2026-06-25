import { TransitionBlocker, type TransitionBlockerShape } from './errors.js';
import { type TransitionSnapshot } from './transition.js';
import { cloneRecord, type WorkflowContext } from './types.js';

export { EventNames } from './event-names.js';

export type Listener<T> = (event: WorkflowEvent<T>) => void;

export interface WorkflowEvent<T> {
	workflowName(): string;
	subject(): T;
	transition(): TransitionSnapshot;
	marking(): Record<string, number>;
	context(): WorkflowContext;
}

export class BaseEvent<T> implements WorkflowEvent<T> {
	readonly #machine: string;
	readonly #subject: T;
	readonly #transition: TransitionSnapshot;
	readonly #tokens: Record<string, number>;
	readonly #context: WorkflowContext;

	public constructor(machine: string, subject: T, transition: TransitionSnapshot, tokens: Record<string, number>, context: WorkflowContext = {}) {
		this.#machine = machine;
		this.#subject = subject;
		this.#transition = {
			name: transition.name,
			from: [...transition.from],
			to: [...transition.to],
		};
		this.#tokens = { ...tokens };
		this.#context = cloneRecord(context);
	}

	public workflowName(): string {
		return this.#machine;
	}

	public subject(): T {
		return this.#subject;
	}

	public transition(): TransitionSnapshot {
		return {
			name: this.#transition.name,
			from: [...this.#transition.from],
			to: [...this.#transition.to],
		};
	}

	public marking(): Record<string, number> {
		return { ...this.#tokens };
	}

	public context(): WorkflowContext {
		return cloneRecord(this.#context);
	}
}

export class GuardEvent<T> extends BaseEvent<T> {
	#blocked = false;
	readonly #blockers: TransitionBlocker[] = [];

	public constructor(base: BaseEvent<T>) {
		super(base.workflowName(), base.subject(), base.transition(), base.marking(), base.context());
	}

	public setBlocked(blocked: boolean, message = ''): this {
		this.#blocked = blocked;

		if (blocked && message !== '') {
			this.#blockers.push(new TransitionBlocker(message));
		}

		return this;
	}

	public addTransitionBlocker(blocker: TransitionBlocker | TransitionBlockerShape): this {
		this.#blocked = true;
		this.#blockers.push(TransitionBlocker.from(blocker));

		return this;
	}

	public blocked(): boolean {
		return this.#blocked;
	}

	public blockers(): TransitionBlocker[] {
		return this.#blockers.map((blocker) => TransitionBlocker.from(blocker));
	}
}

export class LeaveEvent<T> extends BaseEvent<T> {
	public readonly place: string;

	public constructor(base: BaseEvent<T>, place: string) {
		super(base.workflowName(), base.subject(), base.transition(), base.marking(), base.context());
		this.place = place;
	}
}

export class TransitionEvent<T> extends BaseEvent<T> {
	public constructor(base: BaseEvent<T>) {
		super(base.workflowName(), base.subject(), base.transition(), base.marking(), base.context());
	}
}

export class EnterEvent<T> extends BaseEvent<T> {
	public readonly place: string;

	public constructor(base: BaseEvent<T>, place: string) {
		super(base.workflowName(), base.subject(), base.transition(), base.marking(), base.context());
		this.place = place;
	}
}

export class EnteredEvent<T> extends BaseEvent<T> {
	public readonly place: string;

	public constructor(base: BaseEvent<T>, place: string) {
		super(base.workflowName(), base.subject(), base.transition(), base.marking(), base.context());
		this.place = place;
	}
}

export class CompletedEvent<T> extends BaseEvent<T> {
	public constructor(base: BaseEvent<T>) {
		super(base.workflowName(), base.subject(), base.transition(), base.marking(), base.context());
	}
}

export class AnnounceEvent<T> extends BaseEvent<T> {
	public readonly enabled: TransitionSnapshot[];

	public constructor(base: BaseEvent<T>, enabled: TransitionSnapshot[]) {
		super(base.workflowName(), base.subject(), base.transition(), base.marking(), base.context());
		this.enabled = enabled.map((transition) => ({
			name: transition.name,
			from: [...transition.from],
			to: [...transition.to],
		}));
	}
}

export class Dispatcher<T> {
	readonly #listeners = new Map<string, Listener<T>[]>();

	public on(name: string, listener: Listener<T>): this {
		const listeners = this.#listeners.get(name) ?? [];

		listeners.push(listener);
		this.#listeners.set(name, listeners);

		return this;
	}

	public onGuard(name: string, listener: (event: GuardEvent<T>) => void): this {
		return this.on(name, (event) => {
			if (event instanceof GuardEvent) {
				listener(event);
			}
		});
	}

	public dispatch(name: string, event: WorkflowEvent<T>): void {
		for (const listener of this.#listeners.get(name) ?? []) {
			listener(event);
		}
	}
}
