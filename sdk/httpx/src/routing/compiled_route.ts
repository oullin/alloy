export interface Token {
	readonly kind: 'text' | 'variable';
	readonly prefix: string;
	readonly regexp?: string;
	readonly name?: string;
	readonly utf8?: boolean;
	readonly important?: boolean;
}

export class CompiledRoute {
	readonly compiledRegex: RegExp;
	readonly compiledHostRegex: RegExp | null;

	constructor(
		readonly staticPrefix: string,
		readonly regex: string,
		readonly tokens: readonly Token[],
		readonly pathVariables: readonly string[],
		readonly hostRegex: string = '',
		readonly hostTokens: readonly Token[] = [],
		readonly hostVariables: readonly string[] = [],
		readonly variables: readonly string[] = [],
	) {
		this.compiledRegex = regex !== '' ? new RegExp(regex, 's') : new RegExp('^$');
		this.compiledHostRegex = hostRegex !== '' ? new RegExp(hostRegex, 'si') : null;
	}
}
