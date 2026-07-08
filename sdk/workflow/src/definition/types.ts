import type { Marking } from '#workflow/marking';
import type { Transition } from '#workflow/transition';
import type { Metadata, NestedMetadata } from '#workflow/types';

export interface DefinitionInput {
	places?: string[];
	transitions?: Transition[];
	initialMarking?: Marking;
	metadata?: Metadata;
	placeMetadata?: NestedMetadata;
	transitionMetadata?: NestedMetadata;
}

export interface MutableDefinition {
	places: string[];
	transitions: Transition[];
	initialMarking: Marking;
	metadata: Metadata;
	placeMetadata: NestedMetadata;
	transitionMetadata: NestedMetadata;
}
