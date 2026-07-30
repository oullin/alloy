import { describe, expect, it } from 'vite-plus/test';

import {
	AuditTrail,
	Definition,
	DefinitionBuilder,
	Dispatcher,
	EventNames,
	Marking,
	SingleStateStore,
	StateMachine,
	Transition,
	TransitionError,
	TransitionNotFoundError,
	WorkflowRegistry,
	WorkflowRegistryEntry,
	WorkflowValidator,
} from '@hara/sdk-workflow';

interface Subscription {
	id: string;
	state: string;
}

const subscriptionDefinition = (): Definition =>
	new DefinitionBuilder()
		.addPlace('trial')
		.addPlace('active')
		.addPlace('cancelled')
		.setInitialPlaces('trial')
		.addTransition('activate', ['trial'], ['active'])
		.addTransition('cancel', ['active'], ['cancelled'])
		.build();

const subscriptionStore = (): SingleStateStore<Subscription> =>
	new SingleStateStore(
		(subscription) => subscription.state,
		(subscription, place) => {
			subscription.state = place;
		},
	);

describe('workflow core', () => {
	it('mutates markings and reports active places deterministically', () => {
		const marking = Marking.fromPlaces('b', 'a', 'a');

		expect(marking.tokens('a')).toBe(2);
		expect(marking.has('b')).toBe(true);
		expect(marking.activePlaces()).toEqual(['a', 'b']);

		marking.remove('a').add('c');

		expect(marking.toJSON()).toEqual({ a: 1, b: 1, c: 1 });
	});

	it('reads definition metadata and validates invalid definitions', () => {
		const definition = new DefinitionBuilder().addPlace('trial').setInitialPlaces('trial').setMetadata('title', 'Subscription').setPlaceMetadata('trial', 'label', 'Free trial').build();

		expect(definition.metadataValue('title')).toBe('Subscription');
		expect(definition.hasMetadataValue('title')).toBe(true);
		expect(definition.placeMetadataValue('trial', 'label')).toBe('Free trial');

		expect(() =>
			new Definition({
				places: ['a'],
				initialMarking: Marking.fromPlaces('a'),
				transitions: [new Transition('go', ['a'], ['ghost'])],
			}).validate(),
		).toThrow('unknown place');

		expect(() => new Definition({ places: ['a'] }).validate()).toThrow('initial marking');
		expect(() => new Definition({ places: ['a', 'orphan'], initialMarking: Marking.fromPlaces('a') }).validate()).toThrow('unreachable');
	});

	it('applies state-machine transitions and dispatches lifecycle events', () => {
		const dispatcher = new Dispatcher<Subscription>();
		const order: string[] = [];

		dispatcher.on(EventNames.leave('subscription'), () => order.push('leave'));
		dispatcher.on(EventNames.transition('subscription'), () => order.push('transition'));
		dispatcher.on(EventNames.enter('subscription'), () => order.push('enter'));
		dispatcher.on(EventNames.entered('subscription'), () => order.push('entered'));
		dispatcher.on(EventNames.completed('subscription'), () => order.push('completed'));
		dispatcher.on(EventNames.announce('subscription'), () => order.push('announce'));

		const machine = new StateMachine('subscription', subscriptionDefinition(), subscriptionStore(), dispatcher);
		const subscription = { id: 's1', state: 'trial' };
		const marking = machine.apply(subscription, 'activate', { reason: 'trial-end' });

		expect(marking.has('active')).toBe(true);
		expect(subscription.state).toBe('active');
		expect(order).toEqual(['leave', 'transition', 'enter', 'entered', 'completed', 'announce']);
	});

	it('returns guard veto errors without mutating the subject', () => {
		const dispatcher = new Dispatcher<Subscription>();

		dispatcher.onGuard(EventNames.guardNamed('subscription', 'activate'), (event) => event.setBlocked(true, 'billing not configured'));

		const machine = new StateMachine('subscription', subscriptionDefinition(), subscriptionStore(), dispatcher);
		const subscription = { id: 's2', state: 'trial' };

		expect(() => machine.apply(subscription, 'activate')).toThrow(TransitionError);
		expect(subscription.state).toBe('trial');

		try {
			machine.apply(subscription, 'activate');
		} catch (error) {
			expect(error).toBeInstanceOf(TransitionError);
			expect((error as TransitionError).blockers[0]?.message).toBe('billing not configured');
		}
	});

	it('reports unknown transitions and enabled/disabled transitions', () => {
		const machine = new StateMachine('subscription', subscriptionDefinition(), subscriptionStore());
		const subscription = { id: 's3', state: 'trial' };

		expect(() => machine.apply(subscription, 'bogus')).toThrow(TransitionNotFoundError);
		expect(machine.enabledTransitions(subscription).map((transition) => transition.name)).toEqual(['activate']);
		expect(machine.disabledTransitions(subscription).map((transition) => transition.name)).toEqual(['cancel']);
	});

	it('looks up workflows from registry and records audit entries', () => {
		const dispatcher = new Dispatcher<Subscription>();
		const trail = new AuditTrail<Subscription>().attach('subscription', dispatcher);
		const machine = new StateMachine('subscription', subscriptionDefinition(), subscriptionStore(), dispatcher);
		const registry = new WorkflowRegistry<Subscription>().add(new WorkflowRegistryEntry({ name: 'subscription', machine, supports: (subject) => subject.id !== '' }));
		const subscription = { id: 's4', state: 'trial' };

		expect(registry.get(subscription, 'subscription').name()).toBe('subscription');

		machine.apply(subscription, 'activate', { reason: 'trial-end' });

		expect(trail.entries).toHaveLength(1);
		expect(trail.entries[0]?.transition).toBe('activate');
		expect(trail.entries[0]?.context.reason).toBe('trial-end');
	});

	it('validates definitions through the validator subpath class', async () => {
		const { WorkflowValidator: SubpathValidator } = await import('@hara/sdk-workflow/validator');

		const validator = new WorkflowValidator();
		const subpathValidator = new SubpathValidator();
		const definition = subscriptionDefinition();

		expect(() => validator.validateDefinition(definition)).not.toThrow();
		expect(() => subpathValidator.validateStateMachine(definition)).not.toThrow();
		expect(() => validator.validateDefinition(undefined)).toThrow('definition is required');
	});
});
