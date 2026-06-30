import { describe, expect, it } from 'vite-plus/test';
import { confirm, createMemoryOutput, createScriptedInput, form, Key, outro, Stream, text, withPromptEnvironment } from '#console/index';

const testLabel = (value: unknown, fallback: string): string => (typeof value === 'string' ? value : fallback);

describe('form builder', () => {
	it('runs chained steps and returns positional responses', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['A', 'd', 'a', Key.enter, Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().text('Name').select('Runtime', ['TS', 'JS']).confirm('Active').submit(),
		);

		expect(responses[0]).toBe('Ada');
		expect(responses[1]).toBe('TS');
		expect(responses[2]).toBe(true);
	});

	it('keys named responses and passes prior responses and names to custom steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['A', 'd', 'a', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.text('Name', '', '', false, undefined, '', 'name')
					.add((values, previous, name) => `${testLabel(name, 'missing')}:${testLabel(previous, 'new')}:Hello ${testLabel(values.name, '')}`, 'greeting')
					.submit(),
		);

		expect(responses.name).toBe('Ada');
		expect(responses.greeting).toBe('greeting:new:Hello Ada');
	});

	it('stores null for skipped conditional steps', async () => {
		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['n']),
				output: createMemoryOutput(),
				error: createMemoryOutput(),
				interactive: true,
			},
			() =>
				form()
					.confirm('Include details?', false, 'Yes', 'No', false, undefined, '', 'include')
					.addIf(
						(values) => values.include === true,
						() => 'details',
						'details',
					)
					.submit(),
		);

		expect(responses.include).toBe(false);
		expect(responses.details).toBeNull();
	});

	it('awaits asynchronous conditional steps', async () => {
		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['y', Key.enter]),
				output: createMemoryOutput(),
				error: createMemoryOutput(),
				interactive: true,
			},
			() =>
				form()
					.confirm('Include details?', true, 'Yes', 'No', false, undefined, '', 'include')
					.addIf(
						async (values) => values.include === true,
						() => 'details',
						'details',
					)
					.submit(),
		);

		expect(responses.include).toBe(true);
		expect(responses.details).toBe('details');
	});

	it('reverts to the previous form step and reuses the prior response', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['A', Key.enter, 'B', Key.ctrlU, Key.backspace, 'C', Key.enter, 'D', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().text('First').text('Second').submit(),
		);

		expect(responses[0]).toBe('C');
		expect(responses[1]).toBe('D');
	});

	it('clears responses of reverted steps to prevent state leaks', async () => {
		const output = createMemoryOutput();
		const runHistory: any[] = [];

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['A', Key.enter, 'B', Key.ctrlU, Key.backspace, 'C', Key.enter, 'D', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.text('First', '', '', false, undefined, '', 'first')
					.add(
						(res) => {
							runHistory.push({ ...res });

							return 'side-effect';
						},
						'side',
						true,
					)
					.text('Second', '', '', false, undefined, '', 'second')
					.submit(),
		);

		expect(responses.first).toBe('C');
		expect(responses.second).toBe('D');
		expect(responses.side).toBe('side-effect');

		expect(runHistory[0]).toEqual({ first: 'A' });
		expect(runHistory[1]).toEqual({ first: 'C' });
	});

	it('reuses numeric previous responses when reverting prompt builder steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['4', Key.enter, Key.ctrlU, '5', Key.enter, 'D', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().number('Count').text('Done').submit(),
		);

		expect(responses[0]).toBe(45);
		expect(responses[1]).toBe('D');
	});

	it('applies number transforms from form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['4', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.number('Count', '', '', false, undefined, '', undefined, undefined, undefined, 'count', (value) => Number(value) * 2)
					.submit(),
		);

		expect(responses.count).toBe(8);
	});

	it('preserves empty number defaults in form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().number('Count', '', undefined, false, undefined, '', undefined, undefined, undefined, 'labelCount').number({ message: 'Size' }, 'objectCount').submit(),
		);

		expect(responses.labelCount).toBe('');
		expect(responses.objectCount).toBe('');
	});

	it('runs object-option basic prompt form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter, Key.enter, Key.enter, Key.ctrlD]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.text({ message: 'Name', default: 'Ada', transform: (value) => value.toUpperCase() }, 'name')
					.number({ message: 'Count', default: 2, transform: (value) => Number(value) * 2 }, 'count')
					.password({ message: 'Secret', default: 'shh', transform: (value) => `${value}!` }, 'secret')
					.textarea({ message: 'Body', default: 'Line', rows: 3, transform: (value) => `${value}\nDone` }, 'body')
					.submit(),
		);

		expect(responses.name).toBe('ADA');
		expect(responses.count).toBe(4);
		expect(responses.secret).toBe('shh!');
		expect(responses.body).toBe('Line\nDone');
	});

	it('runs object-option choice prompt form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter, Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.confirm({ message: 'Continue?', default: false, transform: (value) => !value }, 'confirmed')
					.select({ message: 'Color', options: { red: 'Red', green: 'Green' }, default: 'green', transform: (value) => value.toUpperCase() }, 'color')
					.multiselect({ message: 'Colors', options: { red: 'Red', blue: 'Blue' }, default: ['red'], transform: (value) => [...value, 'blue'] }, 'colors')
					.submit(),
		);

		expect(responses.confirmed).toBe(true);
		expect(responses.color).toBe('GREEN');
		expect(responses.colors).toEqual(['red', 'blue']);
	});

	it('stores unnamed object-option confirm form steps positionally', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().confirm({ message: 'Continue?' }).submit(),
		);

		expect(responses[0]).toBe(true);
	});

	it('reuses array previous responses when reverting prompt builder steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.space, Key.enter, Key.ctrlU, Key.enter, 'D', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().multiselect('Flags', ['alpha', 'beta']).text('Done').submit(),
		);

		expect(responses[0]).toEqual(['alpha']);
		expect(responses[1]).toBe('D');
	});

	it('reuses boolean previous responses when reverting confirm steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['n', Key.enter, Key.ctrlU, Key.enter, 'D', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().confirm('Active?').text('Done').submit(),
		);

		expect(responses[0]).toBe(false);
		expect(responses[1]).toBe('D');
	});

	it('reverts past ignored side-effect steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['A', Key.enter, 'B', Key.ctrlU, Key.backspace, 'C', Key.enter, 'D', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().text('First').note('Skipping back over this', null, 'note').text('Second').submit(),
		);

		expect(responses[0]).toBe('C');
		expect(responses.note).toBeNull();
		expect(responses[2]).toBe('D');
	});

	it('will not skip over the first step when reverting', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlU, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().info('This should not be skipped').confirm('Are you sure?').submit(),
		);

		expect(responses[0]).toBeNull();
		expect(responses[1]).toBe(true);
	});

	it('stops custom step execution at the moment of reverting', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['2', '7', Key.enter, Key.down, Key.ctrlU, Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.text('Age')
					.add(async () => {
						const confirmed = await confirm('Are you sure?');

						if (!confirmed) {
							outro('This should not appear!');
						}
					})
					.submit(),
		);

		expect(output.text()).not.toContain('This should not appear!');
	});

	it('re-evaluates conditional steps after reverting prior responses', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['A', 'd', 'a', Key.enter, Key.down, Key.enter, Key.ctrlU, Key.up, Key.enter, '1', Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.text('Name')
					.select('Runtime', ['TS', 'JS'])
					.addIf(
						(values) => values[1] === 'TS',
						() => text('Version'),
					)
					.confirm('Are you sure?')
					.submit(),
		);

		expect(responses[0]).toBe('Ada');
		expect(responses[1]).toBe('TS');
		expect(responses[2]).toBe('1');
		expect(responses[3]).toBe(true);
	});

	it('re-evaluates asynchronous conditional steps after reverting prior responses', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['A', 'd', 'a', Key.enter, Key.down, Key.enter, Key.ctrlU, Key.up, Key.enter, '1', Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.text('Name')
					.select('Runtime', ['TS', 'JS'])
					.addIf(
						async (values) => values[1] === 'TS',
						() => text('Version'),
					)
					.confirm('Are you sure?')
					.submit(),
		);

		expect(responses[0]).toBe('Ada');
		expect(responses[1]).toBe('TS');
		expect(responses[2]).toBe('1');
		expect(responses[3]).toBe(true);
	});

	it('uses declared defaults when named choice steps follow skipped conditional responses', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.addIf(false, () => 'skipped', 'choice')
					.select('Choice', { a: 'A', b: 'B' }, 'b', 5, undefined, '', true, 'choice')
					.addIf(false, () => 'skipped', 'searched')
					.search({ message: 'Search', options: { red: 'Red', blue: 'Blue' }, default: 'blue' }, 'searched')
					.submit(),
		);

		expect(responses.choice).toBe('b');
		expect(responses.searched).toBe('blue');
	});

	it('does not revert the first form step', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlU, 'A', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().text('First').submit(),
		);

		expect(responses[0]).toBe('A');
		expect(output.text()).toContain('This cannot be reverted.');
	});

	it('does not allow reverting a normal prompt after form submit', async () => {
		const output = createMemoryOutput();

		const confirmed = await withPromptEnvironment(
			{
				input: createScriptedInput(['A', Key.enter, Key.ctrlU, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			async () => {
				await form().text('Name').submit();

				return confirm('Ready?');
			},
		);

		expect(confirmed).toBe(true);
		expect(output.text()).toContain('This cannot be reverted.');
	});

	it('runs suggest, search, and multisearch form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.tab, Key.enter, 'g', Key.down, Key.enter, 'r', Key.down, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.suggest('Suggested color', ['Red', 'Green', 'Blue'], '', '', 5, false, undefined, '', 'suggested')
					.search(
						{
							message: 'Searched color',
							options: (value) => {
								const options = { green: 'Green', blue: 'Blue', red: 'Red' };

								return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
							},
						},
						'searched',
					)
					.multisearch(
						{
							message: 'Many colors',
							options: (value) => {
								const options = { red: 'Red', green: 'Green', blue: 'Blue' };

								return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
							},
						},
						'many',
					)
					.submit(),
		);

		expect(responses.suggested).toBe('Blue');
		expect(responses.searched).toBe('green');
		expect(responses.many).toEqual(['red']);
	});

	it('passes suggest placeholders and defaults from form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().suggest('Suggested color', ['Red', 'Green', 'Blue'], 'Type a color', 'Blue', 5, false, undefined, '', 'suggested').submit(),
		);

		expect(responses.suggested).toBe('Blue');
	});

	it('passes autocomplete placeholders from form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().autocomplete('Auto color', ['Red', 'Green', 'Blue'], 'Type a color', '', false, undefined, '', 'auto').submit(),
		);

		expect(responses.auto).toBe('');
		expect(output.text()).toContain('Type a color');
	});

	it('passes suggest and autocomplete info from form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.down, Key.enter, 'g', Key.tab, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.suggest('Suggested color', ['Red', 'Green', 'Blue'], '', '', 5, false, undefined, '', 'suggested', undefined, (value) => `About suggested ${value ?? 'none'}`)
					.autocomplete('Auto color', ['Red', 'Green', 'Blue'], '', '', false, undefined, '', 'auto', undefined, (value) => `About auto ${value ?? 'none'}`)
					.submit(),
		);

		expect(responses.suggested).toBe('Blue');
		expect(responses.auto).toBe('Green');
		expect(output.text()).toContain('About suggested Blue');
		expect(output.text()).toContain('About auto Green');
	});

	it('runs object-option suggest and autocomplete form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.suggest({ message: 'Suggested color', options: ['Red', 'Green', 'Blue'], default: 'Blue' }, 'suggested')
					.autocomplete({ message: 'Auto color', options: ['Red', 'Green', 'Blue'], default: 'Green' }, 'auto')
					.submit(),
		);

		expect(responses.suggested).toBe('Blue');
		expect(responses.auto).toBe('Green');
	});

	it('runs label-first search and multisearch form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['g', Key.down, Key.enter, 'r', Key.down, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.search(
						'Searched color',
						(value) => {
							const options = { green: 'Green', blue: 'Blue', red: 'Red' };

							return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
						},
						'',
						5,
						undefined,
						'',
						true,
						'searched',
					)
					.multisearch(
						'Many colors',
						(value) => {
							const options = { red: 'Red', green: 'Green', blue: 'Blue' };

							return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
						},
						'',
						5,
						false,
						undefined,
						'Use the space bar to select options.',
						'many',
					)
					.submit(),
		);

		expect(responses.searched).toBe('green');
		expect(responses.many).toEqual(['red']);
	});

	it('passes label-first choice info through form steps', async () => {
		const output = createMemoryOutput();

		const colorOptions = (value: string): Record<string, string> => {
			const options = { red: 'Red', green: 'Green', blue: 'Blue' };

			return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
		};

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter, Key.down, Key.space, Key.enter, 'g', Key.down, Key.enter, 'r', Key.down, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.select('Choice', ['first', 'second'], undefined, 5, undefined, '', true, 'choice', undefined, (value) => `About choice ${value ?? 'none'}`)
					.multiselect('Choices', ['first', 'second'], [], 5, false, undefined, 'Use the space bar to select options.', 'choices', undefined, (value) => `About choices ${value ?? 'none'}`)
					.search('Searched color', colorOptions, '', 5, undefined, '', true, 'searched', undefined, (value) => `About search ${value ?? 'none'}`)
					.multisearch('Many colors', colorOptions, '', 5, false, undefined, 'Use the space bar to select options.', 'many', undefined, (value) => `About multisearch ${value ?? 'none'}`)
					.submit(),
		);

		expect(responses.choice).toBe('second');
		expect(responses.choices).toEqual(['second']);
		expect(responses.searched).toBe('green');
		expect(responses.many).toEqual(['red']);
		expect(output.text()).toContain('About choice second');
		expect(output.text()).toContain('About choices second');
		expect(output.text()).toContain('About search green');
		expect(output.text()).toContain('About multisearch red');
	});

	it('reuses previous search responses when reverting object-option form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['g', Key.down, Key.enter, Key.ctrlU]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.search(
						{
							message: 'Searched color',
							options: (value) => {
								const options = { green: 'Green', blue: 'Blue', red: 'Red' };

								return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
							},
						},
						'searched',
					)
					.confirm('Done?')
					.submit(),
		);

		expect(responses.searched).toBe('green');
	});

	it('reuses previous multisearch responses when reverting object-option form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['r', Key.down, Key.space, Key.enter, Key.ctrlU]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.multisearch(
						{
							message: 'Many colors',
							options: (value) => {
								const options = { red: 'Red', green: 'Green', blue: 'Blue' };

								return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
							},
						},
						'many',
					)
					.confirm('Done?')
					.submit(),
		);

		expect(responses.many).toEqual(['red']);
	});

	it('runs password and textarea form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['s', 'e', 'c', 'r', 'e', 't', Key.enter, 'L', '1', Key.enter, 'L', '2', Key.ctrlD]),
				output,
				error: output,
				interactive: true,
			},
			() => form().password('Secret', '', false, undefined, '', 'secret').textarea('Body', '', '', false, undefined, '', 3, 'body').submit(),
		);

		expect(responses.secret).toBe('secret');
		expect(responses.body).toBe('L1\nL2');
	});

	it('reuses password previous responses when reverting label-first form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['s', 'e', 'c', 'r', 'e', 't', Key.enter, Key.ctrlU, Key.enter, 'D', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => form().password('Secret').text('Done').submit(),
		);

		expect(responses[0]).toBe('secret');
		expect(responses[1]).toBe('D');
	});

	it('runs task, pause, and stream form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.task(
						'Build',
						(logger) => {
							logger.log('done');

							return 1;
						},
						10,
						false,
						'',
						'task',
					)
					.pause('Continue')
					.stream(['line one\n'], 'streamed')
					.submit(),
		);

		expect(responses.task).toBe(1);
		expect(responses[1]).toBe(true);
		expect(responses.streamed).toBeNull();
		expect(output.text()).toContain('Build');
		expect(output.text()).toContain('line one');
	});

	it('runs spin and progress form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				output,
				error: output,
			},
			() =>
				form()
					.spin(async () => 'spun', undefined, 'spin')
					.progress('Files', [1, 2], (step) => Number(step) * 2, '', 'progress')
					.submit(),
		);

		expect(responses.spin).toBe('spun');
		expect(responses.progress).toEqual([2, 4]);
		expect(output.text()).toContain(' ⠶ ');
		expect(output.text()).not.toContain('Loading');
		expect(output.text()).not.toContain('Done:');
		expect(output.text()).toContain('Files');
		expect(output.text()).toContain('2 / 2');
	});

	it('runs root-aligned status overloads from form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				output,
				error: output,
			},
			() =>
				form()
					.spin('Loading', async () => 'spun', 'spin')
					.task(
						{
							keepSummary: true,
							task: (logger) => {
								logger.info('done');

								return 2;
							},
							title: 'Build',
						},
						'task',
					)
					.submit(),
		);

		expect(responses.spin).toBe('spun');
		expect(responses.task).toBe(2);
		expect(output.text()).toContain('Loading');
		expect(output.text()).toContain('Build');
		expect(output.text()).toContain('done');
	});

	it('creates manual progress bars from form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				output,
				error: output,
			},
			() => form().progress(2, 'Manual', 'bar').submit(),
		);

		expect(responses.bar).toBeNull();
		expect(output.text()).toContain('Manual');
		expect(output.text()).toContain('0 / 2');
	});

	it('creates manual streams from form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				output,
				error: output,
			},
			async () => {
				const values = await form().stream().submit();

				const outputStream = values[0] as Stream;

				outputStream.write('hello');
				outputStream.close();

				return values;
			},
		);

		expect(responses[0]).toBeInstanceOf(Stream);
		expect((responses[0] as Stream).value()).toBe('hello');
		expect(output.text()).toContain('hello');
		expect(output.text()).toContain('\u001B[?25h');
	});

	it('stores null for display-only output helper responses', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				output,
				error: output,
			},
			() =>
				form()
					.note('Saved', null, 'note')
					.table(['Name'], [['Ollin']], 'table')
					.dataTable(['Alias'], [['Display']], 'dataTable')
					.grid(['A', 'B'], 80, 'grid')
					.title('Prompt', 'title')
					.clear('clear')
					.submit(),
		);

		expect(responses.note).toBeNull();
		expect(responses.table).toBeNull();
		expect(responses.dataTable).toBeNull();
		expect(responses.grid).toBeNull();
		expect(responses.title).toBeNull();
		expect(responses.clear).toBeNull();
		expect(output.text()).toContain('Ollin');
		expect(output.text()).toContain('Display');
		expect(output.text()).toContain('A');
		expect(output.text()).toContain('B');
		expect(output.text()).toContain('\u001B]0;Prompt\u0007');
		expect(output.text()).toContain('\u001B[H\u001B[J');
	});

	it('stores null for named object-option table output form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				output,
				error: output,
			},
			() =>
				form()
					.table({ headers: ['Name'], rows: [['Ada']] }, 'tableOptions')
					.dataTable({ headers: ['Alias'], rows: [['Display']] }, 'dataTableOptions')
					.submit(),
		);

		expect(responses.tableOptions).toBeNull();
		expect(responses.dataTableOptions).toBeNull();
		expect(responses[0]).toBeUndefined();
		expect(output.text()).toContain('Ada');
		expect(output.text()).toContain('Display');
	});

	it('runs label-first data table form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'slow', Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.datatable(
						['Name', 'Tag'],
						[
							{ Name: 'Alpha', Tag: 'fast', value: 'alpha' },
							{ Name: 'Beta', Tag: 'slow', value: 'beta' },
						],
						10,
						'Pick project',
						'',
						false,
						undefined,
						(value) => String(value).toUpperCase(),
						(query, row: any) => !Array.isArray(row) && row.Tag === query,
						'project',
					)
					.submit(),
		);

		expect(responses.project).toBe('BETA');
		expect(output.text()).toContain('Pick project slow');
		expect(output.text()).toContain('Beta');
	});

	it('reuses previous data table responses when reverting object-option form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter, Key.ctrlU, Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.datatable(
						{
							message: 'Pick project',
							rows: [
								{ Name: 'Alpha', value: 'alpha' },
								{ Name: 'Beta', value: 'beta' },
							],
						},
						'project',
					)
					.confirm('Done?')
					.submit(),
		);

		expect(responses.project).toBe('beta');
		expect(responses[1]).toBe(true);
	});

	it('reuses previous data table responses when reverting label-first form steps', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter, Key.ctrlU, Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.datatable(
						['Name'],
						[
							{ Name: 'Alpha', value: 'alpha' },
							{ Name: 'Beta', value: 'beta' },
						],
						10,
						'Pick project',
						'',
						false,
						undefined,
						undefined,
						undefined,
						'project',
					)
					.confirm('Done?')
					.submit(),
		);

		expect(responses.project).toBe('beta');
		expect(responses[1]).toBe(true);
	});
});
