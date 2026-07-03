# Console TypeScript

The TypeScript package lives in `packages/console` and exposes
`@alloy/console`. It is a terminal UI toolkit for interactive CLIs:
async prompts (text, select, confirm, search, and friends), output helpers
(tables, grids, notes, notifications), live status rendering (spinners,
progress bars, tasks, streams), ANSI-aware string utilities, and a
swappable prompt environment that makes everything scriptable in tests.

This is a private workspace package: it is consumed by sibling packages via
`workspace:*` and is never published to npm.

Prompts and status helpers are plain async functions:

```ts
import { confirm, intro, outro, select, spin, task, text } from '@alloy/console';

intro('Project setup');

const name = await text('Project name', 'my-app');
const framework = await select('Pick a framework', ['vue', 'react', 'svelte']);

if (await confirm('Install dependencies?')) {
	await spin('Installing...', async () => installDependencies(name, framework));
}

await task('Scaffolding', async (logger) => {
	logger.line('writing files');
	logger.success('scaffolded');
});

outro('Done');
```

Every prompt also accepts an options object (`text({ message, default, validate, ... })`).

For tests, swap the environment for scripted input and in-memory output:

```ts
import { createMemoryOutput, createScriptedInput, text, withPromptEnvironment } from '@alloy/console';

const output = createMemoryOutput();

await withPromptEnvironment({ input: createScriptedInput(['my-app']), output, interactive: true }, async () => {
	const name = await text('Project name');
});
```

## API overview

| Entry point | Main exports | Purpose |
| --- | --- | --- |
| `@alloy/console` | everything below | root export |
| `@alloy/console/prompts` | `text`, `textarea`, `password`, `number`, `select`, `multiselect`, `confirm`, `suggest`, `autocomplete`, `search`, `multisearch`, `pause` | interactive prompts |
| `@alloy/console/prompt` | `validateUsing`, `fallbackUsing`, `fallbackWhen`, `cancelUsing` | global validation, non-interactive fallbacks, cancel handling |
| `@alloy/console/output` | `intro`, `outro`, `info`, `error`, `warning`, `note`, `table`, `dataTable`, `grid`, `notify` | formatted output and desktop notifications |
| `@alloy/console/status` | `spin`, `task`, `progress`, `stream`, `Logger`, `Progress`, `Stream` | spinners, tasks, progress bars, live streams |
| `@alloy/console/strings` | `visibleWidth`, `truncate`, `wrap`, `parseAnsiText`, `parseAnsiSegments` | ANSI-aware string measurement and slicing |
| `@alloy/console/environment` | `configurePrompts`, `withPromptEnvironment`, `createScriptedInput`, `createMemoryOutput` | environment swapping for tests and embedding |
| `@alloy/console/key` | `Key`, `keyFromEvent`, `oneOf` | key constants and matching |
| `@alloy/console/terminal` | `terminalSize`, `clearTerminal`, `hideCursor`, `showCursor`, ... | low-level terminal control |
| `@alloy/console/form` | `form`, `FormBuilder` | multi-step prompt flows |

Individual prompts are also exposed as subpaths (e.g.
`@alloy/console/prompts/select`). Tests live in `packages/console/tests`
and run with `pnpm --filter @alloy/console test`.
