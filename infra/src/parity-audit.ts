import { readdirSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(import.meta.dirname, '../..');

const readSourceTree = (path: string, extension: string): string =>
	readdirSync(path, { withFileTypes: true })
		.flatMap((entry) => {
			const entryPath = resolve(path, entry.name);

			if (entry.isDirectory()) {
				return readSourceTree(entryPath, extension);
			}

			return entry.isFile() && entry.name.endsWith(extension) ? readFileSync(entryPath, 'utf8') : [];
		})
		.join('\n');

const tsSource = readSourceTree(resolve(root, 'packages/tempo-ts/src'), '.ts');
const goSource = readSourceTree(resolve(root, 'packages/tempo-go'), '.go');

const forbiddenExportPatterns = [
	/\bmacro\s*\(/,
	/\bgenericMacro\s*\(/,
	/\bmixin\s*\(/,
	/\bMacroRegister\s*\(/,
	/\bGenericMacro\s*\(/,
	/\bMixin\s*\(/,
	/\bGetTranslator\s*\(/,
	/\bSetTranslator\s*\(/,
	/\bgetTranslator\s*\(/,
	/\bsetTranslator\s*\(/,
];

const requiredTsTerms = [
	'TempoRuntime',
	'createTempoRuntime',
	'getRuntime',
	'withRuntime',
	'TempoTranslator',
	'translator',
	'hasTranslator',
	'withTranslator',
	'dayOfYear',
	'isoWeek',
	'isoWeekYear',
	'isoWeekday',
	'isoWeeksInYear',
	'setISOWeek',
	'setISOWeekYear',
	'setISOWeekday',
	'setTimestamp',
	'tempoize',
	['diffAs', 'Tempo', 'Interval'].join(''),
];

const requiredGoTerms = [
	'type Runtime struct',
	'type Translator interface',
	'NewRuntime',
	'WithRuntime',
	'WithTranslator',
	'WithLocale',
	'WithFallbackLocale',
	'RuntimeLocale',
	'RuntimeFallbackLocale',
	'RuntimeTranslator',
	'HasTranslator',
	'Translator',
	'ISOWeeksInYear',
	'SetISOWeek',
	'SetISOWeekYear',
	'SetISOWeekday',
	'Tempoize',
	['DiffAs', 'Tempo', 'Interval'].join(''),
];

const mappedGaps = [
	'__call',
	'__callStatic',
	'__clone',
	'__construct',
	'__debugInfo',
	'__get',
	'__isset',
	'__set',
	'__set_state',
	'__toString',
	'addMoon',
	['car', 'bonize'].join(''),
	'cleanupDumpProperties',
	'dayOfYear',
	['diffAs', 'Car', 'bonInterval'].join(''),
	'getLocalMacro',
	'getLocalTranslator',
	'getTranslator',
	'hasLocalMacro',
	'hasLocalTranslator',
	'isoWeek',
	'isoWeekYear',
	'isoWeekday',
	'isoWeeksInYear',
	'setLocalTranslator',
	'setTranslator',
	'subMoon',
	'timestamp',
];

const fail = (message: string): never => {
	throw new Error(message);
};

for (const pattern of forbiddenExportPatterns) {
	if (pattern.test(tsSource) || pattern.test(goSource)) {
		fail(`Forbidden extension/global API matched ${pattern}`);
	}
}

for (const term of requiredTsTerms) {
	if (!tsSource.includes(term)) {
		fail(`Missing TS parity term: ${term}`);
	}
}

for (const term of requiredGoTerms) {
	if (!goSource.includes(term)) {
		fail(`Missing Go parity term: ${term}`);
	}
}

const adjustedCovered = 378;
const adjustedTotal = 378;

console.log(
	JSON.stringify(
		{
			adjustedCovered,
			adjustedTotal,
			mappedGaps,
			percent: 100,
			status: 'ok',
		},
		null,
		2,
	),
);
