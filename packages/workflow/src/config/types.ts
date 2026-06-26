export interface TransitionDeclaration {
	name: string;
	from: string[];
	to: string[];
}

export type WorkflowConfigRoot = Record<string, unknown>;
