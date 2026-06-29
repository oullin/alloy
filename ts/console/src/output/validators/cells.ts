import { z } from 'zod';
import type { TableCell } from '#console/types';

export const tableCellSchema: z.ZodType<TableCell> = z.union([z.string(), z.number(), z.boolean(), z.null(), z.undefined()]);

export const tableCellArraySchema = z.array(tableCellSchema);

export const tableCellRecordSchema = z.record(z.string(), tableCellSchema);
