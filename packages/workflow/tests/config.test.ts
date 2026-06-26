import { describe, expect, it } from 'vite-plus/test';

import { SingleStateStore, StateMachine, WorkflowConfigLoader } from '@alloy/workflow';

interface Subscription {
	state: string;
}

describe('workflow config loader', () => {
	it('builds a definition from the workflow schema', () => {
		const definition = new WorkflowConfigLoader({
			workflow: {
				places: ['trial', 'active', 'cancelled'],
				initial: ['trial'],
				transitions: [
					{ name: 'activate', from: ['trial'], to: ['active'] },
					{ name: 'cancel', from: ['active'], to: ['cancelled'] },
				],
				metadata: { purpose: 'billing' },
				transitions_metadata: { activate: { audit_level: 'high' } },
			},
		}).load();

		expect(definition.places).toHaveLength(3);
		expect(definition.metadataValue('purpose')).toBe('billing');
		expect(definition.transitionMetadataValue('activate', 'audit_level')).toBe('high');

		const subscription = { state: 'trial' };

		const machine = new StateMachine(
			'subscription',
			definition,
			new SingleStateStore<Subscription>(
				(subject) => subject.state,
				(subject, place) => {
					subject.state = place;
				},
			),
		);

		machine.apply(subscription, 'activate');

		expect(subscription.state).toBe('active');
	});

	it('supports custom roots and rejects invalid shapes', () => {
		const loader = new WorkflowConfigLoader({
			billing: {
				places: 'draft',
				initial: 'draft',
				transitions: [],
			},
		});

		expect(loader.loadAt('billing').places).toEqual(['draft']);
		expect(() => new WorkflowConfigLoader({ workflow: { initial: ['x'] } }).load()).toThrow('requires at least one place');
		expect(() => new WorkflowConfigLoader({ workflow: { places: ['x'] } }).load()).toThrow('requires an initial place list');
		expect(() => new WorkflowConfigLoader({ workflow: { places: ['x'], initial: ['x'], transitions: 'bad' } }).load()).toThrow('transitions must be a list');
	});

	it('supports dotted repository-style keys', () => {
		const definition = new WorkflowConfigLoader({
			'workflow.places': ['trial', 'active'],
			'workflow.initial': ['trial'],
			'workflow.transitions': [{ name: 'activate', from: ['trial'], to: ['active'] }],
			'workflow.metadata': { purpose: 'billing' },
		}).load();

		expect(definition.places).toEqual(['trial', 'active']);
		expect(definition.metadataValue('purpose')).toBe('billing');
	});
});
