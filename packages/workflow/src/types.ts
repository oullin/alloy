export type WorkflowContext = Record<string, unknown>;

export type Metadata = Record<string, unknown>;

export type NestedMetadata = Record<string, Metadata>;

export interface Sink {
	debug(message: string, ...args: unknown[]): void;
	info(message: string, ...args: unknown[]): void;
	warn(message: string, ...args: unknown[]): void;
	error(message: string, ...args: unknown[]): void;
}

export interface MetadataStore {
	workflowMetadata(key: string): unknown;
	hasWorkflowMetadata(key: string): boolean;
	placeMetadata(place: string, key: string): unknown;
	hasPlaceMetadata(place: string, key: string): boolean;
	transitionMetadata(transition: string, key: string): unknown;
	hasTransitionMetadata(transition: string, key: string): boolean;
}

export const cloneRecord = <T>(source: Record<string, T> | undefined): Record<string, T> => {
	if (source === undefined) {
		return {};
	}

	return { ...source };
};

export const cloneNestedRecord = <T>(source: Record<string, Record<string, T>> | undefined): Record<string, Record<string, T>> => {
	const output: Record<string, Record<string, T>> = {};

	for (const [key, value] of Object.entries(source ?? {})) {
		output[key] = { ...value };
	}

	return output;
};
