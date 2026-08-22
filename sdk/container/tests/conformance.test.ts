import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vite-plus/test';
import type { Container } from '@hara/sdk-container';

import {
	AliasCycleError,
	Application,
	CircularResolutionError,
	MissingBindingError,
	MissingMethodBindingError,
	ProviderCycleError,
	SelfAliasError,
	createServiceToken,
	type ContainerErrorCode,
	type ResolutionParameters,
	type ServiceToken,
} from '@hara/sdk-container';

type Scalar = string | number | boolean | null;

type Primitive = 'constant' | 'increment-counter' | 'resolve-token' | 'read-parameter' | 'append-suffix' | 'return-instance';

type OperationKind =
	| 'bind'
	| 'bind-if'
	| 'singleton-if'
	| 'scoped-if'
	| 'instance'
	| 'resolve'
	| 'resolve-with-parameters'
	| 'get'
	| 'forget-scoped'
	| 'forget-instance'
	| 'flush'
	| 'alias'
	| 'contextual-value'
	| 'contextual-factory'
	| 'contextual-tagged'
	| 'tag'
	| 'tagged'
	| 'extend'
	| 'callback'
	| 'rebinding'
	| 'method-bind'
	| 'method-call'
	| 'call'
	| 'wrap'
	| 'factory-func'
	| 'provider-register'
	| 'provider-register-many'
	| 'provider-boot'
	| 'observe-counter'
	| 'observe-events'
	| 'observe-bound'
	| 'observe-has'
	| 'observe-resolved'
	| 'observe-is-shared'
	| 'observe-bindings'
	| 'observe-providers'
	| 'observe-has-provider'
	| 'observe-provider-for'
	| 'observe-booted'
	| 'observe-has-method';

type TokenSpec = { readonly id: string; readonly kind?: 'string' };

type ProviderSpec = {
	readonly id: string;
	readonly provides: readonly string[];
	readonly dependsOn: readonly string[];
	readonly deferred: boolean;
	readonly registerEvent?: string;
	readonly bootEvent?: string;
	readonly registerValue?: { readonly token: string; readonly value: Scalar };
	readonly registerResolve?: string;
};

type Operation = {
	readonly kind: OperationKind;
	readonly token?: string;
	readonly target?: string;
	readonly alias?: string;
	readonly primitive?: Primitive;
	readonly lifetime?: 'transient' | 'singleton' | 'scoped';
	readonly counter?: string;
	readonly parameter?: string;
	readonly value?: Scalar;
	readonly suffix?: string;
	readonly parameters?: Readonly<Record<string, Scalar>>;
	readonly observe?: string;
	readonly consumer?: string;
	readonly consumers?: readonly string[];
	readonly needs?: string;
	readonly tokens?: readonly string[];
	readonly tag?: string;
	readonly phase?: 'before' | 'resolving' | 'after';
	readonly event?: string;
	readonly method?: string;
	readonly instance?: Scalar;
	readonly provider?: string;
	readonly providers?: readonly string[];
};

type FixtureCase = {
	readonly id: string;
	readonly note: string;
	readonly tokens: readonly TokenSpec[];
	readonly providers: readonly ProviderSpec[];
	readonly operations: readonly Operation[];
	readonly expected?: readonly string[];
	readonly error?: ContainerErrorCode;
};

type Fixture = { readonly schemaVersion: 1; readonly cases: readonly FixtureCase[] };

const primitives = new Set<Primitive>(['constant', 'increment-counter', 'resolve-token', 'read-parameter', 'append-suffix', 'return-instance']);

const kinds = new Set<OperationKind>([
	'bind',
	'bind-if',
	'singleton-if',
	'scoped-if',
	'instance',
	'resolve',
	'resolve-with-parameters',
	'get',
	'forget-scoped',
	'forget-instance',
	'flush',
	'alias',
	'contextual-value',
	'contextual-factory',
	'contextual-tagged',
	'tag',
	'tagged',
	'extend',
	'callback',
	'rebinding',
	'method-bind',
	'method-call',
	'call',
	'wrap',
	'factory-func',
	'provider-register',
	'provider-register-many',
	'provider-boot',
	'observe-counter',
	'observe-events',
	'observe-bound',
	'observe-has',
	'observe-resolved',
	'observe-is-shared',
	'observe-bindings',
	'observe-providers',
	'observe-has-provider',
	'observe-provider-for',
	'observe-booted',
	'observe-has-method',
]);

const errors = new Set<ContainerErrorCode>(['ALIAS_CYCLE', 'MISSING_BINDING', 'CIRCULAR_RESOLUTION', 'SELF_ALIAS', 'MISSING_METHOD_BINDING', 'PROVIDER_CYCLE']);

const object = (value: unknown, where: string): Readonly<Record<string, unknown>> => {
	if (typeof value !== 'object' || value === null || Array.isArray(value)) {
		throw new TypeError(`${where} must be an object`);
	}

	return value as Readonly<Record<string, unknown>>;
};

const string = (value: unknown, where: string): string => {
	if (typeof value !== 'string' || value.trim() === '') {
		throw new TypeError(`${where} must be a non-empty string`);
	}

	return value;
};

const scalar = (value: unknown, where: string): Scalar => {
	if (value === null || typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
		return value;
	}

	throw new TypeError(`${where} must be a scalar`);
};

const strings = (value: unknown, where: string): readonly string[] => {
	if (!Array.isArray(value)) {
		throw new TypeError(`${where} must be an array`);
	}

	return value.map((entry, index) => string(entry, `${where}[${index}]`));
};

const optionalString = (record: Readonly<Record<string, unknown>>, key: string, where: string): string | undefined =>
	record[key] === undefined ? undefined : string(record[key], `${where}.${key}`);

const scalarMap = (value: unknown, where: string): Readonly<Record<string, Scalar>> => {
	const record = object(value, where);

	return Object.fromEntries(Object.entries(record).map(([key, entry]) => [key, scalar(entry, `${where}.${key}`)]));
};

const requireToken = (token: string | undefined, tokenIds: ReadonlySet<string>, where: string): string => {
	const value = string(token, where);

	if (!tokenIds.has(value)) {
		throw new TypeError(`${where} references unknown token "${value}"`);
	}

	return value;
};

/** Parse and validate the language-neutral fixture before either runtime executes it. */
export const parseContainerFixture = (input: unknown): Fixture => {
	const root = object(input, 'fixture');

	if (root.schemaVersion !== 1) {
		throw new TypeError('fixture.schemaVersion must be 1');
	}

	if (!Array.isArray(root.cases) || root.cases.length === 0) {
		throw new TypeError('fixture.cases must be a non-empty array');
	}

	const ids = new Set<string>();

	const cases = root.cases.map((rawCase, caseIndex): FixtureCase => {
		const where = `fixture.cases[${caseIndex}]`;
		const record = object(rawCase, where);
		const id = string(record.id, `${where}.id`);

		if (ids.has(id)) {
			throw new TypeError(`${where}.id duplicates "${id}"`);
		}

		ids.add(id);

		const note = string(record.note, `${where}.note`);

		if (!Array.isArray(record.tokens)) {
			throw new TypeError(`${where}.tokens must be an array`);
		}

		const tokenIds = new Set<string>();

		const tokens = record.tokens.map((raw, index): TokenSpec => {
			const token = object(raw, `${where}.tokens[${index}]`);
			const tokenId = string(token.id, `${where}.tokens[${index}].id`);

			if (tokenIds.has(tokenId)) {
				throw new TypeError(`${where}.tokens duplicates "${tokenId}"`);
			}

			tokenIds.add(tokenId);

			if (token.kind !== undefined && token.kind !== 'string') {
				throw new TypeError(`${where}.tokens[${index}].kind is unknown`);
			}

			return { id: tokenId, ...(token.kind === 'string' ? { kind: 'string' as const } : {}) };
		});

		const providerIds = new Set<string>();

		const providers = (record.providers === undefined ? [] : stringsOrObjects(record.providers, `${where}.providers`)).map((raw, index): ProviderSpec => {
			const provider = object(raw, `${where}.providers[${index}]`);
			const providerId = string(provider.id, `${where}.providers[${index}].id`);

			if (providerIds.has(providerId)) {
				throw new TypeError(`${where}.providers duplicates "${providerId}"`);
			}

			providerIds.add(providerId);

			const provides = provider.provides === undefined ? [] : strings(provider.provides, `${where}.providers[${index}].provides`);
			const dependsOn = provider.dependsOn === undefined ? [] : strings(provider.dependsOn, `${where}.providers[${index}].dependsOn`);

			for (const token of [...provides, ...dependsOn]) {
				requireToken(token, tokenIds, `${where}.providers[${index}]`);
			}

			let registerValue: ProviderSpec['registerValue'];

			if (provider.registerValue !== undefined) {
				const value = object(provider.registerValue, `${where}.providers[${index}].registerValue`);

				registerValue = {
					token: requireToken(value.token as string | undefined, tokenIds, `${where}.providers[${index}].registerValue.token`),
					value: scalar(value.value, `${where}.providers[${index}].registerValue.value`),
				};
			}

			const registerResolve = provider.registerResolve === undefined ? undefined : requireToken(provider.registerResolve as string, tokenIds, `${where}.providers[${index}].registerResolve`);

			return {
				id: providerId,
				provides,
				dependsOn,
				deferred: provider.deferred === true,
				...(optionalString(provider, 'registerEvent', `${where}.providers[${index}]`) === undefined
					? {}
					: { registerEvent: optionalString(provider, 'registerEvent', `${where}.providers[${index}]`) }),
				...(optionalString(provider, 'bootEvent', `${where}.providers[${index}]`) === undefined ? {} : { bootEvent: optionalString(provider, 'bootEvent', `${where}.providers[${index}]`) }),
				...(registerValue === undefined ? {} : { registerValue }),
				...(registerResolve === undefined ? {} : { registerResolve }),
			};
		});

		if (!Array.isArray(record.operations)) {
			throw new TypeError(`${where}.operations must be an array`);
		}

		const operations = record.operations.map((raw, index): Operation => parseOperation(raw, `${where}.operations[${index}]`, tokenIds, providerIds));
		const hasExpected = record.expected !== undefined;
		const hasError = record.error !== undefined;

		if (hasExpected === hasError) {
			throw new TypeError(`${where} must declare exactly one of expected or error`);
		}

		if (hasExpected && !Array.isArray(record.expected)) {
			throw new TypeError(`${where}.expected must be an array`);
		}

		if (hasError && !errors.has(string(record.error, `${where}.error`) as ContainerErrorCode)) {
			throw new TypeError(`${where}.error is unknown`);
		}

		return {
			id,
			note,
			tokens,
			providers,
			operations,
			...(hasExpected
				? { expected: (record.expected as readonly unknown[]).map((entry, index) => string(entry, `${where}.expected[${index}]`)) }
				: { error: string(record.error, `${where}.error`) as ContainerErrorCode }),
		};
	});

	return { schemaVersion: 1, cases };
};

const stringsOrObjects = (value: unknown, where: string): readonly unknown[] => {
	if (!Array.isArray(value)) {
		throw new TypeError(`${where} must be an array`);
	}

	return value;
};

const parseOperation = (raw: unknown, where: string, tokenIds: ReadonlySet<string>, providerIds: ReadonlySet<string>): Operation => {
	const record = object(raw, where);
	const kind = string(record.kind, `${where}.kind`) as OperationKind;

	if (!kinds.has(kind)) {
		throw new TypeError(`${where}.kind is unknown`);
	}

	const token = record.token === undefined ? undefined : requireToken(record.token as string, tokenIds, `${where}.token`);
	const target = record.target === undefined ? undefined : requireToken(record.target as string, tokenIds, `${where}.target`);
	const alias = record.alias === undefined ? undefined : requireToken(record.alias as string, tokenIds, `${where}.alias`);
	const consumer = record.consumer === undefined ? undefined : requireToken(record.consumer as string, tokenIds, `${where}.consumer`);

	const consumers = record.consumers === undefined ? undefined : strings(record.consumers, `${where}.consumers`).map((entry) => requireToken(entry, tokenIds, `${where}.consumers`));

	const needs = record.needs === undefined ? undefined : requireToken(record.needs as string, tokenIds, `${where}.needs`);
	const primitive = record.primitive === undefined ? undefined : (string(record.primitive, `${where}.primitive`) as Primitive);

	if (primitive !== undefined && !primitives.has(primitive)) {
		throw new TypeError(`${where}.primitive is unknown`);
	}

	const tokens = record.tokens === undefined ? undefined : strings(record.tokens, `${where}.tokens`).map((entry) => requireToken(entry, tokenIds, `${where}.tokens`));

	const providers =
		record.providers === undefined
			? undefined
			: strings(record.providers, `${where}.providers`).map((entry) => {
					if (!providerIds.has(entry)) {
						throw new TypeError(`${where}.providers references unknown provider "${entry}"`);
					}

					return entry;
				});

	const provider = record.provider === undefined ? undefined : string(record.provider, `${where}.provider`);

	if (provider !== undefined && !providerIds.has(provider)) {
		throw new TypeError(`${where}.provider references unknown provider "${provider}"`);
	}

	const lifetime = record.lifetime === undefined ? undefined : string(record.lifetime, `${where}.lifetime`);

	if (lifetime !== undefined && lifetime !== 'transient' && lifetime !== 'singleton' && lifetime !== 'scoped') {
		throw new TypeError(`${where}.lifetime is unknown`);
	}

	const phase = record.phase === undefined ? undefined : string(record.phase, `${where}.phase`);

	if (phase !== undefined && phase !== 'before' && phase !== 'resolving' && phase !== 'after') {
		throw new TypeError(`${where}.phase is unknown`);
	}

	if (
		(kind === 'bind' ||
			kind === 'bind-if' ||
			kind === 'singleton-if' ||
			kind === 'scoped-if' ||
			kind === 'extend' ||
			kind === 'method-bind' ||
			kind === 'contextual-factory' ||
			kind === 'call' ||
			kind === 'wrap') &&
		primitive === undefined
	) {
		throw new TypeError(`${where} requires primitive`);
	}

	if (primitive === 'constant' && record.value === undefined) {
		throw new TypeError(`${where}.constant requires value`);
	}

	if (primitive === 'increment-counter' && record.counter === undefined) {
		throw new TypeError(`${where}.increment-counter requires counter`);
	}

	if (primitive === 'resolve-token' && target === undefined) {
		throw new TypeError(`${where}.resolve-token requires target`);
	}

	if (primitive === 'read-parameter' && record.parameter === undefined) {
		throw new TypeError(`${where}.read-parameter requires parameter`);
	}

	if (primitive === 'append-suffix' && record.suffix === undefined) {
		throw new TypeError(`${where}.append-suffix requires suffix`);
	}

	if ((kind === 'contextual-value' || kind === 'contextual-factory' || kind === 'contextual-tagged') && consumer !== undefined && consumers !== undefined) {
		throw new TypeError(`${where} has ambiguous consumer-versus-consumers representation`);
	}

	if ((kind === 'contextual-value' || kind === 'contextual-factory' || kind === 'contextual-tagged') && consumer === undefined && consumers === undefined) {
		throw new TypeError(`${where} requires consumer`);
	}

	if (kind === 'contextual-value' && record.value === undefined) {
		throw new TypeError(`${where} requires value`);
	}

	if (kind === 'contextual-value' && primitive !== undefined) {
		throw new TypeError(`${where} has ambiguous value-versus-factory representation`);
	}

	for (const required of requiredFields(kind)) {
		if (record[required] === undefined) {
			throw new TypeError(`${where} requires ${required}`);
		}
	}

	return {
		kind,
		...(token === undefined ? {} : { token }),
		...(target === undefined ? {} : { target }),
		...(alias === undefined ? {} : { alias }),
		...(primitive === undefined ? {} : { primitive }),
		...(lifetime === undefined ? {} : { lifetime: lifetime as 'transient' | 'singleton' | 'scoped' }),
		...(record.counter === undefined ? {} : { counter: string(record.counter, `${where}.counter`) }),
		...(record.parameter === undefined ? {} : { parameter: string(record.parameter, `${where}.parameter`) }),
		...(record.value === undefined ? {} : { value: scalar(record.value, `${where}.value`) }),
		...(record.suffix === undefined ? {} : { suffix: string(record.suffix, `${where}.suffix`) }),
		...(record.parameters === undefined ? {} : { parameters: scalarMap(record.parameters, `${where}.parameters`) }),
		...(record.observe === undefined ? {} : { observe: string(record.observe, `${where}.observe`) }),
		...(consumer === undefined ? {} : { consumer }),
		...(consumers === undefined ? {} : { consumers }),
		...(needs === undefined ? {} : { needs }),
		...(tokens === undefined ? {} : { tokens }),
		...(record.tag === undefined ? {} : { tag: string(record.tag, `${where}.tag`) }),
		...(phase === undefined ? {} : { phase: phase as 'before' | 'resolving' | 'after' }),
		...(record.event === undefined ? {} : { event: string(record.event, `${where}.event`) }),
		...(record.method === undefined ? {} : { method: string(record.method, `${where}.method`) }),
		...(record.instance === undefined ? {} : { instance: scalar(record.instance, `${where}.instance`) }),
		...(provider === undefined ? {} : { provider }),
		...(providers === undefined ? {} : { providers }),
	};
};

const requiredFields = (kind: OperationKind): readonly string[] =>
	({
		bind: ['token'],
		'bind-if': ['token'],
		'singleton-if': ['token'],
		'scoped-if': ['token'],
		instance: ['token', 'value'],
		resolve: ['token'],
		'resolve-with-parameters': ['token', 'parameters'],
		get: ['token'],
		'forget-scoped': [],
		'forget-instance': ['token'],
		flush: [],
		alias: ['target', 'alias'],
		'contextual-value': ['needs', 'value'],
		'contextual-factory': ['needs'],
		'contextual-tagged': ['needs', 'tag'],
		tag: ['tag', 'tokens'],
		tagged: ['tag'],
		extend: ['token'],
		callback: ['phase', 'event'],
		rebinding: ['token', 'event'],
		'method-bind': ['method'],
		'method-call': ['method', 'instance'],
		call: [],
		wrap: [],
		'factory-func': ['token'],
		'provider-register': ['provider'],
		'provider-register-many': ['providers'],
		'provider-boot': [],
		'observe-counter': ['counter'],
		'observe-events': [],
		'observe-bound': ['token'],
		'observe-has': ['token'],
		'observe-resolved': ['token'],
		'observe-is-shared': ['token'],
		'observe-bindings': [],
		'observe-providers': [],
		'observe-has-provider': ['token'],
		'observe-provider-for': ['token'],
		'observe-booted': [],
		'observe-has-method': ['method'],
	})[kind];

const fixturePath = new URL('../../../conformance/container.json', import.meta.url);

const fixture = parseContainerFixture(JSON.parse(readFileSync(fixturePath, 'utf8')) as unknown);

const render = (value: unknown): string => {
	if (value === undefined || value === null) {
		return '<nil>';
	}

	if (Array.isArray(value)) {
		return value.map(render).join(',');
	}

	if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
		return String(value);
	}

	return JSON.stringify(value);
};

const token = (tokens: ReadonlyMap<string, ServiceToken<unknown>>, id: string): ServiceToken<unknown> => {
	const value = tokens.get(id);

	if (value === undefined) {
		throw new Error(`fixture token ${id} was not created`);
	}

	return value;
};

const consumersOf = (operation: Operation, tokens: ReadonlyMap<string, ServiceToken<unknown>>): readonly ServiceToken<unknown>[] => {
	if (operation.consumers !== undefined) {
		return operation.consumers.map((id) => token(tokens, id));
	}

	return [token(tokens, operation.consumer ?? '')];
};

const runContainerCase = (testCase: FixtureCase): readonly string[] => {
	const tokens = new Map(testCase.tokens.map((spec) => [spec.id, createServiceToken<unknown>(spec.id)]));
	const counters = new Map<string, number>();
	const events: string[] = [];
	const observations: string[] = [];
	const subject = new Application();

	const runPrimitive = (operation: Operation, resolved: Container, parameters: ResolutionParameters): unknown => {
		switch (operation.primitive) {
			case 'constant':
				return operation.value;

			case 'increment-counter': {
				const next = (counters.get(operation.counter ?? '') ?? 0) + 1;

				counters.set(operation.counter ?? '', next);

				return next;
			}

			case 'resolve-token':
				return resolved.make(token(tokens, operation.target ?? ''));

			case 'read-parameter':
				return parameters[operation.parameter ?? ''] ?? '<nil>';

			case 'return-instance':
				return parameters._instance;

			default:
				throw new Error(`primitive ${operation.primitive ?? ''} cannot construct a factory`);
		}
	};

	const factory =
		(operation: Operation) =>
		(resolved: Container, parameters: ResolutionParameters): unknown =>
			runPrimitive(operation, resolved, parameters);

	const extender =
		(operation: Operation) =>
		(value: unknown): unknown =>
			operation.primitive === 'append-suffix' ? `${render(value)}${operation.suffix ?? ''}` : value;

	const providers = new Map(
		testCase.providers.map((spec) => [
			spec.id,
			{
				provides: spec.provides.map((id) => token(tokens, id)),
				dependsOn: spec.dependsOn.map((id) => token(tokens, id)),
				deferred: spec.deferred,
				register: (application: Application): void => {
					if (spec.registerEvent !== undefined) {
						events.push(spec.registerEvent);
					}

					if (spec.registerResolve !== undefined) {
						application.make(token(tokens, spec.registerResolve));
					}

					if (spec.registerValue !== undefined) {
						application.instance(token(tokens, spec.registerValue.token), spec.registerValue.value);
					}
				},
				...(spec.bootEvent === undefined
					? {}
					: {
							boot: (): void => {
								events.push(spec.bootEvent ?? '');
							},
						}),
			},
		]),
	);

	const bindWithLifetime = (operation: Operation, conditional: boolean): void => {
		const service = token(tokens, operation.token ?? '');

		if (operation.lifetime === 'singleton') {
			if (conditional) {
				subject.singletonIf(service, factory(operation));
			} else {
				subject.singleton(service, factory(operation));
			}
		} else if (operation.lifetime === 'scoped') {
			if (conditional) {
				subject.scopedIf(service, factory(operation));
			} else {
				subject.scoped(service, factory(operation));
			}
		} else if (conditional) {
			subject.bindIf(service, factory(operation));
		} else {
			subject.bind(service, factory(operation));
		}
	};

	for (const operation of testCase.operations) {
		const observe = (value: unknown): void => {
			if (operation.observe !== undefined) {
				observations.push(`${operation.observe}=${render(value)}`);
			}
		};

		switch (operation.kind) {
			case 'bind':
				bindWithLifetime(operation, false);
				break;

			case 'bind-if':
				bindWithLifetime(operation, true);
				break;

			case 'singleton-if':
				subject.singletonIf(token(tokens, operation.token ?? ''), factory(operation));
				break;

			case 'scoped-if':
				subject.scopedIf(token(tokens, operation.token ?? ''), factory(operation));
				break;

			case 'instance':
				subject.instance(token(tokens, operation.token ?? ''), operation.value);
				break;

			case 'resolve':
				observe(subject.make(token(tokens, operation.token ?? '')));
				break;

			case 'resolve-with-parameters':
				observe(subject.makeWith(token(tokens, operation.token ?? ''), operation.parameters ?? {}));
				break;

			case 'get':
				observe(subject.get(token(tokens, operation.token ?? '')));
				break;

			case 'forget-scoped':
				subject.forgetScopedInstances();
				break;

			case 'forget-instance':
				subject.forgetInstance(token(tokens, operation.token ?? ''));
				break;

			case 'flush':
				subject.flush();
				break;

			case 'alias':
				subject.alias(token(tokens, operation.target ?? ''), token(tokens, operation.alias ?? ''));
				break;

			case 'contextual-value':
				subject
					.when(...consumersOf(operation, tokens))
					.needs(token(tokens, operation.needs ?? ''))
					.give(operation.value);
				break;

			case 'contextual-factory':
				subject
					.when(...consumersOf(operation, tokens))
					.needs(token(tokens, operation.needs ?? ''))
					.giveFactory(factory(operation));
				break;

			case 'contextual-tagged':
				subject
					.when(...consumersOf(operation, tokens))
					.needs(token(tokens, operation.needs ?? ''))
					.giveTagged(operation.tag ?? '');
				break;

			case 'tag':
				subject.tag(
					(operation.tokens ?? []).map((id) => token(tokens, id)),
					operation.tag ?? '',
				);
				break;

			case 'tagged':
				observe(subject.tagged(operation.tag ?? ''));
				break;

			case 'extend':
				subject.extend(token(tokens, operation.token ?? ''), extender(operation));
				break;

			case 'callback':
				if (operation.phase === 'before') {
					subject.beforeResolving(operation.token === undefined ? undefined : token(tokens, operation.token), () => events.push(operation.event ?? ''));
				} else if (operation.phase === 'resolving') {
					subject.resolving(operation.token === undefined ? undefined : token(tokens, operation.token), () => events.push(operation.event ?? ''));
				} else {
					subject.afterResolving(operation.token === undefined ? undefined : token(tokens, operation.token), () => events.push(operation.event ?? ''));
				}

				break;

			case 'rebinding':
				subject.rebinding(token(tokens, operation.token ?? ''), (value) => events.push(`${operation.event ?? ''}:${render(value)}`));
				break;

			case 'method-bind':
				subject.bindMethod(operation.method ?? '', factory(operation));
				break;

			case 'method-call':
				observe(subject.callMethodBinding(operation.method ?? '', operation.instance));
				break;

			case 'call':
				observe(subject.call(factory(operation), operation.parameters ?? {}));
				break;

			case 'wrap':
				observe(subject.wrap(factory(operation), operation.parameters ?? {})());
				break;

			case 'factory-func':
				observe(subject.factoryFunc(token(tokens, operation.token ?? ''))());
				break;

			case 'provider-register':
				subject.register(
					providers.get(operation.provider ?? '') ??
						(() => {
							throw new Error('missing provider');
						})(),
				);
				break;

			case 'provider-register-many':
				subject.registerMany(
					(operation.providers ?? []).map(
						(id) =>
							providers.get(id) ??
							(() => {
								throw new Error('missing provider');
							})(),
					),
				);
				break;

			case 'provider-boot':
				subject.boot();
				break;

			case 'observe-counter':
				observations.push(`${operation.counter ?? ''}=${counters.get(operation.counter ?? '') ?? 0}`);
				break;

			case 'observe-events':
				observations.push(`${operation.observe ?? 'events'}=${events.join(',')}`);
				break;

			case 'observe-bound':
				observe(subject.bound(token(tokens, operation.token ?? '')));
				break;

			case 'observe-has':
				observe(subject.has(token(tokens, operation.token ?? '')));
				break;

			case 'observe-resolved':
				observe(subject.resolved(token(tokens, operation.token ?? '')));
				break;

			case 'observe-is-shared':
				observe(subject.isShared(token(tokens, operation.token ?? '')));
				break;

			case 'observe-bindings': {
				const bindings = [...subject.bindings().entries()].map(([service, binding]) => `${service.name}:${binding.lifetime}`).sort((left, right) => left.localeCompare(right));

				observe(bindings.join(','));
				break;
			}

			case 'observe-providers': {
				const ids = subject.providers().map((provider) => {
					for (const [id, value] of providers) {
						if (value === provider) {
							return id;
						}
					}

					return '<unknown>';
				});

				observe(ids.join(','));
				break;
			}

			case 'observe-has-provider':
				observe(subject.hasProvider(token(tokens, operation.token ?? '')));
				break;

			case 'observe-provider-for': {
				const found = subject.providerFor(token(tokens, operation.token ?? ''));

				if (found === undefined) {
					observe(null);
					break;
				}

				for (const [id, value] of providers) {
					if (value === found) {
						observe(id);
						break;
					}
				}

				break;
			}

			case 'observe-booted':
				observe(subject.booted());
				break;

			case 'observe-has-method':
				observe(subject.hasMethodBinding(operation.method ?? ''));
				break;
		}
	}

	return observations;
};

type ErrorCtor = new (...args: never[]) => Error;

const errorType = (identity: ContainerErrorCode): ErrorCtor => {
	switch (identity) {
		case 'ALIAS_CYCLE':
			return AliasCycleError;

		case 'CIRCULAR_RESOLUTION':
			return CircularResolutionError;

		case 'SELF_ALIAS':
			return SelfAliasError;

		case 'MISSING_METHOD_BINDING':
			return MissingMethodBindingError;

		case 'PROVIDER_CYCLE':
			return ProviderCycleError;

		default:
			return MissingBindingError;
	}
};

describe('container cross-runtime conformance', () => {
	for (const testCase of fixture.cases) {
		it(testCase.id, () => {
			if (testCase.error !== undefined) {
				expect(() => runContainerCase(testCase), testCase.note).toThrow(errorType(testCase.error));

				return;
			}

			expect(runContainerCase(testCase), testCase.note).toEqual(testCase.expected);
		});
	}
});

describe('container conformance fixture validation', () => {
	const valid = (): unknown => ({
		schemaVersion: 1,
		cases: [{ id: 'valid', note: 'valid fixture', tokens: [{ id: 'token' }], operations: [{ kind: 'resolve', token: 'token' }], error: 'MISSING_BINDING' }],
	});

	it.each([
		[
			'unknown operation',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ ...((fixture.cases as readonly unknown[])[0] as object), operations: [{ kind: 'unknown' }] }] }),
		],
		[
			'unknown error identity',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ ...((fixture.cases as readonly unknown[])[0] as object), error: 'UNKNOWN' }] }),
		],
		[
			'both expected and error',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ ...((fixture.cases as readonly unknown[])[0] as object), expected: [] }] }),
		],
		['neither expected nor error', (fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ id: 'valid', note: 'valid fixture', tokens: [], operations: [] }] })],
		[
			'duplicate case id',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [(fixture.cases as readonly unknown[])[0], (fixture.cases as readonly unknown[])[0]] }),
		],
		['missing note', (fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ id: 'valid', tokens: [], operations: [], error: 'MISSING_BINDING' }] })],
		[
			'invalid token reference',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({
				...fixture,
				cases: [{ id: 'valid', note: 'valid fixture', tokens: [], operations: [{ kind: 'resolve', token: 'missing' }], error: 'MISSING_BINDING' }],
			}),
		],
		[
			'missing tokens array',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ id: 'valid', note: 'valid fixture', operations: [], error: 'MISSING_BINDING' }] }),
		],
		[
			'null tokens array',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ id: 'valid', note: 'valid fixture', tokens: null, operations: [], error: 'MISSING_BINDING' }] }),
		],
		[
			'missing operations array',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ id: 'valid', note: 'valid fixture', tokens: [], error: 'MISSING_BINDING' }] }),
		],
		[
			'null operations array',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({ ...fixture, cases: [{ id: 'valid', note: 'valid fixture', tokens: [], operations: null, error: 'MISSING_BINDING' }] }),
		],
		[
			'invalid lifetime',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({
				...fixture,
				cases: [
					{
						id: 'valid',
						note: 'valid fixture',
						tokens: [{ id: 'token' }],
						operations: [{ kind: 'bind', token: 'token', lifetime: 'request', primitive: 'constant', value: 'v' }],
						error: 'MISSING_BINDING',
					},
				],
			}),
		],
		[
			'rebinding without event',
			(fixture: Readonly<Record<string, unknown>>): unknown => ({
				...fixture,
				cases: [{ id: 'valid', note: 'valid fixture', tokens: [{ id: 'token' }], operations: [{ kind: 'rebinding', token: 'token' }], error: 'MISSING_BINDING' }],
			}),
		],
	])('%s is rejected', (_name, mutate) => expect(() => parseContainerFixture(mutate(valid() as Readonly<Record<string, unknown>>))).toThrow(TypeError));
});
