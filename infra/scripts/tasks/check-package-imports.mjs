#!/usr/bin/env node
import { execFileSync } from 'node:child_process';
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import ts from 'typescript';

const rootPath = process.cwd();
const packagesPath = path.join(rootPath, 'packages');
const relativeSpecifierPattern = /^\.{1,2}\//u;

const trackedPackageFiles = () => {
	try {
		return execFileSync('git', ['-C', rootPath, 'ls-files', '-z', '--', 'packages'], { encoding: 'utf8' })
			.split('\0')
			.filter((file) => file.endsWith('.ts'));
	} catch {
		return [];
	}
};

const discoveredPackageFiles = (directory) => {
	if (!existsSync(directory)) {
		return [];
	}

	const files = [];
	const entries = readdirSync(directory, { withFileTypes: true });

	for (const entry of entries) {
		const absolutePath = path.join(directory, entry.name);

		if (entry.isDirectory()) {
			if (entry.name !== 'dist' && entry.name !== 'node_modules') {
				files.push(...discoveredPackageFiles(absolutePath));
			}

			continue;
		}

		if (entry.isFile() && entry.name.endsWith('.ts')) {
			files.push(path.relative(rootPath, absolutePath));
		}
	}

	return files;
};

const trackedFiles = trackedPackageFiles();
const packageFiles = trackedFiles.length > 0 ? trackedFiles : discoveredPackageFiles(packagesPath);
const violations = [];

const reportViolation = (sourceFile, specifier, message) => {
	const location = sourceFile.getLineAndCharacterOfPosition(specifier.getStart(sourceFile));
	violations.push(`${sourceFile.fileName}:${location.line + 1}:${location.character + 1} ${message}: ${specifier.text}`);
};

const checkSpecifier = (sourceFile, specifier, message) => {
	if (ts.isStringLiteralLike(specifier) && relativeSpecifierPattern.test(specifier.text)) {
		reportViolation(sourceFile, specifier, message);
	}
};

const visit = (sourceFile, node) => {
	if (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) {
		if (node.moduleSpecifier !== undefined) {
			checkSpecifier(sourceFile, node.moduleSpecifier, 'relative package import/export is not allowed');
		}
	} else if (ts.isImportEqualsDeclaration(node) && ts.isExternalModuleReference(node.moduleReference)) {
		checkSpecifier(sourceFile, node.moduleReference.expression, 'relative package import is not allowed');
	} else if (ts.isCallExpression(node)) {
		const [firstArgument] = node.arguments;

		if (firstArgument !== undefined && node.expression.kind === ts.SyntaxKind.ImportKeyword) {
			checkSpecifier(sourceFile, firstArgument, 'relative dynamic import is not allowed');
		}

		if (firstArgument !== undefined && ts.isIdentifier(node.expression) && node.expression.text === 'require') {
			checkSpecifier(sourceFile, firstArgument, 'relative require is not allowed');
		}
	}

	ts.forEachChild(node, (child) => visit(sourceFile, child));
};

for (const file of packageFiles) {
	const absolutePath = path.join(rootPath, file);

	if (!existsSync(absolutePath) || !statSync(absolutePath).isFile()) {
		continue;
	}

	const sourceFile = ts.createSourceFile(file, readFileSync(absolutePath, 'utf8'), ts.ScriptTarget.Latest, true);
	visit(sourceFile, sourceFile);
}

if (violations.length > 0) {
	console.error('Relative TypeScript imports are not allowed under packages/. Use package aliases instead.');
	console.error(violations.join('\n'));
	process.exit(1);
}

console.log(`Checked ${packageFiles.length} package TypeScript files for alias-only imports.`);
