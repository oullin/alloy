import { describe, expect, it } from 'vite-plus/test';

import {
	AuditTrail,
	AnnounceEvent,
	Definition,
	DefinitionBuilder,
	Dispatcher,
	EventNames,
	Marking,
	SingleStateStore,
	StateMachine,
	Transition,
	TransitionBlocker,
	TransitionError,
	TransitionNotFoundError,
	WorkflowConfigLoader,
} from '@alloy/workflow';

interface Ticket {
	state: string;
}

const store = (): SingleStateStore<Ticket> =>
	new SingleStateStore(
		(ticket) => ticket.state,
		(ticket, place) => {
			ticket.state = place;
		},
	);

const definition = (): Definition =>
	new DefinitionBuilder()
		.addPlace('draft')
		.addPlace('review')
		.addPlace('done')
		.setInitialPlaces('draft')
		.addTransition('submit', ['draft'], ['review'])
		.addTransition('approve', ['review'], ['done'])
		.build();

describe('workflow guards and errors', () => {
	it('collects transition blockers and preserves typed error classes', () => {
		const dispatcher = new Dispatcher<Ticket>();

		dispatcher.onGuard(EventNames.guard('ticket'), (event) => event.addTransitionBlocker(new TransitionBlocker('account disabled', 'disabled')));
		dispatcher.onGuard(EventNames.guardNamed('ticket', 'submit'), (event) => event.addTransitionBlocker({ message: 'missing approval', code: 'approval' }));

		const machine = new StateMachine('ticket', definition(), store(), dispatcher);
		const ticket = { state: 'draft' };

		try {
			machine.apply(ticket, 'submit');
			throw new Error('expected transition to be blocked');
		} catch (error) {
			expect(error).toBeInstanceOf(TransitionError);
			expect(error).not.toBeInstanceOf(TransitionNotFoundError);
			expect((error as TransitionError).blockers).toHaveLength(2);
			expect((error as TransitionError).blockers[0]).toBeInstanceOf(TransitionBlocker);
			expect((error as TransitionError).blockers.map((blocker) => blocker.code)).toEqual(['disabled', 'approval']);
		}

		expect(ticket.state).toBe('draft');
	});

	it('reports transition not found distinctly from disabled transitions', () => {
		const machine = new StateMachine('ticket', definition(), store());
		const ticket = { state: 'draft' };

		expect(() => machine.apply(ticket, 'missing')).toThrow(TransitionNotFoundError);
		expect(() => machine.apply(ticket, 'approve')).toThrow(TransitionError);
		expect(machine.can(ticket, 'submit')).toBe(true);
		expect(machine.cannot(ticket, 'approve')).toBe(true);
	});

	it('emits lifecycle events in state-store order and records completed audit entries', () => {
		const dispatcher = new Dispatcher<Ticket>();
		const order: string[] = [];
		const trail = new AuditTrail<Ticket>().attach('ticket', dispatcher);

		dispatcher.on(EventNames.leave('ticket'), (event) => order.push(`leave:${'place' in event ? event.place : ''}:${event.marking().draft ?? 0}`));
		dispatcher.on(EventNames.transition('ticket'), (event) => order.push(`transition:${event.transition().name}`));
		dispatcher.on(EventNames.enter('ticket'), (event) => order.push(`enter:${'place' in event ? event.place : ''}`));
		dispatcher.on(EventNames.entered('ticket'), (event) => order.push(`entered:${'place' in event ? event.place : ''}`));
		dispatcher.on(EventNames.completed('ticket'), (event) => order.push(`completed:${event.marking().review ?? 0}`));
		dispatcher.on(EventNames.announce('ticket'), (event) => order.push(`announce:${event instanceof AnnounceEvent ? event.enabled.map((transition) => transition.name).join(',') : ''}`));

		const machine = new StateMachine('ticket', definition(), store(), dispatcher);
		const ticket = { state: 'draft' };

		machine.apply(ticket, 'submit', { actor: 'ada' });

		expect(order).toEqual(['leave:draft:1', 'transition:submit', 'enter:review', 'entered:review', 'completed:1', 'announce:approve']);
		expect(trail.entries).toHaveLength(1);
		expect(trail.entries[0]?.context.actor).toBe('ada');
		expect(trail.entries[0]?.marking.activePlaces()).toEqual(['review']);
	});

	it('rejects invalid definitions and config shapes with clear validation failures', () => {
		expect(() => new StateMachine('ticket', new Definition({ places: ['draft'], initialMarking: Marking.fromPlaces('ghost') }), store())).toThrow('unknown place');
		expect(() => new Definition({ places: ['draft', 'review'], initialMarking: Marking.fromPlaces('draft'), transitions: [new Transition('', ['draft'], ['review'])] }).validate()).toThrow(
			'transition name cannot be empty',
		);
		expect(() => new Definition({ places: ['draft'], initialMarking: Marking.fromPlaces('draft'), transitions: [new Transition('submit', [], ['draft'])] }).validate()).toThrow(
			'requires at least one from place',
		);
		expect(() => new WorkflowConfigLoader({ workflow: { places: ['draft'], initial: 1 } }).load()).toThrow('requires an initial place list');
		expect(() => new WorkflowConfigLoader({ workflow: { places: ['draft'], initial: ['draft'], transitions: [{ from: ['draft'], to: ['draft'] }] } }).load()).toThrow('transition name');
	});
});
