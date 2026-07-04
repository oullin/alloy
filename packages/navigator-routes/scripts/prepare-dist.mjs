// Renames tsc output to the .mjs/.d.mts names declared in package.json#exports
// so the package can be consumed as a git dependency via the prepare script,
// matching what `vp pack` produces.
import { renameSync } from 'node:fs';
import { join } from 'node:path';

const dist = new URL('../dist', import.meta.url).pathname;

renameSync(join(dist, 'index.js'), join(dist, 'index.mjs'));
renameSync(join(dist, 'index.d.ts'), join(dist, 'index.d.mts'));
