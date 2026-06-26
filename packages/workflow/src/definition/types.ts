import type { Marking } from '../marking.js';
import type { Transition } from '../transition.js';
import type { Metadata, NestedMetadata } from '../types.js';

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
