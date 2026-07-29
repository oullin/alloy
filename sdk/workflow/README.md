# Workflow TypeScript

The TypeScript package lives in `sdk/workflow` and exposes
`@hara/sdk-workflow`. It implements workflow-net and state-machine engines in
the style of Symfony Workflow: a `Definition` describes places and
transitions, a `Machine` (or single-state `StateMachine`) applies
transitions to your own subjects, and a `MarkingStore` maps the current
marking onto whatever field your subject already has. Transitions dispatch
guard/leave/transition/enter lifecycle events, and companions cover audit
trails, config loading, registries, and multi-step job graphs.

This is a private workspace package: it is consumed by sibling packages via
`workspace:*` and is never published to npm.

Definitions are assembled with the builder and driven by a machine:

```ts
import { DefinitionBuilder, SingleStateStore, StateMachine } from '@hara/sdk-workflow';

interface Subscription {
	id: string;
	state: string;
}

const definition = new DefinitionBuilder()
	.addPlace('trial')
	.addPlace('active')
	.addPlace('cancelled')
	.setInitialPlaces('trial')
	.addTransition('activate', ['trial'], ['active'])
	.addTransition('cancel', ['active'], ['cancelled'])
	.build();

const store = new SingleStateStore<Subscription>(
	(subscription) => subscription.state,
	(subscription, place) => {
		subscription.state = place;
	},
);

const machine = new StateMachine('subscription', definition, store);
const subscription: Subscription = { id: 's-1', state: '' };

machine.can(subscription, 'activate'); // true
machine.apply(subscription, 'activate');
subscription.state; // "active"
machine.enabledTransitions(subscription).map((transition) => transition.name); // ["cancel"]
```

Guards and lifecycle hooks attach through the event dispatcher:

```ts
import { Dispatcher, EventNames } from '@hara/sdk-workflow';

const dispatcher = new Dispatcher<Subscription>();

dispatcher.on(EventNames.guardNamed('subscription', 'cancel'), (event) => {
	// event.addTransitionBlocker(...) to veto the transition
});

const guarded = new StateMachine('subscription', definition, store, dispatcher);
```

## API overview

| Entry point                     | Main exports                                                                          | Purpose                                  |
| ------------------------------- | ------------------------------------------------------------------------------------- | ---------------------------------------- |
| `@hara/sdk-workflow`            | `Machine`, `StateMachine`, `Definition`, `DefinitionBuilder`, `Marking`, `Transition` | core engine, definitions, markings       |
| `@hara/sdk-workflow`            | `Dispatcher`, `EventNames`, `TransitionError`, `TransitionNotFoundError`              | lifecycle events and coded failures      |
| `@hara/sdk-workflow/stores`     | `SingleStateStore`, `MultiStateStore`, `MarkingStore`                                 | persisting markings on subjects          |
| `@hara/sdk-workflow/events`     | `GuardEvent`, `TransitionEvent`, `EnteredEvent`, ...                                  | typed lifecycle event classes            |
| `@hara/sdk-workflow/audit`      | `AuditTrail`, `AuditEntry`                                                            | recording applied transitions            |
| `@hara/sdk-workflow/registry`   | `WorkflowRegistry`, `WorkflowRegistryEntry`                                           | resolving machines per subject           |
| `@hara/sdk-workflow/config`     | `WorkflowConfigLoader`                                                                | building definitions from config sources |
| `@hara/sdk-workflow/validator`  | `WorkflowValidator`                                                                   | definition validation rules              |
| `@hara/sdk-workflow/multisteps` | `MultiStepWorkflow`, job/engine helpers                                               | dependency-graph job execution           |

Acceptance tests live in `sdk/workflow/tests` and run with
`pnpm --filter @hara/sdk-workflow test`.
