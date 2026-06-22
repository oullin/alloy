import { z } from 'zod';

const renderableTableColumnCountSchema = z.number().finite().positive();

export const isRenderableTableColumnCount = (count: number): boolean => {
	return renderableTableColumnCountSchema.safeParse(count).success;
};
