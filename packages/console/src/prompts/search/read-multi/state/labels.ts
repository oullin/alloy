import { multiSearchSelectedLabels } from '#console/prompts/search/read-multi/state/selection';
import type { SearchSelection } from '#console/prompts/search/selection';

export const selectedMultiSearchLabels = <T>(selected: SearchSelection<T>): string[] => {
	return multiSearchSelectedLabels(selected);
};
