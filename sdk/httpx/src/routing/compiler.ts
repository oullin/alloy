import { RouteCompilationError } from '#httpx/errors';
import { CompiledRoute, type Token } from '#httpx/routing/compiled_route';

export interface SourceRoute {
	path(): string;
	host(): string;
	requirements(): Readonly<Record<string, string>>;
	hasDefault(name: string): boolean;
	defaults?(): Readonly<Record<string, unknown>>;
}

export const SEPARATORS = '/,;.:-_~+*=@|';

export const VARIABLE_MAXIMUM_LENGTH = 32;

const VARIABLE_RE = /\{(!)?([\w\x80-\xff]+)\}/g;
const VALID_GROUP_NAME = /^[A-Za-z_]\w*$/;

interface CompileResult {
	staticPrefix: string;
	regex: string;
	tokens: Token[];
	variables: string[];
}

export function compileRoute(route: SourceRoute): CompiledRoute {
	let hostVariables: string[] = [];

	const variables: string[] = [];

	let hostRegex = '';
	let hostTokens: Token[] = [];

	const host = route.host();

	if (host !== '') {
		const res = compilePattern(route, host, true);

		hostVariables = res.variables;
		variables.push(...hostVariables);
		hostTokens = res.tokens;
		hostRegex = res.regex;
	}

	const path = route.path();
	const res = compilePattern(route, path, false);

	for (const v of res.variables) {
		if (v === '_fragment') {
			throw new RouteCompilationError(`route pattern "${path}" cannot contain "_fragment" as a path parameter`);
		}
	}

	const pathVariables = res.variables;

	variables.push(...pathVariables);

	const uniqueVars = Array.from(new Set(variables));

	return new CompiledRoute(res.staticPrefix, res.regex, res.tokens, pathVariables, hostRegex, hostTokens, hostVariables, uniqueVars);
}

function compilePattern(route: SourceRoute, pattern: string, isHost: boolean): CompileResult {
	const tokens: Token[] = [];
	const variables: string[] = [];

	let pos = 0;

	const defaultSeparator = isHost ? '.' : '/';

	const matches: Array<{
		matchStart: number;
		matchEnd: number;
		important: boolean;
		varName: string;
	}> = [];

	const regexForMatch = new RegExp(VARIABLE_RE.source, 'g');

	let m: RegExpExecArray | null;

	while ((m = regexForMatch.exec(pattern)) !== null) {
		matches.push({
			matchStart: m.index,
			matchEnd: regexForMatch.lastIndex,
			important: m[1] !== undefined,
			varName: m[2],
		});
	}

	for (const match of matches) {
		const matchStart = match.matchStart;
		const matchEnd = match.matchEnd;
		const important = match.important;
		const varName = match.varName;

		const precedingText = pattern.slice(pos, matchStart);

		pos = matchEnd;

		let precedingChar = '';

		if (precedingText !== '') {
			precedingChar = precedingText.slice(-1);
		}

		const isSeparator = precedingChar !== '' && SEPARATORS.includes(precedingChar);

		if (varName.length > 0 && varName[0] >= '0' && varName[0] <= '9') {
			throw new RouteCompilationError(`variable name "${varName}" cannot start with a digit in route pattern "${pattern}"`);
		}

		if (!VALID_GROUP_NAME.test(varName)) {
			throw new RouteCompilationError(`variable name "${varName}" contains characters not supported by RE2/named groups in route pattern "${pattern}"`);
		}

		if (variables.includes(varName)) {
			throw new RouteCompilationError(`route pattern "${pattern}" cannot reference variable name "${varName}" more than once`);
		}

		if (varName.length > VARIABLE_MAXIMUM_LENGTH) {
			throw new RouteCompilationError(`variable name "${varName}" cannot be longer than ${VARIABLE_MAXIMUM_LENGTH} characters in route pattern "${pattern}"`);
		}

		if (isSeparator && precedingText !== precedingChar) {
			tokens.push({ kind: 'text', prefix: precedingText.slice(0, precedingText.length - precedingChar.length) });
		} else if (!isSeparator && precedingText !== '') {
			tokens.push({ kind: 'text', prefix: precedingText });
		}

		const reqs = route.requirements();

		let regexp = reqs[varName] ?? '';

		if (regexp === '') {
			const followingPattern = pattern.slice(pos);
			const nextSeparator = findNextSeparator(followingPattern);

			let extra = '';

			if (defaultSeparator !== nextSeparator && nextSeparator !== '') {
				extra = escapeRegex(nextSeparator);
			}

			regexp = `[^${escapeRegex(defaultSeparator)}${extra}]+`;
		} else {
			regexp = transformCapturingGroupsToNonCapturing(regexp);
		}

		const tokenPrefix = isSeparator ? precedingChar : '';

		tokens.push({
			kind: 'variable',
			prefix: tokenPrefix,
			regexp,
			name: varName,
			important,
		});
		variables.push(varName);
	}

	if (pos < pattern.length) {
		tokens.push({ kind: 'text', prefix: pattern.slice(pos) });
	}

	let firstOptional = -1;

	if (!isHost) {
		for (let i = tokens.length - 1; i >= 0; i--) {
			const t = tokens[i];

			if (t.kind === 'variable' && !t.important && route.hasDefault(t.name ?? '')) {
				firstOptional = i;
			} else {
				break;
			}
		}
	}

	let regexpBuilder = '';

	for (let i = 0; i < tokens.length; i++) {
		regexpBuilder += computeRegexp(tokens, i, firstOptional);
	}

	const regexpStr = `^${regexpBuilder}$`;

	const reversed: Token[] = [];

	for (let i = 0; i < tokens.length; i++) {
		reversed[tokens.length - 1 - i] = tokens[i];
	}

	return {
		staticPrefix: determineStaticPrefix(route, tokens),
		regex: regexpStr,
		tokens: reversed,
		variables,
	};
}

function determineStaticPrefix(route: SourceRoute, tokens: Token[]): string {
	if (tokens.length === 0) {
		return '';
	}

	if (tokens[0].kind !== 'text') {
		if (route.hasDefault(tokens[0].name ?? '') || tokens[0].prefix === '/') {
			return '';
		}

		return tokens[0].prefix;
	}

	let prefix = tokens[0].prefix;

	if (tokens.length > 1 && tokens[1].prefix !== '/' && !route.hasDefault(tokens[1].name ?? '')) {
		prefix += tokens[1].prefix;
	}

	return prefix;
}

function findNextSeparator(pattern: string): string {
	if (pattern === '') {
		return '';
	}

	const stripped = pattern.replace(new RegExp(VARIABLE_RE.source, 'g'), '');

	if (stripped === '') {
		return '';
	}

	const first = stripped.charAt(0);

	if (SEPARATORS.includes(first)) {
		return first;
	}

	return '';
}

function computeRegexp(tokens: Token[], index: number, firstOptional: number): string {
	const t = tokens[index];

	if (t.kind === 'text') {
		return escapeRegex(t.prefix);
	}

	if (index === 0 && firstOptional === 0) {
		return `${escapeRegex(t.prefix)}(?<${t.name}>${t.regexp})?`;
	}

	let regex = `${escapeRegex(t.prefix)}(?<${t.name}>${t.regexp})`;

	if (firstOptional >= 0 && index >= firstOptional) {
		const nbOptional = index - firstOptional;

		if (index === firstOptional) {
			regex = '(?:' + regex;
		}

		if (index === tokens.length - 1) {
			regex += ')?'.repeat(nbOptional + 1);
		}
	}

	return regex;
}

function transformCapturingGroupsToNonCapturing(pattern: string): string {
	let result = '';

	for (let i = 0; i < pattern.length; i++) {
		if (pattern[i] === '\\' && i + 1 < pattern.length) {
			result += pattern[i] + pattern[i + 1];
			i++;
			continue;
		}

		if (pattern[i] === '(' && i + 1 < pattern.length && pattern[i + 1] !== '?') {
			result += '(?:';
			continue;
		}

		result += pattern[i];
	}

	return result;
}

export function escapeRegex(s: string): string {
	return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
