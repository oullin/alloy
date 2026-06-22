import { outputBuilderMethods } from '#console/form/builder/output';
import { promptBuilderMethods } from '#console/form/builder/prompts';
import { statusBuilderMethods } from '#console/form/builder/status';
import type { FormBuilder } from '#console/form/builder/index';
import type { OutputBuilderMethods } from '#console/form/builder/output';
import type { PromptBuilderMethods } from '#console/form/builder/prompts';
import type { StatusBuilderMethods } from '#console/form/builder/status';

export type FormBuilderMethods = OutputBuilderMethods & PromptBuilderMethods & StatusBuilderMethods;

export abstract class FormBuilderMethodSurface {
	declare readonly alert: OutputBuilderMethods['alert'];
	declare readonly autocomplete: PromptBuilderMethods['autocomplete'];
	declare readonly confirm: PromptBuilderMethods['confirm'];
	declare readonly clear: OutputBuilderMethods['clear'];
	declare readonly dataTable: OutputBuilderMethods['dataTable'];
	declare readonly datatable: OutputBuilderMethods['datatable'];
	declare readonly error: OutputBuilderMethods['error'];
	declare readonly grid: OutputBuilderMethods['grid'];
	declare readonly info: OutputBuilderMethods['info'];
	declare readonly intro: OutputBuilderMethods['intro'];
	declare readonly multisearch: PromptBuilderMethods['multisearch'];
	declare readonly multiselect: PromptBuilderMethods['multiselect'];
	declare readonly note: OutputBuilderMethods['note'];
	declare readonly notify: OutputBuilderMethods['notify'];
	declare readonly number: PromptBuilderMethods['number'];
	declare readonly outro: OutputBuilderMethods['outro'];
	declare readonly password: PromptBuilderMethods['password'];
	declare readonly pause: PromptBuilderMethods['pause'];
	declare readonly progress: StatusBuilderMethods['progress'];
	declare readonly search: PromptBuilderMethods['search'];
	declare readonly select: PromptBuilderMethods['select'];
	declare readonly spin: StatusBuilderMethods['spin'];
	declare readonly stream: StatusBuilderMethods['stream'];
	declare readonly suggest: PromptBuilderMethods['suggest'];
	declare readonly table: OutputBuilderMethods['table'];
	declare readonly task: StatusBuilderMethods['task'];
	declare readonly text: PromptBuilderMethods['text'];
	declare readonly textarea: PromptBuilderMethods['textarea'];
	declare readonly title: OutputBuilderMethods['title'];
	declare readonly warning: OutputBuilderMethods['warning'];
}

const formBuilderMethods: FormBuilderMethods & ThisType<FormBuilder> = {
	...outputBuilderMethods,
	...promptBuilderMethods,
	...statusBuilderMethods,
};

export const applyFormBuilderMethods = (builder: FormBuilderMethodSurface): void => {
	Object.assign(builder, formBuilderMethods);
};
